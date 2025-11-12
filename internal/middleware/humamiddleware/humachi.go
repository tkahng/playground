package humamiddleware

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/tkahng/playground/internal/contextstore"
)

func HumaChiMiddleware(mw func(http.Handler) http.Handler) func(ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		r, w := humachi.Unwrap(ctx)
		mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx = humachi.NewContext(ctx.Operation(), r, w)
			next(ctx)
		})).ServeHTTP(w, r)
	}
}

func HumaChiMiddlewares(mws ...func(http.Handler) http.Handler) []func(ctx huma.Context, next func(huma.Context)) {
	var middlewares []func(ctx huma.Context, next func(huma.Context))
	for _, mw := range mws {
		middlewares = append(middlewares, HumaChiMiddleware(mw))
	}
	return middlewares
}

func HumaOperationSecurityMiddleware() func(ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		opSec := ctx.Operation().Security
		if opSec == nil {
			next(ctx)
			return
		}
		r, w := humachi.Unwrap(ctx)
		rawCtx := r.Context()
		rawCtx = contextstore.SetContextOperationSecurity(rawCtx, opSec)
		r = r.WithContext(rawCtx)
		ctx = humachi.NewContext(ctx.Operation(), r, w)
		next(ctx)
	}
}
