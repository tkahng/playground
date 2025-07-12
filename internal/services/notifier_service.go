package services

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/tools/notifier"
)

type NotifierService interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Subscribe(channel string) notifier.Subscription
}

type DbNotifierService struct {
	notifier notifier.Notifier
	listener notifier.Listener
}

func NewDbNotifierService(ctx context.Context, dbx database.Dbx, l *slog.Logger) *DbNotifierService {
	li := notifier.NewListener(dbx)
	err := li.Connect(ctx)
	if err != nil {
		panic(fmt.Errorf("error connecting to database: %w", err))
	}

	n := notifier.NewNotifier(l, li)
	return &DbNotifierService{
		listener: li,
		notifier: n,
	}
}

// Listen implements NotifierService.
func (d *DbNotifierService) Subscribe(channel string) notifier.Subscription {
	return d.notifier.Subscribe(channel)
}

// Stope implements NotifierService.
func (d *DbNotifierService) Stop(ctx context.Context) error {
	return d.listener.Close(ctx)
}

// Start implements NotifierService.
func (d *DbNotifierService) Start(ctx context.Context) (err error) {
	err = d.listener.Connect(ctx)
	if err != nil {
		slog.ErrorContext(
			ctx,
			"error connecting to database",
			slog.Any("error", err),
		)
		if newerr := d.listener.Close(ctx); newerr != nil {
			slog.ErrorContext(
				ctx,
				"error closing connection to database",
				slog.Any("error", newerr),
			)
		}
		return
	}
	d.notifier = notifier.NewNotifier(slog.Default(), d.listener)
	go func() {
		for {
			if runErr := d.notifier.Run(context.Background()); runErr != nil {
				slog.ErrorContext(
					ctx,
					"error running notifier",
					slog.Any("error", runErr),
				)
				err = runErr
				return
			}
		}
	}()
	return
}
