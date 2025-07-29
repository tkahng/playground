package humamiddleware

import (
	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/playground/internal/core"
	"github.com/tkahng/playground/internal/middleware"
)

func CheckTaskOwnerMiddleware(api huma.API, app core.App) func(ctx huma.Context, next func(huma.Context)) {
	return HumaChiMiddleware(middleware.CheckTaskOwnerMiddleware(app))
}
