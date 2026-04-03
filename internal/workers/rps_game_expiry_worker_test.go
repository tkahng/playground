package workers

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/tkahng/playground/internal/conf"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/database/repository"
	"github.com/tkahng/playground/internal/jobs"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/tools/types"
)

// --- mocks ---

type mockRpsGameService struct {
	calls     int
	processed int
	err       error
}

func (m *mockRpsGameService) ExpireGamesAndRefundBets(_ context.Context) (int, error) {
	m.calls++
	return m.processed, m.err
}

type mockEnqueuer struct {
	enqueued []*jobs.EnqueueParams
	err      error
}

func (m *mockEnqueuer) Enqueue(_ context.Context, p *jobs.EnqueueParams) error {
	if m.err != nil {
		return m.err
	}
	m.enqueued = append(m.enqueued, p)
	return nil
}

func (m *mockEnqueuer) EnqueueMany(_ context.Context, params ...*jobs.EnqueueParams) error {
	for _, p := range params {
		m.enqueued = append(m.enqueued, p)
	}
	return nil
}

// makeJob returns a minimal *jobs.Job[RpsGameExpiryJobArgs] for test use.
func makeJob() *jobs.Job[RpsGameExpiryJobArgs] {
	return &jobs.Job[RpsGameExpiryJobArgs]{
		JobRow: &models.JobRow{},
		Args:   RpsGameExpiryJobArgs{},
	}
}

// newTestDb creates a real DB connection for integration tests.
func newTestDb(t *testing.T) database.Dbx {
	t.Helper()
	ctx := context.Background()
	cfg := conf.ZeroEnvConfig()
	db := database.CreateNewQueriesContext(ctx, cfg.Db.GetDatabaseUrl())
	t.Cleanup(func() { db.Close() })
	return db
}

// --- RpsGameExpiryWorker.Work ---

func TestRpsGameExpiryWorker_Work_Success_ReEnqueues(t *testing.T) {
	svc := &mockRpsGameService{processed: 3}
	enq := &mockEnqueuer{}
	worker := NewRpsGameExpiryWorker(svc, enq)

	if err := worker.Work(context.Background(), makeJob()); err != nil {
		t.Fatalf("Work() error = %v", err)
	}

	if svc.calls != 1 {
		t.Errorf("ExpireGamesAndRefundBets called %d times, want 1", svc.calls)
	}

	// Must re-enqueue exactly once.
	if len(enq.enqueued) != 1 {
		t.Fatalf("re-enqueue count = %d, want 1", len(enq.enqueued))
	}

	next := enq.enqueued[0]

	// Kind must match.
	if next.Args.Kind() != RpsGameExpirySweepKind {
		t.Errorf("re-enqueued job kind = %q, want %q", next.Args.Kind(), RpsGameExpirySweepKind)
	}

	// RunAfter must be approximately now + interval.
	wantAfter := time.Now().Add(RpsGameExpirySweepInterval)
	delta := next.RunAfter.Sub(wantAfter)
	if delta < -2*time.Second || delta > 2*time.Second {
		t.Errorf("re-enqueued RunAfter = %v, want ~%v (delta %v)", next.RunAfter, wantAfter, delta)
	}

	// UniqueKey must be set.
	if next.UniqueKey == nil || *next.UniqueKey != RpsGameExpirySweepKind {
		t.Errorf("re-enqueued UniqueKey = %v, want %q", next.UniqueKey, RpsGameExpirySweepKind)
	}

	// MaxAttempts should be 1 (sweep is idempotent, no retries needed).
	if next.MaxAttempts != 1 {
		t.Errorf("re-enqueued MaxAttempts = %d, want 1", next.MaxAttempts)
	}
}

func TestRpsGameExpiryWorker_Work_ZeroProcessed_StillReEnqueues(t *testing.T) {
	svc := &mockRpsGameService{processed: 0}
	enq := &mockEnqueuer{}
	worker := NewRpsGameExpiryWorker(svc, enq)

	if err := worker.Work(context.Background(), makeJob()); err != nil {
		t.Fatalf("Work() error = %v", err)
	}

	// Even when nothing expired, the next sweep must be scheduled.
	if len(enq.enqueued) != 1 {
		t.Errorf("re-enqueue count = %d, want 1 (sweep should always re-schedule)", len(enq.enqueued))
	}
}

func TestRpsGameExpiryWorker_Work_ServiceError_DoesNotReEnqueue(t *testing.T) {
	svc := &mockRpsGameService{err: errors.New("db down")}
	enq := &mockEnqueuer{}
	worker := NewRpsGameExpiryWorker(svc, enq)

	err := worker.Work(context.Background(), makeJob())
	if err == nil {
		t.Fatal("Work() expected error, got nil")
	}
	if err.Error() != "db down" {
		t.Errorf("Work() error = %q, want %q", err.Error(), "db down")
	}

	// No re-enqueue on failure — the caller (poller) handles retry/failure marking.
	if len(enq.enqueued) != 0 {
		t.Errorf("re-enqueue count = %d, want 0 on service error", len(enq.enqueued))
	}
}

