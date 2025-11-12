package middleware

import (
	"log/slog"
	"net"
	"net/http"
	"strings"

	"github.com/tkahng/playground/internal/contextstore"
	"github.com/tkahng/playground/internal/tools/logger"
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

func IpAddressMiddleware() HttpMiddelwareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			var ip string
			for idx, header := range headers {
				index := idx
				ipHeader := r.Header.Get(header.Header)
				if header.Split {
					ipBefore, _, _ := strings.Cut(ipHeader, ",")
					ipHeader = ipBefore
				}
				if ipHeader != "" && net.ParseIP(ipHeader) != nil {
					slog.DebugContext(ctx, "found ip", slog.Int("index", index), slog.String("ip", ipHeader))
					ip = ipHeader
					break
				}
			}
			if ip == "" {
				next.ServeHTTP(w, r)
				return
			}
			logger.SetAttrs(ctx, slog.String("ip", ip))

			next.ServeHTTP(w, r.WithContext(contextstore.SetContextIPAddress(ctx, ip)))
		})
	}
}
