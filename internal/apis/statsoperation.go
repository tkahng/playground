package apis

import (
	"context"
	"errors"

	"github.com/tkahng/authgo/internal/core"
	"github.com/tkahng/authgo/internal/queries"
	"github.com/tkahng/authgo/internal/shared"
)

type StatsResponse struct {
	Body *shared.UserStats `json:"body"`
}

func (api *Api) Stats(ctx context.Context, input *struct{}) (*StatsResponse, error) {
	db := api.app.Db()
	user := core.GetContextUserInfo(ctx)
	if user == nil {
		return nil, errors.New("user not found")
	}
	stats, err := queries.GetUserTaskStats(ctx, db, user.User.ID)
	if err != nil {
		return nil, err
	}
	return &StatsResponse{
		Body: &shared.UserStats{
			Task: *stats,
		},
	}, nil
}
