package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/playground/internal/contextstore"
	"github.com/tkahng/playground/internal/core"
	appHttp "github.com/tkahng/playground/internal/tools/http"
	"github.com/tkahng/playground/internal/tools/http/queryparam"
	"github.com/tkahng/playground/internal/tools/logger"
)

func TokenFromHeader(r *http.Request, w http.ResponseWriter) string {
	bearer := r.Header.Get("Authorization")
	if len(bearer) > 7 && strings.ToUpper(bearer[0:6]) == "BEARER" {
		return bearer[7:]
	}
	return ""
}
func TokenFromQuery(r *http.Request, w http.ResponseWriter) string {
	return queryparam.Get(r.URL.RawQuery, "access_token")
}

var HttpTokenFuncs = []func(r *http.Request, w http.ResponseWriter) string{
	TokenFromHeader,
	TokenFromQuery,
}

type HttpMiddelwareFunc func(next http.Handler) http.Handler

func Unwrap(ctx huma.Context) (*http.Request, http.ResponseWriter) {
	for {
		if c, ok := ctx.(interface{ Unwrap() huma.Context }); ok {
			ctx = c.Unwrap()
			continue
		}
		break
	}
	if c, ok := ctx.(interface {
		Unwrap() (*http.Request, http.ResponseWriter)
	}); ok {
		return c.Unwrap()
	}
	panic("this context does not implement Unwrap")
}

func HttpEmailVerifiedMiddleware(app core.App) HttpMiddelwareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rawCtx := r.Context()
			userInfo := contextstore.GetContextUserInfo(rawCtx)
			if userInfo == nil {
				_ = appHttp.WriteErr(w, r, http.StatusUnauthorized, "unauthorized. user info not found", nil)
				return
			}
			if userInfo.User.EmailVerifiedAt == nil {
				_ = appHttp.WriteErr(w, r, http.StatusUnauthorized, "email not verified", nil)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func HttpAuthMiddleware(app core.App) HttpMiddelwareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			// check if already has user claims
			if claims := contextstore.GetContextUserInfo(ctx); claims != nil {
				next.ServeHTTP(w, r)
				return
			}
			var token string
			for _, f := range HttpTokenFuncs {
				token = f(r, w)
				if len(token) > 0 {
					break
				}
			}
			if len(token) == 0 {
				next.ServeHTTP(w, r)
				return
			}
			user, err := app.Auth().VerifyAccessToken(ctx, token)
			if err != nil {
				slog.ErrorContext(ctx, "failed to handle access token", slog.Any("error", err))
				next.ServeHTTP(w, r)
				return
			}
			ctx = contextstore.SetContextUserInfo(ctx, user)
			ctx = logger.AppendCtx(
				ctx,
				slog.String("user_id", user.User.ID.String()),
				slog.String("email", user.User.Email),
			)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}

func HttpRequireAuthMiddleware(app core.App) HttpMiddelwareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			// check if already has user claims
			if claims := contextstore.GetContextUserInfo(ctx); claims != nil {
				slog.InfoContext(ctx, "user already authenticated")
				next.ServeHTTP(w, r)
				return
			}
			slog.InfoContext(ctx, "user not authenticated")
			_ = appHttp.WriteErr(w, r, http.StatusUnauthorized, "you are not authenticated.", nil)
		})
	}
}

func HttpCheckPermissionsMiddleware(app core.App, requiredPermissions ...string) HttpMiddelwareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if claims := contextstore.GetContextUserInfo(r.Context()); claims != nil {
				if len(requiredPermissions) == 0 {
					next.ServeHTTP(w, r)
					return
				}
				for _, p := range claims.Permissions {
					if slices.Contains(requiredPermissions, p) {
						next.ServeHTTP(w, r)
						return
					}
				}
			}
			_ = appHttp.WriteErr(w, r, http.StatusForbidden, fmt.Sprintf("You do not have the required permissions: %v", requiredPermissions))
		})
	}
}
