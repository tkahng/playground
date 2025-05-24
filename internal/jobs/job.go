package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/sync/errgroup"
)

// 'pending', 'processing', 'done', 'failed'
type JobStatus string

const (
	JobStatusPending    JobStatus = "pending"
	JobStatusProcessing JobStatus = "processing"
	JobStatusDone       JobStatus = "done"
	JobStatusFailed     JobStatus = "failed"
)

type JobRow struct {
	_           struct{}  `db:"jobs" json:"-"`
	ID          uuid.UUID `db:"id" json:"id"`
	Kind        string    `db:"kind" json:"kind"`
	UniqueKey   *string   `db:"unique_key" json:"unique_key"`
	Payload     []byte    `db:"payload" json:"payload"`
	Status      JobStatus `db:"status" json:"status"`
	RunAfter    time.Time `db:"run_after" json:"run_after"`
	Attempts    int64     `db:"attempts" json:"attempts"`
	MaxAttempts int64     `db:"max_attempts" json:"max_attempts"`
	LastError   *string   `db:"last_error" json:"last_error"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}

// Job represents a single unit of work, holding both the arguments and
// information for a job with args of type T.
type Job[T JobArgs] struct {
	*JobRow

	// Args are the arguments for the job.
	Args T
}

type JobArgs interface {
	// Kind is a string that uniquely identifies the type of job. This must be
	// provided on your job arguments struct. Jobs are identified by a string
	// instead of being based on type names so that previously inserted jobs
	// can be worked across deploys even if job/worker types are renamed.
	//
	// Kinds should be formatted without spaces like `my_custom_job`,
	// `mycustomjob`, or `my-custom-job`. Many special characters like colons,
	// dots, hyphens, and underscores are allowed, but those like spaces and
	// commas, which would interfere with UI functionality, are invalid.
	//
	// After initially deploying a job, it's generally not safe to rename its
	// kind (unless the database is completely empty) because River won't know
	// which worker should work the old kind. Job kinds can be renamed safely
	// over multiple deploys using the JobArgsWithKindAliases interface.
	Kind() string
}

// Dispatcher routes jobs to their appropriate handlers based on job kind.
type Dispatcher interface {
	// Dispatch executes the job with the appropriate handler.
	Dispatch(ctx context.Context, row *JobRow) error

	// SetHandler registers a handler for a specific job kind.
	// Panics if a handler is already registered for the kind.
	SetHandler(kind string, handler func(context.Context, *JobRow) error)
}

type dispatcher struct {
	handlers map[string]func(context.Context, *JobRow) error
}

func (d *dispatcher) SetHandler(kind string, handler func(context.Context, *JobRow) error) {
	if _, exists := d.handlers[kind]; exists {
		panic("duplicate worker kind: " + kind)
	}
	d.handlers[kind] = handler
}

func NewDispatcher() Dispatcher {
	return &dispatcher{
		handlers: make(map[string]func(context.Context, *JobRow) error),
	}
}

func RegisterWorker[T JobArgs](d *dispatcher, worker Worker[T]) {
	var zero T
	kind := zero.Kind()
	d.SetHandler(
		kind,
		func(ctx context.Context, row *JobRow) error {
			var args T
			if err := json.Unmarshal(row.Payload, &args); err != nil {
				return fmt.Errorf("unmarshal payload: %w", err)
			}
			return execute(ctx, worker, &Job[T]{JobRow: row, Args: args})
		},
	)
}

func (d *dispatcher) Dispatch(ctx context.Context, row *JobRow) error {
	handler, ok := d.handlers[row.Kind]
	if !ok {
		return fmt.Errorf("no handler registered for kind: %s", row.Kind)
	}
	return handler(ctx, row)
}

func execute[T JobArgs](ctx context.Context, worker Worker[T], job *Job[T]) (err error) {
	defer func() {
		if r := recover(); r != nil {
			slog.ErrorContext(ctx, "panic in job", "kind", job.Args.Kind(), "panic_recover", r, slog.Any("stack", string(debug.Stack())))
			err = fmt.Errorf("panic: %v", r)
		}
	}()
	slog.InfoContext(ctx, "start job", "kind", job.Args.Kind(), "id", job.ID)
	err = worker.Work(ctx, job)
	if err != nil {
		slog.ErrorContext(ctx, "fail job", "kind", job.Args.Kind(), "id", job.ID, "error", err)
	} else {
		slog.InfoContext(ctx, "done job", "kind", job.Args.Kind(), "id", job.ID)
	}
	return err
}

func ServeWithPoller(ctx context.Context, poller *Poller) error {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		return poller.Run(ctx)
	})

	return g.Wait()
}

type pollerOpts struct {
	Timeout time.Duration
}
type PollerOptsFunc func(*pollerOpts)

func WithTimeout(timeout time.Duration) PollerOptsFunc {
	return func(opts *pollerOpts) {
		opts.Timeout = timeout
	}
}

type Poller struct {
	Store      JobStore
	Dispatcher Dispatcher
	Interval   time.Duration
	opts       pollerOpts
}

func NewPoller(store JobStore, dispatcher Dispatcher, interval time.Duration, opts ...PollerOptsFunc) *Poller {
	p := &Poller{
		Store:      store,
		Dispatcher: dispatcher,
		Interval:   interval,
	}
	for _, opt := range opts {
		opt(&p.opts)
	}
	return p
}

func (p *Poller) Run(ctx context.Context) error {
	ticker := time.NewTicker(p.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := p.pollOnce(ctx); err != nil {
				slog.ErrorContext(ctx, "poller error", "error", err)
			}
		}
	}
}

func (p *Poller) pollOnce(ctx context.Context) error {
	tx, txErr := p.Store.Begin(ctx)
	if txErr != nil {
		return fmt.Errorf("begin tx: %w", txErr)
	}
	defer tx.Rollback(ctx) // Always defer rollback; commit will override

	// Claim only one job
	jobs, err := p.Store.ClaimPendingJobs(ctx, tx, 1) // LIMIT to 1
	if err != nil {
		return fmt.Errorf("claim jobs: %w", err)
	}

	if len(jobs) == 0 {
		// No jobs to process, commit and return
		return tx.Commit(ctx) // Commit an empty transaction to avoid erroring if there are no jobs
	}

	row := jobs[0] // Get the single claimed job
	timeout := p.opts.Timeout
	if timeout == 0 { // Provide a default if not set by options
		timeout = 30 * time.Second
	}

	jobCtx, cancel := context.WithTimeout(
		ctx,
		p.opts.Timeout,
	) // Still consider making this configurable
	defer cancel()

	dispatchErr := p.Dispatcher.Dispatch(jobCtx, &row)
	if dispatchErr != nil {
		slog.ErrorContext(ctx, "there was an error dispatching the job. will attempt to reschedule or mark as failed", slog.Any("error", dispatchErr), slog.String("job_id", row.ID.String()))
		if row.Attempts >= row.MaxAttempts {
			if markFailedErr := p.Store.MarkFailed(ctx, tx, row.ID, dispatchErr.Error()); markFailedErr != nil {
				slog.ErrorContext(ctx, "Error marking job as failed (and rolling back)", slog.Any("error", markFailedErr), slog.String("job_id", row.ID.String()))
				return fmt.Errorf("failed to mark job %s as failed: %w", row.ID, markFailedErr)
			}
		} else {
			delay := time.Duration(math.Pow(2, float64(row.Attempts))) * time.Second
			if rescheduleErr := p.Store.RescheduleJob(ctx, tx, row.ID, delay); rescheduleErr != nil {
				slog.ErrorContext(ctx, "Error rescheduling job (and rolling back)", slog.Any("error", rescheduleErr), slog.String("job_id", row.ID.String()))
				return fmt.Errorf("failed to reschedule job %s: %w", row.ID, rescheduleErr)
			}
		}
	} else {
		if markDoneErr := p.Store.MarkDone(ctx, tx, row.ID); markDoneErr != nil {
			slog.ErrorContext(ctx, "Error marking job as done (and rolling back)", slog.Any("error", markDoneErr), slog.String("job_id", row.ID.String()))
			return fmt.Errorf("failed to mark job %s as done: %w", row.ID, markDoneErr)
		}
	}

	return tx.Commit(ctx)
}

type JobStore interface {
	Begin(ctx context.Context) (pgx.Tx, error)
	ClaimPendingJobs(ctx context.Context, tx pgx.Tx, limit int) ([]JobRow, error)
	MarkDone(ctx context.Context, tx pgx.Tx, id uuid.UUID) error
	MarkFailed(ctx context.Context, tx pgx.Tx, id uuid.UUID, reason string) error
	RescheduleJob(ctx context.Context, tx pgx.Tx, id uuid.UUID, delay time.Duration) error
}
type DbJobStore struct {
	DB Db
}

func (s *DbJobStore) Begin(ctx context.Context) (pgx.Tx, error) {
	return s.DB.Begin(ctx)
}

func (s *DbJobStore) ClaimPendingJobs(ctx context.Context, tx pgx.Tx, limit int) ([]JobRow, error) {
	rows, err := tx.Query(ctx, `
		UPDATE jobs SET status='processing', updated_at=clock_timestamp(), attempts=attempts+1
		WHERE id IN (
			SELECT id FROM jobs
			WHERE status='pending' AND run_after <= clock_timestamp() AND attempts < max_attempts
			ORDER BY run_after
			LIMIT $1
			FOR UPDATE SKIP LOCKED
		)
		RETURNING id, kind, unique_key, payload, status, run_after, attempts, max_attempts, last_error, created_at, updated_at
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []JobRow
	for rows.Next() {
		var row JobRow
		if err := rows.Scan(
			&row.ID, &row.Kind, &row.UniqueKey, &row.Payload, &row.Status, &row.RunAfter,
			&row.Attempts, &row.MaxAttempts, &row.LastError, &row.CreatedAt, &row.UpdatedAt,
		); err != nil {
			return nil, err
		}
		jobs = append(jobs, row)
	}
	return jobs, rows.Err()
}

