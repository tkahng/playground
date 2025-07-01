package jobs

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/tkahng/authgo/internal/models"
	"golang.org/x/sync/errgroup"
)

func ServeWithPoller(ctx context.Context, poller *Poller) {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		return poller.Run(ctx)
	})

	g.Wait()
}

type pollerOpts struct {
	Interval time.Duration
	Timeout  time.Duration
	Size     int
}
type PollerOptsFunc func(*pollerOpts)

func WithTimeout(timeout int64) PollerOptsFunc {
	return func(opts *pollerOpts) {
		opts.Timeout = time.Duration(timeout) * time.Second
	}
}

func WithInterval(interval int64) PollerOptsFunc {
	return func(opts *pollerOpts) {
		opts.Interval = time.Duration(interval) * time.Second
	}
}

func WithSize(size int) PollerOptsFunc {
	return func(opts *pollerOpts) {
		opts.Size = size
	}
}

type Poller struct {
	Store      JobStore
	Dispatcher Dispatcher
	opts       pollerOpts
}

func NewPoller(store JobStore, dispatcher Dispatcher, opts ...PollerOptsFunc) *Poller {
	p := &Poller{
		Store:      store,
		Dispatcher: dispatcher,
		opts: pollerOpts{
			Interval: 5 * time.Second,
			Timeout:  30 * time.Second,
			Size:     1,
		},
	}
	for _, opt := range opts {
		opt(&p.opts)
	}
	return p
}

func (p *Poller) Run(ctx context.Context) error {
	ticker := time.NewTicker(p.opts.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := p.pollOnce(ctx); err != nil {
				slog.ErrorContext(ctx, "poller error", "error", err)
			}
		}
	}
}

func (p *Poller) pollOnce(ctx context.Context) error {
	return p.Store.RunInTx(
		ctx,
		func(js JobStore) error {

			// Claim only one job
			jobs, err := js.ClaimPendingJobs(ctx, p.opts.Size) // LIMIT to 1
			if err != nil {
				return fmt.Errorf("claim jobs: %w", err)
			}

			if len(jobs) == 0 {
				// No jobs to process, commit and return
				return nil // Commit an empty transaction to avoid erroring if there are no jobs
			}

			timeout := p.opts.Timeout
			if timeout == 0 { // Provide a default if not set by options
				timeout = 30 * time.Second
			}

			jobCtx, cancel := context.WithTimeout(
				ctx,
				timeout,
			)
			defer cancel()

			for _, job := range jobs {
				if err := p.dispatch(jobCtx, job, js); err != nil {
					slog.ErrorContext(ctx, "dispatch error", "error", err)
					return err
				}
			}
			return nil
		},
	)
}

func (p *Poller) dispatch(ctx context.Context, row *models.JobRow, js JobStore) error {
	dispatchErr := p.Dispatcher.Dispatch(ctx, row)
	if dispatchErr != nil {
		slog.ErrorContext(ctx, "there was an error dispatching the job. will attempt to reschedule or mark as failed", slog.Any("error", dispatchErr), slog.String("job_id", row.ID.String()))
		if row.Attempts >= row.MaxAttempts {
			if markFailedErr := js.MarkFailed(ctx, row.ID, dispatchErr.Error()); markFailedErr != nil {
				slog.ErrorContext(ctx, "Error marking job as failed (and rolling back)", slog.Any("error", markFailedErr), slog.String("job_id", row.ID.String()))
				return fmt.Errorf("failed to mark job %s as failed: %w", row.ID, markFailedErr)
			}
		} else {
			delay := time.Duration(math.Pow(2, float64(row.Attempts))) * time.Second
			if rescheduleErr := js.RescheduleJob(ctx, row.ID, delay); rescheduleErr != nil {
				slog.ErrorContext(ctx, "Error rescheduling job (and rolling back)", slog.Any("error", rescheduleErr), slog.String("job_id", row.ID.String()))
				return fmt.Errorf("failed to reschedule job %s: %w", row.ID, rescheduleErr)
			}
		}
	} else {
		if markDoneErr := js.MarkDone(ctx, row.ID); markDoneErr != nil {
			slog.ErrorContext(ctx, "Error marking job as done (and rolling back)", slog.Any("error", markDoneErr), slog.String("job_id", row.ID.String()))
			return fmt.Errorf("failed to mark job %s as done: %w", row.ID, markDoneErr)
		}
	}

	return nil
}
