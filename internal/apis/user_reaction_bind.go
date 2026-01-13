package apis

func bindUserReactionApi(appApi *Api) {
	api := appApi.Api()
	appApi.bindCreateUserReaction(api)
	appApi.bindUserReactionSse(api)
	appApi.bindGetLatestUserReactionStats(api)
}
