package jobs

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/sync/errgroup"
)

type Poller struct {
	DB         *pgxpool.Pool
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
	rows, err := p.DB.Query(ctx, `
		UPDATE jobs SET status='processing', updated_at=now()
		WHERE id IN (
			SELECT id FROM jobs
			WHERE status = 'pending' AND run_after <= now()
			ORDER BY run_after
			LIMIT 10 FOR UPDATE SKIP LOCKED
		)
		RETURNING id, kind, unique_key, payload, status, run_after, attempts, max_attempts, last_error, created_at, updated_at
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var row JobRow
		if err := rows.Scan(
			&row.ID, &row.Kind, &row.UniqueKey, &row.Payload, &row.Status, &row.RunAfter,
			&row.Attempts, &row.MaxAttempts, &row.LastError, &row.CreatedAt, &row.UpdatedAt,
		); err != nil {
			log.Printf("scan error: %v", err)
			continue
		}
		go func(r JobRow) {
			ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
			defer cancel()
			_ = p.Dispatcher.Dispatch(ctx, &r)
		}(row)
	}

	return rows.Err()
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