func (s *DbJobStore) MarkDone(ctx context.Context, tx pgx.Tx, id uuid.UUID) error {
	_, err := tx.Exec(ctx, `
		UPDATE jobs SET status='done', updated_at=clock_timestamp() WHERE id=$1
	`, id)
	return err
}

func (s *DbJobStore) MarkFailed(ctx context.Context, tx pgx.Tx, id uuid.UUID, reason string) error {
	_, err := tx.Exec(ctx, `
		UPDATE jobs SET status='failed', last_error=$2, updated_at=clock_timestamp()
		WHERE id=$1 AND attempts >= max_attempts
	`, id, reason)
	return err
}

func (s *DbJobStore) RescheduleJob(ctx context.Context, tx pgx.Tx, id uuid.UUID, delay time.Duration) error {
	_, err := tx.Exec(ctx, `
		UPDATE jobs SET run_after = clock_timestamp() + $2, updated_at = clock_timestamp(), status = 'pending'
		WHERE id = $1
	`, id, delay)
	return err
}

// with the client using the AddWorker function.
type Worker[T JobArgs] interface {

	// Work performs the job and returns an error if the job failed. The context
	// will be configured with a timeout according to the worker settings and may
	// be cancelled for other reasons.
	//
	// If no error is returned, the job is assumed to have succeeded and will be
	// marked completed.
	//
	// It is important for any worker to respect context cancellation to enable
	// the client to respond to shutdown requests; there is no way to cancel a
	// running job that does not respect context cancellation, other than
	// terminating the process.
	Work(ctx context.Context, job *Job[T]) error
}

