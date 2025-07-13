package apis

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/playground/internal/contextstore"
)

type RequestPasswordResetInput struct {
	Email string `form:"email" json:"email" example:"tkahng+01@gmail.com"`
}

type RequestPasswordResetOutput struct {
}

func (api *Api) RequestPasswordReset(ctx context.Context, input *struct{ Body *RequestPasswordResetInput }) (*RequestPasswordResetOutput, error) {

	checker := api.App().Checker()
	ok, err := checker.CannotBeSuperUserEmail(ctx, input.Body.Email)
	if err != nil {
		return nil, err
	}
	if ok {
		return nil, huma.Error400BadRequest("Cannot reset password for super user")
	}
	action := api.App().Auth()
	err = action.HandlePasswordResetRequest(ctx, input.Body.Email)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (api *Api) CheckPasswordResetGet(ctx context.Context, input *OtpInput) (*struct{}, error) {

	action := api.App().Auth()
	err := action.HandleCheckResetPasswordToken(ctx, input.Token)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

type ConfirmPasswordResetInput struct {
	Token           string `form:"token" json:"token"`
	Password        string `form:"password" json:"password"`
	ConfirmPassword string `form:"confirm_password" json:"confirm_password"`
}

func (api *Api) ConfirmPasswordReset(ctx context.Context, input *struct{ Body *ConfirmPasswordResetInput }) (*RequestPasswordResetOutput, error) {

	action := api.App().Auth()
	err := action.HandlePasswordResetToken(ctx, input.Body.Token, input.Body.Password)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

type PasswordResetInput struct {
	PreviousPassword string `form:"previous_password" json:"previous_password"`
	NewPassword      string `form:"new_password" json:"new_password"`
}

func (api *Api) ResetPassword(ctx context.Context, input *struct{ Body PasswordResetInput }) (*struct{}, error) {

	claims := contextstore.GetContextUserInfo(ctx)
	if claims == nil {
		return nil, huma.Error404NotFound("User not found")
	}
	checker := api.App().Checker()
	ok, err := checker.CannotBeSuperUserEmail(ctx, claims.User.Email)
	if err != nil {
		return nil, err
	}
	if ok {
		return nil, huma.Error400BadRequest("Cannot reset password for super user")
	}
	action := api.App().Auth()
	err = action.ResetPassword(ctx, claims.User.ID, input.Body.PreviousPassword, input.Body.NewPassword)
	if err != nil {
		return nil, err
	}
	return nil, nil
}
