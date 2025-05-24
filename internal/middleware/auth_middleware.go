package middleware

import (
	"fmt"
	"log"
	"net/http"
	"slices"

	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/authgo/internal/contextstore"
	"github.com/tkahng/authgo/internal/core"
	"github.com/tkahng/authgo/internal/shared"
)

// Auth creates a middleware that will authorize requests based on the required scopes for the operation.
func AuthMiddleware(api huma.API, app core.App) func(ctx huma.Context, next func(huma.Context)) {

	return func(ctx huma.Context, next func(huma.Context)) {
		fmt.Println("AuthMiddleware")
		ctxx := ctx.Context()
		// check if already has user claims
		if claims := contextstore.GetContextUserInfo(ctxx); claims != nil {
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
		user, err := app.Auth().HandleAccessToken(ctxx, token)
		if err != nil {
			log.Println(err)
			next(ctx)
			return
		}
		ctxx = contextstore.SetContextUserInfo(ctxx, user)
		ctx = huma.WithContext(ctx, ctxx)

		next(ctx)
	}
}

func RequireAuthMiddleware(api huma.API) func(ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		fmt.Println("RequireAuthMiddleware")
		if ctx.Operation().Security == nil {
			next(ctx)
			return
		}
		var anyOfNeededScopes []string
		isAuthorizationRequired := false

		for _, opScheme := range ctx.Operation().Security {
			var ok bool
			if anyOfNeededScopes, ok = opScheme[shared.BearerAuthSecurityKey]; ok {
				isAuthorizationRequired = true
				break
			}
		}

		if !isAuthorizationRequired {
			next(ctx)
			return
		}

		c := contextstore.GetContextUserInfo(ctx.Context())
		if c != nil {
			if len(anyOfNeededScopes) == 0 {
				next(ctx)
				return
			}
			for _, scope := range c.Permissions {
				if slices.Contains(anyOfNeededScopes, string(scope)) {
					next(ctx)
					return
				}
			}
		}
		huma.WriteErr(api, ctx, http.StatusForbidden, "Forbidden")
	}
}

func CheckPermissionsMiddleware(api huma.API, permissions ...string) func(ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {

		if claims := contextstore.GetContextUserInfo(ctx.Context()); claims != nil {
			if len(permissions) == 0 {
				next(ctx)
				return
			}
			for _, permission := range claims.Permissions {
				if slices.Contains(permissions, permission) {
					next(ctx)
					return
				}
			}
		}
		huma.WriteErr(api, ctx, http.StatusForbidden, fmt.Sprintf("You do not have the required permissions: %v", permissions))
	}
}