type Db interface {
	SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Begin(ctx context.Context) (pgx.Tx, error)
}

// Enqueuer provides methods for adding jobs to the queue
type Enqueuer interface {
	// Enqueue adds a single job to the queue and returns its time-ordered UUIDv7
	Enqueue(ctx context.Context, args JobArgs, uniqueKey *string, runAfter time.Time, maxAttempts int) error

	// EnqueueMany efficiently adds multiple jobs in batches using transactions
	// Processes jobs in chunks to prevent overwhelming the database
	EnqueueMany(ctx context.Context, jobs ...EnqueueParams) error
}

// DBEnqueuer implements Enqueuer using a PostgreSQL connection pool
type DBEnqueuer struct {
	db Db
}

// NewDBEnqueuer creates a new database-backed job enqueuer
func NewDBEnqueuer(db Db) *DBEnqueuer {
	return &DBEnqueuer{db: db}
}

// EnqueueParams contains all parameters needed to enqueue a job
type EnqueueParams struct {
	Args        JobArgs   // Job arguments (must implement JobArgs interface)
	UniqueKey   *string   // Optional unique key for deduplication
	RunAfter    time.Time // When the job should become available for processing
	MaxAttempts int       // Maximum number of attempts before marking as failed
}

// maxBatchSize defines how many jobs to insert in a single database operation
// Adjust based on your database's performance characteristics
const maxBatchSize = 1000

