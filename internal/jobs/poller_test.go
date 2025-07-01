package jobs

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/stores"
	"github.com/tkahng/authgo/internal/test"
)

func TestPoller_Run(t *testing.T) {
	test.WithTx(t, func(ctx context.Context, dbx database.Dbx) {
		testJobs := setupJobs(ctx, dbx)

		type fields struct {
			Store      JobStore
			Dispatcher Dispatcher
			opts       pollerOpts
		}
		type args struct {
			ctx  context.Context
			wg   *sync.WaitGroup
			args JobArgs
		}
		tests := []struct {
			name    string
			fields  fields
			args    args
			wantErr bool
			setup   func(args)
		}{
			{
				name:    "",
				fields:  fields{},
				args:    args{},
				wantErr: false,
				setup: func(args args) {
					testJobs.Clear()
					if args.wg != nil {
						testJobs.UseWg(args.wg)
					}
					testJobs.EmailWorker.WorkFunc = func(ctx context.Context, job *Job[EmailJobArgs]) error {
						testJobs.EmailWorker.Job = job
						testJobs.EmailWorker.Success = true
						return nil
					}
					testJobs.ReportWorker.WorkFunc = func(ctx context.Context, job *Job[ReportJobArgs]) error {
						testJobs.ReportWorker.Job = job
						testJobs.ReportWorker.Success = true
						return nil
					}
				},
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				tt.setup(tt.args)
				ServeWithPoller(tt.args.ctx, testJobs.Poller)

				if err := testJobs.Enqueuer.Enqueue(tt.args.ctx, tt.args.args, nil, time.Now(), 1); (err != nil) != tt.wantErr {
					t.Errorf("Poller.Run() error = %v, wantErr %v", err, tt.wantErr)
				}
				if tt.args.wg != nil {
					tt.args.wg.Wait()
				}
			})
		}
	})
}

type TestJobService struct {
	Adapter      stores.StorageAdapterInterface
	Store        JobStore
	Dispatcher   Dispatcher
	Poller       *Poller
	Enqueuer     Enqueuer
	EmailWorker  *EmailWorker
	ReportWorker *ReportWorker
	Wg           *sync.WaitGroup
}

func (s *TestJobService) UseWg(wg *sync.WaitGroup) {
	s.Wg = wg
	s.EmailWorker.Wg = wg
	s.ReportWorker.Wg = wg
}

func (s *TestJobService) Clear() {
	s.EmailWorker.Clear()
	s.ReportWorker.Clear()
}

// type

func setupJobs(ctx context.Context, dbx database.Dbx) *TestJobService {
	store := NewDbJobStore(dbx)
	adapter := stores.NewStorageAdapter(dbx)
	dispatcher := NewDispatcher()
	poller := NewPoller(store, dispatcher)
	enqueuer := NewDBEnqueuer(dbx)
	emailWorker := &EmailWorker{}
	reportWorker := &ReportWorker{}

	RegisterWorker(dispatcher, emailWorker)
	RegisterWorker(dispatcher, reportWorker)
	return &TestJobService{
		Adapter:      adapter,
		Store:        store,
		Dispatcher:   dispatcher,
		Poller:       poller,
		Enqueuer:     enqueuer,
		EmailWorker:  emailWorker,
		ReportWorker: reportWorker,
	}
}
