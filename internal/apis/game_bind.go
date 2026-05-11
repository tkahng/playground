package apis

import (
	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/playground/internal/middleware"
	"github.com/tkahng/playground/internal/middleware/humamiddleware"
)

func bindGameApi(api *Api) {
	protectedGameGroup := huma.NewGroup(api.Api())
	protectedGameGroup.UseMiddleware(
		humamiddleware.HumaChiMiddlewares(
			middleware.SetCurrentPlayerMiddleware(api.App()),
		)...,
	)
	bindGetMyPlayerApi(protectedGameGroup, api.App())
	bindPutMyPlayerApi(protectedGameGroup, api.App())
	bindFindPlayersApi(protectedGameGroup, api.App())
	bindFindRegisteredPlayerByEmailApi(protectedGameGroup, api.App())
	bindSendGameRequestToRegisteredPlayerApi(protectedGameGroup, api.App())
	bindSubmitMoveToRpsGameApi(protectedGameGroup, api.App())
	bindSendGameRequestToUnRegisteredPlayerApi(protectedGameGroup, api.App())
	bindVerifyRpsGameInviteApi(api.Api(), api.App())
	bindSubmitMoveWithTokenApi(api.Api(), api.App())
	bindFindCurrentPlayersRpsGamesApi(protectedGameGroup, api.App())
	bindChallengeHouseApi(protectedGameGroup, api.App())
	bindCancelRpsGameApi(protectedGameGroup, api.App())
	bindRpsRematchApi(protectedGameGroup, api.App())
	bindFriendApi(protectedGameGroup, api.App())
	bindGetPlayerOnlineStatusApi(protectedGameGroup, api.App())
	bindPlayerSseApi(api)
}
