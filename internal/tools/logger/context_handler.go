package logger

import (
	"context"
	"log/slog"
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
	if attrs, ok := ctx.Value(slogFields).([]slog.Attr); ok {
		for _, v := range attrs {
			r.AddAttrs(v)
		}
	}
	return h.Handler.Handle(ctx, r)
}

// WithAttributes adds slog attributes to context
func WithAttributes(parent context.Context, attrs ...slog.Attr) context.Context {
	if parent == nil {
		parent = context.Background()
	}

	if existing, ok := parent.Value(slogFields).([]slog.Attr); ok {
		return context.WithValue(parent, slogFields, append(existing, attrs...))
	}
	return context.WithValue(parent, slogFields, attrs)
}

// Initialize default logger at package init
func init() {
	slog.SetDefault(GetDefaultLogger(slog.LevelInfo))
}
