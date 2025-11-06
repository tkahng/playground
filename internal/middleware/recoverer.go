package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	apphttp "github.com/tkahng/playground/internal/tools/http"
)

func RecovererMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rvr := recover(); rvr != nil {
					if rvr == http.ErrAbortHandler {
						// we don't recover http.ErrAbortHandler so the response
						// to the client is aborted, this should not be logged
						panic(rvr)
					}

					slog.ErrorContext(
						r.Context(),
						"recovered from panic",
						slog.Any("panic", rvr),
						slog.Any("stack", string(debug.Stack())),
					)

					if r.Header.Get("Connection") != "Upgrade" {
						_ = apphttp.WriteErr(w, r, http.StatusInternalServerError, "internal server error")
					}
				}
			}()

			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(fn)
	}
}
