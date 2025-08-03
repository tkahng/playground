package middleware

import (
	"log/slog"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/playground/internal/contextstore"
)

type HeadersInput struct {
	Header string
	Split  bool
}

var headers = []HeadersInput{
	{
		Header: "X-Client-IP",
	},
	{
		Header: "X-Forwarded-For",
		Split:  true,
	},
	{
		Header: "X-Forwarded",
		Split:  true,
	},
	{
		Header: "Forwarded-For",
		Split:  true,
	},
	{
		Header: "Forwarded",
		Split:  true,
	},
	{
		Header: "CF-Connecting-IP",
	},
	{
		Header: "Fastly-Client-Ip",
	},
	{
		Header: "True-Client-Ip",
	},
	{
		Header: "X-Real-IP",
	},
	{
		Header: "X-Cluster-Client-IP",
	},
}

func IpAddressMiddleware(api huma.API) HttpMiddelwareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			var ip string
			for idx, header := range headers {
				index := idx
				ipHeader := r.Header.Get(header.Header)
				if len(ipHeader) > 0 {
					slog.InfoContext(ctx, "found ip", slog.Int("index", index), slog.String("ip", ip))
					ip = ipHeader
					break
				}
			}
			if len(ip) == 0 {
				next.ServeHTTP(w, r)
				return
			}
			ctxx := contextstore.SetContextIPAddress(ctx, ip)
			r = r.WithContext(ctxx)
			next.ServeHTTP(w, r)
		})
	}
}
