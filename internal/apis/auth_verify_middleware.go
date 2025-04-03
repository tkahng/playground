package apis

import (
	"log"

	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/authgo/internal/core"
)

// Auth creates a middleware that will authorize requests based on the required scopes for the operation.
func RequireVerifyToken(api huma.API, app core.App) func(ctx huma.Context, next func(huma.Context)) {

	return func(ctx huma.Context, next func(huma.Context)) {
		log.Println("verify middleware")
		// check if already has user claims

		if claims := core.GetContextUserClaims(ctx.Context()); claims != nil {
			log.Println("already has user claims")
			next(ctx)
			return
		}
		var token string
		for _, f := range HumaTokenFuncs {
			token = f(ctx)
			if len(token) > 0 {
				break
			}
		}
		if len(token) == 0 {
			next(ctx)
			return
		}
		user, err := app.HandleAuthToken(ctx.Context(), token)
		if err != nil {
			log.Println(err)
			next(ctx)
			return
		}
		ctx = huma.WithValue(ctx, core.ContextKeyUserClaims, user)

		next(ctx)
	}
}
