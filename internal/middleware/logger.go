package middleware

import (
	"log/slog"
	"net/http"

	"github.com/tkahng/playground/internal/tools/logger"

	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

func SetContextAttrs(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := logger.InitContextAttrs(r.Context())
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func SetRequestIDAttrs(next http.Handler) http.Handler {
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
