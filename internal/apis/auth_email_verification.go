package apis

import (
	"context"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/playground/internal/contextstore"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/shared"
	"github.com/tkahng/playground/internal/tools/types"
)

type EmailVerificationPostInput struct {
	Token string `json:"token" form:"token" required:"true"`
}

func (a *Api) bindRequestVerification(api huma.API) {
	huma.Register(
		api,
		huma.Operation{
			OperationID: "request-verification",
			Method:      http.MethodPost,
			Path:        "/auth/request-verification",
			Summary:     "Email verification request",
			Description: "Request email verification",
			Tags:        []string{"Auth", "Verify"},
			Errors:      []int{http.StatusNotFound},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
		},
		a.RequestVerification,
	)
}
func (api *Api) RequestVerification(ctx context.Context, input *struct{}) (*struct{}, error) {
	claims := contextstore.GetContextUserInfo(ctx)
	if claims == nil {
		return nil, huma.Error404NotFound("User not found")
	}
	if claims.User.EmailVerifiedAt != nil {
		return nil, huma.Error409Conflict("Email already verified")
	}
	err := api.App().Auth().SendEmailVerification(ctx, claims.User.Email)

	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (a *Api) bindVerifyEmail(api huma.API) {
	huma.Register(
		api,
		huma.Operation{
			OperationID: "confirm-verification",
			Method:      http.MethodPost,
			Path:        "/auth/confirm-verification",
			Summary:     "Confirm Email verification request",
			Description: "Confirm Request email verification",
			Tags:        []string{"Auth", "Verify"},
			Errors:      []int{http.StatusNotFound},
		},
		a.VerifyEmail,
	)
}

// VerifyEmail checks whether the user has been verified through the email verification link.
//
// If the user has been verified, it will do the following operations:
//
// 1. Validate the verificaiton token, then delete it.
//
// 1. Create a stripe customer for the user
//
// 2. Update the user's email_verified_at to the current time if it has not been set.
//
// these two operations will run in a transaction, and if any of them fails, the transaction will be rolled back
// to prevent data inconsistency. In case of a roll back, the only thing that will persist is the stripe customer on stripe's side,
// which will not be accessed since each team creation attempt will simply create a new customer.
func (api *Api) VerifyEmail(ctx context.Context, input *struct{ Body EmailVerificationPostInput }) (*struct{}, error) {
	runInTxErr := api.App().Adapter().RunInTxCtx(ctx, func(txCtx context.Context) error {
		// validate the verification token
		email, err := api.App().Token().ValidateToken(txCtx, input.Body.Token, models.TokenTypesVerificationToken)
		if err != nil {
			return err
		}
		// get userinfo
		userInfo, err := api.App().Adapter().User().GetUserInfo(txCtx, email)
		if err != nil {
			return err
		}
		if userInfo == nil {
			return huma.Error404NotFound("userInfo not found")
		}
		user := &userInfo.User
		if user.EmailVerifiedAt != nil {
			return huma.Error409Conflict("Email already verified")
		}
		// update user's email_verified_at if it has not been set
		user.EmailVerifiedAt = types.Pointer(time.Now())
		err = api.App().Adapter().User().UpdateUser(txCtx, user)
		if err != nil {
			return err
		}
		// create user customer
		_, err = api.App().Payment().CreateUserCustomer(
			txCtx,
			user,
		)
		if err != nil {
			return err
		}

		return nil
	})
	if runInTxErr != nil {
		return nil, runInTxErr
	}
	return nil, nil
}
