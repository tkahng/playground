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

	"github.com/tkahng/playground/internal/models"
	"golang.org/x/sync/errgroup"
)

func ServeWithPoller(ctx context.Context, poller *DbPoller) {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		return poller.Run(ctx)
	})

	if err := g.Wait(); err != nil {
		slog.ErrorContext(ctx, "poller error", "error", err)
	}
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

func WithIntervalS(interval int64) PollerOptsFunc {
	return func(opts *pollerOpts) {
		opts.Interval = time.Duration(interval) * time.Second
	}
}

func WithIntervalMs(interval int64) PollerOptsFunc {
	return func(opts *pollerOpts) {
		opts.Interval = time.Duration(interval) * time.Millisecond
	}
}

func WithSize(size int) PollerOptsFunc {
	return func(opts *pollerOpts) {
		opts.Size = size
	}
}

type Poller interface {
	Run(ctx context.Context) error
	PollOnce(ctx context.Context) error
}

type DbPoller struct {
	Store      JobStore
	Dispatcher Dispatcher
	opts       pollerOpts
}

var _ Poller = (*DbPoller)(nil)

func NewDbPoller(store JobStore, dispatcher Dispatcher, opts ...PollerOptsFunc) *DbPoller {
	p := &DbPoller{
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

func (p *DbPoller) Run(ctx context.Context) error {
	ticker := time.NewTicker(p.opts.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := p.PollOnce(ctx); err != nil {
				slog.ErrorContext(ctx, "poller error", "error", err)
			}
		}
	}
}

func (p *DbPoller) PollOnce(ctx context.Context) error {
	// Use a timeout for the transaction itself
	firstCancelCtx, firstCancelFunc := context.WithTimeout(ctx, p.opts.Timeout)
	defer firstCancelFunc()

	var claimedJobs []*models.JobRow
	err := p.Store.RunInTx(firstCancelCtx, func(js JobStore) error {
		jobs, err := js.ClaimPendingJobs(firstCancelCtx, p.opts.Size)
		if err != nil {
			return fmt.Errorf("claim jobs: %w", err)
		}
		claimedJobs = jobs
		return nil // commit the transaction even if no jobs
	})
	if err != nil {
		return fmt.Errorf("run tx: %w", err)
	}
	if len(claimedJobs) == 0 {
		return nil // nothing to do
	}

	sem := make(chan struct{}, p.opts.Size) // Limit concurrency to `Size`
	g, gctx := errgroup.WithContext(ctx)

	for _, job := range claimedJobs {
		sem <- struct{}{}

		g.Go(func() error {
			defer func() { <-sem }()

			// Set timeout for this job
			jobTimeoutCtx, cancel := context.WithTimeout(gctx, p.opts.Timeout)
			defer cancel()

			dispatchErr := p.Dispatcher.Dispatch(jobTimeoutCtx, job)

			// Use new transaction to mark result
			markErr := p.Store.RunInTx(jobTimeoutCtx, func(js JobStore) error {
				if dispatchErr != nil {
					slog.ErrorContext(jobTimeoutCtx, "job failed", "error", dispatchErr, "job_id", job.ID.String())

					if job.Attempts >= job.MaxAttempts {
						return js.MarkFailed(jobTimeoutCtx, job.ID, dispatchErr.Error())
					}
					// Reschedule with exponential backoff
					delay := time.Duration(math.Pow(2, float64(job.Attempts))) * time.Second
					return js.RescheduleJob(jobTimeoutCtx, job.ID, delay)
				}

				return js.MarkDone(jobTimeoutCtx, job.ID)

			})
			if markErr != nil {
				slog.ErrorContext(jobTimeoutCtx, "error updating job status", "error", markErr, "job_id", job.ID.String())
			}

			return nil // always return nil to allow others to proceed
		})
	}

	return g.Wait()
}
