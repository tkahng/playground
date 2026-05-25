package workers

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/database/repository"
	"github.com/tkahng/playground/internal/jobs"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/tools/types"
)

// mockRematchSvc wraps mockRpsGameService to track ExpireRematches calls separately.
type mockRematchSvc struct {
	calls     int
	processed int
	err       error
}

func (m *mockRematchSvc) ExpireGamesAndRefundBets(_ context.Context) (int, error) {
	return 0, nil
}

func (m *mockRematchSvc) ExpireRematches(_ context.Context) (int, error) {
	m.calls++
	return m.processed, m.err
}

func makeRematchJob() *jobs.Job[RpsRematchExpiryJobArgs] {
	return &jobs.Job[RpsRematchExpiryJobArgs]{
		JobRow: &models.JobRow{},
		Args:   RpsRematchExpiryJobArgs{},
	}
}

func TestRpsRematchExpiryWorker_Work_Success_ReEnqueues(t *testing.T) {
	svc := &mockRematchSvc{processed: 2}
	enq := &mockEnqueuer{}
	worker := NewRpsRematchExpiryWorker(svc, enq)

	if err := worker.Work(context.Background(), makeRematchJob()); err != nil {
		t.Fatalf("Work() error = %v", err)
	}

	if svc.calls != 1 {
		t.Errorf("ExpireRematches called %d times, want 1", svc.calls)
	}
	if len(enq.enqueued) != 1 {
		t.Fatalf("re-enqueue count = %d, want 1", len(enq.enqueued))
	}

	next := enq.enqueued[0]
	if next.Args.Kind() != RpsRematchExpirySweepKind {
		t.Errorf("re-enqueued kind = %q, want %q", next.Args.Kind(), RpsRematchExpirySweepKind)
	}

	wantAfter := time.Now().Add(RpsRematchExpirySweepInterval)
	delta := next.RunAfter.Sub(wantAfter)
	if delta < -2*time.Second || delta > 2*time.Second {
		t.Errorf("re-enqueued RunAfter delta = %v, want within ±2s", delta)
	}

	if next.UniqueKey == nil || *next.UniqueKey != RpsRematchExpirySweepKind {
		t.Errorf("re-enqueued UniqueKey = %v, want %q", next.UniqueKey, RpsRematchExpirySweepKind)
	}
	if next.MaxAttempts != 1 {
		t.Errorf("re-enqueued MaxAttempts = %d, want 1", next.MaxAttempts)
	}
}

func TestRpsRematchExpiryWorker_Work_ZeroExpired_StillReEnqueues(t *testing.T) {
	svc := &mockRematchSvc{processed: 0}
	enq := &mockEnqueuer{}
	worker := NewRpsRematchExpiryWorker(svc, enq)

	if err := worker.Work(context.Background(), makeRematchJob()); err != nil {
		t.Fatalf("Work() error = %v", err)
	}
	if len(enq.enqueued) != 1 {
		t.Errorf("re-enqueue count = %d, want 1 even when nothing expired", len(enq.enqueued))
	}
}

func TestRpsRematchExpiryWorker_Work_ServiceError_DoesNotReEnqueue(t *testing.T) {
	svc := &mockRematchSvc{err: errors.New("db failure")}
	enq := &mockEnqueuer{}
	worker := NewRpsRematchExpiryWorker(svc, enq)

	err := worker.Work(context.Background(), makeRematchJob())
	if err == nil {
		t.Fatal("Work() expected error, got nil")
	}
	if len(enq.enqueued) != 0 {
		t.Errorf("re-enqueue count = %d, want 0 on service error", len(enq.enqueued))
	}
}

func TestRpsRematchExpiryWorker_Work_EnqueueError_DoesNotSurfaceError(t *testing.T) {
	svc := &mockRematchSvc{processed: 1}
	enq := &mockEnqueuer{err: errors.New("enqueue failed")}
	worker := NewRpsRematchExpiryWorker(svc, enq)

	if err := worker.Work(context.Background(), makeRematchJob()); err != nil {
		t.Errorf("Work() returned %v, want nil (enqueue failure is non-fatal)", err)
	}
}

