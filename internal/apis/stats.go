package apis

import (
	"context"
	"errors"

	"github.com/tkahng/playground/internal/contextstore"
	"github.com/tkahng/playground/internal/models"
)

type UserStats struct {
	Task models.TaskStats `json:"task_stats" db:"task_stats"`
}

type StatsResponse struct {
	Body *UserStats `json:"body"`
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
