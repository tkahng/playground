package apis

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/contextstore"
	"github.com/tkahng/playground/internal/core"
	"github.com/tkahng/playground/internal/middleware"
	"github.com/tkahng/playground/internal/middleware/humamiddleware"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/services"
	"github.com/tkahng/playground/internal/shared"
	"github.com/tkahng/playground/internal/stores"
	"github.com/tkahng/playground/internal/tools/security"
	"github.com/tkahng/playground/internal/tools/types"
	"github.com/tkahng/playground/internal/workers"
)

type SubmitMoveToGameInput struct {
	Move   RpsParticipantMove `json:"move" required:"true" enum:"rock,paper,scissors"`
	Status RpsGameStatus      `json:"status" required:"true" enum:"pending,completed,cancelled"`
}

func bindSubmitMoveToRpsGameApi(api huma.API, app core.App) {
	huma.Register(
		api,
		huma.Operation{
			OperationID: "submit-move-to-rps-game",
			Method:      http.MethodPost,
			Path:        "/games/rps/{game-id}/submit-move",
			Summary:     "submit move to rps game",
			Description: "submit move to rps game",
			Tags:        []string{"Games", "Player"},
			Errors:      []int{http.StatusUnauthorized},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
			Middlewares: humamiddleware.HumaChiMiddlewares(
				middleware.RequireCurrentPlayerMiddelware(),
			),
		},
		func(ctx context.Context, input *struct {
			GameID string `path:"game-id" required:"true" format:"uuid"`
			Body   SubmitMoveToGameInput
		}) (*ApiSingleOutput[*RpsGameWithParticipants], error) {
			currentPlayer := contextstore.GetContextCurrentPlayer(ctx)
			if currentPlayer == nil {
				return nil, huma.Error401Unauthorized("no player found.")
			}
			gameId, err := uuid.Parse(input.GameID)
			if err != nil {
				return nil, huma.Error400BadRequest("error parsing game id")
			}
			game, err := app.RpsGame().FindRpsGameWithParticipants(ctx, gameId)
			if err != nil {
				return nil, err
			}
			if game == nil {
				return nil, huma.Error404NotFound("game not found")
			}
			if currentPlayer.ID != game.InvitedParticipant.PlayerID {
				return nil, huma.Error401Unauthorized("not invited player")
			}
			var rpsGameWithParticipants *services.RpsGameWithParticipants
			txErr := app.Adapter().RunInTxCtx(ctx, func(txCtx context.Context) error {
				rpsGameWithParticipants, err = app.RpsGame().RespondToGameRequest(txCtx, &services.GameRequestResponse{
					GameID:          game.RpsGame.ID,
					InvitedPlayerID: currentPlayer.ID,
					Move:            models.RpsParticipantMove(input.Body.Move),
					Status:          models.RpsGameStatus(input.Body.Status),
				})
				return err
			})
			if txErr != nil {
				return nil, txErr
			}
			return &ApiSingleOutput[*RpsGameWithParticipants]{
				Body: ApiSingleResponse[*RpsGameWithParticipants]{
					Data: &RpsGameWithParticipants{
						RpsGame:               toApiRpsGame(rpsGameWithParticipants.RpsGame),
						RequestingParticipant: ToApiRpsParticipant(rpsGameWithParticipants.RequestingParticipant),
						InvitedParticipant:    ToApiRpsParticipant(rpsGameWithParticipants.InvitedParticipant),
					},
				},
			}, nil
		},
	)
}

type UnregisteredPlayerInput struct {
	InvitingPlayerEmail string             `json:"inviting-player-email" required:"true" format:"email"`
	Move                RpsParticipantMove `json:"move" required:"true" enum:"rock,paper,scissors"`
}

