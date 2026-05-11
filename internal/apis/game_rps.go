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
	"github.com/tkahng/playground/internal/database/queries"
	"github.com/tkahng/playground/internal/middleware"
	"github.com/tkahng/playground/internal/middleware/humamiddleware"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/notification"
	"github.com/tkahng/playground/internal/populator"
	"github.com/tkahng/playground/internal/services"
	"github.com/tkahng/playground/internal/shared"
	"github.com/tkahng/playground/internal/stores"
	"github.com/tkahng/playground/internal/tools/mapper"
	"github.com/tkahng/playground/internal/tools/security"
	"github.com/tkahng/playground/internal/tools/sse"
	"github.com/tkahng/playground/internal/tools/types"
	"github.com/tkahng/playground/internal/tools/utils"
	"github.com/tkahng/playground/internal/workers"
)

type RpsGameFilter struct {
	PaginatedInput
	SortParams
	Ids           []string                       `query:"ids,omitempty" required:"false" minimum:"1" maximum:"100" format:"uuid"`
	Statuses      []models.RpsGameStatus         `query:"statuses,omitempty" required:"false" minimum:"1" maximum:"100" enum:"pending,cancelled,completed"`
	CompletedAt   types.OptionalParam[time.Time] `query:"completed_at,omitempty" required:"false"`
	CompletedAtOp queries.FilterOperator         `query:"completed_at_op,omitempty" required:"false" enum:"eq,gt,gte,lt,lte"`
	ExpiresAt     types.OptionalParam[time.Time] `query:"expires_at,omitempty" required:"false"`
	ExpiresAtOp   queries.FilterOperator         `query:"expires_at_op,omitempty" required:"false" enum:"eq,gt,gte,lt,lte"`
}

func bindFindCurrentPlayersRpsGamesApi(api huma.API, app core.App) {
	huma.Register(
		api,
		huma.Operation{
			OperationID: "find-current-players-rps-games",
			Method:      http.MethodGet,
			Path:        "/players/current-player/games/rps",
			Summary:     "find current players rps games",
			Description: "find current players rps games",
			Tags:        []string{"Games", "Player"},
			Errors:      []int{http.StatusUnauthorized},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
			Middlewares: humamiddleware.HumaChiMiddlewares(
				middleware.RequireCurrentPlayerMiddelware(),
			),
		},
		func(ctx context.Context, input *RpsGameFilter) (*ApiPaginatedOutput[*RpsGameWithParticipants], error) {
			currentPlayer := contextstore.GetContextCurrentPlayer(ctx)
			if currentPlayer == nil {
				return nil, huma.Error401Unauthorized("no player found.")
			}
			filter := &stores.RpsGameFilter{}
			filter.Page = input.Page
			filter.PerPage = input.PerPage
			filter.CompletedAt = input.CompletedAt
			filter.CompletedAtOp = input.CompletedAtOp
			filter.ExpiresAt = input.ExpiresAt
			filter.ExpiresAtOp = input.ExpiresAtOp
			filter.Ids = utils.ParseValidUUIDs(input.Ids...)
			filter.ParticipantIds = []uuid.UUID{currentPlayer.ID}
			filter.SortBy = input.SortBy
			filter.SortOrder = input.SortOrder
			filter.Statuses = input.Statuses
			games, err := app.Adapter().Gaming().FindRpsGames(ctx, filter)
			if err != nil {
				return nil, err
			}
			pop := populator.New(app.Adapter())
			gamesWithParticipants := []*services.RpsGameWithParticipants{}
			for _, game := range games {
				var gameWithPartipants *services.RpsGameWithParticipants = &services.RpsGameWithParticipants{
					RpsGame: game,
				}
				participants, err := app.Adapter().Gaming().FindRpsParticipants(ctx, &stores.RpsParticipantFilter{
					RpsGameIds: []uuid.UUID{game.ID},
				})
				if err != nil {
					return nil, err
				}
				for _, p := range participants {
					player, err := pop.GetPlayerByID(ctx, p.PlayerID)
					if err != nil {
						return nil, err
					}
					p.Player = player
					if p.Type == models.RpsParticipantTypeHost {
						gameWithPartipants.RequestingParticipant = p
					}
					if p.Type == models.RpsParticipantTypeGuest {
						gameWithPartipants.InvitedParticipant = p
					}
				}
				gamesWithParticipants = append(gamesWithParticipants, gameWithPartipants)
			}
			count, err := app.Adapter().Gaming().CountRpsGames(ctx, filter)
			if err != nil {
				return nil, err
			}
			return &ApiPaginatedOutput[*RpsGameWithParticipants]{
				Body: ApiPaginatedResponse[*RpsGameWithParticipants]{
					Data: mapper.Map(gamesWithParticipants, func(p *services.RpsGameWithParticipants) *RpsGameWithParticipants {
						// For pending games, hide the host's move from the guest. The host
						// chose their move when creating the game; revealing it before the
						// guest responds would let the guest pick the winning move.
						isPending := p.RpsGame != nil && p.RpsGame.Status == models.RpsGameStatusPending
						isViewer := func(participant *models.RpsParticipant) bool {
							return participant != nil && participant.PlayerID == currentPlayer.ID
						}
						viewerIsGuest := isPending &&
							p.InvitedParticipant != nil &&
							isViewer(p.InvitedParticipant)

						requestingAPI := ToApiRpsParticipant(p.RequestingParticipant)
						if viewerIsGuest {
							requestingAPI = ToApiRpsParticipantMasked(p.RequestingParticipant)
						}
						return &RpsGameWithParticipants{
							RpsGame:               toApiRpsGame(p.RpsGame),
							RequestingParticipant: requestingAPI,
							InvitedParticipant:    ToApiRpsParticipant(p.InvitedParticipant),
						}
					}),
					Meta: ApiGenerateMeta(&input.PaginatedInput, count),
				},
			}, nil
		},
	)
}

