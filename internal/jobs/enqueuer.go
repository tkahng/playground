package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

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
