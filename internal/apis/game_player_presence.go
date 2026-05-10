package apis

import (
	"context"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/core"
	"github.com/tkahng/playground/internal/middleware"
	"github.com/tkahng/playground/internal/middleware/humamiddleware"
	"github.com/tkahng/playground/internal/shared"
	"github.com/tkahng/playground/internal/stores"
)

const onlineThreshold = 2 * time.Minute

func IsPlayerOnline(lastSeenAt *time.Time) bool {
	if lastSeenAt == nil {
		return false
	}
	return time.Since(*lastSeenAt) <= onlineThreshold
}

type PlayerOnlineStatusResponse struct {
	Body struct {
		PlayerID uuid.UUID `json:"player_id"`
		IsOnline bool      `json:"is_online"`
		LastSeenAt *time.Time `json:"last_seen_at,omitempty"`
	}
}

func bindGetPlayerOnlineStatusApi(api huma.API, app core.App) {
	huma.Register(
		api,
		huma.Operation{
			OperationID: "get-player-online-status",
			Method:      http.MethodGet,
			Path:        "/players/{player-id}/online-status",
			Summary:     "Get player online status",
			Description: "Returns whether the player has been active within the last 2 minutes.",
			Tags:        []string{"Games", "Player"},
			Errors:      []int{http.StatusUnauthorized, http.StatusNotFound},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
			Middlewares: humamiddleware.HumaChiMiddlewares(
				middleware.RequireCurrentPlayerMiddelware(),
			),
		},
		func(ctx context.Context, input *struct {
			PlayerID uuid.UUID `path:"player-id" required:"true" format:"uuid"`
		}) (*PlayerOnlineStatusResponse, error) {
			player, err := app.Adapter().Gaming().FindPlayer(ctx, &stores.PlayersFilter{
				Ids: []uuid.UUID{input.PlayerID},
			})
			if err != nil {
				return nil, err
			}
			if player == nil {
				return nil, huma.Error404NotFound("player not found")
			}
			resp := &PlayerOnlineStatusResponse{}
			resp.Body.PlayerID = player.ID
			resp.Body.IsOnline = IsPlayerOnline(player.LastSeenAt)
			resp.Body.LastSeenAt = player.LastSeenAt
			return resp, nil
		},
	)
}
