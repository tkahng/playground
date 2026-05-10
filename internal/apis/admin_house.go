package apis

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/playground/internal/core"
	"github.com/tkahng/playground/internal/services"
	"github.com/tkahng/playground/internal/shared"
)

type HouseStatsResponse struct {
	TotalGames     int64 `json:"total_games"`
	BettedGames    int64 `json:"betted_games"`
	HouseWins      int64 `json:"house_wins"`
	UserWins       int64 `json:"user_wins"`
	Ties           int64 `json:"ties"`
	TotalBetAmount int64 `json:"total_bet_amount"`
	Enabled        bool  `json:"enabled"`
}

func bindAdminHouseStatsApi(api huma.API, app core.App) {
	huma.Register(
		api,
		huma.Operation{
			OperationID: "admin-house-stats",
			Method:      http.MethodGet,
			Path:        "/house/stats",
			Summary:     "House player stats",
			Description: "Total games, win/lose/tie counts, and bet amounts for the house player.",
			Tags:     []string{"Admin", "Games"},
			Errors:   []int{http.StatusNotFound},
			Security: []map[string][]string{{shared.BearerAuthSecurityKey: {}}},
		},
		func(ctx context.Context, _ *struct{}) (*ApiSingleOutput[*HouseStatsResponse], error) {
			stats, err := services.GetHousePlayerStats(ctx, app.Adapter())
			if err != nil {
				return nil, err
			}
			return &ApiSingleOutput[*HouseStatsResponse]{
				Body: ApiSingleResponse[*HouseStatsResponse]{
					Data: &HouseStatsResponse{
						TotalGames:     stats.TotalGames,
						BettedGames:    stats.BettedGames,
						HouseWins:      stats.HouseWins,
						UserWins:       stats.UserWins,
						Ties:           stats.Ties,
						TotalBetAmount: stats.TotalBetAmount,
						Enabled:        stats.Enabled,
					},
				},
			}, nil
		},
	)
}

type HouseToggleInput struct {
	Enabled bool `json:"enabled" required:"true"`
}

func bindAdminHouseToggleApi(api huma.API, app core.App) {
	huma.Register(
		api,
		huma.Operation{
			OperationID: "admin-house-toggle",
			Method:      http.MethodPut,
			Path:        "/house/enabled",
			Summary:     "Enable or disable the house player",
			Description: "Sets the house player's enabled flag. When disabled, POST /games/rps/house returns 403.",
			Tags:     []string{"Admin", "Games"},
			Errors:   []int{http.StatusNotFound},
			Security: []map[string][]string{{shared.BearerAuthSecurityKey: {}}},
		},
		func(ctx context.Context, input *struct {
			Body HouseToggleInput
		}) (*ApiSingleOutput[*HouseStatsResponse], error) {
			house, err := services.GetHousePlayer(ctx, app.Adapter())
			if err != nil {
				return nil, err
			}

			updated, err := services.SetHouseEnabled(house.Metadata, input.Body.Enabled)
			if err != nil {
				return nil, err
			}
			house.Metadata = updated
			if _, err = app.Adapter().Gaming().UpdatePlayer(ctx, house); err != nil {
				return nil, err
			}

			stats, err := services.GetHousePlayerStats(ctx, app.Adapter())
			if err != nil {
				return nil, err
			}
			return &ApiSingleOutput[*HouseStatsResponse]{
				Body: ApiSingleResponse[*HouseStatsResponse]{
					Data: &HouseStatsResponse{
						TotalGames:     stats.TotalGames,
						BettedGames:    stats.BettedGames,
						HouseWins:      stats.HouseWins,
						UserWins:       stats.UserWins,
						Ties:           stats.Ties,
						TotalBetAmount: stats.TotalBetAmount,
						Enabled:        stats.Enabled,
					},
				},
			}, nil
		},
	)
}
