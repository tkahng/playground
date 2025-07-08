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

type StartEvent struct {
	hook.Event
	WaitEvent
	App    App
	Server *http.Server
	Api    huma.API
}

type StopEvent struct {
	hook.Event
	App App
}

type lifecycle struct {
	logger *slog.Logger

	onStart *hook.Hook[*StartEvent]

	onStop *hook.Hook[*StopEvent]
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

// Start implements Lifecycle.
func (l *lifecycle) Start(e *StartEvent, oneOffHandlerFuncs ...func(*StartEvent) error) error {
	return l.onStart.Trigger(e, oneOffHandlerFuncs...)
}

// Stop implements Lifecycle.
func (l *lifecycle) Stop(e *StopEvent, oneOffHandlerFuncs ...func(*StopEvent) error) error {
	return l.OnStop().Trigger(e, oneOffHandlerFuncs...)
}

type Lifecycle interface {
	OnStart() *hook.Hook[*StartEvent]
	OnStop() *hook.Hook[*StopEvent]
}

var _ Lifecycle = (*lifecycle)(nil)

func NewLifecycle(logger *slog.Logger) Lifecycle {
	return &lifecycle{
		logger:  logger,
		onStart: &hook.Hook[*StartEvent]{},
		onStop:  &hook.Hook[*StopEvent]{},
	}
}
