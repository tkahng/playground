package jobs

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/stores"
)

func TestPoller_Run(t *testing.T) {
	database.WithNewDatabase(t, func(ctx context.Context, dbx database.Dbx) {
		t.Run("Poller.Run", func(t *testing.T) {
			wg := &sync.WaitGroup{}
			args := EmailJobArgs{
				Recipient: "fail@example.com",
				Subject:   uuid.NewString(),
				Body:      "test email body",
			}
			testJobs := setupJobs(dbx)
			testJobs.Worker.WorkFunc = func(ctx context.Context, job *Job[EmailJobArgs]) error {
				testJobs.Worker.Job = job
				testJobs.Worker.Success = true
				return nil
			}
			testJobs.UseWg(wg)
			wg.Add(1)

			ctx, cancel := context.WithCancel(ctx)
			defer cancel()

			go func() {
				ServeWithPoller(ctx, testJobs.Poller)
			}()

			if err := testJobs.Manager.Enqueue(ctx, &EnqueueParams{
				Args:        args,
				RunAfter:    time.Now(),
				MaxAttempts: 1,
			}); err != nil {
				t.Errorf("Poller.Run() enqueue error = %v", err)
			}

			wg.Wait()
			if !testJobs.Worker.Success {
				t.Errorf("Poller.Run() job success = %v", testJobs.Worker.Success)
			}
			cancel()
		})
	})
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
