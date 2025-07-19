package middleware

import (
	"net/http"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
)

func HumaChiMiddleware(mw func(http.Handler) http.Handler) func(ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		r, w := humachi.Unwrap(ctx)
		mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r = r.WithContext(ctx.Context()) // ✨
			ctx = humachi.NewContext(ctx.Operation(), r, w)
			next(ctx)
		})).ServeHTTP(w, r)
	}
}
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

func HumaTokenFromQuery(ctx huma.Context) string {
	return ctx.Query("access_token")
}

var HumaTokenFuncs = []func(huma.Context) string{
	HumaTokenFromHeader,
	HumaTokenFromQuery,
}

// func
