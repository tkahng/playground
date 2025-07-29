package humamiddleware

import (
	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/playground/internal/middleware"
)

func IpAddressMiddleware(api huma.API) func(ctx huma.Context, next func(huma.Context)) {
	return HumaChiMiddleware(middleware.IpAddressMiddleware(api))
}