// emitRpsGameCompleted sends SSE completion events to both participants (best-effort).
func emitRpsGameCompleted(ctx context.Context, app core.App, result *services.RpsGameWithParticipants) {
	if result == nil {
		return
	}
	type side struct {
		playerID   uuid.UUID
		result     string
		yourMove   string
		oppMove    string
	}
	sides := []side{}
	if result.RequestingParticipant != nil && result.InvitedParticipant != nil {
		sides = append(sides,
			side{
				playerID: result.RequestingParticipant.PlayerID,
				result:   string(result.RequestingParticipant.Result),
				yourMove: string(result.RequestingParticipant.Move),
				oppMove:  string(result.InvitedParticipant.Move),
			},
			side{
				playerID: result.InvitedParticipant.PlayerID,
				result:   string(result.InvitedParticipant.Result),
				yourMove: string(result.InvitedParticipant.Move),
				oppMove:  string(result.RequestingParticipant.Move),
			},
		)
	}
	for _, s := range sides {
		payload := notification.NewNotificationPayload(
			"Game over",
			"Your Rock Paper Scissors result is ready",
			notification.RpsGameCompletedData{
				GameID:       result.RpsGame.ID,
				Result:       s.result,
				YourMove:     s.yourMove,
				OpponentMove: s.oppMove,
			},
		)
		if err := app.SseManager().Send(sse.PlayerChannel(s.playerID.String()), payload); err != nil {
			slog.WarnContext(ctx, "rps completed SSE notify failed", "player_id", s.playerID, "error", err)
		}
	}
}

func getRpsGameInviteFromTokenQuery(app core.App, ctx context.Context, token string) (*models.RpsGameInvite, error) {
	rpsGameInvite, err := app.Adapter().Gaming().FindRpsGameInvite(ctx, &stores.RpsGameInviteFilter{
		Tokens: []string{token},
	})
	if err != nil {
		return nil, err
	}
	if rpsGameInvite == nil {
		return nil, huma.Error400BadRequest("invalid token")
	}
	if rpsGameInvite.ExpiresAt.UTC().Before(time.Now().UTC()) {
		return nil, huma.Error400BadRequest("token expired")
	}
	return rpsGameInvite, nil
}

