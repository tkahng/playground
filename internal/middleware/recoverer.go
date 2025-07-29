package middleware

import (
	"log/slog"
	"net/http"

	"github.com/tkahng/playground/internal/core"
	apphttp "github.com/tkahng/playground/internal/tools/http"
)

func HttpRecovererMiddleware(app core.App) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					slog.Error("Recovered from panic", slog.Any("error", err))
					_ = apphttp.WriteErr(
						w,
						r,
						http.StatusInternalServerError,
						"internal server error",
					)
					return
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
