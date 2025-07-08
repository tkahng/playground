package core

import (
	"log/slog"

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
	App App
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

func (l *lifecycle) Init() {
	l.onStart = &hook.Hook[*StartEvent]{}
	l.onStop = &hook.Hook[*StopEvent]{}
}

type Lifecycle interface {
	Init()
	Start(e *StartEvent, fns ...func(*StartEvent) error) error
	OnStart() *hook.Hook[*StartEvent]
	OnStop() *hook.Hook[*StopEvent]
	Stop(e *StopEvent, fns ...func(*StopEvent) error) error
}

var _ Lifecycle = (*lifecycle)(nil)

func NewLifecycle(logger *slog.Logger) Lifecycle {
	return &lifecycle{
		logger:  logger,
		onStart: &hook.Hook[*StartEvent]{},
		onStop:  &hook.Hook[*StopEvent]{},
	}
}
