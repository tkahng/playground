package logger

import (
	"log/slog"
	"os"
	"sync"

	"github.com/go-chi/httplog/v3"
	"github.com/tkahng/playground/internal/conf"
	"go.opentelemetry.io/contrib/bridges/otelslog"
)

var (
	logger  *slog.Logger
	ctxOnce sync.Once
)

func GetLoggerSingleton(cfg *conf.AppConfig) *slog.Logger {
	ctxOnce.Do(func() {
		logger := getLogger(cfg)
		slog.SetDefault(logger)
	})
	return logger
}

func getLogger(cfg *conf.AppConfig) *slog.Logger {
	level := slog.LevelInfo
	if cfg.Debug {
		level = slog.LevelDebug
	}
	stdoutHandler := ContextHandler{
		Handler: slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level:       level,
			AddSource:   true,
			ReplaceAttr: httplog.SchemaOTEL.Concise(true).ReplaceAttr,
		}),
	}
	if !cfg.OtelEnabled {
		return slog.New(stdoutHandler)
	}
	otelHandler := otelslog.NewHandler("playground",
		otelslog.WithLoggerProvider(nil), // uses global provider
	)
	return slog.New(newMultiHandler(stdoutHandler, otelHandler))
}

func GetDefaultLogger() *slog.Logger {
	opts := conf.GetConfig[conf.AppConfig]()
	logger := getLogger(&opts)
	slog.SetDefault(logger)
	return logger
}
