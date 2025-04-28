package apis

import (
	"context"
	"errors"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/authgo/internal/core"
	"github.com/tkahng/authgo/internal/repository"
	"github.com/tkahng/authgo/internal/shared"
)

func (api *Api) StatsOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "stats-get",
		Method:      http.MethodGet,
		Path:        path,
		Summary:     "Get stats",
		Description: "Get stats",
		Tags:        []string{"Stats"},
		Errors:      []int{http.StatusNotFound},
		Security: []map[string][]string{
			{shared.BearerAuthSecurityKey: {}},
		},
	}
}

type StatsResponse struct {
	Body *shared.UserStats `json:"body"`
}

func (api *Api) Stats(ctx context.Context, input *struct{}) (*StatsResponse, error) {
	db := api.app.Db()
	user := core.GetContextUserInfo(ctx)
	if user == nil {
		return nil, errors.New("user not found")
	}
	stats, err := repository.GetUserTaskStats(ctx, db, user.User.ID)
	if err != nil {
		return nil, err
	}
	return &StatsResponse{
		Body: &shared.UserStats{
			Task: *stats,
		},
	}, nil
}
