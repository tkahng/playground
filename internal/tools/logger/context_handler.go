package logger

import (
	"context"
	"log/slog"
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

const (
	ErrorKey = "error"
)

type ctxKeyLogAttrs struct{}

func (c *ctxKeyLogAttrs) String() string {
	return "logger attrs context"
}

// SetAttrs sets the attributes on the request log.
func SetAttrs(ctx context.Context, attrs ...slog.Attr) {
	if ptr, ok := ctx.Value(ctxKeyLogAttrs{}).(*[]slog.Attr); ok && ptr != nil {
		*ptr = append(*ptr, attrs...)
	}
}

func InitContextAttrs(ctx context.Context) context.Context {
	return context.WithValue(ctx, ctxKeyLogAttrs{}, &[]slog.Attr{})
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
