package middleware

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/tkahng/playground/internal/core"
	"github.com/tkahng/playground/internal/shared"
)

func HumaEmailVerifiedMiddleware(api huma.API, a core.App) func(ctx huma.Context, next func(huma.Context)) {
	return HumaChiMiddleware(HttpEmailVerifiedMiddleware(a))
	// return func(ctx huma.Context, next func(huma.Context)) {
	// 	rawCtx := ctx.Context()
	// 	userInfo := contextstore.GetContextUserInfo(rawCtx)
	// 	if userInfo == nil {
	// 		huma.WriteErr(api, ctx, http.StatusUnauthorized, "unauthorized", nil)
	// 		return
	// 	}
	// 	if userInfo.User.EmailVerifiedAt == nil {
	// 		huma.WriteErr(api, ctx, http.StatusUnauthorized, "email not verified", nil)
	// 		return
	// 	}
	// 	next(ctx)
	// }
}

// Auth creates a middleware that will authorize requests based on the required scopes for the operation.
func HumaAuthMiddleware(api huma.API, app core.App) func(ctx huma.Context, next func(huma.Context)) {
	return HumaChiMiddleware(HttpAuthMiddleware(app))
	// return func(ctx huma.Context, next func(huma.Context)) {
	// 	ctxx := ctx.Context()
	// 	// check if already has user claims
	// 	if claims := contextstore.GetContextUserInfo(ctxx); claims != nil {
	// 		next(ctx)
	// 		return
	// 	}
	// 	var token string
	// 	for idx, f := range HumaTokenFuncs {
	// 		index := idx
	// 		token = f(ctx)
	// 		if len(token) > 0 {
	// 			slog.InfoContext(ctxx, "found token", slog.Int("index", index), slog.String("token", token))
	// 			break
	// 		}
	// 	}
	// 	if len(token) == 0 {
	// 		next(ctx)
	// 		return
	// 	}
	// 	user, err := app.Auth().HandleAccessToken(ctxx, token)
	// 	if err != nil {
	// 		slog.ErrorContext(ctxx, "failed to handle access token", slog.Any("error", err))
	// 		next(ctx)
	// 		return
	// 	}
	// 	ctxx = contextstore.SetContextUserInfo(ctxx, user)
	// 	ctx = huma.WithContext(ctx, ctxx)
	// 	next(ctx)
	// }
}

func HumaRequireAuthMiddleware(api huma.API, app core.App) func(ctx huma.Context, next func(huma.Context)) {
	// return HumaChiMiddleware(HttpRequireAuthMiddleware(app))
	return func(ctx huma.Context, next func(huma.Context)) {
		if ctx.Operation().Security == nil {
			next(ctx)
			return
		}
		isAuthorizationRequired := false

		for _, opScheme := range ctx.Operation().Security {
			var ok bool
			if _, ok = opScheme[shared.BearerAuthSecurityKey]; ok {
				isAuthorizationRequired = true
				break
			}
		}

		if !isAuthorizationRequired {
			next(ctx)
			return
		}
		mw := HttpRequireAuthMiddleware(app)
		r, w := humachi.Unwrap(ctx)
		mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// r = r.WithContext(ctx.Context())
			ctx = humachi.NewContext(ctx.Operation(), r, w)
			next(ctx)
		})).ServeHTTP(w, r)
	}
	// return func(ctx huma.Context, next func(huma.Context)) {
	// 	if ctx.Operation().Security == nil {
	// 		next(ctx)
	// 		return
	// 	}
	// 	var anyOfNeededScopes []string
	// 	isAuthorizationRequired := false

	// 	for _, opScheme := range ctx.Operation().Security {
	// 		var ok bool
	// 		if anyOfNeededScopes, ok = opScheme[shared.BearerAuthSecurityKey]; ok {
	// 			isAuthorizationRequired = true
	// 			break
	// 		}
	// 	}

	// 	if !isAuthorizationRequired {
	// 		next(ctx)
	// 		return
	// 	}

	// 	c := contextstore.GetContextUserInfo(ctx.Context())
	// 	if c != nil {
	// 		if len(anyOfNeededScopes) == 0 {
	// 			next(ctx)
	// 			return
	// 		}
	// 		for _, scope := range c.Permissions {
	// 			if slices.Contains(anyOfNeededScopes, string(scope)) {
	// 				next(ctx)
	// 				return
	// 			}
	// 		}
	// 	}
	// 	huma.WriteErr(api, ctx, http.StatusForbidden, "Forbidden")
	// }
}

func HumaCheckPermissionsMiddleware(api huma.API, app core.App, permissions ...string) func(ctx huma.Context, next func(huma.Context)) {
	return HumaChiMiddleware(HttpCheckPermissionsMiddleware(app, permissions...))
	// return func(ctx huma.Context, next func(huma.Context)) {

	// 	if claims := contextstore.GetContextUserInfo(ctx.Context()); claims != nil {
	// 		if len(permissions) == 0 {
	// 			next(ctx)
	// 			return
	// 		}
	// 		for _, permission := range claims.Permissions {
	// 			if slices.Contains(permissions, permission) {
	// 				next(ctx)
	// 				return
	// 			}
	// 		}
	// 	}
	// 	huma.WriteErr(api, ctx, http.StatusForbidden, fmt.Sprintf("You do not have the required permissions: %v", permissions))
	// }
}
