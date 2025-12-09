package apis

import (
	"context"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/contextstore"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/shared"
	"github.com/tkahng/playground/internal/stores"
)

type ApiPlayer struct {
	_           struct{}   `db:"players" schema:"gaming" json:"-"`
	ID          uuid.UUID  `db:"id,pk" json:"id"`
	Email       string     `db:"email" json:"email"`
	DisplayName *string    `db:"display_name" json:"display_name,omitempty" required:"false"`
	UserID      *uuid.UUID `db:"user_id" json:"user_id,omitempty" required:"false"`
	Metadata    []byte     `db:"metadata" json:"metadata"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at" json:"updated_at"`
	User        *ApiUser   `db:"user" src:"user_id" dest:"id" table:"auth.users" json:"user,omitempty"`
}

func ToApiPlayer(player *models.Player) *ApiPlayer {
	if player == nil {
		return nil
	}
	return &ApiPlayer{
		ID:          player.ID,
		Email:       player.Email,
		DisplayName: player.DisplayName,
		UserID:      player.UserID,
		Metadata:    player.Metadata,
		CreatedAt:   player.CreatedAt,
		UpdatedAt:   player.UpdatedAt,
		User:        fromUserModel(player.User),
	}
}

func bindGameGetUserPlayerApi(appApi *Api) {
	huma.Register(
		appApi.Api(),
		huma.Operation{
			OperationID: "get-user-player",
			Method:      http.MethodGet,
			Path:        "/games/players",
			Summary:     "Put player.",
			Description: "Gets a player for the user. Returns the player if there is one.",
			Tags:        []string{"Games", "Player"},
			Errors:      []int{http.StatusUnauthorized},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
		},
		func(ctx context.Context, input *struct{}) (*ApiOutput[*ApiPlayer], error) {
			user := contextstore.GetContextUserInfo(ctx)
			if user == nil {
				return nil, huma.Error401Unauthorized("Unauthorized. No user info")
			}
			player, err := appApi.App().Adapter().Gaming().FindPlayer(ctx, &stores.PlayersFilter{
				Emails: []string{user.User.Email},
			})
			if err != nil {
				return nil, err
			}
			return &ApiOutput[*ApiPlayer]{
				Body: ToApiPlayer(player),
			}, nil
		},
	)
}

type GamePutUserPlayerArgs struct {
	DisplayName *string `json:"display_name" required:"true" nullable:"true" minLength:"1" maxLength:"80"`
}

func bindGamePutUserPlayerApi(appApi *Api) {
	huma.Register(
		appApi.Api(),
		huma.Operation{
			OperationID: "put-user-player",
			Method:      http.MethodPut,
			Path:        "/games/players",
			Summary:     "Put user player.",
			Description: "Creates a player for the user if there is none, otherwise updates the player. Returns the player.",
			Tags:        []string{"Games", "Player"},
			Errors:      []int{http.StatusUnauthorized},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
		},
		func(ctx context.Context, input *struct {
			Body GamePutUserPlayerArgs
		}) (*ApiOutput[*ApiPlayer], error) {
			user := contextstore.GetContextUserInfo(ctx)
			if user == nil {
				return nil, huma.Error401Unauthorized("Unauthorized. No user info")
			}
			player, err := appApi.App().Adapter().Gaming().FindPlayer(ctx, &stores.PlayersFilter{
				Emails: []string{user.User.Email},
			})
			if err != nil {
				return nil, err
			}
			if player == nil {
				player, err = appApi.App().Adapter().Gaming().CreatePlayer(ctx, &models.Player{
					Email:       user.User.Email,
					UserID:      &user.User.ID,
					DisplayName: input.Body.DisplayName,
				})
				if err != nil {
					return nil, err
				}
			} else {
				player, err = appApi.App().Adapter().Gaming().UpdatePlayer(ctx, &models.Player{
					ID:          player.ID,
					UserID:      &user.User.ID,
					DisplayName: user.User.Name,
					Email:       user.User.Email,
					Metadata:    player.Metadata,
				})
				if err != nil {
					return nil, err
				}
			}
			return &ApiOutput[*ApiPlayer]{
				Body: ToApiPlayer(player),
			}, nil
		},
	)
}
