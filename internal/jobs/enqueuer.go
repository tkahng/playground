package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Enqueuer interface {
	// Enqueue returns the created job's UUIDv7
	Enqueue(ctx context.Context, args JobArgs, uniqueKey *string, runAfter time.Time, maxAttempts int) (uuid.UUID, error)

	// EnqueueMany uses variadic parameters with chunking
	EnqueueMany(ctx context.Context, jobs ...EnqueueParams) error
}

type EnqueueParams struct {
	Args        JobArgs
	UniqueKey   *string
	RunAfter    time.Time
	MaxAttempts int
}

// DBEnqueuer implements Enqueuer using a database connection
type DBEnqueuer struct {
	DB *pgxpool.Pool
}

func NewDBEnqueuer(db *pgxpool.Pool) *DBEnqueuer {
	return &DBEnqueuer{DB: db}
}

const maxBatchSize = 1000 // Adjust based on your database's limits

func (e *DBEnqueuer) Enqueue(ctx context.Context, args JobArgs, uniqueKey *string, runAfter time.Time, maxAttempts int) (uuid.UUID, error) {
	payload, err := json.Marshal(args)
	if err != nil {
		return uuid.Nil, fmt.Errorf("marshal args: %w", err)
	}

	id, err := uuid.NewV7()
	if err != nil {
		return uuid.Nil, fmt.Errorf("generate uuid: %w", err)
	}

	_, err = e.DB.Exec(ctx, `
        INSERT INTO jobs (id, kind, unique_key, payload, status, run_after, attempts, max_attempts, created_at, updated_at)
        VALUES ($1, $2, $3, $4, 'pending', $5, 0, $6, clock_timestamp(), clock_timestamp())
        ON CONFLICT (unique_key)
        WHERE status IN ('pending', 'processing')
        DO UPDATE SET
            payload = EXCLUDED.payload,
            run_after = EXCLUDED.run_after,
            updated_at = clock_timestamp()
    `, id, args.Kind(), uniqueKey, payload, runAfter, maxAttempts)

	return id, err
}

func (e *DBEnqueuer) EnqueueMany(ctx context.Context, jobs ...EnqueueParams) error {
	if len(jobs) == 0 {
		return nil
	}

	// Process in chunks
	for i := 0; i < len(jobs); i += maxBatchSize {
		end := min(i+maxBatchSize, len(jobs))

		if err := e.enqueueBatch(ctx, jobs[i:end]); err != nil {
			return err
		}
	}

	return nil
}

func (e *DBEnqueuer) enqueueBatch(ctx context.Context, jobs []EnqueueParams) error {
	tx, err := e.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	batch := &pgx.Batch{}

	for _, job := range jobs {
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
	}

	br := tx.SendBatch(ctx, batch)
	defer br.Close()

	for range jobs {
		_, err := br.Exec()
		if err != nil {
			return fmt.Errorf("exec batch: %w", err)
		}
	}

	return tx.Commit(ctx)
}
