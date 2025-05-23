package apis

import (
	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/authgo/internal/middleware"
)

func BindTeamsApi(api huma.API, appApi *Api) {
	teamInfoMiddleware := middleware.TeamInfoFromParamMiddleware(api, appApi.app)
	// requiredTeamMember := middleware.RequireTeamMemberRolesMiddleware(api)
	// requiredOwnerMember := middleware.RequireTeamMemberRolesMiddleware(api, models.TeamMemberRoleOwner)
	teamsGroup := huma.NewGroup(api)

	teamsGroup.UseMiddleware(
		teamInfoMiddleware,
	)

	// huma.Register(
	// 	teamsGroup,
	// 	huma.Operation{
	// 		OperationID: "create-team",
	// 		// Method:
	// 		// },
	// 	},

	// )
}
