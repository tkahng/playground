package apis

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/playground/internal/shared"
)

func bindAuthApi(api huma.API, appApi *Api) {

	// signup -------------------------------------------------------------
	appApi.bindSingup(api)
	// signin -------------------------------------------------------------
	appApi.bindSignin(api)
	//  me get ---------------------------------------------------------------
	appApi.bindMe(api)
	// me update -------------------------------------------------------------
	appApi.bindMeUpdate(api)
	// me delete -------------------------------------------------------------
	appApi.bindMeDelete(api)
	// refresh token -------------------------------------------------------------
	appApi.bindRefreshToken(api)
	// signout -------------------------------------------------------------
	appApi.bindSignout(api)
	// verify email -------------------------------------------------------------
	huma.Register(
		api,
		huma.Operation{
			OperationID: "verify-get",
			Method:      http.MethodGet,
			Path:        "/auth/verify",
			Summary:     "Verify",
			Description: "Verify",
			Deprecated:  true,
			Tags:        []string{"Auth", "Verify"},
			Errors:      []int{http.StatusNotFound, http.StatusBadRequest},
		},
		appApi.Verify,
	)

	// verify email post -------------------------------------------------------------
	huma.Register(
		api,
		huma.Operation{
			OperationID: "verify-post",
			Method:      http.MethodPost,
			Path:        "/auth/verify",
			Summary:     "Verify",
			Description: "Verify",
			Deprecated:  true,
			Tags:        []string{"Auth", "Verify"},
			Errors:      []int{http.StatusNotFound, http.StatusBadRequest},
		},
		appApi.VerifyPost,
	)
	// request verification -------------------------------------------------------------
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
		appApi.RequestVerification,
	)
	// confirm verification -------------------------------------------------------------
	appApi.bindVerifyEmail(api)
	// request password reset -------------------------------------------------------------
	huma.Register(
		api,
		huma.Operation{
			OperationID: "request-password-reset",
			Method:      http.MethodPost,
			Path:        "/auth/request-password-reset",
			Summary:     "Request password reset",
			Description: "Request password reset",
			Tags:        []string{"Auth"},
			Errors:      []int{http.StatusNotFound},
		},
		appApi.RequestPasswordReset,
	)
	// confirm password reset -------------------------------------------------------------
	huma.Register(
		api,
		huma.Operation{
			OperationID: "confirm-password-reset",
			Method:      http.MethodPost,
			Path:        "/auth/confirm-password-reset",
			Summary:     "Confirm password reset",
			Description: "Confirm password reset",
			Tags:        []string{"Auth"},
			Errors:      []int{http.StatusNotFound},
		},
		appApi.ConfirmPasswordReset,
	)
	// check password reset -------------------------------------------------------------
	huma.Register(
		api,
		huma.Operation{
			OperationID: "check-password-reset",
			Method:      http.MethodPost,
			Path:        "/auth/check-password-reset",
			Summary:     "Check password reset",
			Description: "Check password reset",
			Tags:        []string{"Auth"},
			Errors:      []int{http.StatusNotFound},
		},
		appApi.CheckPasswordReset,
	)
	// password reset
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
		appApi.ResetPassword,
	)

	huma.Register(
		api,
		huma.Operation{
			OperationID: "oauth2-callback-get",
			Method:      http.MethodGet,
			Path:        "/auth/callback",
			Summary:     "OAuth2 Callback (GET)",
			Description: "Handle OAuth2 callback (GET)",
			Tags:        []string{"Auth", "OAuth2"},
			Errors:      []int{http.StatusNotFound},
		},
		appApi.OAuth2CallbackGet,
	)

	huma.Register(
		api,
		huma.Operation{
			OperationID: "oauth2-callback-post",
			Method:      http.MethodPost,
			Path:        "/auth/callback",
			Summary:     "OAuth2 Callback (POST)",
			Description: "Handle OAuth2 callback (POST)",
			Tags:        []string{"Auth", "OAuth2"},
			Errors:      []int{http.StatusNotFound},
		},
		appApi.OAuth2CallbackPost,
	)

	huma.Register(
		api,
		huma.Operation{
			OperationID: "oauth2-authorization-url",
			Method:      http.MethodGet,
			Path:        "/auth/authorization-url",
			Summary:     "OAuth2 Authorization URL",
			Description: "Get OAuth2 authorization URL",
			Tags:        []string{"Auth", "OAuth2"},
			Errors:      []int{http.StatusNotFound},
		},
		appApi.OAuth2AuthorizationUrl,
	)
}
