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
	selectCustomerFromUser HumaMiddlewareFunc
	selectCustomerFromTeam HumaMiddlewareFunc
	// team info middlewares
	teamInfoFromParam           HumaMiddlewareFunc
	teamInfoFromTeamSlug        HumaMiddlewareFunc
	teamInfoFromUserAndMemberID HumaMiddlewareFunc
	teamInfoFromTask            HumaMiddlewareFunc
	teamInfoFromTaskProject     HumaMiddlewareFunc
	// check middlewares
	memberIdBelongsToUser   HumaMiddlewareFunc
	teamCanDelete           HumaMiddlewareFunc
	emailVerified           HumaMiddlewareFunc
	teamRequiredOwnerMember HumaMiddlewareFunc
	TeamRequiredAnyMember   HumaMiddlewareFunc
	// auth middlewares
	auth        HumaMiddlewareFunc
	requireAuth HumaMiddlewareFunc
	// common middlewares
	recoverer HumaMiddlewareFunc
}

func (a *ApiMiddlewares) GetMemberIdBelongsToUser() HumaMiddlewareFunc {
	return a.memberIdBelongsToUser
}

func (a *ApiMiddlewares) GetRequireAuth() HumaMiddlewareFunc {
	return a.requireAuth
}

func (a *ApiMiddlewares) GetAuth() HumaMiddlewareFunc {
	return a.auth
}

func (a *ApiMiddlewares) GetRecoverer() HumaMiddlewareFunc {
	return a.recoverer
}

func (a *ApiMiddlewares) GetTeamInfoFromUserAndMemberID() HumaMiddlewareFunc {
	return a.teamInfoFromUserAndMemberID
}

func (a *ApiMiddlewares) GetEmailVerified() HumaMiddlewareFunc {
	return a.emailVerified
}

func (a *ApiMiddlewares) GetTeamCanDelete() HumaMiddlewareFunc {
	return a.teamCanDelete
}

func (a *ApiMiddlewares) GetTeamRequiredAnyMember() HumaMiddlewareFunc {
	return a.TeamRequiredAnyMember
}

func (a *ApiMiddlewares) GetTeamRequiredOwnerMember() HumaMiddlewareFunc {
	return a.teamRequiredOwnerMember
}

func (a *ApiMiddlewares) GetTeamInfoFromTeamSlug() HumaMiddlewareFunc {
	return a.teamInfoFromTeamSlug
}

func (a *ApiMiddlewares) GetTeamInfoFromParam() HumaMiddlewareFunc {
	return a.teamInfoFromParam
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
		selectCustomerFromUser: HumaChiMiddleware(middleware.SelectCustomerFromUser(app)),
		selectCustomerFromTeam: HumaChiMiddleware(middleware.SelectCustomerFromTeam(app)),
		// team info middlewares
		teamInfoFromParam:           HumaChiMiddleware(middleware.TeamInfoFromParam(app)),
		teamInfoFromTeamSlug:        HumaChiMiddleware(middleware.TeamInfoFromTeamSlug(app)),
		teamInfoFromUserAndMemberID: HumaChiMiddleware(middleware.TeamInfoFromUserAndMemberID(app)),
		teamInfoFromTask:            HumaChiMiddleware(middleware.TeamInfoFromTask(app)),
		teamInfoFromTaskProject:     HumaChiMiddleware(middleware.TeamInfoFromTaskProject(app)),
		// check middlewares
		memberIdBelongsToUser:   HumaChiMiddleware(middleware.MemberIdBelongsToUser(app)),
		teamCanDelete:           HumaChiMiddleware(middleware.TeamCanDelete(app)),
		emailVerified:           HumaChiMiddleware(middleware.HttpEmailVerifiedMiddleware()),
		teamRequiredOwnerMember: HumaChiMiddleware(middleware.RequireTeamMemberRolesMiddleware(models.TeamMemberRoleOwner)),
		TeamRequiredAnyMember:   HumaChiMiddleware(middleware.RequireTeamMemberRolesMiddleware()),
		// auth middlewares
		auth:        HumaChiMiddleware(middleware.HttpAuthMiddleware(app)),
		requireAuth: HumaChiMiddleware(middleware.HttpRequireAuthMiddleware()),
		// common middlewares
		recoverer: HumaChiMiddleware(middleware.RecovererMiddleware()),
	}
}
