package jobs

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type JobStore struct {
	Conn *pgx.Conn
}

func NewJobStore(conn *pgx.Conn) *JobStore {
	return &JobStore{Conn: conn}
}

func (s *JobStore) FetchNextJob(ctx context.Context) (*JobRow, error) {
	tx, err := s.Conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	row := tx.QueryRow(ctx, `
		UPDATE jobs
		SET status = 'processing', updated_at = now(), attempts = attempts + 1
		WHERE id = (
			SELECT id FROM jobs
			WHERE status = 'pending'
			  AND run_after <= now()
			  AND attempts < max_attempts
			ORDER BY created_at
			FOR UPDATE SKIP LOCKED
			LIMIT 1
		)
		RETURNING id, kind, unique_key, payload, status, run_after, attempts, max_attempts, last_error, created_at, updated_at
	`)

	var job JobRow
	err = row.Scan(
		&job.ID, &job.Kind, &job.UniqueKey, &job.Payload,
		&job.Status, &job.RunAfter, &job.Attempts, &job.MaxAttempts,
		&job.LastError, &job.CreatedAt, &job.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &job, nil
}