func TestSeedRpsRematchExpiryJob_InsertsJobWithCorrectKind(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		manager := jobs.NewDbJobManager(db)

		if err := SeedRpsRematchExpiryJob(ctx, manager); err != nil {
			t.Fatalf("SeedRpsRematchExpiryJob() error = %v", err)
		}

		rows, err := repository.Job.Get(ctx, db, &map[string]any{
			"kind": RpsRematchExpirySweepKind,
		}, nil, nil, nil)
		if err != nil {
			t.Fatalf("get job rows: %v", err)
		}
		if len(rows) != 1 {
			t.Fatalf("job count = %d, want 1", len(rows))
		}
		if rows[0].Kind != RpsRematchExpirySweepKind {
			t.Errorf("job kind = %q, want %q", rows[0].Kind, RpsRematchExpirySweepKind)
		}
		if rows[0].UniqueKey == nil || *rows[0].UniqueKey != RpsRematchExpirySweepKind {
			t.Errorf("job UniqueKey = %v, want %q", rows[0].UniqueKey, RpsRematchExpirySweepKind)
		}
	})
}

func TestSeedRpsRematchExpiryJob_Idempotent(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		manager := jobs.NewDbJobManager(db)
		for range 3 {
			if err := SeedRpsRematchExpiryJob(ctx, manager); err != nil {
				t.Fatalf("SeedRpsRematchExpiryJob() error = %v", err)
			}
		}
		rows, err := repository.Job.Get(ctx, db, &map[string]any{
			"kind": RpsRematchExpirySweepKind,
		}, nil, nil, nil)
		if err != nil {
			t.Fatalf("get job rows: %v", err)
		}
		if len(rows) != 1 {
			t.Errorf("job count = %d, want 1 (idempotent)", len(rows))
		}
	})
}

func TestRpsRematchExpiryWorker_Integration_SelfScheduling(t *testing.T) {
	ctx := context.Background()
	db := newTestDb(t)
	t.Cleanup(func() {
		repository.Job.Delete(ctx, db, &map[string]any{"kind": RpsRematchExpirySweepKind}) //nolint:errcheck
	})

	svc := &mockRematchSvc{processed: 1}
	manager := jobs.NewDbJobManager(db)
	dispatcher := jobs.NewDispatcher()
	poller := jobs.NewDbPoller(jobs.NewDbJobStore(db), dispatcher, jobs.WithIntervalMs(50), jobs.WithSize(1), jobs.WithTimeout(5))

	worker := NewRpsRematchExpiryWorker(svc, manager)
	jobs.RegisterWorker(dispatcher, worker)

	if err := SeedRpsRematchExpiryJob(ctx, manager); err != nil {
		t.Fatalf("SeedRpsRematchExpiryJob() error = %v", err)
	}

	runCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	go poller.Run(runCtx) //nolint:errcheck
	<-runCtx.Done()

	if svc.calls == 0 {
		t.Error("ExpireRematches was never called")
	}

	rows, err := repository.Job.Get(ctx, db, &map[string]any{
		"kind":       RpsRematchExpirySweepKind,
		"unique_key": types.Pointer(RpsRematchExpirySweepKind),
	}, nil, nil, nil)
	if err != nil {
		t.Fatalf("find re-enqueued job: %v", err)
	}
	if len(rows) == 0 {
		t.Error("expected a re-enqueued sweep job after execution, found none")
	}
	for _, row := range rows {
		if row.Status == models.JobStatusDone {
			continue
		}
		if row.Status != models.JobStatusPending {
			t.Errorf("re-enqueued job status = %q, want %q", row.Status, models.JobStatusPending)
		}
	}
}
