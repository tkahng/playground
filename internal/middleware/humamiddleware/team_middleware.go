package humamiddleware

import (
	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/playground/internal/core"
	"github.com/tkahng/playground/internal/middleware"
	"github.com/tkahng/playground/internal/models"
)

func UserIsMemberWithID(api huma.API, app core.App) func(ctx huma.Context, next func(huma.Context)) {
	return HumaChiMiddleware(middleware.UserIsMemberWithID(app))
}

func TeamInfoFromUserAndMemberID(api huma.API, app core.App) func(ctx huma.Context, next func(huma.Context)) {
	return HumaChiMiddleware(middleware.TeamInfoFromUserAndMemberID(app))
}
func TeamCanDelete(api huma.API, app core.App) func(ctx huma.Context, next func(huma.Context)) {
	return HumaChiMiddleware(middleware.TeamCanDelete(app))
}

func TeamInfoFromTask(api huma.API, app core.App) func(ctx huma.Context, next func(huma.Context)) {
	return HumaChiMiddleware(middleware.TeamInfoFromTask(app))
}

func TeamInfoFromTaskProject(api huma.API, app core.App) func(ctx huma.Context, next func(huma.Context)) {
	return HumaChiMiddleware(middleware.TeamInfoFromTaskProject(app))
}

func TeamInfoFromTeamSlug(api huma.API, app core.App) func(ctx huma.Context, next func(huma.Context)) {
	return HumaChiMiddleware(middleware.TeamInfoFromTeamSlug(app))
}

func TeamInfoFromParam(api huma.API, app core.App) func(ctx huma.Context, next func(huma.Context)) {
	return HumaChiMiddleware(middleware.TeamInfoFromParam(app))
}

func RequireTeamMemberRolesMiddleware(api huma.API, roles ...models.TeamMemberRole) func(ctx huma.Context, next func(huma.Context)) {
	return HumaChiMiddleware(middleware.RequireTeamMemberRolesMiddleware(roles...))
}
