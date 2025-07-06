package logger

import (
	"log/slog"
	"os"
	"path/filepath"
)

func SetDefaultLogger() {
	slog.SetDefault(GetDefaultLogger())
}

func GetDefaultLogger() *slog.Logger {
	return slog.New(ContextHandler{
		Handler: slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			AddSource: true,
			Level:     nil,
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				if a.Key == slog.SourceKey {
					source, _ := a.Value.Any().(*slog.Source)
					if source != nil {
						source.Function = ""
						source.File = filepath.Base(source.File)
					}
				}
				return a
			},
		}),
	})

}
