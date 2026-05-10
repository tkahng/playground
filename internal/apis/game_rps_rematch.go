package apis

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/contextstore"
	"github.com/tkahng/playground/internal/core"
	"github.com/tkahng/playground/internal/middleware"
	"github.com/tkahng/playground/internal/middleware/humamiddleware"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/notification"
	"github.com/tkahng/playground/internal/services"
	"github.com/tkahng/playground/internal/shared"
	"github.com/tkahng/playground/internal/stores"
	"github.com/tkahng/playground/internal/tools/sse"
)

type RpsRematchStatus string

const (
	RpsRematchStatusPending  RpsRematchStatus = "pending"
	RpsRematchStatusAccepted RpsRematchStatus = "accepted"
	RpsRematchStatusDeclined RpsRematchStatus = "declined"
	RpsRematchStatusExpired  RpsRematchStatus = "expired"
)

type RpsRematchRequest struct {
	ID                 uuid.UUID        `json:"id"`
	OriginalGameID     uuid.UUID        `json:"original_game_id"`
	RequestingPlayerID uuid.UUID        `json:"requesting_player_id"`
	InvitedPlayerID    uuid.UUID        `json:"invited_player_id"`
	Status             RpsRematchStatus `json:"status" enum:"pending,accepted,declined,expired"`
	NewGameID          *uuid.UUID       `json:"new_game_id,omitempty"`
	ExpiresAt          time.Time        `json:"expires_at"`
	CreatedAt          time.Time        `json:"created_at"`
}

func toApiRpsRematchRequest(r *models.RpsRematchRequest) *RpsRematchRequest {
	if r == nil {
		return nil
	}
	return &RpsRematchRequest{
		ID:                 r.ID,
		OriginalGameID:     r.OriginalGameID,
		RequestingPlayerID: r.RequestingPlayerID,
		InvitedPlayerID:    r.InvitedPlayerID,
		Status:             RpsRematchStatus(r.Status),
		NewGameID:          r.NewGameID,
		ExpiresAt:          r.ExpiresAt,
		CreatedAt:          r.CreatedAt,
	}
}

func bindRpsRematchApi(api huma.API, app core.App) {
	bindRequestRematchApi(api, app)
	bindAcceptRematchApi(api, app)
	bindDeclineRematchApi(api, app)
	bindGetRematchApi(api, app)
}

func bindGetRematchApi(api huma.API, app core.App) {
	huma.Register(
		api,
		huma.Operation{
			OperationID: "get-rps-rematch-request",
			Method:      http.MethodGet,
			Path:        "/games/rps/{game-id}/rematch",
			Summary:     "Get pending rematch request for a game",
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
			GameID uuid.UUID `path:"game-id" required:"true" format:"uuid"`
		}) (*ApiSingleOutput[*RpsRematchRequest], error) {
			rematch, err := app.Adapter().Gaming().FindRpsRematchRequest(ctx, &stores.RpsRematchFilter{
				OriginalGameIDs: []uuid.UUID{input.GameID},
				Statuses:        []models.RpsRematchStatus{models.RpsRematchStatusPending},
			})
			if err != nil {
				return nil, err
			}
			if rematch == nil {
				return nil, huma.Error404NotFound("no pending rematch request for this game")
			}
			return &ApiSingleOutput[*RpsRematchRequest]{
				Body: ApiSingleResponse[*RpsRematchRequest]{
					Data: toApiRpsRematchRequest(rematch),
				},
			}, nil
		},
	)
}

func bindRequestRematchApi(api huma.API, app core.App) {
	huma.Register(
		api,
		huma.Operation{
			OperationID: "request-rps-rematch",
			Method:      http.MethodPost,
			Path:        "/games/rps/{game-id}/rematch",
			Summary:     "Request a rematch",
			Description: "Request a rematch for a completed game. Invited player has 45 seconds to accept.",
			Tags:        []string{"Games", "Player"},
			Errors:      []int{http.StatusUnauthorized, http.StatusNotFound, http.StatusConflict},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
			Middlewares: humamiddleware.HumaChiMiddlewares(
				middleware.RequireCurrentPlayerMiddelware(),
			),
		},
		func(ctx context.Context, input *struct {
			GameID uuid.UUID `path:"game-id" required:"true" format:"uuid"`
		}) (*ApiSingleOutput[*RpsRematchRequest], error) {
			currentPlayer := contextstore.GetContextCurrentPlayer(ctx)
			if currentPlayer == nil {
				return nil, huma.Error401Unauthorized("no player found")
			}
			// Determine the opponent (invited player) from game participants.
			game, err := app.RpsGame().FindRpsGameWithParticipants(ctx, input.GameID)
			if err != nil {
				return nil, err
			}
			if game == nil {
				return nil, huma.Error404NotFound("game not found")
			}
			var invitedPlayerID uuid.UUID
			var invitedPlayer *models.Player
			switch {
			case game.RequestingParticipant != nil && game.RequestingParticipant.PlayerID == currentPlayer.ID:
				invitedPlayerID = game.InvitedParticipant.PlayerID
				invitedPlayer = game.InvitedParticipant.Player
			case game.InvitedParticipant != nil && game.InvitedParticipant.PlayerID == currentPlayer.ID:
				invitedPlayerID = game.RequestingParticipant.PlayerID
				invitedPlayer = game.RequestingParticipant.Player
			default:
				return nil, huma.Error403Forbidden("you are not a participant in this game")
			}

			rematch, err := app.RpsGame().RequestRematch(ctx, &services.RematchRequestInput{
				OriginalGameID:     input.GameID,
				RequestingPlayerID: currentPlayer.ID,
				InvitedPlayerID:    invitedPlayerID,
			})
			if err != nil {
				return nil, err
			}

			// Notify opponent via SSE (best-effort).
			payload := notification.NewNotificationPayload(
				"Rematch requested",
				currentPlayer.Email+" wants a rematch!",
				notification.RpsRematchRequestedData{
					RematchRequestID:   rematch.ID,
					OriginalGameID:     rematch.OriginalGameID,
					RequestingPlayerID: rematch.RequestingPlayerID,
					RequestingEmail:    currentPlayer.Email,
					ExpiresAt:          rematch.ExpiresAt.Format(time.RFC3339),
				},
			)
			if invitedPlayer != nil {
				if err := app.SseManager().Send(sse.PlayerChannel(invitedPlayerID.String()), payload); err != nil {
					slog.WarnContext(ctx, "rematch request SSE notify failed", "error", err)
				}
			}

			return &ApiSingleOutput[*RpsRematchRequest]{
				Body: ApiSingleResponse[*RpsRematchRequest]{
					Data: toApiRpsRematchRequest(rematch),
				},
			}, nil
		},
	)
}

