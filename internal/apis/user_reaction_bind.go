package apis

import "github.com/danielgtaylor/huma/v2"

func BindUserReactionApi(api huma.API, appApi *Api) {
	appApi.BindCreateUserReaction(api)
}
