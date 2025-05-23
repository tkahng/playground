package apis

import (
	"github.com/danielgtaylor/huma/v2"
)

func BindTeamsApi(api huma.API, appApi *Api) {
	teamInfoMiddleware := TeamInfoFromParamMiddleware(api, appApi.app)

	teamsGroup := huma.NewGroup(api)

	teamsGroup.UseMiddleware(
		teamInfoMiddleware,
	)
}
