package apis

import (
	"context"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/contextstore"
	"github.com/tkahng/playground/internal/core"
	"github.com/tkahng/playground/internal/middleware"
	"github.com/tkahng/playground/internal/middleware/humamiddleware"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/shared"
	"github.com/tkahng/playground/internal/stores"
	"github.com/tkahng/playground/internal/tools/mapper"
	"github.com/tkahng/playground/internal/tools/types"
	"github.com/tkahng/playground/internal/tools/utils"
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

func bindGetMyPlayerApi(api huma.API, app core.App) {
	huma.Register(
		api,
		huma.Operation{
			OperationID: "get-my-player",
			Method:      http.MethodGet,
			Path:        "/games/players/me",
			Summary:     "get my player.",
			Description: "Gets a player for the user. Returns the player if there is one.",
			Tags:        []string{"Games", "Player"},
			Errors:      []int{http.StatusUnauthorized},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
		},
		func(ctx context.Context, input *struct{}) (*ApiSingleOutput[*ApiPlayer], error) {
			user := contextstore.GetContextUserInfo(ctx)
			if user == nil {
				return nil, huma.Error401Unauthorized("Unauthorized. No user info")
			}
			player, err := app.Adapter().Gaming().FindPlayer(ctx, &stores.PlayersFilter{
				Emails: []string{user.User.Email},
			})
			if err != nil {
				return nil, err
			}
			return &ApiSingleOutput[*ApiPlayer]{
				Body: ApiSingleResponse[*ApiPlayer]{
					Data: ToApiPlayer(player),
				},
			}, nil
		},
	)
}

type GamePutPlayerMeArgs struct {
	DisplayName *string `json:"display_name" required:"true" nullable:"true" minLength:"1" maxLength:"80"`
}

func bindPutMyPlayerApi(api huma.API, app core.App) {
	huma.Register(
		api,
		huma.Operation{
			OperationID: "put-my-player",
			Method:      http.MethodPut,
			Path:        "/games/players/me",
			Summary:     "Put user player.",
			Description: "Creates a player for the user if there is none, otherwise updates the player. Returns the player.",
			Tags:        []string{"Games", "Player"},
			Errors:      []int{http.StatusUnauthorized},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
		},
		func(ctx context.Context, input *struct {
			Body GamePutPlayerMeArgs
		}) (*ApiSingleOutput[*ApiPlayer], error) {
			user := contextstore.GetContextUserInfo(ctx)
			if user == nil {
				return nil, huma.Error401Unauthorized("Unauthorized. No user info")
			}
			player, err := app.Adapter().Gaming().FindPlayer(ctx, &stores.PlayersFilter{
				Emails: []string{user.User.Email},
			})
			if err != nil {
				return nil, err
			}
			if player == nil {
				player, err = app.Adapter().Gaming().CreatePlayer(ctx, &models.Player{
					Email:       user.User.Email,
					UserID:      &user.User.ID,
					DisplayName: input.Body.DisplayName,
				})
				if err != nil {
					return nil, err
				}
			} else {
				player, err = app.Adapter().Gaming().UpdatePlayer(ctx, &models.Player{
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
			return &ApiSingleOutput[*ApiPlayer]{
				Body: ApiSingleResponse[*ApiPlayer]{
					Data: ToApiPlayer(player),
				},
			}, nil
		},
	)
}

type PlayersFilter struct {
	PaginatedInput
	SortParams
	Ids          []string                  `query:"ids,omitempty" required:"false" minimum:"1" maximum:"100" format:"uuid"`
	Q            string                    `query:"q,omitempty" required:"false"`
	Emails       []string                  `query:"emails,omitempty" required:"false" minimum:"1" maximum:"100" format:"email"`
	DisplayNames []string                  `query:"display_names,omitempty" required:"false" minimum:"1" maximum:"100" format:"email"`
	UserIds      []string                  `query:"user_ids,omitempty" required:"false" minimum:"1" maximum:"100" format:"uuid"`
	Registered   types.OptionalParam[bool] `query:"registered,omitempty" required:"false"`
}

func bindFindPlayersApi(api huma.API, app core.App) {
	huma.Register(
		api,
		huma.Operation{
			OperationID: "get-players",
			Method:      http.MethodGet,
			Path:        "/games/players",
			Summary:     "get players.",
			Description: "gets a list of players.",
			Tags:        []string{"Games", "Player"},
			Errors:      []int{http.StatusUnauthorized},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
		},
		func(ctx context.Context, input *PlayersFilter) (*ApiPaginatedOutput[*ApiPlayer], error) {
			user := contextstore.GetContextUserInfo(ctx)
			if user == nil {
				return nil, huma.Error401Unauthorized("Unauthorized. No user info")
			}
			filter := &stores.PlayersFilter{}
			filter.DisplayNames = input.DisplayNames
			filter.Emails = input.Emails
			filter.UserIds = utils.ParseValidUUIDs(input.UserIds...)
			filter.Registered = input.Registered
			filter.Ids = utils.ParseValidUUIDs(input.Ids...)
			filter.Q = input.Q
			filter.Page = input.Page
			filter.PerPage = input.PerPage
			filter.SortBy = input.SortBy
			filter.SortOrder = input.SortOrder
			players, err := app.Adapter().Gaming().FindPlayers(ctx, filter)
			if err != nil {
				return nil, err
			}
			count, err := app.Adapter().Gaming().CountPlayers(ctx, filter)
			if err != nil {
				return nil, err
			}
			return &ApiPaginatedOutput[*ApiPlayer]{
				Body: ApiPaginatedResponse[*ApiPlayer]{
					Data: mapper.Map(players, ToApiPlayer),
					Meta: ApiGenerateMeta(&input.PaginatedInput, count),
				},
			}, nil
		},
	)
}
func bindFindRegisteredPlayerByEmailApi(api huma.API, app core.App) {
	huma.Register(
		api,
		huma.Operation{
			OperationID: "search-registered-player",
			Method:      http.MethodGet,
			Path:        "/games/players/registered/email/{inviting-player-email}",
			Summary:     "search registered player by email",
			Description: "search registered player by email",
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
			Email string `path:"inviting-player-email" required:"true" format:"email"`
		}) (*ApiSingleOutput[*ApiPlayer], error) {
			user := contextstore.GetContextUserInfo(ctx)
			if user == nil {
				return nil, huma.Error401Unauthorized("Unauthorized. No user info")
			}
			filter := &stores.PlayersFilter{}
			filter.Emails = []string{input.Email}
			filter.Registered = types.OptionalParam[bool]{
				Value: true,
				IsSet: true,
			}
			filter.Page = 0
			filter.PerPage = 1
			player, err := app.Adapter().Gaming().FindPlayer(ctx, filter)
			if err != nil {
				return nil, err
			}
			return &ApiSingleOutput[*ApiPlayer]{
				Body: ApiSingleResponse[*ApiPlayer]{
					Data: ToApiPlayer(player),
				},
			}, nil
		},
	)
}
