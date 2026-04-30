package apis

import (
	"context"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/playground/internal/contextstore"
)

type RequestPasswordResetInput struct {
	Email string `form:"email" json:"email" example:"tkahng+01@gmail.com"`
}

type RequestPasswordResetOutput struct {
}

func (a *Api) bindRequestPasswordReset(api huma.API) {
	rl := newAuthRateLimiter(5, time.Minute)
	huma.Register(
		api,
		huma.Operation{
			OperationID: "request-password-reset",
			Method:      http.MethodPost,
			Path:        "/auth/request-password-reset",
			Summary:     "Request password reset",
			Description: "Request password reset",
			Tags:        []string{"Auth"},
			Errors:      []int{http.StatusNotFound, http.StatusTooManyRequests},
			Middlewares: huma.Middlewares{rl},
		},
		a.RequestPasswordReset,
	)
}
func (api *Api) RequestPasswordReset(ctx context.Context, input *struct {
	Body *RequestPasswordResetInput `json:"body" required:"true"`
}) (*RequestPasswordResetOutput, error) {
	checker := api.App().Checker()
	ok, err := checker.CannotBeSuperUserEmail(ctx, input.Body.Email)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, huma.Error400BadRequest("Cannot reset password for super user")
	}
	action := api.App().Auth()
	err = action.RequestPasswordReset(ctx, input.Body.Email)
	if err != nil {
		return nil, err
	}
	return nil, nil
}
func (a *Api) bindCheckPasswordReset(api huma.API) {
	rl := newAuthRateLimiter(5, time.Minute)
	huma.Register(
		api,
		huma.Operation{
			OperationID: "check-password-reset",
			Method:      http.MethodPost,
			Path:        "/auth/check-password-reset",
			Summary:     "Check password reset",
			Description: "Check password reset",
			Tags:        []string{"Auth"},
			Errors:      []int{http.StatusNotFound, http.StatusTooManyRequests},
			Middlewares: huma.Middlewares{rl},
		},
		a.CheckPasswordReset,
	)
}

func (api *Api) CheckPasswordReset(ctx context.Context, input *struct {
	Body *struct {
		Token string `json:"token" required:"true"`
	}
}) (*struct{}, error) {
	action := api.App().Auth()
	err := action.CheckPasswordResetToken(ctx, input.Body.Token)
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

func (a *Api) bindConfirmPasswordReset(api huma.API) {
	rl := newAuthRateLimiter(5, time.Minute)
	huma.Register(
		api,
		huma.Operation{
			OperationID: "confirm-password-reset",
			Method:      http.MethodPost,
			Path:        "/auth/confirm-password-reset",
			Summary:     "Confirm password reset",
			Description: "Confirm password reset",
			Tags:        []string{"Auth"},
			Errors:      []int{http.StatusNotFound, http.StatusTooManyRequests},
			Middlewares: huma.Middlewares{rl},
		},
		a.ConfirmPasswordReset,
	)
}

func (api *Api) ConfirmPasswordReset(ctx context.Context, input *struct {
	Body *ConfirmPasswordResetInput `json:"body" required:"true"`
}) (*RequestPasswordResetOutput, error) {
	err := api.App().Auth().ConfirmPasswordReset(ctx, input.Body.Token, input.Body.Password)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

type PasswordResetInput struct {
	PreviousPassword string `form:"previous_password" json:"previous_password"`
	NewPassword      string `form:"new_password" json:"new_password"`
}

func (a *Api) bindResetPassword(api huma.API) {
	huma.Register(
		api,
		huma.Operation{
			OperationID: "reset-password",
			Method:      http.MethodPost,
			Path:        "/auth/password-reset",
			Summary:     "Reset Password",
			Description: "Reset Password",
			Tags:        []string{"Auth"},
			Errors:      []int{http.StatusNotFound},
		},
		a.ResetPassword,
	)
}

func (api *Api) ResetPassword(ctx context.Context, input *struct {
	Body PasswordResetInput `json:"body" required:"true"`
}) (*struct{}, error) {
	claims := contextstore.GetContextUserInfo(ctx)
	if claims == nil {
		return nil, huma.Error404NotFound("User not found")
	}
	checker := api.App().Checker()
	ok, err := checker.CannotBeSuperUserEmail(ctx, claims.User.Email)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, huma.Error400BadRequest("Cannot reset password for super user")
	}
	err = api.App().Auth().UpdatePassword(ctx, claims.User.ID, input.Body.PreviousPassword, input.Body.NewPassword)
	if err != nil {
		return nil, err
	}
	return nil, nil
}
