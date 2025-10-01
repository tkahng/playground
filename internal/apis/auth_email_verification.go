package apis

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/playground/internal/contextstore"
)

type EmailVerificationInput struct {
	Token string `json:"token" form:"token" query:"token" required:"true"`
}

func (api *Api) RequestVerification(ctx context.Context, input *struct{}) (*struct{}, error) {
	claims := contextstore.GetContextUserInfo(ctx)
	if claims == nil {
		return nil, huma.Error404NotFound("User not found")
	}
	if claims.User.EmailVerifiedAt != nil {
		return nil, huma.Error404NotFound("Email already verified")
	}
	err := api.App().Auth2().SendEmailVerification(ctx, claims.User.Email)

	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (api *Api) VerifyEmail(ctx context.Context, input *EmailVerificationInput) (*struct{}, error) {
	err := api.App().Auth2().ValidateEmailVerification(ctx, input.Token)
	if err != nil {
		return nil, err
	}
	return nil, nil
}
