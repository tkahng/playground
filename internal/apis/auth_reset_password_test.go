package apis_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/tkahng/playground/internal/apis"
	"github.com/tkahng/playground/internal/auth"
	"github.com/tkahng/playground/internal/core"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/test"
	"github.com/tkahng/playground/internal/tools/mailer"
)

func TestApi_ResetPassword(t *testing.T) {
	// t.Parallel()
	test.SkipIfShort(t)
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := SetupApi(t, ctx, db)
		userInfo := CreateUserWithOptions(t, testApi.App, UserWithPassword("Password123!"))
		var testMailer *mailer.TestMailer
		if m, ok := testApi.App.Mailer().(*mailer.TestMailer); ok {
			testMailer = m
		} else {
			t.Fatal("mailer is not a TestMailer")
		}
		tests := []ApiScenario{
			{
				Name:           "Test reset password request success",
				Method:         http.MethodPost,
				URL:            "/auth/request-password-reset",
				ExpectedStatus: http.StatusNoContent,
				TestAppFactory: func(t testing.TB) *TestApi {
					return testApi
				},
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario) {
					dto := apis.RequestPasswordResetInput{
						Email: userInfo.User.Email,
					}
					data, err := json.Marshal(dto)
					if err != nil {
						t.Errorf("Error marshalling input: %v", err)
					}
					scenario.Body = strings.NewReader(string(data))

				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario, res *httptest.ResponseRecorder) {
					err := app.JobManager().PollOnce(ctx)
					if err != nil {
						t.Fatalf("Error polling job: %v", err)
					}
					var message *mailer.Message
					if len(testMailer.Messages) > 0 {
						message = testMailer.Messages[0]
						testMailer.Messages = nil
						testMailer.Wg = nil
					} else {
						t.Fatalf("No message found for user")
					}
					stoken, err := test.GetLinkParam(message.Body, "token")
					if err != nil {
						t.Fatalf("Error getting token from email: %v", err)
					}
					if stoken == "" {
						t.Fatalf("No token found in email. Body: %s", message.Body)
					}
					err = app.Auth2().CheckPasswordResetToken(ctx, stoken)
					if err != nil {
						t.Fatalf("Error verifying token: %v, body: %s", err, message.Body)
					}
				},
			},
			{
				Name:           "Test reset password check success",
				Method:         http.MethodPost,
				URL:            "/auth/check-password-reset",
				ExpectedStatus: http.StatusNoContent,
				TestAppFactory: func(t testing.TB) *TestApi {
					return testApi
				},
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario) {
					checktoken, err := app.Token().GenerateToken(ctx, userInfo.User.Email, models.TokenTypesPasswordResetToken)
					if err != nil {
						t.Fatalf("Error getting token from email: %v", err)
					}
					dto := struct {
						Token string `json:"token"`
					}{
						Token: checktoken,
					}
					data, err := json.Marshal(dto)
					if err != nil {
						t.Fatalf("Error marshalling input: %v", err)
					}
					scenario.Body = strings.NewReader(string(data))
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario, res *httptest.ResponseRecorder) {

				},
			},
			{
				Name:           "Test reset password confirm success",
				Method:         http.MethodPost,
				URL:            "/auth/confirm-password-reset?token",
				ExpectedStatus: http.StatusNoContent,
				TestAppFactory: func(t testing.TB) *TestApi {
					return testApi
				},
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario) {
					checktoken, err := app.Token().GenerateToken(ctx, userInfo.User.Email, models.TokenTypesPasswordResetToken)
					if err != nil {
						t.Fatalf("Error getting token from email: %v", err)
					}
					dto := apis.ConfirmPasswordResetInput{
						Token:           checktoken,
						Password:        "SomePassword123!",
						ConfirmPassword: "SomePassword123!",
					}
					data, err := json.Marshal(dto)
					if err != nil {
						t.Fatalf("Error marshalling input: %v", err)
					}
					scenario.Body = strings.NewReader(string(data))

				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario, res *httptest.ResponseRecorder) {
					authTokens, err := app.Auth2().Signup(ctx, &auth.SignupInput{
						Email:    userInfo.User.Email,
						Password: "SomePassword123!",
					})
					if err != nil {
						t.Fatalf("Error authenticating user: %v", err)
					}
					if authTokens == nil {
						t.Fatalf("Error getting auth tokens")
					}
				},
			},
		}
		for _, tt := range tests {
			tt.Test(t)
		}
	})
}
