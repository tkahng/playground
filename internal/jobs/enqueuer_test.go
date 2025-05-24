package jobs_test

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/tkahng/authgo/internal/jobs"
	"github.com/tkahng/authgo/internal/test"
)

type testJob struct {
	Message string
}

func (j testJob) Kind() string { return "test_job" }

func TestEnqueuer(t *testing.T) {
	test.DbSetup()
	t.Run("Enqueue single job", func(t *testing.T) {
		test.WithTx(t, func(ctx context.Context, tx pgx.Tx) {
			enqueuer := jobs.NewDBEnqueuer(tx)
			job := testJob{Message: "hello"}
			runAfter := time.Now().Add(1 * time.Hour)

			err := enqueuer.Enqueue(ctx, job, nil, runAfter, 3)
			assert.NoError(t, err)

			// Verify job was inserted
			var storedJob jobs.JobRow
			err = tx.QueryRow(ctx, `
			SELECT id, kind, payload, status, run_after, attempts, max_attempts
			FROM jobs WHERE kind = $1
		`, job.Kind()).Scan(
				&storedJob.ID, &storedJob.Kind, &storedJob.Payload,
				&storedJob.Status, &storedJob.RunAfter,
				&storedJob.Attempts, &storedJob.MaxAttempts,
			)
			assert.NoError(t, err)
			assert.Equal(t, jobs.JobStatusPending, storedJob.Status)
			assert.Equal(t, int64(0), storedJob.Attempts)
			assert.Equal(t, int64(3), storedJob.MaxAttempts)
		})
	})

	t.Run("Enqueue with unique key", func(t *testing.T) {
		test.WithTx(t, func(ctx context.Context, tx pgx.Tx) {
			defer tx.Rollback(ctx)
			enqueuer := jobs.NewDBEnqueuer(tx)

			uniqueKey := "unique_123"
			job := testJob{Message: "unique"}

			// First insert
			err := enqueuer.Enqueue(ctx, job, &uniqueKey, time.Now(), 1)
			assert.NoError(t, err)

			// Second insert should update existing
			err = enqueuer.Enqueue(ctx, testJob{Message: "updated"}, &uniqueKey, time.Now(), 1)
			assert.NoError(t, err)

			// Verify payload was updated
			var payload []byte
			err = tx.QueryRow(ctx, "SELECT payload FROM jobs WHERE unique_key = $1", uniqueKey).Scan(&payload)
			assert.NoError(t, err)
			assert.Contains(t, string(payload), `"updated"`)
		})
	})

	t.Run("EnqueueMany batch insert", func(t *testing.T) {
		test.WithTx(t, func(ctx context.Context, tx pgx.Tx) {
			enqueuer := jobs.NewDBEnqueuer(tx)

			jobs := []jobs.EnqueueParams{
				{
					Args:        testJob{Message: "batch1"},
					RunAfter:    time.Now(),
					MaxAttempts: 1,
				},
				{
					Args:        testJob{Message: "batch2"},
					UniqueKey:   strPtr("batch_key"),
					RunAfter:    time.Now().Add(1 * time.Hour),
					MaxAttempts: 3,
				},
			}

			err := enqueuer.EnqueueMany(ctx, jobs...)
			assert.NoError(t, err)

			var count int
			err = tx.QueryRow(ctx, "SELECT COUNT(*) FROM jobs").Scan(&count)
			assert.NoError(t, err)
			assert.Equal(t, 2, count)
		})
	})
}

func strPtr(s string) *string { return &s }
