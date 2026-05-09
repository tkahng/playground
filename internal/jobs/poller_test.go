package jobs

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/conf"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/database/repository"
	"github.com/tkahng/playground/internal/stores"
)

func TestPoller_Run(t *testing.T) {
	ctx := context.Background()
	cfg := conf.ZeroEnvConfig()
	dbx, err := database.CreateNewQueriesContext(ctx, cfg.Db.GetDatabaseURL())
	if err != nil {
		t.Fatalf("failed to create database pool: %v", err)
	}
	t.Cleanup(func() {
		_, err := repository.Job.Delete(ctx, dbx, &map[string]any{})
		dbx.Close()
		if err != nil {
			t.Error(err)
		}
	})
	t.Run("Poller.Run", func(t *testing.T) {
		wantErr := false
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

		done := make(chan struct{})

		// Run poller in background
		go func() {
			ServeWithPoller(ctx, testJobs.Poller)
		}()
		// tt.args.args, nil, time.Now(), 1
		if err := testJobs.Manager.Enqueue(ctx, &EnqueueParams{
			Args:        args,
			RunAfter:    time.Now(),
			MaxAttempts: 1,
		}); (err != nil) != wantErr {
			t.Errorf("Poller.Run() error = %v, wantErr %v", err, wantErr)
		}

		wg.Wait()
		if !testJobs.Worker.Success {
			t.Errorf("Poller.Run() job success = %v", testJobs.Worker.Success)
		}
		// Remove committed jobs immediately so parallel package tests cannot
		// claim them via PollOnce before the outer t.Cleanup runs.
		if _, err := repository.Job.Delete(ctx, dbx, &map[string]any{}); err != nil {
			t.Errorf("subtest job cleanup error: %v", err)
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
