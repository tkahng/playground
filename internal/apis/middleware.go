package apis

import (
	"fmt"
	"log"
	"net/http"
	"slices"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/authgo/internal/core"
	"github.com/tkahng/authgo/internal/shared"
)

// TokenFromCookie tries to retreive the token string from a cookie named
// "jwt".
func HumaTokenFromCookie(ctx huma.Context) string {
	cookie, err := huma.ReadCookie(ctx, "access_token")
	//  ctx.Header()
	if err != nil {
		return ""
	}
	return cookie.Value
}

// TokenFromHeader tries to retreive the token string from the
// "Authorization" reqeust header: "Authorization: BEARER T".
func HumaTokenFromHeader(ctx huma.Context) string {
	// Get token from authorization header.
	bearer := ctx.Header("Authorization")
	if len(bearer) > 7 && strings.ToUpper(bearer[0:6]) == "BEARER" {
		return bearer[7:]
	}
	return ""
}

var HumaTokenFuncs = []func(huma.Context) string{
	HumaTokenFromHeader,
	HumaTokenFromCookie,
}

func CheckRolesMiddleware(api huma.API, roles ...string) func(ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {

		if claims := core.GetContextUserClaims(ctx.Context()); claims != nil {
			if len(roles) == 0 {
				next(ctx)
				return
			}
			for _, role := range claims.Roles {
				if slices.Contains(roles, role) {
					next(ctx)
					return
				}
			}
		}
		huma.WriteErr(api, ctx, http.StatusForbidden, fmt.Sprintf("You do not have the required roles: %v", roles))
	}
}

func CheckPermissionsMiddleware(api huma.API, permissions ...string) func(ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {

		if claims := core.GetContextUserClaims(ctx.Context()); claims != nil {
			if len(permissions) == 0 {
				next(ctx)
				return
			}
			for _, role := range claims.Permissions {
				if slices.Contains(permissions, role) {
					next(ctx)
					return
				}
			}
		}
		huma.WriteErr(api, ctx, http.StatusForbidden, fmt.Sprintf("You do not have the required permissions: %v", permissions))
	}
}

func RequireAuthMiddleware(api huma.API) func(ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
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

		log.Println("RequireAuthMiddleware")
		c := core.GetContextUserClaims(ctx.Context())
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

// Auth creates a middleware that will authorize requests based on the required scopes for the operation.
func AuthMiddleware(api huma.API, app core.App) func(ctx huma.Context, next func(huma.Context)) {

	return func(ctx huma.Context, next func(huma.Context)) {
		log.Println("auth middleware")
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
