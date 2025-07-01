package jobs

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/tkahng/authgo/internal/database"
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
