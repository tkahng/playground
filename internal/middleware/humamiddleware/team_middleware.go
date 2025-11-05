package humamiddleware

import (
	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/playground/internal/core"
	"github.com/tkahng/playground/internal/middleware"
	"github.com/tkahng/playground/internal/models"
)

// MemberIdBelongsToUser middleware ensures that the user is the member with id {team-member-id}
func MemberIdBelongsToUser(api huma.API, app core.App) func(ctx huma.Context, next func(huma.Context)) {
	return HumaChiMiddleware(middleware.MemberIdBelongsToUser(app))
}

// TeamInfoFromUserAndMemberID finds the team info from the userId and teamId of the member of {team-member-id}
func TeamInfoFromUserAndMemberID(api huma.API, app core.App) func(ctx huma.Context, next func(huma.Context)) {
	return HumaChiMiddleware(middleware.TeamInfoFromUserAndMemberID(app))
}

// TeamCanDelete middleware checks whether the team can be deleted, i.e. it has no valid subscriptions
func TeamCanDelete(api huma.API, app core.App) func(ctx huma.Context, next func(huma.Context)) {
	return HumaChiMiddleware(middleware.TeamCanDelete(app))
}

// TeamInfoFromTask captures the {task-id} path param to query its teamId, and along with the user info, queries the teamInfo membership.
// if the user has membership in the team of the task, that teamInfo is added to the context, and the request is forwarded to the next middleware,
// otherwise it returns an error
func TeamInfoFromTask(api huma.API, app core.App) func(ctx huma.Context, next func(huma.Context)) {
	return HumaChiMiddleware(middleware.TeamInfoFromTask(app))
}

// TeamInfoFromTaskProject captures the {"task-project-id"} path param to query its teamId, and along with the user info, queries the teamInfo membership.
// if the user has membership in the team of the task project, that teamInfo is added to the context and the request is forwarded to the next middleware, otherwise it returns an error
func TeamInfoFromTaskProject(api huma.API, app core.App) func(ctx huma.Context, next func(huma.Context)) {
	return HumaChiMiddleware(middleware.TeamInfoFromTaskProject(app))
}

// TeamInfoFromTeamSlug captures the {team-slug} path param, and along with the user info, queries the teamInfo.
// If the user has membership in the team of {team-slug}, that teamInfo is added to the context, otherwise it returns an error
func TeamInfoFromTeamSlug(api huma.API, app core.App) func(ctx huma.Context, next func(huma.Context)) {
	return HumaChiMiddleware(middleware.TeamInfoFromTeamSlug(app))
}

// TeamInfoFromParam captures the {team-id} path param, and along with the user info, queries the teamInfo.
// If the user has membership in the team of the task project, that teamInfo is added to the context, otherwise it returns an error
func TeamInfoFromParam(api huma.API, app core.App) func(ctx huma.Context, next func(huma.Context)) {
	return HumaChiMiddleware(middleware.TeamInfoFromParam(app))
}

// RequireTeamMemberRolesMiddleware checks if the member has the required team member roles
func RequireTeamMemberRolesMiddleware(api huma.API, roles ...models.TeamMemberRole) func(ctx huma.Context, next func(huma.Context)) {
	return HumaChiMiddleware(middleware.RequireTeamMemberRolesMiddleware(roles...))
}
