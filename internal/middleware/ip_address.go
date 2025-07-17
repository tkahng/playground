package middleware

import (
	"log/slog"

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

func IpAddressMiddleware(api huma.API) func(ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		var ip string
		for idx, header := range headers {
			index := idx
			ipHeader := ctx.Header(header.Header)
			if len(ipHeader) > 0 {
				slog.InfoContext(ctx.Context(), "found ip", slog.Int("index", index), slog.String("ip", ip))
				ip = ipHeader
				break
			}
		}
		if len(ip) == 0 {
			next(ctx)
			return
		}
		ctxx := contextstore.SetContextIPAddress(ctx.Context(), ip)
		ctx = huma.WithContext(ctx, ctxx)
		next(ctx)
	}
}
