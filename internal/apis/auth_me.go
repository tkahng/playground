package apis

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/authgo/internal/core"
	"github.com/tkahng/authgo/internal/db/models"
	"github.com/tkahng/authgo/internal/shared"
)

func (api *Api) MeOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "me",
		Method:      http.MethodGet,
		Path:        path,
		Summary:     "Me",
		Description: "Me",
		Tags:        []string{"Auth"},
		Errors:      []int{http.StatusUnauthorized, http.StatusNotFound},
		Security: []map[string][]string{
			{shared.BearerAuthSecurityKey: {}},
		},
	}
}

type MeOutput struct {
	Body *models.User
}

func (api *Api) Me(ctx context.Context, input *struct{ Body *SigninDto }) (*MeOutput, error) {
	claims := core.GetUserClaims(ctx)
	if claims == nil {
		return nil, huma.Error404NotFound("User not found")
	}
	return &MeOutput{
		Body: claims.User,
	}, nil

}
