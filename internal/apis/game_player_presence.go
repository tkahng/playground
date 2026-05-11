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
	"github.com/tkahng/playground/internal/tools/sse"
)

// recentActivityThreshold is used only as a secondary "recently active" signal
// when the player has no live SSE connection. It reflects the window within
// which an HTTP-request-based last_seen_at stamp is still meaningful.
const recentActivityThreshold = 5 * time.Minute

type PlayerPresenceResponse struct {
	Body struct {
		PlayerID   uuid.UUID  `json:"player_id"`
		IsConnected bool      `json:"is_connected"`
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
			Summary:     "Get player presence",
			Description: "Returns whether the player has an active SSE connection and when they were last active.",
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
		}) (*PlayerPresenceResponse, error) {
			player, err := app.Adapter().Gaming().FindPlayer(ctx, &stores.PlayersFilter{
				Ids: []uuid.UUID{input.PlayerID},
			})
			if err != nil {
				return nil, err
			}
			if player == nil {
				return nil, huma.Error404NotFound("player not found")
			}
			resp := &PlayerPresenceResponse{}
			resp.Body.PlayerID = player.ID
			resp.Body.IsConnected = app.SseManager().IsChannelConnected(sse.PlayerChannel(player.ID.String()))
			resp.Body.LastSeenAt = player.LastSeenAt
			return resp, nil
		},
	)
}
