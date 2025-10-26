package apis

import (
	"context"
	"fmt"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/playground/internal/shared"
)

type SignoutDto struct {
	RefreshToken string `json:"refresh_token" cookie:"refresh_token" form:"refresh_token" required:"true"`
}

func (a *Api) bindSignout(api huma.API) {
	huma.Register(
		api,
		huma.Operation{
			OperationID: "signout",
			Method:      http.MethodPost,
			Path:        "/auth/signout",
			Summary:     "Signout",
			Description: "Signout",
			Tags:        []string{"Auth"},
			Errors:      []int{http.StatusUnauthorized, http.StatusNotFound},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
		},
		a.Signout,
	)
}

func (api *Api) Signout(ctx context.Context, input *struct{ Body SignoutDto }) (*struct{}, error) {
	action := api.App().Auth()
	err := action.Signout(ctx, input.Body.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("error signing out: %w", err)
	}
	return nil, nil
}
