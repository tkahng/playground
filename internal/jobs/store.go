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

type JobStore struct {
	DB *pgxpool.Pool
}

func (s *JobStore) Enqueue(ctx context.Context, args JobArgs, uniqueKey *string, runAfter time.Time, maxAttempts int) error {
	payload, err := json.Marshal(args)
	if err != nil {
		return fmt.Errorf("marshal args: %w", err)
	}
	_, err = s.DB.Exec(ctx, `
		INSERT INTO jobs (id, kind, unique_key, payload, status, run_after, attempts, max_attempts, created_at, updated_at)
		VALUES ($1, $2, $3, $4, 'pending', $5, 0, $6, clock_timestamp(), clock_timestamp())
		ON CONFLICT (unique_key)
		WHERE status IN ('pending', 'processing')
		DO UPDATE SET
			payload = EXCLUDED.payload,
			run_after = EXCLUDED.run_after,
			updated_at = clock_timestamp()
	`, uuid.New(), args.Kind(), uniqueKey, payload, runAfter, maxAttempts)
	return err
}

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
