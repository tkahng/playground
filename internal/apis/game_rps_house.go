package apis

import (
	"context"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/playground/internal/contextstore"
	"github.com/tkahng/playground/internal/core"
	"github.com/tkahng/playground/internal/middleware"
	"github.com/tkahng/playground/internal/middleware/humamiddleware"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/services"
	"github.com/tkahng/playground/internal/shared"
)

type ChallengeHouseInput struct {
	Move      RpsParticipantMove `json:"move" required:"true" enum:"rock,paper,scissors"`
	BetAmount *int64             `json:"bet_amount,omitempty" required:"false" minimum:"1" maximum:"500" doc:"Optional points wager (max 500)."`
}

type ChallengeHouseResponse struct {
	RpsGame               *RpsGame        `json:"rps_game"`
	RequestingParticipant *RpsParticipant `json:"requesting_participant"`
	InvitedParticipant    *RpsParticipant `json:"invited_participant"`
	HouseMessage          *string         `json:"house_message,omitempty"`
	CooldownEndsAt        time.Time       `json:"cooldown_ends_at"`
}

func toChallengeHouseResponse(r *services.ChallengeHouseResult) *ChallengeHouseResponse {
	return &ChallengeHouseResponse{
		RpsGame:               toApiRpsGame(r.Game.RpsGame),
		RequestingParticipant: ToApiRpsParticipant(r.Game.RequestingParticipant),
		InvitedParticipant:    ToApiRpsParticipant(r.Game.InvitedParticipant),
		HouseMessage:          r.HouseMessage,
		CooldownEndsAt:        r.CooldownEndsAt,
	}
}

func bindChallengeHouseApi(api huma.API, app core.App) {
	huma.Register(
		api,
		huma.Operation{
			OperationID: "challenge-house",
			Method:      http.MethodPost,
			Path:        "/games/rps/house",
			Summary:     "challenge the house",
			Description: "Play RPS against the house bot. Result is immediate. Cooldown applies between games.",
			Tags:        []string{"Games", "Player"},
			Errors:      []int{http.StatusUnauthorized, http.StatusConflict, http.StatusTooManyRequests, http.StatusBadRequest},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
			Middlewares: humamiddleware.HumaChiMiddlewares(
				middleware.RequireCurrentPlayerMiddelware(),
			),
		},
		func(ctx context.Context, input *struct {
			Body ChallengeHouseInput
		}) (*ApiSingleOutput[*ChallengeHouseResponse], error) {
			user := contextstore.GetContextUserInfo(ctx)
			if user == nil {
				return nil, huma.Error401Unauthorized("unauthorized")
			}
			currentPlayer := contextstore.GetContextCurrentPlayer(ctx)
			if currentPlayer == nil {
				return nil, huma.Error401Unauthorized("no player found")
			}

			svcInput := &services.ChallengeHouseInput{
				RequestingPlayerID:   currentPlayer.ID,
				RequestingPlayerMove: models.RpsParticipantMove(input.Body.Move),
				BetAmount:            input.Body.BetAmount,
			}
			if input.Body.BetAmount != nil {
				svcInput.HostUserID = &user.User.ID
			}

			result, err := app.RpsGame().ChallengeHouse(ctx, svcInput)
			if err != nil {
				return nil, err
			}

			return &ApiSingleOutput[*ChallengeHouseResponse]{
				Body: ApiSingleResponse[*ChallengeHouseResponse]{
					Data: toChallengeHouseResponse(result),
				},
			}, nil
		},
	)
}
