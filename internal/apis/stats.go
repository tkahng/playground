package apis

import (
	"context"
	"errors"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/playground/internal/contextstore"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/shared"
)

type UserStats struct {
	Task models.TaskStats `json:"task_stats" db:"task_stats"`
}

type StatsResponse struct {
	Body *UserStats `json:"body"`
}

func bindStatsApi(api huma.API, appApi *Api) {
	statsGroup := huma.NewGroup(api)
	huma.Register(
		statsGroup,
		huma.Operation{
			OperationID: "stats-get",
			Method:      http.MethodGet,
			Path:        "/stats",
			Summary:     "Get stats",
			Description: "Get stats",
			Tags:        []string{"Stats"},
			Errors:      []int{http.StatusNotFound},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
		},
		appApi.Stats,
	)
}
func (api *Api) Stats(ctx context.Context, input *struct{}) (*StatsResponse, error) {
	user := contextstore.GetContextUserInfo(ctx)
	if user == nil {
		return nil, errors.New("user not found")
	}
	stats, err := api.App().Adapter().Task().GetTeamTaskStats(ctx, user.User.ID)
	if err != nil {
		return nil, err
	}
	return &StatsResponse{
		Body: &UserStats{
			Task: *stats,
		},
	}, nil
}
