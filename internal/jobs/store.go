package jobs

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/models"
)

type JobStore interface {
	ClaimPendingJobs(ctx context.Context, limit int) ([]*models.JobRow, error)
	MarkDone(ctx context.Context, id uuid.UUID) error
	MarkFailed(ctx context.Context, id uuid.UUID, reason string) error
	RescheduleJob(ctx context.Context, id uuid.UUID, delay time.Duration) error
	RunInTx(ctx context.Context, fn func(JobStore) error) error
}
type DbJobStore struct {
	db Db
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
		UPDATE jobs SET status='done', updated_at=clock_timestamp() WHERE id=$1
	`, id)
	return err
}

func (s *DbJobStore) MarkFailed(ctx context.Context, id uuid.UUID, reason string) error {
	_, err := s.db.Exec(ctx, `
		UPDATE jobs SET status='failed', last_error=$2, updated_at=clock_timestamp()
		WHERE id=$1 AND attempts >= max_attempts
	`, id, reason)
	return err
}

func (s *DbJobStore) RescheduleJob(ctx context.Context, id uuid.UUID, delay time.Duration) error {
	_, err := s.db.Exec(ctx, `
		UPDATE jobs SET run_after = clock_timestamp() + $2, updated_at = clock_timestamp(), status = 'pending'
		WHERE id = $1
	`, id, delay)
	return err
}

type JobStoreDecorator struct {
	Job                  *models.JobRow
	Delegate             JobStore
	RunInTxFunc          func(ctx context.Context, fn func(JobStore) error) error
	ClaimPendingJobsFunc func(ctx context.Context, limit int) ([]*models.JobRow, error)
	MarkDoneFunc         func(ctx context.Context, id uuid.UUID) error
	MarkFailedFunc       func(ctx context.Context, id uuid.UUID, reason string) error
	RescheduleJobFunc    func(ctx context.Context, id uuid.UUID, delay time.Duration) error
}

var _ JobStore = (*JobStoreDecorator)(nil)

// RunInTx implements JobStore.
func (d *JobStoreDecorator) RunInTx(ctx context.Context, fn func(JobStore) error) error {
	if d.RunInTxFunc != nil {
		return d.RunInTxFunc(ctx, fn)
	}
	return d.Delegate.RunInTx(ctx, fn)
}

func NewJobStoreDecorator() *JobStoreDecorator {
	return &JobStoreDecorator{}
}

var _ JobStore = (*JobStoreDecorator)(nil)

func (d *JobStoreDecorator) ClaimPendingJobs(ctx context.Context, limit int) ([]*models.JobRow, error) {
	if d.ClaimPendingJobsFunc != nil {
		return d.ClaimPendingJobsFunc(ctx, limit)
	}
	return d.Delegate.ClaimPendingJobs(ctx, limit)
}

func (d *JobStoreDecorator) MarkDone(ctx context.Context, id uuid.UUID) error {
	if d.MarkDoneFunc != nil {
		return d.MarkDoneFunc(ctx, id)
	}
	return d.Delegate.MarkDone(ctx, id)
}

func (d *JobStoreDecorator) MarkFailed(ctx context.Context, id uuid.UUID, reason string) error {
	if d.MarkFailedFunc != nil {
		return d.MarkFailedFunc(ctx, id, reason)
	}
	return d.Delegate.MarkFailed(ctx, id, reason)
}

func (d *JobStoreDecorator) RescheduleJob(ctx context.Context, id uuid.UUID, delay time.Duration) error {
	if d.RescheduleJobFunc != nil {
		return d.RescheduleJobFunc(ctx, id, delay)
	}
	return d.Delegate.RescheduleJob(ctx, id, delay)
}