// Enqueue adds a single job to the queue
func (e *DBEnqueuer) Enqueue(ctx context.Context, args JobArgs, uniqueKey *string, runAfter time.Time, maxAttempts int) error {
	payload, err := json.Marshal(args)
	if err != nil {
		return fmt.Errorf("marshal args: %w", err)
	}

	// Generate time-ordered UUIDv7 for better database performance
	id, err := uuid.NewV7()
	if err != nil {
		return fmt.Errorf("generate uuid: %w", err)
	}

	_, err = e.db.Exec(ctx, `
		INSERT INTO jobs (id, kind, unique_key, payload, status, run_after, attempts, max_attempts, created_at, updated_at)
		VALUES ($1, $2, $3, $4, 'pending', $5, 0, $6, clock_timestamp(), clock_timestamp())
		ON CONFLICT (unique_key)
		WHERE status IN ('pending', 'processing')
		DO UPDATE SET
			payload = EXCLUDED.payload,
			run_after = EXCLUDED.run_after,
			updated_at = clock_timestamp()
	`, id, args.Kind(), uniqueKey, payload, runAfter, maxAttempts)

	return err
}

// EnqueueMany efficiently processes multiple jobs in batches
func (e *DBEnqueuer) EnqueueMany(ctx context.Context, jobs ...EnqueueParams) error {
	if len(jobs) == 0 {
		return nil
	}

	// Process in chunks to prevent overwhelming the database
	for i := 0; i < len(jobs); i += maxBatchSize {
		end := min(i+maxBatchSize, len(jobs))

		if err := e.processBatch(ctx, jobs[i:end]); err != nil {
			return fmt.Errorf("batch %d-%d: %w", i, end, err)
		}
	}

	return nil
}

// processBatch handles a single chunk of jobs in a transaction
func (e *DBEnqueuer) processBatch(ctx context.Context, jobs []EnqueueParams) error {
	tx, err := e.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	batch := &pgx.Batch{}

	// Prepare all insert statements for this batch
	for _, job := range jobs {
		if err := e.addJobToBatch(batch, job); err != nil {
			return err
		}
	}

	// Execute the batch and check for errors
	if err := e.executeBatch(ctx, tx, batch, len(jobs)); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// addJobToBatch adds a single job to the batch operation
func (e *DBEnqueuer) addJobToBatch(batch *pgx.Batch, job EnqueueParams) error {
	payload, err := json.Marshal(job.Args)
	if err != nil {
		return fmt.Errorf("marshal args: %w", err)
	}

	id, err := uuid.NewV7()
	if err != nil {
		return fmt.Errorf("generate uuid: %w", err)
	}

	batch.Queue(`
		INSERT INTO jobs (id, kind, unique_key, payload, status, run_after, attempts, max_attempts, created_at, updated_at)
		VALUES ($1, $2, $3, $4, 'pending', $5, 0, $6, clock_timestamp(), clock_timestamp())
		ON CONFLICT (unique_key)
		WHERE status IN ('pending', 'processing')
		DO UPDATE SET
			payload = EXCLUDED.payload,
			run_after = EXCLUDED.run_after,
			updated_at = clock_timestamp()
	`, id, job.Args.Kind(), job.UniqueKey, payload, job.RunAfter, job.MaxAttempts)

	return nil
}

// executeBatch sends the batch to the database and verifies all operations succeeded
func (e *DBEnqueuer) executeBatch(ctx context.Context, tx pgx.Tx, batch *pgx.Batch, expectedResults int) error {
	br := tx.SendBatch(ctx, batch)
	defer br.Close()

	// Verify all operations completed successfully
	for i := range expectedResults {
		_, err := br.Exec()
		if err != nil {
			return fmt.Errorf("job %d in batch: %w", i, err)
		}
	}

	return nil
}
