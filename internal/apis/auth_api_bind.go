package apis

import (
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/go-chi/httprate"
	"github.com/tkahng/playground/internal/middleware/humamiddleware"
	apphttp "github.com/tkahng/playground/internal/tools/http"
)

// newAuthRateLimiter returns a per-IP huma middleware that allows at most
// limit requests per window. Exceeding it returns 429.
func newAuthRateLimiter(limit int, window time.Duration) func(ctx huma.Context, next func(huma.Context)) {
	return humamiddleware.HumaChiMiddleware(httprate.Limit(
		limit,
		window,
		httprate.WithKeyFuncs(httprate.KeyByRealIP),
		httprate.WithLimitHandler(func(w http.ResponseWriter, r *http.Request) {
			apphttp.WriteErr(w, r, http.StatusTooManyRequests, "too many requests")
		}),
	))
}

func bindAuthApi(appApi *Api) {
	api := appApi.Api()

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
	appApi.RequestVerificationBind(api)
	// confirm verification -------------------------------------------------------------
	appApi.VerifyEmailBind(api)
	// confirm verification otp -------------------------------------------------------------
	appApi.VerifyEmailOtpBind(api)
	// request password reset -------------------------------------------------------------
	appApi.bindRequestPasswordReset(api)
	// confirm password reset -------------------------------------------------------------
	appApi.bindConfirmPasswordReset(api)
	// check password reset -------------------------------------------------------------
	appApi.bindCheckPasswordReset(api)
	// password reset
	appApi.bindResetPassword(api)
	// oauth2 callback get
	appApi.bindOath2CallbackGet(api)
	// oauth2 callback post
	appApi.bindOAuth2CallbackPost(api)
	// oauth2 authorization url
	appApi.OAuth2AuthorizationUrlBind(api)
}
