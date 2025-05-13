package apis

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/authgo/internal/core"
)

type EmailVerificationInput struct {
	Token string `json:"token" form:"token" query:"token" required:"true"`
}

type EmailVerificationRequestInput struct {
	Email string `json:"email" form:"email" required:"true"`
}

func (api *Api) RequestVerification(ctx context.Context, input *struct{}) (*struct{}, error) {
	action := api.app.Auth()
	claims := core.GetContextUserInfo(ctx)
	if claims == nil {
		return nil, huma.Error404NotFound("User not found")
	}
	if claims.User.EmailVerifiedAt != nil {
		return nil, huma.Error404NotFound("Email already verified")
	}
	err := action.SendOtpEmail(core.EmailTypeVerify, ctx, &claims.User)
	if err != nil {
		return nil, err
	}
	return nil, nil
}
