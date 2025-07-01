package jobs

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/repository"
	"github.com/tkahng/authgo/internal/test"
)

func TestDbJobStore_SaveJob(t *testing.T) {
	test.WithTx(t, func(ctx context.Context, db database.Dbx) {
		type fields struct {
			db Db
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
				var payload Job[EmailJobArgs]
				if err := json.Unmarshal(firstJob.Payload, &payload); err != nil {
					t.Error(err)
				}
				argPayload, ok := tt.args.job.Args.(EmailJobArgs)
				if !ok {
					t.Error("job args is not email job args")
				}
				if argPayload.Recipient != payload.Args.Recipient {
					t.Errorf("DbJobStore.SaveJob() argPayload.Recipient = %v, want %v", argPayload.Recipient, payload.Args.Recipient)
				}
				if argPayload.Subject != payload.Args.Subject {
					t.Errorf("DbJobStore.SaveJob() argPayload.Subject = %v, want %v", argPayload.Subject, payload.Args.Subject)
				}
				if argPayload.Body != payload.Args.Body {
					t.Errorf("DbJobStore.SaveJob() argPayload.Body = %v, want %v", argPayload.Body, payload.Args.Body)
				}
			})
		}
	})
}

func TestDbJobStore_SaveManyJobs(t *testing.T) {
	test.WithTx(t, func(ctx context.Context, db database.Dbx) {
		type fields struct {
			db Db
		}
		type args struct {
			ctx  context.Context
			jobs []EnqueueParams
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
					jobs: []EnqueueParams{
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
	test.WithTx(t, func(ctx context.Context, db database.Dbx) {
		type fields struct {
			db Db
		}
		type args struct {
			jobs  []EnqueueParams
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
					jobs: []EnqueueParams{
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
	test.WithTx(t, func(ctx context.Context, db database.Dbx) {
		type fields struct {
			db database.Dbx
		}
		type args struct {
			jobs []EnqueueParams
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
					jobs: []EnqueueParams{
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