func bindAcceptRematchApi(api huma.API, app core.App) {
	huma.Register(
		api,
		huma.Operation{
			OperationID: "accept-rps-rematch",
			Method:      http.MethodPost,
			Path:        "/games/rps/rematches/{rematch-id}/accept",
			Summary:     "Accept a rematch request",
			Tags:        []string{"Games", "Player"},
			Errors:      []int{http.StatusUnauthorized, http.StatusNotFound, http.StatusBadRequest},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
			Middlewares: humamiddleware.HumaChiMiddlewares(
				middleware.RequireCurrentPlayerMiddelware(),
			),
		},
		func(ctx context.Context, input *struct {
			RematchID uuid.UUID `path:"rematch-id" required:"true" format:"uuid"`
		}) (*ApiSingleOutput[*RpsRematchRequest], error) {
			currentPlayer := contextstore.GetContextCurrentPlayer(ctx)
			if currentPlayer == nil {
				return nil, huma.Error401Unauthorized("no player found")
			}
			rematch, err := app.RpsGame().AcceptRematch(ctx, input.RematchID, currentPlayer.ID)
			if err != nil {
				return nil, err
			}

			// Notify the requester via SSE (best-effort).
			payload := notification.NewNotificationPayload(
				"Rematch accepted",
				"Your rematch request was accepted",
				notification.RpsRematchAcceptedData{
					RematchRequestID: rematch.ID,
					NewGameID:        *rematch.NewGameID,
				},
			)
			if err := app.SseManager().Send(sse.PlayerChannel(rematch.RequestingPlayerID.String()), payload); err != nil {
				slog.WarnContext(ctx, "rematch accept SSE notify failed", "error", err)
			}

			return &ApiSingleOutput[*RpsRematchRequest]{
				Body: ApiSingleResponse[*RpsRematchRequest]{
					Data: toApiRpsRematchRequest(rematch),
				},
			}, nil
		},
	)
}

func bindDeclineRematchApi(api huma.API, app core.App) {
	huma.Register(
		api,
		huma.Operation{
			OperationID: "decline-rps-rematch",
			Method:      http.MethodPost,
			Path:        "/games/rps/rematches/{rematch-id}/decline",
			Summary:     "Decline a rematch request",
			Tags:        []string{"Games", "Player"},
			Errors:      []int{http.StatusUnauthorized, http.StatusNotFound, http.StatusBadRequest},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
			Middlewares: humamiddleware.HumaChiMiddlewares(
				middleware.RequireCurrentPlayerMiddelware(),
			),
		},
		func(ctx context.Context, input *struct {
			RematchID uuid.UUID `path:"rematch-id" required:"true" format:"uuid"`
		}) (*ApiSingleOutput[*RpsRematchRequest], error) {
			currentPlayer := contextstore.GetContextCurrentPlayer(ctx)
			if currentPlayer == nil {
				return nil, huma.Error401Unauthorized("no player found")
			}
			rematch, err := app.RpsGame().DeclineRematch(ctx, input.RematchID, currentPlayer.ID)
			if err != nil {
				return nil, err
			}

			// Notify the requester via SSE (best-effort).
			payload := notification.NewNotificationPayload(
				"Rematch declined",
				"Your rematch request was declined",
				notification.RpsRematchDeclinedData{
					RematchRequestID: rematch.ID,
					OriginalGameID:   rematch.OriginalGameID,
				},
			)
			if err := app.SseManager().Send(sse.PlayerChannel(rematch.RequestingPlayerID.String()), payload); err != nil {
				slog.WarnContext(ctx, "rematch decline SSE notify failed", "error", err)
			}

			return &ApiSingleOutput[*RpsRematchRequest]{
				Body: ApiSingleResponse[*RpsRematchRequest]{
					Data: toApiRpsRematchRequest(rematch),
				},
			}, nil
		},
	)
}
