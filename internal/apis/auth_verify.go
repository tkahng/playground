package apis

import (
	"context"
)

type OtpInput struct {
	Token string `query:"token" json:"token" required:"true"`
}

func (api *Api) Verify(ctx context.Context, input *OtpInput) (*struct{}, error) {
	return verify(api, ctx, input)
}

func (api *Api) VerifyPost(ctx context.Context, input *struct{ Body *OtpInput }) (*struct{}, error) {
	return verify(api, ctx, input.Body)
}

func verify(api *Api, ctx context.Context, input *OtpInput) (*struct{}, error) {
	action := api.App().Auth()
	err := action.HandleVerificationToken(ctx, input.Token)
	if err != nil {
		return nil, err
	}
	return nil, nil

}
