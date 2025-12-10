package apis

import (
	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/playground/internal/middleware"
	"github.com/tkahng/playground/internal/middleware/humamiddleware"
)

func bindGameApi(api *Api) {
	gameGroup := huma.NewGroup(api.Api())
	gameGroup.UseMiddleware(
		humamiddleware.HumaChiMiddlewares(
			middleware.SetCurrentPlayerMiddleware(api.App()),
		)...,
	)
	bindGetMyPlayerApi(gameGroup, api.App())
	bindPutMyPlayerApi(gameGroup, api.App())
	bindFindPlayersApi(gameGroup, api.App())
	bindFindRegisteredPlayerByEmailApi(gameGroup, api.App())
}
