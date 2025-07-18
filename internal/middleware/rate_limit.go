package middleware

import (
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/httprate"
)

func RateLimitByIp(api huma.API, amount int, duration time.Duration) func(ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		// Unwrap the request and response objects.
		r, w := humachi.Unwrap(ctx)

		// Do something with the request and response objects.
		httprate.LimitByIP(amount, duration)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next(ctx)
		})).ServeHTTP(w, r)
	}
}
