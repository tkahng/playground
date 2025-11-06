package humamiddleware

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/tkahng/playground/internal/contextstore"
	"github.com/tkahng/playground/internal/core"
	"github.com/tkahng/playground/internal/middleware"
	"github.com/tkahng/playground/internal/shared"
)

func HumaEmailVerifiedMiddleware(api huma.API, a core.App) func(ctx huma.Context, next func(huma.Context)) {
	return HumaChiMiddleware(middleware.HttpEmailVerifiedMiddleware(a))

}

// Auth creates a middleware that will authorize requests based on the required scopes for the operation.
func HumaAuthMiddleware(api huma.API, app core.App) func(ctx huma.Context, next func(huma.Context)) {
	return HumaChiMiddleware(middleware.HttpAuthMiddleware(app))
}

func HumaRequireAuthMiddleware(api huma.API, app core.App) func(ctx huma.Context, next func(huma.Context)) {
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
		mw := middleware.HttpRequireAuthMiddleware(app)
		r, w := humachi.Unwrap(ctx)
		mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx = humachi.NewContext(ctx.Operation(), r, w)
			next(ctx)
		})).ServeHTTP(w, r)
	}

}

func HumaOperationSecurityMiddleware() func(ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		opSec := ctx.Operation().Security
		if opSec == nil {
			next(ctx)
			return
		}
		rawCtx := ctx.Context()
		rawCtx = contextstore.SetContextOperationSecurity(rawCtx, opSec)
		ctx = huma.WithContext(ctx, rawCtx)
		next(ctx)
	}
}

func HumaCheckPermissionsMiddleware(api huma.API, app core.App, permissions ...string) func(ctx huma.Context, next func(huma.Context)) {
	return HumaChiMiddleware(middleware.HttpCheckPermissionsMiddleware(app, permissions...))
}
