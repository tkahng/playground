package apis

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	humasse "github.com/danielgtaylor/huma/v2/sse"
	"github.com/go-chi/httprate"
	"github.com/tkahng/playground/internal/contextstore"
	"github.com/tkahng/playground/internal/middleware"
	"github.com/tkahng/playground/internal/middleware/humamiddleware"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/stores"
	"github.com/tkahng/playground/internal/tools/mapper"
	"github.com/tkahng/playground/internal/tools/sse"
	"github.com/tkahng/playground/internal/userreaction"
)

type GeolocationCoordinate struct {
	Latitude  float64 `json:"latitude" required:"true"`
	Longitude float64 `json:"longitude" required:"true"`
}

type UserReactionDto struct {
	Type        string                 `json:"type" required:"true"`
	Coordinates *GeolocationCoordinate `json:"coordinates" required:"false"`
}
type UserReactionInput struct {
	Body UserReactionDto
}

func (api *Api) bindCreateUserReaction(aapi huma.API) {
	ipMiddleware := humamiddleware.HumaChiMiddleware(middleware.IpAddressMiddleware())
	rateLimitByIp := humamiddleware.HumaChiMiddleware(httprate.LimitByIP(1, 1*time.Second))
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
				rateLimitByIp,
				ipMiddleware,
			},
		},
		func(ctx context.Context, input *UserReactionInput) (*struct{}, error) {
			ip := contextstore.GetContextIPAddress(ctx)
			userInfo := contextstore.GetContextUserInfo(ctx)

			reaction := new(models.UserReaction)
			reaction.IpAddress = &ip
			if userInfo != nil {
				reaction.UserID = &userInfo.User.ID
			}
			var place *userreaction.Location
			if input.Body.Coordinates != nil {
				place = userreaction.GetLocationFromBody(ctx, input.Body.Coordinates.Longitude, input.Body.Coordinates.Latitude)
				if place == nil {
					if ip != "" {
						place = userreaction.GetLocationFromIp(ctx, ip)
					}
				}
			} else if ip != "" {
				place = userreaction.GetLocationFromIp(ctx, ip)
			}
			if place == nil {
				return nil, huma.Error500InternalServerError("failed to get location")
			}
			reaction.City = &place.City
			reaction.Country = &place.Country

			reaction.Type = input.Body.Type
			reaction, err := api.App().Adapter().UserReaction().CreateUserReaction(ctx, reaction)
			if err != nil {
				return nil, err
			}
			// utils.PrettyPrintJSON(reaction)
			err = api.App().EventManager().EventBus().Publish(ctx, userreaction.UserReactionCreated{
				UserReaction: reaction,
			})
			if err != nil {
				return nil, err
			}
			return nil, nil
		},
	)
}

func (api *Api) bindGetLatestUserReactionStats(aapi huma.API) {
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
			latest, err := api.App().Adapter().UserReaction().GetLastReaction(ctx)
			if err != nil {
				return nil, err
			}

			stats := new(userreaction.UserReactionStats)
			stats.LastCreated = userreaction.FromModelUserReaction(latest)
			recent, err := api.App().Adapter().UserReaction().CountByCountry(ctx, &stores.UserReactionFilter{
				PaginatedInput: stores.PaginatedInput{
					PerPage: 5,
				},
			})
			if err != nil {
				api.App().Logger().Error("failed to get recent user reactions", slog.Any("error", err))
			}
			stats.TopFiveCountries = mapper.Map(recent, func(r *stores.ReactionByCountry) userreaction.ReactionByCountry {
				return userreaction.ReactionByCountry{
					Country:        r.Country,
					TotalReactions: r.TotalReactions,
				}
			})
			count, err := api.App().Adapter().UserReaction().CountUserReactions(ctx, nil)
			if err != nil {
				api.App().Logger().Error("failed to get recent user reactions", slog.Any("error", err))
			}
			stats.TotalReactions = count
			return &ApiOutput[*userreaction.UserReactionStats]{Body: stats}, nil
		},
	)
}

type UserReactionSseInput struct {
}

func (api *Api) bindUserReactionSse(humapi huma.API) {
	handler := sse.ServeSSE(
		func(ctx context.Context, f func(any) error, input *UserReactionSseInput) sse.Client {
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
		handler,
	)
}
