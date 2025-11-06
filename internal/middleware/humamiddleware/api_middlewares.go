package humamiddleware

import (
	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/playground/internal/core"
	"github.com/tkahng/playground/internal/middleware"
	"github.com/tkahng/playground/internal/models"
)

type HumaMiddlewareFunc func(ctx huma.Context, next func(huma.Context))

type ApiMiddlewares struct {
	// customer middlewares
	SelectCustomerFromUser HumaMiddlewareFunc
	SelectCustomerFromTeam HumaMiddlewareFunc
	// team info middlewares
	TeamInfoFromParam           HumaMiddlewareFunc
	TeamInfoFromTeamSlug        HumaMiddlewareFunc
	TeamInfoFromUserAndMemberID HumaMiddlewareFunc
	TeamInfoFromTask            HumaMiddlewareFunc
	TeamInfoFromTaskProject     HumaMiddlewareFunc
	// check middlewares
	MemberIdBelongsToUser   HumaMiddlewareFunc
	TeamCanDelete           HumaMiddlewareFunc
	EmailVerified           HumaMiddlewareFunc
	TeamRequiredOwnerMember HumaMiddlewareFunc
	TeamRequiredAnyMember   HumaMiddlewareFunc
	// auth middlewares
	Auth        HumaMiddlewareFunc
	RequireAuth HumaMiddlewareFunc
	// common middlewares
	Recoverer HumaMiddlewareFunc
}

func (a *ApiMiddlewares) GetEmailVerified() HumaMiddlewareFunc {
	return a.EmailVerified
}

func (a *ApiMiddlewares) GetTeamCanDelete() HumaMiddlewareFunc {
	return a.TeamCanDelete
}

func (a *ApiMiddlewares) GetTeamRequiredAnyMember() HumaMiddlewareFunc {
	return a.TeamRequiredAnyMember
}

func (a *ApiMiddlewares) GetTeamRequiredOwnerMember() HumaMiddlewareFunc {
	return a.TeamRequiredOwnerMember
}

func (a *ApiMiddlewares) GetTeamInfoFromTeamSlug() HumaMiddlewareFunc {
	return a.TeamInfoFromTeamSlug
}

func (a *ApiMiddlewares) GetTeamInfoFromParam() HumaMiddlewareFunc {
	return a.TeamInfoFromParam
}

func (a *ApiMiddlewares) GetSelectCustomerFromTeam() HumaMiddlewareFunc {
	return a.SelectCustomerFromTeam
}

func (a *ApiMiddlewares) GetSelectCustomerFromUser() HumaMiddlewareFunc {
	return a.SelectCustomerFromUser
}

func NewApiMiddlewares(app core.App) *ApiMiddlewares {
	return &ApiMiddlewares{
		// customer middlewares
		SelectCustomerFromUser: HumaChiMiddleware(middleware.SelectCustomerFromUser(app)),
		SelectCustomerFromTeam: HumaChiMiddleware(middleware.SelectCustomerFromTeam(app)),
		// team info middlewares
		TeamInfoFromParam:           HumaChiMiddleware(middleware.TeamInfoFromParam(app)),
		TeamInfoFromTeamSlug:        HumaChiMiddleware(middleware.TeamInfoFromTeamSlug(app)),
		TeamInfoFromUserAndMemberID: HumaChiMiddleware(middleware.TeamInfoFromUserAndMemberID(app)),
		TeamInfoFromTask:            HumaChiMiddleware(middleware.TeamInfoFromTask(app)),
		TeamInfoFromTaskProject:     HumaChiMiddleware(middleware.TeamInfoFromTaskProject(app)),
		// check middlewares
		MemberIdBelongsToUser:   HumaChiMiddleware(middleware.MemberIdBelongsToUser(app)),
		TeamCanDelete:           HumaChiMiddleware(middleware.TeamCanDelete(app)),
		EmailVerified:           HumaChiMiddleware(middleware.HttpEmailVerifiedMiddleware()),
		TeamRequiredOwnerMember: HumaChiMiddleware(middleware.RequireTeamMemberRolesMiddleware(models.TeamMemberRoleOwner)),
		TeamRequiredAnyMember:   HumaChiMiddleware(middleware.RequireTeamMemberRolesMiddleware()),
		// auth middlewares
		Auth:        HumaChiMiddleware(middleware.HttpAuthMiddleware(app)),
		RequireAuth: HumaChiMiddleware(middleware.HttpRequireAuthMiddleware()),
		// common middlewares
		Recoverer: HumaChiMiddleware(middleware.RecovererMiddleware()),
	}
}
