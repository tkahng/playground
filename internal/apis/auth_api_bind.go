package apis

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/playground/internal/shared"
)

func BindAuthApi(api huma.API, appApi *Api) {
	huma.Register(
		api,
		huma.Operation{
			OperationID: "signup",
			Method:      http.MethodPost,
			Path:        "/auth/signup",
			Summary:     "Sign up",
			Description: "Count the number of colors for all themes",
			Tags:        []string{"Auth"},
			Errors:      []int{http.StatusNotFound},
		},
		appApi.SignUp,
	)
	// signin -------------------------------------------------------------
	huma.Register(
		api,
		huma.Operation{
			OperationID: "signin",
			Method:      http.MethodPost,
			Path:        "/auth/signin",
			Summary:     "Sign in",
			Description: "Count the number of colors for all themes",
			Tags:        []string{"Auth"},
			Errors:      []int{http.StatusNotFound},
		},
		appApi.SignIn,
	)
	//  me get ---------------------------------------------------------------
	huma.Register(
		api,
		huma.Operation{
			OperationID: "me",
			Method:      http.MethodGet,
			Path:        "/auth/me",
			Summary:     "Me",
			Description: "Me",
			Tags:        []string{"Auth"},
			Errors:      []int{http.StatusUnauthorized, http.StatusNotFound},
			Security: []map[string][]string{
				{shared.BearerAuthSecurityKey: {}},
			},
		},
		appApi.Me,
	)
	// me update -------------------------------------------------------------
	huma.Register(
		api,
		huma.Operation{
			OperationID: "meUpdate",
			Method:      http.MethodPut,
			Path:        "/auth/me",
			Summary:     "Me Update",
			Description: "Me Update",
			Tags:        []string{"Auth"},
			Errors:      []int{http.StatusUnauthorized, http.StatusNotFound},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
		},
		appApi.MeUpdate,
	)
	// me delete -------------------------------------------------------------
	huma.Register(
		api,
		huma.Operation{
			OperationID: "me-delete",
			Method:      http.MethodDelete,
			Path:        "/auth/me",
			Summary:     "Me delete",
			Description: "Me delete",
			Tags:        []string{"Auth", "Me"},
			Errors:      []int{http.StatusUnauthorized, http.StatusNotFound},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
		},
		appApi.MeDelete,
	)
	// refresh token -------------------------------------------------------------
	huma.Register(
		api,
		huma.Operation{
			OperationID: "refresh-token",
			Method:      http.MethodPost,
			Path:        "/auth/refresh-token",
			Summary:     "Refresh token",
			Description: "Count the number of colors for all themes",
			Tags:        []string{"Auth"},
			Errors:      []int{http.StatusNotFound},
		},
		appApi.RefreshToken,
	)
	// signout -------------------------------------------------------------
	huma.Register(
		api,
		huma.Operation{
			OperationID: "signout",
			Method:      http.MethodPost,
			Path:        "/auth/signout",
			Summary:     "Signout",
			Description: "Signout",
			Tags:        []string{"Auth"},
			Errors:      []int{http.StatusUnauthorized, http.StatusNotFound},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
		},
		appApi.Signout,
	)
	// verify email -------------------------------------------------------------
	huma.Register(
		api,
		huma.Operation{
			OperationID: "verify-get",
			Method:      http.MethodGet,
			Path:        "/auth/verify",
			Summary:     "Verify",
			Description: "Verify",
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
			Method:      http.MethodGet,
			Path:        "/auth/check-password-reset",
			Summary:     "Check password reset",
			Description: "Check password reset",
			Tags:        []string{"Auth"},
			Errors:      []int{http.StatusNotFound},
		},
		appApi.CheckPasswordResetGet,
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
