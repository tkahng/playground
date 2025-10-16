package jobs

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/tkahng/playground/internal/models"
)

type EnqueueParams struct {
	Args        JobArgs   // Job arguments (must implement JobArgs interface)
	UniqueKey   *string   // Optional unique key for deduplication
	RunAfter    time.Time // When the job should become available for processing
	MaxAttempts int       // Maximum number of attempts before marking as failed
}
type Enqueuer interface {
	// Enqueue adds a single job to the queue and returns its time-ordered UUIDv7
	Enqueue(ctx context.Context, args *EnqueueParams) error

	// EnqueueMany efficiently adds multiple jobs in batches using transactions
	// Processes jobs in chunks to prevent overwhelming the database
	EnqueueMany(ctx context.Context, jobs ...*EnqueueParams) error
}

const maxBatchSize = 1000

type JobStore interface {
	SaveJob(ctx context.Context, args *EnqueueParams) error
	SaveManyJobs(ctx context.Context, jobs ...*EnqueueParams) error
	ClaimPendingJobs(ctx context.Context, limit int) ([]*models.JobRow, error)
	MarkDone(ctx context.Context, id uuid.UUID) error
	MarkFailed(ctx context.Context, id uuid.UUID, reason string) error
	RescheduleJob(ctx context.Context, id uuid.UUID, delay time.Duration) error
	RunInTx(ctx context.Context, fn func(JobStore) error) error
}
type DbJobStore struct {
	db Db
}

const query string = `
		INSERT INTO app.jobs (id, kind, unique_key, payload, status, run_after, attempts, max_attempts, created_at, updated_at)
		VALUES ($1, $2, $3, $4, 'pending', $5, 0, $6, clock_timestamp(), clock_timestamp())
		ON CONFLICT (unique_key)
		WHERE status IN ('pending', 'processing')
		DO UPDATE SET
			payload = EXCLUDED.payload,
			status = EXCLUDED.status,
			run_after = EXCLUDED.run_after,
			attempts = EXCLUDED.attempts,
			max_attempts = EXCLUDED.max_attempts,
			created_at = EXCLUDED.created_at,
			updated_at = clock_timestamp()
	`

// SaveJob implements JobStore.
func (s *DbJobStore) SaveJob(ctx context.Context, job *EnqueueParams) error {
	payload, err := json.Marshal(job.Args)
	if err != nil {
		return fmt.Errorf("marshal args: %w", err)
	}

	// Generate time-ordered UUIDv7 for better database performance
	id, err := uuid.NewV7()
	if err != nil {
		return fmt.Errorf("generate uuid: %w", err)
	}

	_, err = s.db.Exec(ctx, query, id, job.Args.Kind(), job.UniqueKey, payload, job.RunAfter, job.MaxAttempts)

	return err
}

// SaveManyJobs implements JobStore.
func (e *DbJobStore) SaveManyJobs(ctx context.Context, jobs ...*EnqueueParams) error {
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
func (e *DbJobStore) processBatch(ctx context.Context, jobs []*EnqueueParams) error {
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
func (e *DbJobStore) addJobToBatch(batch *pgx.Batch, job *EnqueueParams) error {
	if job == nil {
		return errors.New("job is nil")
	}
	payload, err := json.Marshal(job.Args)
	if err != nil {
		return fmt.Errorf("marshal args: %w", err)
	}

	id, err := uuid.NewV7()
	if err != nil {
		return fmt.Errorf("generate uuid: %w", err)
	}

	batch.Queue(query, id, job.Args.Kind(), job.UniqueKey, payload, job.RunAfter, job.MaxAttempts)

	return nil
}

// executeBatch sends the batch to the database and verifies all operations succeeded
func (e *DbJobStore) executeBatch(ctx context.Context, tx pgx.Tx, batch *pgx.Batch, expectedResults int) error {
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

var _ JobStore = (*DbJobStore)(nil)

func (s *DbJobStore) RunInTx(ctx context.Context, fn func(JobStore) error) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	err = fn(&DbJobStore{db: tx})
	if err == nil {
		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("commit transaction: %w", err)
		}
	}
	return err
}

func NewDbJobStore(db Db) *DbJobStore {
	return &DbJobStore{
		db: db,
	}
}

func (s *DbJobStore) ClaimPendingJobs(ctx context.Context, limit int) ([]*models.JobRow, error) {
	rows, err := s.db.Query(ctx, `
		UPDATE app.jobs SET status='processing', updated_at=clock_timestamp(), attempts=attempts+1
		WHERE id IN (
			SELECT id FROM app.jobs
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

	var jobs []*models.JobRow
	for rows.Next() {
		var row models.JobRow
		if err := rows.Scan(
			&row.ID, &row.Kind, &row.UniqueKey, &row.Payload, &row.Status, &row.RunAfter,
			&row.Attempts, &row.MaxAttempts, &row.LastError, &row.CreatedAt, &row.UpdatedAt,
		); err != nil {
			return nil, err
		}
		jobs = append(jobs, &row)
	}
	return jobs, rows.Err()
}

func (s *DbJobStore) MarkDone(ctx context.Context, id uuid.UUID) error {
	_, err := s.db.Exec(ctx, `
		UPDATE app.jobs SET status='done', updated_at=clock_timestamp() WHERE id=$1
	`, id)
	return err
}

func (s *DbJobStore) MarkFailed(ctx context.Context, id uuid.UUID, reason string) error {
	_, err := s.db.Exec(ctx, `
		UPDATE app.jobs SET status='failed', last_error=$2, updated_at=clock_timestamp()
		WHERE id=$1 AND attempts >= max_attempts
	`, id, reason)
	return err
}

func (s *DbJobStore) RescheduleJob(ctx context.Context, id uuid.UUID, delay time.Duration) error {
	_, err := s.db.Exec(ctx, `
		UPDATE app.jobs SET run_after = clock_timestamp() + $2, updated_at = clock_timestamp(), status = 'pending'
		WHERE id = $1
	`, id, delay)
	return err
}

func strPtr(s string) *string { return &s }
