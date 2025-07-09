package core

import (
	"log/slog"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/authgo/internal/tools/hook"
	"golang.org/x/sync/errgroup"
)

type WaitEvent struct {
	wg *errgroup.Group
}

func (e *WaitEvent) Go(f func() error) {
	e.wg.Go(f)
}

type BootstrapEvent struct {
	hook.Event
	App App
}

type StartEvent struct {
	hook.Event
	WaitEvent
	App    App
	Server *http.Server
	Api    huma.API
}

type StopEvent struct {
	hook.Event
	App       App
	IsRestart bool
}
type Lifecycle interface {
	OnBootstrap() *hook.Hook[*BootstrapEvent]
	OnStart() *hook.Hook[*StartEvent]
	OnStop() *hook.Hook[*StopEvent]
}
type lifecycle struct {
	logger *slog.Logger

	onBootstrap *hook.Hook[*BootstrapEvent]

	onStart *hook.Hook[*StartEvent]

	onStop *hook.Hook[*StopEvent]
}

// OnBootstrap implements Lifecycle.
func (l *lifecycle) OnBootstrap() *hook.Hook[*BootstrapEvent] {
	if l.onBootstrap == nil {
		l.onBootstrap = &hook.Hook[*BootstrapEvent]{}
	}
	return l.onBootstrap
}

// OnStart implements Lifecycle.
func (l *lifecycle) OnStart() *hook.Hook[*StartEvent] {
	if l.onStart == nil {
		l.onStart = &hook.Hook[*StartEvent]{}
	}
	return l.onStart
}

// OnStop implements Lifecycle.
func (l *lifecycle) OnStop() *hook.Hook[*StopEvent] {
	if l.onStop == nil {
		l.onStop = &hook.Hook[*StopEvent]{}
	}
	return l.onStop
}

var _ Lifecycle = (*lifecycle)(nil)

func NewLifecycle(logger *slog.Logger) Lifecycle {
	return &lifecycle{
		logger:      logger,
		onStart:     &hook.Hook[*StartEvent]{},
		onStop:      &hook.Hook[*StopEvent]{},
		onBootstrap: &hook.Hook[*BootstrapEvent]{},
	}
}
