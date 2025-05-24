package jobs

import (
	"context"
	"encoding/json"
	"fmt"
)

type Dispatcher struct {
	handlers map[string]func(context.Context, *JobRow) error
}

func NewDispatcher() *Dispatcher {
	return &Dispatcher{
		handlers: make(map[string]func(context.Context, *JobRow) error),
	}
}

func RegisterWorker[T JobArgs](d *Dispatcher, worker Worker[T]) {
	var zero T
	kind := zero.Kind()
	if _, exists := d.handlers[kind]; exists {
		panic(fmt.Sprintf("duplicate worker registered for kind %q", kind))
	}
	d.handlers[kind] = func(ctx context.Context, row *JobRow) error {
		var args T
		if err := json.Unmarshal(row.Payload, &args); err != nil {
			return fmt.Errorf("failed to decode payload for kind %q: %w", kind, err)
		}
		job := &Job[T]{JobRow: row, Args: args}
		return Execute(ctx, worker, job)
	}
}

func (d *Dispatcher) Dispatch(ctx context.Context, row *JobRow) error {
	handler, ok := d.handlers[row.Kind]
	if !ok {
		return fmt.Errorf("no handler for job kind \"%s\"", row.Kind)
	}
	return handler(ctx, row)
}
