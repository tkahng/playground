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

func (h *Api) VerifyPost(ctx context.Context, input *struct{ Body *OtpInput }) (*struct{}, error) {
	return verify(h, ctx, input.Body)
}

func verify(api *Api, ctx context.Context, input *OtpInput) (*struct{}, error) {
	action := api.app.NewAuthActions()
	err := action.HandleVerificationToken(ctx, input.Token)
	if err != nil {
		return nil, err
	}
	return nil, nil

}