func TestRpsGameExpiryWorker_Work_EnqueueError_DoesNotSurfaceError(t *testing.T) {
	// If re-enqueue fails, Work should still return nil — the current sweep succeeded.
	// The seed on next startup will recover the schedule.
	svc := &mockRpsGameService{processed: 1}
	enq := &mockEnqueuer{err: errors.New("enqueue failed")}
	worker := NewRpsGameExpiryWorker(svc, enq)

	if err := worker.Work(context.Background(), makeJob()); err != nil {
		t.Errorf("Work() returned error %v, want nil (enqueue failure is non-fatal)", err)
	}
}

// --- SeedRpsGameExpiryJob ---

func TestSeedRpsGameExpiryJob_InsertsJobWithCorrectKind(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		manager := jobs.NewDbJobManager(db)

		if err := SeedRpsGameExpiryJob(ctx, manager); err != nil {
			t.Fatalf("SeedRpsGameExpiryJob() error = %v", err)
		}

		rows, err := repository.Job.Get(ctx, db, &map[string]any{
			"kind": RpsGameExpirySweepKind,
		}, nil, nil, nil)
		if err != nil {
			t.Fatalf("get job rows: %v", err)
		}
		if len(rows) != 1 {
			t.Fatalf("job count = %d, want 1", len(rows))
		}
		if rows[0].Kind != RpsGameExpirySweepKind {
			t.Errorf("job kind = %q, want %q", rows[0].Kind, RpsGameExpirySweepKind)
		}
		if rows[0].UniqueKey == nil || *rows[0].UniqueKey != RpsGameExpirySweepKind {
			t.Errorf("job UniqueKey = %v, want %q", rows[0].UniqueKey, RpsGameExpirySweepKind)
		}
	})
}

func TestSeedRpsGameExpiryJob_Idempotent(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		manager := jobs.NewDbJobManager(db)

		// Two calls should not create two rows.
		for range 2 {
			if err := SeedRpsGameExpiryJob(ctx, manager); err != nil {
				t.Fatalf("SeedRpsGameExpiryJob() error = %v", err)
			}
		}

		rows, err := repository.Job.Get(ctx, db, &map[string]any{
			"kind": RpsGameExpirySweepKind,
		}, nil, nil, nil)
		if err != nil {
			t.Fatalf("get job rows: %v", err)
		}
		if len(rows) != 1 {
			t.Errorf("job count = %d, want 1 (idempotent seed)", len(rows))
		}
	})
}

func TestSeedRpsGameExpiryJob_MaxAttemptsIsOne(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		manager := jobs.NewDbJobManager(db)

		if err := SeedRpsGameExpiryJob(ctx, manager); err != nil {
			t.Fatalf("SeedRpsGameExpiryJob() error = %v", err)
		}

		rows, err := repository.Job.Get(ctx, db, &map[string]any{"kind": RpsGameExpirySweepKind}, nil, nil, nil)
		if err != nil {
			t.Fatalf("get job rows: %v", err)
		}
		if len(rows) != 1 {
			t.Fatalf("job count = %d, want 1", len(rows))
		}
		if rows[0].MaxAttempts != 1 {
			t.Errorf("MaxAttempts = %d, want 1", rows[0].MaxAttempts)
		}
	})
}

// --- Integration: seed → poller executes → re-enqueue appears in DB ---

func TestRpsGameExpiryWorker_Integration_SelfScheduling(t *testing.T) {
	ctx := context.Background()
	db := newTestDb(t)
	t.Cleanup(func() {
		repository.Job.Delete(ctx, db, &map[string]any{"kind": RpsGameExpirySweepKind}) //nolint:errcheck
	})

	svc := &mockRpsGameService{processed: 2}

	manager := jobs.NewDbJobManager(db)
	dispatcher := jobs.NewDispatcher()
	poller := jobs.NewDbPoller(jobs.NewDbJobStore(db), dispatcher, jobs.WithIntervalMs(50), jobs.WithSize(1), jobs.WithTimeout(5))

	worker := NewRpsGameExpiryWorker(svc, manager)
	jobs.RegisterWorker(dispatcher, worker)

	// Seed the first job.
	if err := SeedRpsGameExpiryJob(ctx, manager); err != nil {
		t.Fatalf("SeedRpsGameExpiryJob() error = %v", err)
	}

	// Run the poller briefly — long enough for one execution.
	runCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	go poller.Run(runCtx) //nolint:errcheck
	<-runCtx.Done()

	// Service should have been called.
	if svc.calls == 0 {
		t.Error("ExpireGamesAndRefundBets was never called")
	}

	// A new pending job should have been re-enqueued by the worker.
	rows, err := repository.Job.Get(ctx, db, &map[string]any{
		"kind":       RpsGameExpirySweepKind,
		"unique_key": types.Pointer(RpsGameExpirySweepKind),
	}, nil, nil, nil)
	if err != nil {
		t.Fatalf("find re-enqueued job: %v", err)
	}
	// At least one row should exist (the re-scheduled next run).
	if len(rows) == 0 {
		t.Error("expected a re-enqueued sweep job after execution, found none")
	}
	// The re-enqueued job should be pending (not processing/done), as its RunAfter is in the future.
	for _, row := range rows {
		if row.Status == models.JobStatusDone {
			continue // the completed one is fine
		}
		if row.Status != models.JobStatusPending {
			t.Errorf("re-enqueued job status = %q, want %q", row.Status, models.JobStatusPending)
		}
	}
}
