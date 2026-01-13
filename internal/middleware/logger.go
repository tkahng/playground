package middleware

import (
	"log/slog"
	"net/http"

	"github.com/tkahng/playground/internal/tools/logger"

	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

// InitContextAttrsMiddleware is a middleware that adds a pointer to a slice of slog.Attr
// to the context. Whenever the logger handler is called, it will add the
// attributes to the slog.Record. This middleware should be called early in the
// middleware chain.
func InitContextAttrsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := logger.InitContextAttrs(r.Context())
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// SetRequestIDAttrsMiddleware sets the request_id to the ContextAttrs.
// it gets the request_id from context set by the request_id middleware,
// therefore it should be called after the request_id middleware
func SetRequestIDAttrsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		requestID := chimiddleware.GetReqID(ctx)
		if requestID == "" {
			next.ServeHTTP(w, r)
			return
		}
		logger.SetAttrs(ctx, slog.String("request_id", requestID))
		next.ServeHTTP(w, r)
	})
}
