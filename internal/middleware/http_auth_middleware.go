package middleware

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/render"
	"github.com/tkahng/playground/internal/contextstore"
	"github.com/tkahng/playground/internal/core"
	appHttp "github.com/tkahng/playground/internal/tools/http"
)

type MiddelwareFunc func(http.Handler) http.Handler

func HttpAuthMiddleware(app core.App) MiddelwareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			// check if already has user claims
			if claims := contextstore.GetContextUserInfo(ctx); claims != nil {
				next.ServeHTTP(w, r)
				return
			}
			var token string
			for idx, f := range HttpTokenFuncs {
				index := idx
				token = f(r, w)
				if len(token) > 0 {
					slog.InfoContext(ctx, "found token", slog.Int("index", index), slog.String("token", token))
					break
				}
			}
			if len(token) == 0 {
				next.ServeHTTP(w, r)
				return
			}
			user, err := app.Auth().HandleAccessToken(ctx, token)
			if err != nil {
				slog.ErrorContext(ctx, "failed to handle access token", slog.Any("error", err))
				next.ServeHTTP(w, r)
				return
			}
			ctx = contextstore.SetContextUserInfo(ctx, user)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}

func HttpRequireAuthMiddleware(app core.App) MiddelwareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			// check if already has user claims
			if claims := contextstore.GetContextUserInfo(ctx); claims != nil {
				next.ServeHTTP(w, r)
				return
			}
			_ = render.Render(w, r, appHttp.NewError(http.StatusUnauthorized, "unauthorized"))
		})
	}
}
