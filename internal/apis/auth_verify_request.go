package apis

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/authgo/internal/core"
	"github.com/tkahng/authgo/internal/shared"
)

type EmailVerificationInput struct {
	Token string `json:"token" form:"token" query:"token" required:"true"`
}

type EmailVerificationRequestInput struct {
	Email string `json:"email" form:"email" required:"true"`
}

func (api *Api) RequestVerificationOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "request-verification",
		Method:      http.MethodPost,
		Path:        path,
		Summary:     "Email verification request",
		Description: "Request email verification",
		Tags:        []string{"Auth", "Verify"},
		Errors:      []int{http.StatusNotFound},
		Security: []map[string][]string{
			{shared.BearerAuthSecurityKey: {}},
		},
	}
}

func (api *Api) RequestVerification(ctx context.Context, input *struct{}) (*struct{}, error) {
	db := api.app.Db()
	claims := core.GetContextUserClaims(ctx)
	if claims == nil {
		return nil, huma.Error404NotFound("User not found")
	}
	if !claims.User.EmailVerifiedAt.IsNull() {
		return nil, huma.Error404NotFound("Email already verified")
	}
	err := api.app.SendVerificationEmail(ctx, db, &claims.User, api.app.Settings().Meta.AppURL)
	if err != nil {
		return nil, err
	}
	// TODO: send email
	return nil, nil
}
