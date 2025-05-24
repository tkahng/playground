package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
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
			log.Printf("panic in job kind=%s: %v\n%s", job.Args.Kind(), r, debug.Stack())
			err = fmt.Errorf("panic: %v", r)
		}
	}()
	log.Printf("start job kind=%s id=%s", job.Args.Kind(), job.ID)
	err = worker.Work(ctx, job)
	if err != nil {
		log.Printf("fail job kind=%s id=%s err=%v", job.Args.Kind(), job.ID, err)
	} else {
		log.Printf("done job kind=%s id=%s", job.Args.Kind(), job.ID)
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

type Poller struct {
	Store      *JobStore
	Dispatcher Dispatcher
	Interval   time.Duration
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
				log.Printf("poller error: %v", err)
			}
		}
	}
}

func (p *Poller) pollOnce(ctx context.Context) error {
	tx, err := p.Store.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	jobs, err := p.Store.ClaimPendingJobs(ctx, tx, 10)
	if err != nil {
		return fmt.Errorf("claim jobs: %w", err)
	}

	for _, row := range jobs {
		row := row
		func() {
			jobCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
			defer cancel()

			err := p.Dispatcher.Dispatch(jobCtx, &row)
			if err != nil {
				if row.Attempts >= row.MaxAttempts {
					_ = p.Store.MarkFailed(ctx, tx, row.ID, err.Error())
				} else {
					delay := time.Duration(math.Pow(2, float64(row.Attempts))) * time.Second
					_ = p.Store.RescheduleJob(ctx, tx, row.ID, delay)
				}
			} else {
				_ = p.Store.MarkDone(ctx, tx, row.ID)
			}
		}()
	}

	return tx.Commit(ctx)
}

type JobStore struct {
	DB *pgxpool.Pool
}

// func (s *JobStore) Enqueue(ctx context.Context, args JobArgs, uniqueKey *string, runAfter time.Time, maxAttempts int) error {
// 	payload, err := json.Marshal(args)
// 	if err != nil {
// 		return fmt.Errorf("marshal args: %w", err)
// 	}
// 	_, err = s.DB.Exec(ctx, `
// 		INSERT INTO jobs (id, kind, unique_key, payload, status, run_after, attempts, max_attempts, created_at, updated_at)
// 		VALUES ($1, $2, $3, $4, 'pending', $5, 0, $6, clock_timestamp(), clock_timestamp())
// 		ON CONFLICT (unique_key)
// 		WHERE status IN ('pending', 'processing')
// 		DO UPDATE SET
// 			payload = EXCLUDED.payload,
// 			run_after = EXCLUDED.run_after,
// 			updated_at = clock_timestamp()
// 	`, uuid.New(), args.Kind(), uniqueKey, payload, runAfter, maxAttempts)
// 	return err
// }

func (s *JobStore) ClaimPendingJobs(ctx context.Context, tx pgx.Tx, limit int) ([]JobRow, error) {
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

func (s *JobStore) MarkDone(ctx context.Context, tx pgx.Tx, id uuid.UUID) error {
	_, err := tx.Exec(ctx, `
		UPDATE jobs SET status='done', updated_at=clock_timestamp() WHERE id=$1
	`, id)
	return err
}

func (s *JobStore) MarkFailed(ctx context.Context, tx pgx.Tx, id uuid.UUID, reason string) error {
	_, err := tx.Exec(ctx, `
		UPDATE jobs SET status='failed', last_error=$2, updated_at=clock_timestamp()
		WHERE id=$1 AND attempts >= max_attempts
	`, id, reason)
	return err
}

func (s *JobStore) RescheduleJob(ctx context.Context, tx pgx.Tx, id uuid.UUID, delay time.Duration) error {
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
