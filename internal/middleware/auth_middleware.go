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
	"github.com/tkahng/playground/internal/shared"
	appHttp "github.com/tkahng/playground/internal/tools/http"
	"github.com/tkahng/playground/internal/tools/logger"
)

type OperationSecurity = []map[string][]string

func TokenFromHeader(r *http.Request, w http.ResponseWriter) string {
	bearer := r.Header.Get("Authorization")
	if len(bearer) > 7 && strings.ToUpper(bearer[0:6]) == "BEARER" {
		return bearer[7:]
	}
	return ""
}
func TokenFromQuery(r *http.Request, w http.ResponseWriter) string {
	return appHttp.GetQuery(r, "access_token")
}

var httpTokenFuncs = []func(r *http.Request, w http.ResponseWriter) string{
	TokenFromHeader,
	TokenFromQuery,
}

type HTTPMiddlewareFunc func(next http.Handler) http.Handler

func Unwrap(ctx huma.Context) (*http.Request, http.ResponseWriter, error) {
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
		r, w := c.Unwrap()
		return r, w, nil
	}
	return nil, nil, fmt.Errorf("huma context does not implement Unwrap")
}

func EmailVerifiedMiddleware() HTTPMiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rawCtx := r.Context()
			userInfo := contextstore.GetContextUserInfo(rawCtx)
			if userInfo == nil {
				appHttp.WriteErr(w, r, http.StatusUnauthorized, "unauthorized. user info not found", nil)
				return
			}
			if userInfo.User.EmailVerifiedAt == nil {
				appHttp.WriteErr(w, r, http.StatusUnauthorized, "email not verified", nil)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func AuthMiddleware(app core.App) HTTPMiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			// check if already has user claims
			if claims := contextstore.GetContextUserInfo(ctx); claims != nil {
				next.ServeHTTP(w, r)
				return
			}
			var token string
			for _, f := range httpTokenFuncs {
				token = f(r, w)
				if len(token) > 0 {
					break
				}
			}
			if len(token) == 0 {
				slog.DebugContext(ctx, "access token not found")
				next.ServeHTTP(w, r)
				return
			}
			userInfo, err := app.Auth().VerifyAccessToken(ctx, token)
			if err != nil {
				slog.ErrorContext(ctx, "failed to handle access token", slog.Any("error", err))
				next.ServeHTTP(w, r)
				return
			}
			if userInfo == nil {
				slog.DebugContext(ctx, "user info not found")
				next.ServeHTTP(w, r)
				return
			}
			logger.SetAttrs(
				ctx,
				slog.String("user_id", userInfo.User.ID.String()),
				slog.Bool("email_verified", userInfo.User.EmailVerifiedAt != nil),
			)
			newCtx := contextstore.SetContextUserInfo(ctx, userInfo)
			r = r.WithContext(newCtx)
			next.ServeHTTP(w, r)
		})
	}
}

func RequireAuthMiddleware() HTTPMiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			opSec := contextstore.GetContextOperationSecurity(ctx)
			if opSec == nil {
				slog.DebugContext(ctx, "operation security not found")
				next.ServeHTTP(w, r)
				return
			}
			isAuthorizationRequired := false

			for _, opScheme := range opSec {
				var ok bool
				if _, ok = opScheme[shared.BearerAuthSecurityKey]; ok {
					slog.DebugContext(ctx, "authorization required")
					isAuthorizationRequired = true
					break
				}
			}

			if !isAuthorizationRequired {
				slog.DebugContext(ctx, "authorization not required.")
				next.ServeHTTP(w, r)
				return
			}
			// check if already has user userInfo
			if userInfo := contextstore.GetContextUserInfo(ctx); userInfo != nil {
				slog.DebugContext(ctx, "user info found")
				next.ServeHTTP(w, r)
				return
			}
			appHttp.WriteErr(w, r, http.StatusUnauthorized, "you are not authenticated.")
		})
	}
}
func CheckPermissionsMiddleware(requiredPermissions ...string) HTTPMiddlewareFunc {
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
			appHttp.WriteErr(w, r, http.StatusForbidden, fmt.Sprintf("You do not have the required permissions: %v", requiredPermissions))
		})
	}
}
