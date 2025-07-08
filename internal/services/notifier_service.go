package services

import (
	"context"
	"log/slog"

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

// Listen implements NotifierService.
func (d *DbNotifierService) Subscribe(channel string) notifier.Subscription {
	return d.notifier.Subscribe(channel)
}

// Stope implements NotifierService.
func (d *DbNotifierService) Stop(ctx context.Context) error {
	panic("unimplemented")
}

// Start implements NotifierService.
func (d *DbNotifierService) Start(ctx context.Context) (err error) {
	err = d.listener.Connect(ctx)
	if err != nil {
		return err
	}
	d.notifier = notifier.NewNotifier(slog.Default(), d.listener)
	go func() {
		for {
			if runErr := d.notifier.Run(context.Background()); runErr != nil {
				err = runErr
				return
			}
		}
	}()
	return
}

var _ NotifierService = (*DbNotifierService)(nil)
