package apis

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/playground/internal/contextstore"
	"github.com/tkahng/playground/internal/core"
	"github.com/tkahng/playground/internal/middleware"
	"github.com/tkahng/playground/internal/middleware/humamiddleware"
	"github.com/tkahng/playground/internal/shared"
)

type PlayerRpsStatsResponse struct {
	TotalGames   int64 `json:"total_games"`
	Wins         int64 `json:"wins"`
	Losses       int64 `json:"losses"`
	Ties         int64 `json:"ties"`
	TotalBetWon  int64 `json:"total_bet_won"`
	TotalBetLost int64 `json:"total_bet_lost"`
}

func bindGetCurrentPlayerRpsStatsApi(api huma.API, app core.App) {
	huma.Register(
		api,
		huma.Operation{
			OperationID: "get-current-player-rps-stats",
			Method:      http.MethodGet,
			Path:        "/players/current-player/rps-stats",
			Summary:     "get current player rps stats",
			Description: "Returns win/loss/tie counts and net bet totals for the authenticated player.",
			Tags:        []string{"Games", "Player"},
			Errors:      []int{http.StatusUnauthorized},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
			Middlewares: humamiddleware.HumaChiMiddlewares(
				middleware.RequireCurrentPlayerMiddelware(),
			),
		},
		func(ctx context.Context, input *struct{}) (*ApiSingleOutput[*PlayerRpsStatsResponse], error) {
			currentPlayer := contextstore.GetContextCurrentPlayer(ctx)
			if currentPlayer == nil {
				return nil, huma.Error401Unauthorized("no player found")
			}
			agg, err := app.Adapter().Gaming().GetPlayerGameAggregates(ctx, currentPlayer.ID)
			if err != nil {
				return nil, err
			}
			return &ApiSingleOutput[*PlayerRpsStatsResponse]{
				Body: ApiSingleResponse[*PlayerRpsStatsResponse]{
					Data: &PlayerRpsStatsResponse{
						TotalGames:   agg.TotalGames,
						Wins:         agg.Wins,
						Losses:       agg.Losses,
						Ties:         agg.Ties,
						TotalBetWon:  agg.TotalBetWon,
						TotalBetLost: agg.TotalBetLost,
					},
				},
			}, nil
		},
	)
}
