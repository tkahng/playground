package humamiddleware

import (
	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/playground/internal/core"
	"github.com/tkahng/playground/internal/middleware"
)

func SelectOrCreateOwnerCustomerFromTeam(api huma.API, app core.App) func(ctx huma.Context, next func(huma.Context)) {
	return HumaChiMiddleware(middleware.SelectOrCreateOwnerCustomerFromTeam(app))
}
func SelectCustomerFromTeam(api huma.API, app core.App) func(ctx huma.Context, next func(huma.Context)) {
	return HumaChiMiddleware(middleware.SelectCustomerFromTeam(app))
}

func SelectCustomerFromUser(api huma.API, app core.App) func(ctx huma.Context, next func(huma.Context)) {
	return HumaChiMiddleware(middleware.SelectCustomerFromUser(app))
}
