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
	"github.com/tkahng/playground/internal/stores"
	"github.com/tkahng/playground/internal/tools/geoip"
	"github.com/tkahng/playground/internal/tools/mapper"
	"github.com/tkahng/playground/internal/tools/sse"
	"github.com/tkahng/playground/internal/userreaction"
)

type UserReactionDto struct {
	Type string `json:"type" required:"true"`
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

				} else if city != nil {
					reaction.City = &city.City.Names.English
					reaction.Country = &city.Country.ISOCode
				}
			}
			reaction.Type = input.Body.Type
			reaction, err := a.App().Adapter().UserReaction().CreateUserReaction(ctx, reaction)
			if err != nil {
				return nil, err
			}
			err = a.App().EventManager().EventBus().Publish(ctx, userreaction.UserReactionCreated{
				UserReaction: reaction,
			})
			if err != nil {
				return nil, err
			}
			return nil, nil
		},
	)

}

func (a *Api) BindGetLatestUserReactionStats(aapi huma.API) {
	huma.Register(
		aapi,
		huma.Operation{
			OperationID: "user-reaction-stats",
			Method:      http.MethodGet,
			Path:        "/user-reactions/stats",
			Summary:     "user-reaction-stats",
			Description: "user-reaction-stats",
			Tags:        []string{"User Reactions"},
			Errors:      []int{http.StatusInternalServerError, http.StatusBadRequest},
		},
		func(ctx context.Context, input *struct{}) (*ApiOutput[*userreaction.UserReactionStats], error) {
			latest, err := a.App().Adapter().UserReaction().GetLastReaction(ctx)
			if err != nil {
				return nil, err
			}

			stats := new(userreaction.UserReactionStats)
			stats.LastCreated = userreaction.FromModelUserReaction(latest)
			fmt.Println(stats.LastCreated != nil)
			recent, err := a.App().Adapter().UserReaction().CountByCountry(ctx, &stores.UserReactionFilter{
				PaginatedInput: stores.PaginatedInput{
					PerPage: 5,
				},
			})
			if err != nil {
				a.App().Logger().Error("failed to get recent user reactions", slog.Any("error", err))
			}
			stats.TopFiveCountries = mapper.Map(recent, func(r *stores.ReactionByCountry) userreaction.ReactionByCountry {
				return userreaction.ReactionByCountry{
					Country:        r.Country,
					TotalReactions: r.TotalReactions,
				}
			})
			fmt.Println(stats.TopFiveCountries)
			count, err := a.App().Adapter().UserReaction().CountUserReactions(ctx, nil)
			if err != nil {
				a.App().Logger().Error("failed to get recent user reactions", slog.Any("error", err))
			}
			stats.TotalReactions = count
			fmt.Println(stats)
			return &ApiOutput[*userreaction.UserReactionStats]{Body: stats}, nil
		},
	)

}

type UserReactionSseInput struct {
}

func (api *Api) BindUserReactionSse(humapi huma.API) {
	hanlder := sse.ServeSSE[UserReactionSseInput](
		func(ctx context.Context, f func(any) error) sse.Client {
			return sse.NewClient(sse.UserReactionsChannel, f, slog.Default(), func() any {
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
		30*time.Second,
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
			Middlewares: huma.Middlewares{},
			Errors:      []int{http.StatusInternalServerError, http.StatusBadRequest},
		},
		map[string]any{
			"latest_user_reaction_stats": &userreaction.LatestUserReactionStatsSseEvent{},
			"ping":                       &PingMessage{},
		},
		hanlder,
	)

}
