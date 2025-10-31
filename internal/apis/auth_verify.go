package apis

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/playground/internal/models"
)

type OtpInput struct {
	Token string            `query:"token" json:"token" required:"true"`
	Type  models.TokenTypes `query:"type" json:"type" required:"true" enum:"invite_token,verification_token,password_reset_token,state_token"`
}

func (api *Api) Verify(ctx context.Context, input *OtpInput) (*struct{}, error) {
	return verify(api, ctx, input)
}

func (api *Api) VerifyPost(ctx context.Context, input *struct{ Body *OtpInput }) (*struct{}, error) {
	return verify(api, ctx, input.Body)
}

func verify(api *Api, ctx context.Context, input *OtpInput) (*struct{}, error) {
	action := api.App().Auth()
	switch input.Type {
	case models.TokenTypesVerificationToken:
		err := action.ValidateEmailVerification(ctx, input.Token)
		if err != nil {
			return nil, err
		}
		return nil, nil
	default:
		return nil, huma.Error400BadRequest("Invalid token type")
	}

}
