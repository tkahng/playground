package apis

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/playground/internal/contextstore"
	"github.com/tkahng/playground/internal/middleware"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/shared"
	"github.com/tkahng/playground/internal/tools/geoip"
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
			Path:        "/users/{user-id}/reactions",
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
