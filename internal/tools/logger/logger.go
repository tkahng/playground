package logger

import (
	"log/slog"
	"os"

	"github.com/go-chi/httplog/v3"
	"github.com/tkahng/playground/internal/conf"
)

func GetDefaultLogger() *slog.Logger {
	opts := conf.GetConfig[conf.AppConfig]()
	logFormat := GetDefaultFormat(&opts)
	logger := slog.New(ContextHandler{
		Handler: slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			AddSource:   !(opts.AppEnv == "development"),
			Level:       slog.LevelInfo,
			ReplaceAttr: logFormat.ReplaceAttr,
		}),
	})
	slog.SetDefault(logger)
	return logger
}

func GetDefaultFormat(opts *conf.AppConfig) *httplog.Schema {
	isdev := opts.AppEnv == "development"
	return httplog.SchemaOTEL.Concise(isdev)
}
