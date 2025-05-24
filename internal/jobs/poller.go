package jobs

import (
	"context"
	"fmt"
	"log"
	"math"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5"
	"golang.org/x/sync/errgroup"
)

func ServeWithPoller(ctx context.Context, poller *Poller) error {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		return poller.Run(ctx)
	})

	return g.Wait()
}

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
	tx, err := p.Store.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	jobs, err := p.Store.ClaimPendingJobs(ctx, tx, 10)
	if err != nil {
		return fmt.Errorf("claim jobs: %w", err)
	}

	for _, row := range jobs {
		row := row
		func() {
			jobCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
			defer cancel()

			err := p.Dispatcher.Dispatch(jobCtx, &row)
			if err != nil {
				if row.Attempts >= row.MaxAttempts {
					_ = p.Store.MarkFailed(ctx, tx, row.ID, err.Error())
				} else {
					delay := time.Duration(math.Pow(2, float64(row.Attempts))) * time.Second
					_ = p.Store.RescheduleJob(ctx, tx, row.ID, delay)
				}
			} else {
				_ = p.Store.MarkDone(ctx, tx, row.ID)
			}
		}()
	}

	return tx.Commit(ctx)
}
