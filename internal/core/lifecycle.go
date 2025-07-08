package core

import (
	"log/slog"
	"sync"

	"github.com/tkahng/authgo/internal/tools/hook"
)

type WaitEvent struct {
	wg *sync.WaitGroup
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
func (l *lifecycle) Start(e *StartEvent) error {
	return l.onStart.Trigger(e)
}

// Stop implements Lifecycle.
func (l *lifecycle) Stop(e *StopEvent) error {
	return l.OnStop().Trigger(e)
}

func (l *lifecycle) Init() {
	l.onStart = &hook.Hook[*StartEvent]{}
	l.onStop = &hook.Hook[*StopEvent]{}
}

type Lifecycle interface {
	Init()
	Start(*StartEvent) error
	OnStart() *hook.Hook[*StartEvent]
	OnStop() *hook.Hook[*StopEvent]
	Stop(*StopEvent) error
}

var _ Lifecycle = (*lifecycle)(nil)

func NewLifecycle(logger *slog.Logger) Lifecycle {
	return &lifecycle{
		logger:  logger,
		onStart: &hook.Hook[*StartEvent]{},
		onStop:  &hook.Hook[*StopEvent]{},
	}
}
