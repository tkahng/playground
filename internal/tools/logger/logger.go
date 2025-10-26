package logger

import (
	"log/slog"
	"os"
	"sync"

	"github.com/go-chi/httplog/v3"
	"github.com/tkahng/playground/internal/conf"
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
	if cfg.AppEnv == "development" {
		level = slog.LevelDebug
	}
	// logger := slog.New(
	// 	NewPrettyHandler(os.Stdout, PrettyHandlerOptions{
	// 		SlogOpts: slog.HandlerOptions{
	// 			Level:       level,
	// 			ReplaceAttr: StackReplaceAttr,
	// 		},
	// 	}),
	// )
	logger := slog.New(ContextHandler{
		Handler: slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level:       level,
			ReplaceAttr: httplog.SchemaOTEL.Concise(true).ReplaceAttr,
		}),
	})
	return logger
}

func GetDefaultLogger() *slog.Logger {
	opts := conf.GetConfig[conf.AppConfig]()
	logger := getLogger(&opts)
	slog.SetDefault(logger)
	return logger
}
