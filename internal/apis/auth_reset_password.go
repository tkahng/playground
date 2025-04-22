package apis

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

type RequestPasswordResetInput struct {
	Email string `form:"email" json:"email" example:"tkahng+01@gmail.com"`
}

type RequestPasswordResetOutput struct {
}

func (api *Api) RequestPasswordResetOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "request-password-reset",
		Method:      http.MethodPost,
		Path:        path,
		Summary:     "Request password reset",
		Description: "Request password reset",
		Tags:        []string{"Auth"},
		Errors:      []int{http.StatusNotFound},
	}
}

func (api *Api) RequestPasswordReset(ctx context.Context, input *struct{ Body *RequestPasswordResetInput }) (*RequestPasswordResetOutput, error) {
	db := api.app.Db()
	action := api.app.NewAuthActions(db)
	err := action.HandlePasswordResetRequest(ctx, input.Body.Email)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (api *Api) CheckPasswordResetOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "check-password-reset",
		Method:      http.MethodGet,
		Path:        path,
		Summary:     "Check password reset",
		Description: "Check password reset",
		Tags:        []string{"Auth"},
		Errors:      []int{http.StatusNotFound},
	}
}

func (api *Api) CheckPasswordResetGet(ctx context.Context, input *OtpInput) (*struct{}, error) {

	db := api.app.Db()
	action := api.app.NewAuthActions(db)
	err := action.CheckResetPasswordToken(ctx, input.Token)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (api *Api) ConfirmPasswordResetOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "confirm-password-reset",
		Method:      http.MethodPost,
		Path:        path,
		Summary:     "Confirm password reset",
		Description: "Confirm password reset",
		Tags:        []string{"Auth"},
		Errors:      []int{http.StatusNotFound},
	}
}

type ConfirmPasswordResetInput struct {
	Token           string `form:"token" json:"token"`
	Password        string `form:"password" json:"password"`
	ConfirmPassword string `form:"confirm_password" json:"confirm_password"`
}

func (api *Api) ConfirmPasswordReset(ctx context.Context, input *struct{ Body *ConfirmPasswordResetInput }) (*RequestPasswordResetOutput, error) {
	db := api.app.Db()
	action := api.app.NewAuthActions(db)
	err := action.HandlePasswordResetToken(ctx, input.Body.Token, input.Body.Password)
	if err != nil {
		return nil, err
	}
	return nil, nil
}
