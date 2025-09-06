package apis

import (
	"context"
)

func (api *Api) ConfirmVerification(ctx context.Context, input *OtpInput) (*struct{}, error) {
	action := api.App().Auth()
	err := action.HandleVerificationToken(ctx, input.Token)
	if err != nil {
		return nil, err
	}
	return nil, nil
}