func bindSendGameRequestToUnRegisteredPlayerApi(api huma.API, app core.App) {
	huma.Register(
		api,
		huma.Operation{
			OperationID: "send-game-request-to-unregistered-player",
			Method:      http.MethodPost,
			Path:        "/games/rps/requests/unregistered",
			Summary:     "send game request to unregistered player",
			Description: "send game request to unregistered player",
			Tags:        []string{"Games", "Player"},
			Errors:      []int{http.StatusUnauthorized},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
			Middlewares: humamiddleware.HumaChiMiddlewares(
				middleware.RequireCurrentPlayerMiddelware(),
			),
		},
		func(ctx context.Context, input *struct {
			Body UnregisteredPlayerInput
		}) (*ApiSingleOutput[*RpsGameWithParticipants], error) {
			// get current player
			currentPlayer := contextstore.GetContextCurrentPlayer(ctx)
			if currentPlayer == nil {
				return nil, huma.Error401Unauthorized("no player found.")
			}
			// find inviting player by email and unregistered.
			player, err := app.Adapter().Gaming().FindPlayer(ctx, &stores.PlayersFilter{
				Emails: []string{input.Body.InvitingPlayerEmail},
				Registered: types.OptionalParam[bool]{
					Value: false,
					IsSet: true,
				},
			})
			if err != nil {
				return nil, err
			}
			if player == nil {
				// if player not found, create player
				player, err = app.Adapter().Gaming().CreatePlayer(ctx, &models.Player{
					Email: input.Body.InvitingPlayerEmail,
				})
				if err != nil {
					return nil, err
				}
			}
			// check if player can play again inviting player
			canPlay, err := app.RpsGame().PlayerCanPlayWithPlayer(ctx, currentPlayer.ID, player.ID)
			if err != nil {
				return nil, err
			}
			if !canPlay {
				return nil, huma.Error400BadRequest("player can't play with invited player")
			}
			//
			var rpsGameWithParticipants *services.RpsGameWithParticipants
			txErr := app.Adapter().RunInTxCtx(ctx, func(txCtx context.Context) error {
				// request game. returns game and participants.
				rpsGameWithParticipants, err = app.RpsGame().RequestGame(txCtx, &services.RpsGameRequestInput{
					RequestingPlayerID:   currentPlayer.ID,
					InvitedPlayerID:      player.ID,
					RequestingPlayerMove: models.RpsParticipantMove(input.Body.Move),
					DurationSeconds:      3 * 24 * 60 * 60,
				})
				if err != nil {
					return err
				}
				// create invite
				invitation, err := app.Adapter().Gaming().CreateRpsGameInvite(txCtx, &models.RpsGameInvite{
					GameID:             rpsGameWithParticipants.RpsGame.ID,
					RequestingPlayerID: currentPlayer.ID,
					InvitedPlayerID:    rpsGameWithParticipants.InvitedParticipant.PlayerID,
					Token:              security.GenerateTokenKey(),
				})
				if err != nil {
					return err
				}
				// send invitation
				err = app.JobService().EnqueueRpsGameInviteJob(txCtx, &workers.RpsGameInvitationJobArgs{
					Email:          player.Email,
					InvitedByEmail: currentPlayer.Email,
					TokenHash:      invitation.Token,
				})
				return err
			})
			if txErr != nil {
				return nil, txErr
			}
			return &ApiSingleOutput[*RpsGameWithParticipants]{
				Body: ApiSingleResponse[*RpsGameWithParticipants]{
					Data: &RpsGameWithParticipants{
						RpsGame:               toApiRpsGame(rpsGameWithParticipants.RpsGame),
						RequestingParticipant: ToApiRpsParticipant(rpsGameWithParticipants.RequestingParticipant),
						InvitedParticipant:    ToApiRpsParticipant(rpsGameWithParticipants.InvitedParticipant),
					},
				},
			}, nil
		},
	)
}

type RpsGameRequestInput struct {
	InvitingPlayerId uuid.UUID          `json:"inviting_player_id" required:"true" format:"uuid"`
	Move             RpsParticipantMove `json:"move" required:"true" enum:"rock,paper,scissors"`
}

func bindSendGameRequestToRegisteredPlayerApi(api huma.API, app core.App) {
	huma.Register(
		api,
		huma.Operation{
			OperationID: "send-game-request-to-registered-player",
			Method:      http.MethodPost,
			Path:        "/games/rps/requests",
			Summary:     "send game request to registered player",
			Description: "send game request to registered player",
			Tags:        []string{"Games", "Player"},
			Errors:      []int{http.StatusUnauthorized},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
			Middlewares: humamiddleware.HumaChiMiddlewares(
				middleware.RequireCurrentPlayerMiddelware(),
			),
		},
		func(ctx context.Context, input *struct {
			Body RpsGameRequestInput
		}) (*ApiSingleOutput[*RpsGameWithParticipants], error) {
			user := contextstore.GetContextUserInfo(ctx)
			if user == nil {
				return nil, huma.Error401Unauthorized("Unauthorized. No user info")
			}
			currentPlayer := contextstore.GetContextCurrentPlayer(ctx)
			if currentPlayer == nil {
				return nil, huma.Error401Unauthorized("no player found.")
			}
			filter := &stores.PlayersFilter{}
			filter.Ids = []uuid.UUID{input.Body.InvitingPlayerId}
			filter.Registered = types.OptionalParam[bool]{
				Value: true,
				IsSet: true,
			}
			player, err := app.Adapter().Gaming().FindPlayer(ctx, filter)
			if err != nil {
				return nil, err
			}
			if player == nil {
				return nil, huma.Error404NotFound("player not found")
			}
			// check if player can play again inviting player
			canPlay, err := app.RpsGame().PlayerCanPlayWithPlayer(ctx, currentPlayer.ID, player.ID)
			if err != nil {
				return nil, err
			}
			if !canPlay {
				return nil, huma.Error400BadRequest("player can't play with invited player")
			}
			var rpsGameWithParticipants *services.RpsGameWithParticipants
			txErr := app.Adapter().RunInTxCtx(ctx, func(txCtx context.Context) error {
				rpsGameWithParticipants, err = app.RpsGame().RequestGame(txCtx, &services.RpsGameRequestInput{
					RequestingPlayerID:   currentPlayer.ID,
					InvitedPlayerID:      player.ID,
					RequestingPlayerMove: models.RpsParticipantMove(input.Body.Move),
					DurationSeconds:      3 * 24 * 60 * 60,
				})
				return err
			})
			if txErr != nil {
				return nil, txErr
			}
			return &ApiSingleOutput[*RpsGameWithParticipants]{
				Body: ApiSingleResponse[*RpsGameWithParticipants]{
					Data: &RpsGameWithParticipants{
						RpsGame:               toApiRpsGame(rpsGameWithParticipants.RpsGame),
						RequestingParticipant: ToApiRpsParticipant(rpsGameWithParticipants.RequestingParticipant),
						InvitedParticipant:    ToApiRpsParticipant(rpsGameWithParticipants.InvitedParticipant),
					},
				},
			}, nil
		},
	)
}
