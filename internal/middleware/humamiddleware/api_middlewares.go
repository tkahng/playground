package humamiddleware

import (
	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/playground/internal/core"
	"github.com/tkahng/playground/internal/middleware"
)

type HumaMiddlewareFunc func(ctx huma.Context, next func(huma.Context))

type ApiMiddlewares struct {
	// customer middlewares
	selectCustomerFromUser  HumaMiddlewareFunc
	selectCustomerFromTeam  HumaMiddlewareFunc
	teamInfoFromTeamIDParam HumaMiddlewareFunc
}

func (a *ApiMiddlewares) GetTeamInfoFromTeamIDParam() HumaMiddlewareFunc {
	return a.teamInfoFromTeamIDParam
}

func (a *ApiMiddlewares) GetSelectCustomerFromTeam() HumaMiddlewareFunc {
	return a.selectCustomerFromTeam
}

func (a *ApiMiddlewares) GetSelectCustomerFromUser() HumaMiddlewareFunc {
	return a.selectCustomerFromUser
}

func NewApiMiddlewares(app core.App) *ApiMiddlewares {
	return &ApiMiddlewares{
		// customer middlewares
		selectCustomerFromUser:  HumaChiMiddleware(middleware.SelectCustomerFromUser(app)),
		selectCustomerFromTeam:  HumaChiMiddleware(middleware.SelectCustomerFromTeam(app)),
		teamInfoFromTeamIDParam: HumaChiMiddleware(middleware.TeamInfoFromTeamIDParam(app)),
	}
}
