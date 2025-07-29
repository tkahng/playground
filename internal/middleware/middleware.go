package middleware

import (
	"net/http"
	"strings"

	"github.com/danielgtaylor/huma/v2"

	"github.com/tkahng/playground/internal/tools/http/queryparam"
)

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

func HttpTokenFromHeader(r *http.Request, w http.ResponseWriter) string {
	return r.Header.Get("Authorization")
}
func HttpTokenFromQuery(r *http.Request, w http.ResponseWriter) string {
	return queryparam.Get(r.URL.RawQuery, "access_token")
}

var HttpTokenFuncs = []func(r *http.Request, w http.ResponseWriter) string{
	HttpTokenFromHeader,
	HttpTokenFromQuery,
}
