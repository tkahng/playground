package apis

import "github.com/danielgtaylor/huma/v2"

func bindUserReactionApi(api huma.API, appApi *Api) {
	appApi.bindCreateUserReaction(api)
	appApi.bindUserReactionSse(api)
	appApi.bindGetLatestUserReactionStats(api)
}
