package apis

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

func (api *Api) RefreshTokenOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "refresh-token",
		Method:      http.MethodPost,
		Path:        path,
		Summary:     "Refresh token",
		Description: "Count the number of colors for all themes",
		Tags:        []string{"Auth"},
		Errors:      []int{http.StatusNotFound},
	}
}

type RefreshTokenInput struct {
	RefreshToken string `json:"refresh_token" cookie:"refresh_token" form:"refresh_token" required:"true"`
}

func (api *Api) RefreshToken(ctx context.Context, input *struct{ Body *RefreshTokenInput }) (*AuthenticatedResponse, error) {
	db := api.app.Db()
	claims, err := api.app.RefreshTokens(ctx, db, input.Body.RefreshToken)
	if err != nil {
		return nil, err
	}
	return &AuthenticatedResponse{
		Body: *claims,
	}, nil
}
