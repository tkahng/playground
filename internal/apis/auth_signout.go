package apis

import (
	"context"
	"fmt"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/authgo/internal/shared"
)

func (api *Api) SignoutOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "signout",
		Method:      http.MethodPost,
		Path:        path,
		Summary:     "Signout",
		Description: "Signout",
		Tags:        []string{"Auth"},
		Errors:      []int{http.StatusUnauthorized, http.StatusNotFound},
		Security: []map[string][]string{
			{shared.BearerAuthSecurityKey: {}},
		},
	}
}

type SignoutDto struct {
	RefreshToken string `json:"refresh_token" cookie:"refresh_token" form:"refresh_token" required:"true"`
}

func (api *Api) Signout(ctx context.Context, input *struct{ Body SignoutDto }) (*struct{}, error) {
	db := api.app.Db()
	action := api.app.NewAuthActions(db)
	err := action.Signout(ctx, input.Body.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("error signing out: %w", err)
	}
	return nil, nil
}
