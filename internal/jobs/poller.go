package jobs

import (
	"context"
	"log"
	"math"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"
)

type Poller struct {
	Store      *JobStore
	Dispatcher *Dispatcher
	Interval   time.Duration
}

func (p *Poller) Run(ctx context.Context) error {
	ticker := time.NewTicker(p.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := p.pollOnce(ctx); err != nil {
				log.Printf("poller error: %v", err)
			}
		}
	}
}

func (p *Poller) pollOnce(ctx context.Context) error {
	jobs, err := p.Store.ClaimPendingJobs(ctx, 10)
	if err != nil {
		return err
	}

	for _, row := range jobs {
		row := row // capture for closure
		go func() {
			ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
			defer cancel()

			err := p.Dispatcher.Dispatch(ctx, &row)
			if err != nil {
				if row.Attempts >= row.MaxAttempts {
					_ = p.Store.MarkFailed(ctx, row.ID, err.Error())
				} else {
					delay := time.Duration(math.Pow(2, float64(row.Attempts))) * time.Second
					_ = p.Store.RescheduleJob(ctx, row.ID, delay)
				}
			} else {
				_ = p.Store.MarkDone(ctx, row.ID)
			}
		}()
	}

	return nil
}

func ServeWithPoller(ctx context.Context, poller *Poller) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer cancel()

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		return poller.Run(ctx)
	})

	if err := g.Wait(); err != nil {
		log.Printf("graceful shutdown: %v", err)
		return err
	}
	return nil
}
