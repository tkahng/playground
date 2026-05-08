package apis

import (
	"context"
	"net/http"
	"time"

	"github.com/tkahng/playground/internal/contextstore"
	"github.com/tkahng/playground/internal/middleware"
	"github.com/tkahng/playground/internal/middleware/humamiddleware"
	"github.com/tkahng/playground/internal/shared"

	"github.com/danielgtaylor/huma/v2"
)

type AiUsageStatus struct {
	Consumed  int64 `json:"consumed"`
	Limit     int64 `json:"limit"`
	Remaining int64 `json:"remaining"`
}

type TeamAiUsageStatusInput struct {
	TeamID string `path:"team-id" required:"true" format:"uuid"`
}

func (api *Api) TeamAiUsageStatusBind(humaApi huma.API) {
	huma.Register(
		humaApi,
		huma.Operation{
			OperationID: "team-ai-usage-status",
			Method:      http.MethodGet,
			Path:        "/teams/{team-id}/ai-usage",
			Summary:     "Team AI usage status",
			Description: "Returns today's consumed tokens, the daily limit, and remaining tokens for the team.",
			Tags:        []string{"Task"},
			Errors:      []int{http.StatusUnauthorized},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
			Middlewares: humamiddleware.HumaChiMiddlewares(
				middleware.RequireTeamInfo(),
			),
		},
		func(ctx context.Context, _ *TeamAiUsageStatusInput) (*struct {
			Body AiUsageStatus
		}, error) {
			teamInfo := contextstore.GetContextTeamInfo(ctx)
			if teamInfo == nil {
				return nil, huma.Error401Unauthorized("no team info")
			}

			teamID := teamInfo.Member.TeamID

			limit, err := api.App().AiUsage().GetDailyLimit(ctx, teamID)
			if err != nil {
				return nil, err
			}

			consumed, err := api.App().Adapter().AiUsage().GetDailyTokensByTeam(ctx, teamID, time.Now().UTC())
			if err != nil {
				return nil, err
			}

			remaining := limit - consumed
			if remaining < 0 {
				remaining = 0
			}

			return &struct {
				Body AiUsageStatus
			}{
				Body: AiUsageStatus{
					Consumed:  consumed,
					Limit:     limit,
					Remaining: remaining,
				},
			}, nil
		},
	)
}
