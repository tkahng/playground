package jobs

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/database/repository"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/test"
)

func TestDbJobStore_SaveJob(t *testing.T) {
	t.Parallel()
	test.SkipIfShort(t)
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		type fields struct {
			db database.Dbx
		}
		type args struct {
			ctx context.Context
			job *EnqueueParams
		}
		tests := []struct {
			name    string
			fields  fields
			args    args
			wantErr bool
		}{
			{
				name: "create email job",
				fields: fields{
					db: db,
				},
				args: args{
					ctx: context.Background(),
					job: &EnqueueParams{
						Args: EmailJobArgs{
							Recipient: "recipient",
							Subject:   "subject",
							Body:      "body",
						},
						UniqueKey:   nil,
						RunAfter:    time.Now(),
						MaxAttempts: 1,
					},
				},
				wantErr: false,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				s := &DbJobStore{
					db: tt.fields.db,
				}
				if err := s.SaveJob(tt.args.ctx, tt.args.job); (err != nil) != tt.wantErr {
					t.Errorf("DbJobStore.SaveJob() error = %v, wantErr %v", err, tt.wantErr)
				}
				firstJob, err := repository.Job.GetOne(tt.args.ctx, db, nil)
				if err != nil {
					t.Error(err)
				}
				if firstJob == nil {
					t.Error("job not found")
				}
				var payload EmailJobArgs
				if err := json.Unmarshal(firstJob.Payload, &payload); err != nil {
					t.Error(err)
				}
				argPayload, ok := tt.args.job.Args.(EmailJobArgs)
				if !ok {
					t.Error("job args is not email job args")
				}
				if argPayload.Recipient != payload.Recipient {
					t.Errorf("DbJobStore.SaveJob() argPayload.Recipient = %v, want %v", argPayload.Recipient, payload.Recipient)
				}
				if argPayload.Subject != payload.Subject {
					t.Errorf("DbJobStore.SaveJob() argPayload.Subject = %v, want %v", argPayload.Subject, payload.Subject)
				}
				if argPayload.Body != payload.Body {
					t.Errorf("DbJobStore.SaveJob() argPayload.Body = %v, want %v", argPayload.Body, payload.Body)
				}
			})
		}
	})
}

func TestDbJobStore_SaveManyJobs(t *testing.T) {
	t.Parallel()
	test.SkipIfShort(t)
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		type fields struct {
			db database.Dbx
		}
		type args struct {
			ctx  context.Context
			jobs []*EnqueueParams
		}
		tests := []struct {
			name    string
			fields  fields
			args    args
			wantErr bool
		}{
			{
				name: "create email job",
				fields: fields{
					db: db,
				},
				args: args{
					ctx: context.Background(),
					jobs: []*EnqueueParams{
						{
							Args: EmailJobArgs{
								Recipient: "recipient",
								Subject:   "subject",
								Body:      "body",
							},
							UniqueKey:   nil,
							RunAfter:    time.Now(),
							MaxAttempts: 1,
						},
						{
							Args: EmailJobArgs{
								Recipient: "recipient2",
								Subject:   "subject2",
								Body:      "body2",
							},
							UniqueKey:   nil,
							RunAfter:    time.Now(),
							MaxAttempts: 1,
						},
					}},
				wantErr: false,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				e := &DbJobStore{
					db: tt.fields.db,
				}
				if err := e.SaveManyJobs(tt.args.ctx, tt.args.jobs...); (err != nil) != tt.wantErr {
					t.Errorf("DbJobStore.SaveManyJobs() error = %v, wantErr %v", err, tt.wantErr)
				}
				count, err := repository.Job.Count(tt.args.ctx, db, nil)
				if err != nil {
					t.Error(err)
				}
				if count != int64(len(tt.args.jobs)) {
					t.Errorf("DbJobStore.SaveManyJobs() count = %v, want %v", count, len(tt.args.jobs))
				}
			})
		}
	})
}

