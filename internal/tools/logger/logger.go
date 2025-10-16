package logger

import (
	"log/slog"
	"os"

	"github.com/go-chi/httplog/v3"
	"github.com/tkahng/playground/internal/conf"
)

func GetDefaultLogger() *slog.Logger {
	opts := conf.GetConfig[conf.AppConfig]()
	isNotProduction := opts.AppEnv != "production"
	logger := slog.New(ContextHandler{
		Handler: slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			AddSource:   isNotProduction,
			Level:       slog.LevelInfo,
			ReplaceAttr: httplog.SchemaOTEL.Concise(isNotProduction).ReplaceAttr,
		}),
	})
	slog.SetDefault(logger)
	return logger
}

func GetDefaultFormat(opts *conf.AppConfig) *httplog.Schema {
	isNotProduction := opts.AppEnv != "production"
	return httplog.SchemaOTEL.Concise(isNotProduction)
}