type SubmitMoveWithTokenInput struct {
	Token  string             `json:"token" required:"true" minlength:"2"`
	Move   RpsParticipantMove `json:"move" required:"true" enum:"rock,paper,scissors"`
	Status RpsGameStatus      `json:"status" required:"true" enum:"pending,completed,cancelled"`
}

func bindSubmitMoveWithTokenApi(api huma.API, app core.App) {
	// 20 req/min per IP — unauthenticated endpoint, token is the only auth factor.
	rl := newAuthRateLimiter(20, time.Minute)
	huma.Register(
		api,
		huma.Operation{
			OperationID: "submit-move-with-token",
			Method:      http.MethodPost,
			Path:        "/games/rps/token/submit-move",
			Summary:     "submit move to rps game with token",
			Description: "submit move to rps game with token",
			Tags:        []string{"Games", "Player"},
			Errors:      []int{http.StatusUnauthorized, http.StatusTooManyRequests},
			Middlewares: huma.Middlewares{rl},
		},
		func(ctx context.Context, input *struct {
			Body SubmitMoveWithTokenInput
		},
		) (*ApiSingleOutput[*RpsGameWithParticipants], error) {
			rpsGameInvite, err := getRpsGameInviteFromTokenQuery(app, ctx, input.Body.Token)
			if err != nil {
				return nil, err
			}
			rpsGameWithParticipants, err := app.RpsGame().FindRpsGameWithParticipants(ctx, rpsGameInvite.GameID)
			if err != nil {
				return nil, err
			}
			txErr := app.Adapter().RunInTxCtx(ctx, func(txCtx context.Context) error {
				rpsGameWithParticipants, err = app.RpsGame().RespondToGameRequest(txCtx, &services.GameRequestResponse{
					GameID:          rpsGameWithParticipants.RpsGame.ID,
					InvitedPlayerID: rpsGameWithParticipants.InvitedParticipant.PlayerID,
					Move:            models.RpsParticipantMove(input.Body.Move),
					Status:          models.RpsGameStatus(input.Body.Status),
				})
				if err != nil {
					return err
				}
				_, err = app.Adapter().Gaming().DeleteRpGameInvites(txCtx, &stores.RpsGameInviteFilter{
					Tokens: []string{input.Body.Token},
				})
				return err
			})
			if txErr != nil {
				return nil, txErr
			}
			emitRpsGameCompleted(ctx, app, rpsGameWithParticipants)
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

type VerifyRpsGameInviteInput struct {
	Token string `json:"token" required:"true" minlength:"2"`
}

func bindVerifyRpsGameInviteApi(api huma.API, app core.App) {
	rl := newAuthRateLimiter(20, time.Minute)
	huma.Register(
		api,
		huma.Operation{
			OperationID: "verify-rps-game-invite",
			Method:      http.MethodPost,
			Path:        "/games/rps/invites/verify",
			Summary:     "verify rps game invite",
			Description: "verify rps game invite",
			Tags:        []string{"Games", "Player"},
			Errors:      []int{http.StatusUnauthorized, http.StatusTooManyRequests},
			Security:    []map[string][]string{{}},
			Middlewares: huma.Middlewares{rl},
		},
		func(ctx context.Context, input *struct {
			Body VerifyRpsGameInviteInput
		},
		) (*ApiSingleOutput[*RpsGameWithParticipants], error) {
			rpsGameInvite, err := getRpsGameInviteFromTokenQuery(app, ctx, input.Body.Token)
			if err != nil {
				return nil, err
			}
			rpsGameWithParticipants, err := app.RpsGame().FindRpsGameWithParticipants(ctx, rpsGameInvite.GameID)
			if err != nil {
				return nil, err
			}
			// This endpoint is called from the guest's perspective via an invite token.
			// Hide the host's move so the guest cannot pick the winning response.
			return &ApiSingleOutput[*RpsGameWithParticipants]{
				Body: ApiSingleResponse[*RpsGameWithParticipants]{
					Data: &RpsGameWithParticipants{
						RpsGame:               toApiRpsGame(rpsGameWithParticipants.RpsGame),
						RequestingParticipant: ToApiRpsParticipantMasked(rpsGameWithParticipants.RequestingParticipant),
						InvitedParticipant:    ToApiRpsParticipant(rpsGameWithParticipants.InvitedParticipant),
					},
				},
			}, nil
		},
	)
}

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
		},
		) (*ApiSingleOutput[*RpsGameWithParticipants], error) {
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
				resp := &services.GameRequestResponse{
					GameID:          game.RpsGame.ID,
					InvitedPlayerID: currentPlayer.ID,
					Move:            models.RpsParticipantMove(input.Body.Move),
					Status:          models.RpsGameStatus(input.Body.Status),
				}
				rpsGameWithParticipants, err = app.RpsGame().RespondToGameRequest(txCtx, resp)
				return err
			})
			if txErr != nil {
				return nil, txErr
			}
			emitRpsGameCompleted(ctx, app, rpsGameWithParticipants)
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