func TestDbJobStore_ClaimPendingJobs(t *testing.T) {
	t.Parallel()
	test.SkipIfShort(t)
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		type fields struct {
			db database.Dbx
		}
		type args struct {
			jobs  []*EnqueueParams
			ctx   context.Context
			limit int
		}
		tests := []struct {
			name      string
			fields    fields
			args      args
			want      []*models.JobRow
			wantCount int64
			wantErr   bool
		}{
			{
				name: "claim jobs",
				fields: fields{
					db: db,
				},
				args: args{
					jobs: []*EnqueueParams{
						{
							Args: EmailJobArgs{
								Recipient: "recipient",
								Subject:   "subject",
								Body:      "body",
							},
							UniqueKey:   nil,
							RunAfter:    time.Now(),
							MaxAttempts: 1,
						},
						{
							Args: EmailJobArgs{
								Recipient: "recipient2",
								Subject:   "subject2",
								Body:      "body2",
							},
							UniqueKey:   nil,
							RunAfter:    time.Now(),
							MaxAttempts: 1,
						},
					},
					ctx:   context.Background(),
					limit: 10,
				},

				wantErr: false,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				s := &DbJobStore{
					db: tt.fields.db,
				}
				if err := s.SaveManyJobs(tt.args.ctx, tt.args.jobs...); (err != nil) != tt.wantErr {
					t.Errorf("DbJobStore.SaveManyJobs() error = %v, wantErr %v", err, tt.wantErr)
				}
				got, err := s.ClaimPendingJobs(tt.args.ctx, tt.args.limit)
				if (err != nil) != tt.wantErr {
					t.Errorf("DbJobStore.ClaimPendingJobs() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if len(got) != len(tt.args.jobs) {
					t.Errorf("DbJobStore.ClaimPendingJobs() got = %v, want %v", len(got), len(tt.args.jobs))
				}
			})
		}
	},
	)
}

func TestDbJobStore_MarkDone(t *testing.T) {
	t.Parallel()
	test.SkipIfShort(t)
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		type fields struct {
			db database.Dbx
		}
		type args struct {
			jobs []*EnqueueParams
			ctx  context.Context
			id   uuid.UUID
		}
		tests := []struct {
			name    string
			fields  fields
			args    args
			wantErr bool
		}{
			{
				name: "mark done",
				fields: fields{
					db: db,
				},
				args: args{
					jobs: []*EnqueueParams{
						{
							Args: EmailJobArgs{
								Recipient: "recipient2",
								Subject:   "subject2",
								Body:      "body2",
							},
							UniqueKey:   nil,
							RunAfter:    time.Now(),
							MaxAttempts: 1,
						},
					},
					ctx: context.Background(),
				},
				wantErr: false,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				s := &DbJobStore{
					db: tt.fields.db,
				}
				if err := s.SaveManyJobs(tt.args.ctx, tt.args.jobs...); (err != nil) != tt.wantErr {
					t.Errorf("DbJobStore.SaveManyJobs() error = %v, wantErr %v", err, tt.wantErr)
				}
				pendingJobs, err := s.ClaimPendingJobs(tt.args.ctx, 1)
				if (err != nil) != tt.wantErr {
					t.Errorf("DbJobStore.ClaimPendingJobs() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if len(pendingJobs) < 1 {
					t.Errorf("DbJobStore.ClaimPendingJobs() got = %v, want %v", len(pendingJobs), 1)
				}
				tt.args.id = pendingJobs[0].ID
				if err := s.MarkDone(tt.args.ctx, tt.args.id); (err != nil) != tt.wantErr {
					t.Errorf("DbJobStore.MarkDone() error = %v, wantErr %v", err, tt.wantErr)
				}
				got, err := repository.Job.GetOne(
					tt.args.ctx,
					tt.fields.db,
					&map[string]any{
						"id": map[string]any{
							"_eq": tt.args.id,
						},
					},
				)
				if (err != nil) != tt.wantErr {
					t.Errorf("DbJobStore.ClaimPendingJobs() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if got.Status != models.JobStatusDone {
					t.Errorf("DbJobStore.MarkDone() got = %v, want %v", got.Status, models.JobStatusDone)
				}
			})
		}
	})
}

