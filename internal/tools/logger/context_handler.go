package logger

import (
	"context"
	"log/slog"

	"github.com/go-chi/httplog/v3"
)

type ctxKey string

const (
	slogFields ctxKey = "slog_fields"
)

// ContextHandler adds contextual attributes to logs
type ContextHandler struct {
	slog.Handler
}

func (h ContextHandler) Handle(ctx context.Context, r slog.Record) error {
	if attrs := GetAttrs(ctx); attrs != nil {
		for _, v := range attrs {
			r.AddAttrs(v)
		}
	}
	return h.Handler.Handle(ctx, r)
}

// AppendCtx adds an slog attribute to the provided context so that it will be
// included in any Record created with such context
func AppendCtx(parent context.Context, attrs ...slog.Attr) context.Context {
	if parent == nil {
		parent = context.Background()
	}

	if v, ok := parent.Value(slogFields).([]slog.Attr); ok {
		v = append(v, attrs...)
		return context.WithValue(parent, slogFields, v)
	}

	v := []slog.Attr{}
	v = append(v, attrs...)
	return context.WithValue(parent, slogFields, v)
}

const (
	ErrorKey = "error"
)

type ctxKeyLogAttrs struct{}

func (c *ctxKeyLogAttrs) String() string {
	return "httplog attrs context"
}

// SetAttrs sets the attributes on the request log.
func SetAttrs(ctx context.Context, attrs ...slog.Attr) {
	if ptr, ok := ctx.Value(ctxKeyLogAttrs{}).(*[]slog.Attr); ok && ptr != nil {
		*ptr = append(*ptr, attrs...)
	}
	httplog.SetAttrs(ctx, attrs...)
}

func GetAttrs(ctx context.Context) []slog.Attr {
	if ptr, ok := ctx.Value(ctxKeyLogAttrs{}).(*[]slog.Attr); ok && ptr != nil {
		return *ptr
	}

	return nil
}

// SetError sets the error attribute on the request log.
func SetError(ctx context.Context, err error) error {
	if err != nil {
		SetAttrs(ctx, slog.Any(ErrorKey, err))
	}

	return err
}
