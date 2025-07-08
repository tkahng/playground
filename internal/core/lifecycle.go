package core

import (
	"log/slog"

	"github.com/tkahng/authgo/internal/tools/hook"
)

type StartEvent struct {
	hook.Event
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
	return l.onStart
}

// OnStop implements Lifecycle.
func (l *lifecycle) OnStop() *hook.Hook[*StopEvent] {
	return l.onStop
}

// Start implements Lifecycle.
func (l *lifecycle) Start(e *StartEvent) error {
	return l.onStart.Trigger(e)
}

// Stop implements Lifecycle.
func (l *lifecycle) Stop(e *StopEvent) error {
	return l.onStop.Trigger(e)
}

type Lifecycle interface {
	Start(*StartEvent) error
	OnStart() *hook.Hook[*StartEvent]
	OnStop() *hook.Hook[*StopEvent]
	Stop(*StopEvent) error
}

var _ Lifecycle = (*lifecycle)(nil)
