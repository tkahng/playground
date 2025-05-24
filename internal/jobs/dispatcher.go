package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"runtime/debug"
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
		panic("duplicate worker kind: " + kind)
	}
	d.handlers[kind] = func(ctx context.Context, row *JobRow) error {
		var args T
		if err := json.Unmarshal(row.Payload, &args); err != nil {
			return fmt.Errorf("unmarshal payload: %w", err)
		}
		return Execute(ctx, worker, &Job[T]{JobRow: row, Args: args})
	}
}

func (d *Dispatcher) Dispatch(ctx context.Context, row *JobRow) error {
	handler, ok := d.handlers[row.Kind]
	if !ok {
		return fmt.Errorf("no handler registered for kind: %s", row.Kind)
	}
	return handler(ctx, row)
}

func Execute[T JobArgs](ctx context.Context, worker Worker[T], job *Job[T]) (err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("panic in job kind=%s: %v\n%s", job.Args.Kind(), r, debug.Stack())
			err = fmt.Errorf("panic: %v", r)
		}
	}()
	log.Printf("start job kind=%s id=%s", job.Args.Kind(), job.ID)
	err = worker.Work(ctx, job)
	if err != nil {
		log.Printf("fail job kind=%s id=%s err=%v", job.Args.Kind(), job.ID, err)
	} else {
		log.Printf("done job kind=%s id=%s", job.Args.Kind(), job.ID)
	}
	return err
}
