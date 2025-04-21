package apis

import (
	"context"
	"fmt"
	"net/http"

	"github.com/alexedwards/argon2id"
	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/authgo/internal/core"
	"github.com/tkahng/authgo/internal/db/models"
	"github.com/tkahng/authgo/internal/repository"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/security"
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
	var user *models.User
	var account *models.UserAccount
	user, err := repository.FindUserByEmail(ctx, db, input.Body.Email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, huma.Error404NotFound("User not found")
	}
	account, err = repository.FindUserAccountByUserIdAndProvider(ctx, db, user.ID, models.ProvidersCredentials)
	if err != nil {
		return nil, err
	}
	if account == nil {
		return nil, huma.Error404NotFound("No credentials cccount found")
	}

	err = api.app.SendPasswordResetEmail(ctx, db, user, api.app.Settings().Meta.AppURL)

	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (api *Api) ConfirmPasswordResetGetOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "confirm-password-reset-get",
		Method:      http.MethodGet,
		Path:        path,
		Summary:     "Confirm password reset",
		Description: "Confirm password reset",
		Tags:        []string{"Auth"},
		Errors:      []int{http.StatusNotFound},
	}
}

func (api *Api) ConfirmPasswordResetGet(ctx context.Context, input *OtpInput) (*struct{}, error) {
	if input.Type != shared.PasswordResetTokenType {
		return nil, fmt.Errorf("invalid token type. want verification_token, got  %v", input.Type)
	}
	opts := api.app.Settings().Auth
	claims, err := core.ParseResetToken(input.Token, opts.PasswordResetToken)
	if err != nil {
		return nil, fmt.Errorf("error at parsing verification token: %w", err)
	}
	if claims == nil {
		return nil, fmt.Errorf("token not found")
	}
	if claims.Type != shared.PasswordResetTokenType {
		return nil, fmt.Errorf("invalid token type. want verification_token, got  %v", claims.Type)
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
	claims, err := api.app.VerifyAndUsePasswordResetToken(ctx, db, input.Body.Token)
	if err != nil {
		return nil, err
	}
	user, err := repository.FindUserByEmail(ctx, db, claims.Email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, huma.Error404NotFound("User not found")
	}
	account, err := repository.FindUserAccountByUserIdAndProvider(ctx, db, user.ID, models.ProvidersCredentials)
	if err != nil {
		return nil, err
	}
	if account == nil {
		return nil, huma.Error404NotFound("No credentials cccount found")
	}

	hash, err := security.CreateHash(input.Body.Password, argon2id.DefaultParams)

	if err != nil {
		return nil, fmt.Errorf("error creating hash: %w", err)
	}

	err = repository.UpdateUserPassword(ctx, db, user.ID, hash)

	if err != nil {
		return nil, fmt.Errorf("error updating user password: %w", err)
	}
	return nil, nil

}
