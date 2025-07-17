package apis

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	humasse "github.com/danielgtaylor/huma/v2/sse"
	"github.com/tkahng/playground/internal/contextstore"
	"github.com/tkahng/playground/internal/middleware"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/shared"
	"github.com/tkahng/playground/internal/tools/geoip"
	"github.com/tkahng/playground/internal/tools/sse"
	"github.com/tkahng/playground/internal/userreaction"
)

type UserReactionDto struct {
	Type string `json:"type"`
}
type UserReactionInput struct {
	Body UserReactionDto
}

func (a *Api) BindCreateUserReaction(aapi huma.API) {
	ipMiddleware := middleware.IpAddressMiddleware(aapi)
	huma.Register(
		aapi,
		huma.Operation{
			OperationID: "create-user-reaction",
			Method:      http.MethodPost,
			Path:        "/user-reactions",
			Summary:     "create-user-reaction",
			Description: "create user reaction",
			Tags:        []string{"User Reactions"},
			Errors:      []int{http.StatusInternalServerError, http.StatusBadRequest},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
			Middlewares: huma.Middlewares{
				ipMiddleware,
			},
		},
		func(ctx context.Context, input *UserReactionInput) (*struct{}, error) {
			ip := contextstore.GetContextIPAddress(ctx)
			userInfo := contextstore.GetContextUserInfo(ctx)

			reaction := new(models.UserReaction)
			if userInfo != nil {
				reaction.UserID = &userInfo.User.ID
			}
			if ip != "" {
				reaction.IpAddress = &ip
				city, err := geoip.City(ip)
				if err != nil {
					slog.ErrorContext(ctx, "error getting city", slog.String("ip", ip), slog.Any("error", err))
				}
				if city != nil {
					reaction.City = &city.City.Names.English
					reaction.Country = &city.Country.Names.English
				}
			}
			reaction.Type = input.Body.Type
			_, err := a.App().Adapter().UserReaction().CreateUserReaction(ctx, reaction)
			if err != nil {
				return nil, err
			}

			return nil, nil
		},
	)

}

type UserReactionSseInput struct {
}

func (api *Api) BindUserReactionSse(humapi huma.API) {
	membermiddleware := middleware.TeamInfoFromTeamMemberID(humapi, api.App())
	hanlder := sse.ServeSSE[TeamMemberSseInput](
		func(ctx context.Context, f func(any) error) sse.Client {
			return sse.NewClient("user-reactions", f, slog.Default(), func() any {
				return &PingMessage{
					Message: "ping",
				}
			})
		},
		func(ctx context.Context, cf context.CancelFunc, c sse.Client) {
			api.app.SseManager().RegisterClient(ctx, cf, c)
		},
		func(c sse.Client) {
			fmt.Println("unregistering client")
			api.app.SseManager().UnregisterClient(c)
		},
		1*time.Second,
	)
	humasse.Register(
		humapi,
		huma.Operation{
			OperationID: "user-reaction-sse",
			Method:      http.MethodGet,
			Path:        "/user-reactions/sse",
			Summary:     "user-reaction-sse",
			Description: "user-reaction-sse",
			Tags:        []string{"User Reactions"},
			Middlewares: huma.Middlewares{
				membermiddleware,
			},
			Errors: []int{http.StatusInternalServerError, http.StatusBadRequest},
		},
		map[string]any{
			"latest_user_reaction_stats": &userreaction.LatestUserReactionStatsSseEvent{},
			"latest_user_reaction":       &userreaction.LatestUserReactionSseEvent{},
			"ping":                       &PingMessage{},
		},
		hanlder,
	)

}