func SendRpsGameRequestToUnregisteredPlayer(app core.App, ctx context.Context, input UnregisteredPlayerInput, currentPlayer *models.Player) (*services.RpsGameWithParticipants, error) {
	// find inviting player by email and unregistered.
	player, err := app.Adapter().Gaming().FindPlayer(ctx, &stores.PlayersFilter{
		Emails: []string{input.InvitingPlayerEmail},
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
			Email: input.InvitingPlayerEmail,
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
			RequestingPlayerMove: models.RpsParticipantMove(input.Move),
			DurationSeconds:      services.GameDurationSeconds,
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
			ExpiresAt:          rpsGameWithParticipants.RpsGame.ExpiresAt,
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
	return rpsGameWithParticipants, nil
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
		},
		) (*ApiSingleOutput[*RpsGameWithParticipants], error) {
			// get current player
			currentPlayer := contextstore.GetContextCurrentPlayer(ctx)
			if currentPlayer == nil {
				return nil, huma.Error401Unauthorized("no player found.")
			}
			rpsGameWithParticipants, err := SendRpsGameRequestToUnregisteredPlayer(app, ctx, input.Body, currentPlayer)
			if err != nil {
				return nil, err
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
	BetAmount        *int64             `json:"bet_amount,omitempty" required:"false" minimum:"1" doc:"Optional points wager. If set, the host must have sufficient balance."`
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
		},
		) (*ApiSingleOutput[*RpsGameWithParticipants], error) {
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
				reqInput := &services.RpsGameRequestInput{
					RequestingPlayerID:   currentPlayer.ID,
					InvitedPlayerID:      player.ID,
					RequestingPlayerMove: models.RpsParticipantMove(input.Body.Move),
					DurationSeconds:      services.GameDurationSeconds,
				}
				if input.Body.BetAmount != nil {
					reqInput.BetAmount = input.Body.BetAmount
					reqInput.HostUserID = &user.User.ID
				}
				rpsGameWithParticipants, err = app.RpsGame().RequestGame(txCtx, reqInput)
				return err
			})
			if txErr != nil {
				return nil, txErr
			}
			// Notify the invited player via SSE (best-effort).
			challengedPayload := notification.NewNotificationPayload(
				"New challenge",
				currentPlayer.Email+" challenged you to Rock Paper Scissors",
				notification.RpsGameChallengedData{
					GameID:             rpsGameWithParticipants.RpsGame.ID,
					RequestingPlayerID: currentPlayer.ID,
					RequestingEmail:    currentPlayer.Email,
				},
			)
			if err := app.SseManager().Send(sse.PlayerChannel(player.ID.String()), challengedPayload); err != nil {
				slog.WarnContext(ctx, "rps challenge SSE notify failed", "error", err)
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
