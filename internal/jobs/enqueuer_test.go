package jobs_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/jobs"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/repository"
	"github.com/tkahng/authgo/internal/test"
)

type testJob struct {
	Message string
}

func (j testJob) Kind() string { return "test_job" }

func TestEnqueuer(t *testing.T) {
	test.DbSetup()
	t.Run("Enqueue single job", func(t *testing.T) {
		test.WithTx(t, func(ctx context.Context, tx database.Dbx) {
			enqueuer := jobs.NewDBEnqueuer(tx)
			job := testJob{Message: "hello"}
			runAfter := time.Now().Add(1 * time.Hour)

			err := enqueuer.Enqueue(ctx, job, nil, runAfter, 3)
			assert.NoError(t, err)
			storedJob, err := repository.Job.GetOne(ctx, tx, &map[string]any{
				"kind": map[string]any{
					"_eq": job.Kind(),
				},
			})
			assert.NoError(t, err)
			assert.Equal(t, models.JobStatusPending, storedJob.Status)
			assert.Equal(t, int64(0), storedJob.Attempts)
			assert.Equal(t, int64(3), storedJob.MaxAttempts)
		})
	})

	t.Run("Enqueue with unique key", func(t *testing.T) {
		test.WithTx(t, func(ctx context.Context, tx database.Dbx) {
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
			queryJob, err := repository.Job.GetOne(ctx, tx, &map[string]any{
				"unique_key": map[string]any{
					"_eq": uniqueKey,
				},
			})
			assert.NoError(t, err)
			assert.Contains(t, string(queryJob.Payload), `"updated"`)
		})
	})

	t.Run("EnqueueMany batch insert", func(t *testing.T) {
		test.WithTx(t, func(ctx context.Context, tx database.Dbx) {
			enqueuer := jobs.NewDBEnqueuer(tx)

			params := []jobs.EnqueueParams{
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

			err := enqueuer.EnqueueMany(ctx, params...)
			assert.NoError(t, err)

			count, err := repository.Job.Count(
				ctx,
				tx,
				nil,
			)
			assert.NoError(t, err)
			assert.Equal(t, int64(2), count)
		})
	})
}

func strPtr(s string) *string { return &s }
