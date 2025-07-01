package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"runtime/debug"

	"github.com/tkahng/authgo/internal/models"
)

// Dispatcher routes jobs to their appropriate handlers based on job kind.
type Dispatcher interface {
	// Dispatch executes the job with the appropriate handler.
	Dispatch(ctx context.Context, row *models.JobRow) error

	// SetHandler registers a handler for a specific job kind.
	// Panics if a handler is already registered for the kind.
	SetHandler(kind string, handler func(context.Context, *models.JobRow) error)
}

type dispatcher struct {
	handlers map[string]func(context.Context, *models.JobRow) error
}

func (d *dispatcher) SetHandler(kind string, handler func(context.Context, *models.JobRow) error) {
	if _, exists := d.handlers[kind]; exists {
		return
		// panic("duplicate worker kind: " + kind)
	}
	d.handlers[kind] = handler
}

func NewDispatcher() Dispatcher {
	return &dispatcher{
		handlers: make(map[string]func(context.Context, *models.JobRow) error),
	}
}

func RegisterWorker[T JobArgs](d Dispatcher, worker Worker[T]) {
	var zero T
	kind := zero.Kind()
	d.SetHandler(
		kind,
		func(ctx context.Context, row *models.JobRow) error {
			var args T
			if err := json.Unmarshal(row.Payload, &args); err != nil {
				return fmt.Errorf("unmarshal payload: %w", err)
			}
			return execute(ctx, worker, &Job[T]{JobRow: row, Args: args})
		},
	)
}

func (d *dispatcher) Dispatch(ctx context.Context, row *models.JobRow) error {
	handler, ok := d.handlers[row.Kind]
	if !ok {
		slog.Error(
			"no handler registered for kind",
			slog.String("kind", row.Kind),
			slog.String("job_id", row.ID.String()),
		)
		return fmt.Errorf("no handler registered for kind: %s", row.Kind)
	}
	return handler(ctx, row)
}

func execute[T JobArgs](ctx context.Context, worker Worker[T], job *Job[T]) (err error) {
	defer func() {
		if r := recover(); r != nil {
			slog.ErrorContext(ctx, "panic in job", "kind", job.Args.Kind(), "panic_recover", r, slog.Any("stack", string(debug.Stack())))
			err = fmt.Errorf("panic: %v", r)
		}
	}()
	slog.InfoContext(ctx, "start job", "kind", job.Args.Kind(), "id", job.ID, slog.Any("args", job.Args))
	if worker == nil {
		return fmt.Errorf("no worker registered for kind: %s", job.Args.Kind())
	}
	err = worker.Work(ctx, job)
	if err != nil {
		slog.ErrorContext(
			ctx,
			"fail job",
			slog.String("kind", job.Args.Kind()),
			slog.String("id", job.ID.String()),
			slog.Any("error", err),
		)
	} else {
		slog.InfoContext(ctx, "done job", "kind", job.Args.Kind(), "id", job.ID)
	}
	return err
}