func TestDbJobStore_MarkFailed(t *testing.T) {
	t.Parallel()
	test.SkipIfShort(t)
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		type fields struct {
			db database.Dbx
		}
		type args struct {
			jobs []*EnqueueParams
			ctx  context.Context
			id   uuid.UUID
		}
		tests := []struct {
			name    string
			fields  fields
			args    args
			wantErr bool
		}{
			{
				name: "mark failed",
				fields: fields{
					db: db,
				},
				args: args{
					jobs: []*EnqueueParams{
						{
							Args: EmailJobArgs{
								Recipient: "recipient2",
								Subject:   "subject2",
								Body:      "body2",
							},
							UniqueKey:   nil,
							RunAfter:    time.Now(),
							MaxAttempts: 1,
						},
					},
					ctx: context.Background(),
				},
				wantErr: false,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				s := &DbJobStore{
					db: tt.fields.db,
				}
				if err := s.SaveManyJobs(tt.args.ctx, tt.args.jobs...); (err != nil) != tt.wantErr {
					t.Errorf("DbJobStore.SaveManyJobs() error = %v, wantErr %v", err, tt.wantErr)
				}
				pendingJobs, err := s.ClaimPendingJobs(tt.args.ctx, 1)
				if (err != nil) != tt.wantErr {
					t.Errorf("DbJobStore.ClaimPendingJobs() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if len(pendingJobs) < 1 {
					t.Errorf("DbJobStore.ClaimPendingJobs() got = %v, want %v", len(pendingJobs), 1)
				}
				tt.args.id = pendingJobs[0].ID
				if err := s.MarkFailed(tt.args.ctx, tt.args.id, "reason"); (err != nil) != tt.wantErr {
					t.Errorf("DbJobStore.MarkDone() error = %v, wantErr %v", err, tt.wantErr)
				}
				got, err := repository.Job.GetOne(
					tt.args.ctx,
					tt.fields.db,
					&map[string]any{
						"id": map[string]any{
							"_eq": tt.args.id,
						},
					},
				)
				if (err != nil) != tt.wantErr {
					t.Errorf("DbJobStore.ClaimPendingJobs() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if got.Status != models.JobStatusFailed {
					t.Errorf("DbJobStore.MarkDone() got = %v, want %v", string(got.Status), string(models.JobStatusFailed))
				}
			})
		}
	})
}

func TestDbJobStore_RescheduleJob(t *testing.T) {
	t.Parallel()
	test.SkipIfShort(t)
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		type fields struct {
			db database.Dbx
		}
		type args struct {
			jobs  []*EnqueueParams
			delay time.Duration
			ctx   context.Context
			id    uuid.UUID
		}
		tests := []struct {
			name    string
			fields  fields
			args    args
			wantErr bool
		}{
			{
				name: "mark RescheduleJob",
				fields: fields{
					db: db,
				},
				args: args{
					jobs: []*EnqueueParams{
						{
							Args: EmailJobArgs{
								Recipient: "recipient2",
								Subject:   "subject2",
								Body:      "body2",
							},
							UniqueKey:   nil,
							RunAfter:    time.Now(),
							MaxAttempts: 1,
						},
					},
					delay: time.Hour,
					ctx:   context.Background(),
				},
				wantErr: false,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				s := &DbJobStore{
					db: tt.fields.db,
				}
				if err := s.SaveManyJobs(tt.args.ctx, tt.args.jobs...); (err != nil) != tt.wantErr {
					t.Errorf("DbJobStore.SaveManyJobs() error = %v, wantErr %v", err, tt.wantErr)
				}
				pendingJobs, err := s.ClaimPendingJobs(tt.args.ctx, 1)
				if (err != nil) != tt.wantErr {
					t.Errorf("DbJobStore.ClaimPendingJobs() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if len(pendingJobs) < 1 {
					t.Errorf("DbJobStore.ClaimPendingJobs() got = %v, want %v", len(pendingJobs), 1)
				}
				tt.args.id = pendingJobs[0].ID
				if err := s.RescheduleJob(tt.args.ctx, tt.args.id, tt.args.delay); (err != nil) != tt.wantErr {
					t.Errorf("DbJobStore.MarkDone() error = %v, wantErr %v", err, tt.wantErr)
				}
				got, err := repository.Job.GetOne(
					tt.args.ctx,
					tt.fields.db,
					&map[string]any{
						"id": map[string]any{
							"_eq": tt.args.id,
						},
					},
				)
				if (err != nil) != tt.wantErr {
					t.Errorf("DbJobStore.ClaimPendingJobs() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if got.Status != models.JobStatusPending {
					t.Errorf("DbJobStore.MarkDone() got = %v, want %v", string(got.Status), string(models.JobStatusFailed))
				}
			})
		}
	})
}

