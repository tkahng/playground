package apis

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/authgo/internal/core"
	"github.com/tkahng/authgo/internal/shared"
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
	checker := api.app.NewChecker(ctx)
	err := checker.CannotBeSuperUserEmail(input.Body.Email)
	if err != nil {
		return nil, err
	}
	action := api.app.NewAuthActions(db)
	err = action.HandlePasswordResetRequest(ctx, input.Body.Email)
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

func (api *Api) ResetPasswordOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "reset-password",
		Method:      http.MethodPost,
		Path:        path,
		Summary:     "Reset password",
		Description: "Reset password",
		Tags:        []string{"Auth"},
		Errors:      []int{http.StatusNotFound},
		Security: []map[string][]string{
			{shared.BearerAuthSecurityKey: {}},
		},
	}
}

type PasswordResetInput struct {
	PreviousPassword string `form:"previous_password" json:"previous_password"`
	NewPassword      string `form:"new_password" json:"new_password"`
}

func (api *Api) ResetPassword(ctx context.Context, input *struct{ Body PasswordResetInput }) (*struct{}, error) {
	db := api.app.Db()
	claims := core.GetContextUserInfo(ctx)
	checker := api.app.NewChecker(ctx)
	err := checker.CannotBeSuperUserEmail(claims.User.Email)
	if err != nil {
		return nil, err
	}
	action := api.app.NewAuthActions(db)
	if claims == nil {
		return nil, huma.Error404NotFound("User not found")
	}
	err = action.ResetPassword(ctx, claims.User.ID, input.Body.PreviousPassword, input.Body.NewPassword)
	if err != nil {
		return nil, err
	}
	return nil, nil
}
