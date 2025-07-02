package jobs

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/repository"
	"github.com/tkahng/authgo/internal/stores"
	"github.com/tkahng/authgo/internal/test"
)

func TestPoller_Run(t *testing.T) {
	ctx, dbx := test.DbSetup()

	// test.WithTx(t, func(ctx context.Context, dbx database.Dbx) {

	t.Cleanup(func() {
		_, err := repository.Job.Delete(ctx, dbx, &map[string]any{})
		if err != nil {
			t.Error(err)
		}
	})
	type fields struct {
		Store      JobStore
		Dispatcher Dispatcher
		opts       pollerOpts
	}
	type args struct {
		ctx     context.Context
		wg      *sync.WaitGroup
		args    JobArgs
		testJob *TestJobService
	}
	tests := []struct {
		name    string
		fields  fields
		args    *args
		wantErr bool
		success bool
		setup   func(*args)
	}{
		{
			name:   "",
			fields: fields{},
			args: &args{
				ctx: context.Background(),
				wg:  &sync.WaitGroup{},
				args: EmailJobArgs{
					Recipient: "fail@example.com",
					Subject:   uuid.NewString(),
					Body:      "test email body",
				},
			},
			wantErr: false,
			success: true,
			setup: func(args *args) {
				args.testJob = setupJobs(dbx)
				// args.testJob = testJobs
				args.testJob.Clear()
				if args.wg != nil {
					args.testJob.UseWg(args.wg)
				}

				args.testJob.Worker.WorkFunc = func(ctx context.Context, job *Job[EmailJobArgs]) error {
					args.testJob.Worker.Job = job
					args.testJob.Worker.Success = true
					return nil
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			tt.setup(tt.args)
			tt.args.wg.Add(1)
			testJobs := tt.args.testJob
			ctx, cancel := context.WithCancel(tt.args.ctx)
			defer cancel()

			done := make(chan struct{})

			// Run poller in background
			go func() {
				ServeWithPoller(ctx, testJobs.Poller)

			}()
			// tt.args.args, nil, time.Now(), 1
			if err := testJobs.Manager.Enqueue(ctx, &EnqueueParams{
				Args:        tt.args.args,
				RunAfter:    time.Now(),
				MaxAttempts: 1,
			}); (err != nil) != tt.wantErr {
				t.Errorf("Poller.Run() error = %v, wantErr %v", err, tt.wantErr)
			}
			// Wait for job(s) to complete
			if tt.args.wg != nil {
				tt.args.wg.Wait()
			}
			if tt.success != testJobs.Worker.Success {
				t.Errorf("Poller.Run() job success = %v, want %v", testJobs.Worker.Success, tt.success)
			}
			// Cancel poller context
			close(done)
			cancel()

			// Optional: Wait for poller shutdown to clean up goroutine
			select {
			case <-done:
			case <-time.After(2 * time.Second):
				t.Errorf("poller did not shut down")
			}

		})
	}
	// })
}

type TestJobService struct {
	Manager    JobManager
	Adapter    stores.StorageAdapterInterface
	Store      JobStore
	Dispatcher Dispatcher
	Poller     *DbPoller
	Worker     *EmailWorker
	Job        *Job[EmailJobArgs]
	Wg         *sync.WaitGroup
}

func (s *TestJobService) UseWg(wg *sync.WaitGroup) {
	s.Wg = wg
	s.Worker.Wg = wg
}

func (s *TestJobService) Clear() {
	s.Wg = nil
	s.Worker.Clear()
}

// type

func setupJobs(dbx database.Dbx) *TestJobService {
	adapter := stores.NewStorageAdapter(dbx)
	store := NewDbJobStore(dbx)
	dispatcher := NewDispatcher()
	poller := NewDbPoller(store, dispatcher,
		WithIntervalMs(100), // 100 ms
		WithSize(1),
		WithTimeout(2),
	)

	emailWorker := &EmailWorker{}
	manager := &DbJobManager{
		store:      store,
		poller:     poller,
		dispatcher: dispatcher,
	}
	RegisterWorker(dispatcher, emailWorker)
	return &TestJobService{
		Manager:    manager,
		Adapter:    adapter,
		Store:      store,
		Dispatcher: dispatcher,
		Poller:     poller,
		Worker:     emailWorker,
	}
}