func TestDbJobStore_ClaimPendingJobs_Buffer(t *testing.T) {
	t.Parallel()
	test.SkipIfShort(t)

	t.Run("claims job within 200ms buffer", func(t *testing.T) {
		t.Parallel()
		database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
			s := &DbJobStore{db: db}
			err := s.SaveJob(ctx, &EnqueueParams{
				Args:        EmailJobArgs{Recipient: "r", Subject: "s", Body: "b"},
				RunAfter:    time.Now().Add(100 * time.Millisecond),
				MaxAttempts: 1,
			})
			assert.NoError(t, err)

			got, err := s.ClaimPendingJobs(ctx, 10)
			assert.NoError(t, err)
			assert.Len(t, got, 1, "job within 200ms buffer should be claimed")
		})
	})

	t.Run("does not claim job beyond 200ms buffer", func(t *testing.T) {
		t.Parallel()
		database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
			s := &DbJobStore{db: db}
			err := s.SaveJob(ctx, &EnqueueParams{
				Args:        EmailJobArgs{Recipient: "r", Subject: "s", Body: "b"},
				RunAfter:    time.Now().Add(500 * time.Millisecond),
				MaxAttempts: 1,
			})
			assert.NoError(t, err)

			got, err := s.ClaimPendingJobs(ctx, 10)
			assert.NoError(t, err)
			assert.Len(t, got, 0, "job beyond 200ms buffer should not be claimed")
		})
	})
}

type testJob struct {
	Message string
}

func (j testJob) Kind() string { return "test_job" }

func TestEnqueuer(t *testing.T) {
	t.Run("Enqueue single job", func(t *testing.T) {
		database.WithNewTestTx(t, func(ctx context.Context, tx database.Dbx) {
			enqueuer := NewDbJobManager(tx)
			job := testJob{Message: "hello"}
			runAfter := time.Now().Add(1 * time.Hour)

			err := enqueuer.Enqueue(ctx, &EnqueueParams{
				Args:        &job,
				RunAfter:    runAfter,
				MaxAttempts: 3,
			})
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
		database.WithNewTestTx(t, func(ctx context.Context, tx database.Dbx) {
			enqueuer := NewDbJobManager(tx)
			uniqueKey := "unique_123"
			job := testJob{Message: "unique"}

			// First insert job, &uniqueKey, time.Now(), 1
			err := enqueuer.Enqueue(ctx, &EnqueueParams{
				Args:        &job,
				UniqueKey:   &uniqueKey,
				RunAfter:    time.Now(),
				MaxAttempts: 1,
			})
			assert.NoError(t, err)

			// Second insert should update existing
			// testJob{Message: "updated"}, &uniqueKey, time.Now(), 1
			err = enqueuer.Enqueue(ctx, &EnqueueParams{
				Args:        &testJob{Message: "updated"},
				UniqueKey:   &uniqueKey,
				RunAfter:    time.Now(),
				MaxAttempts: 1,
			})
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
		database.WithNewTestTx(t, func(ctx context.Context, tx database.Dbx) {
			enqueuer := NewDbJobManager(tx)
			params := []*EnqueueParams{
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
