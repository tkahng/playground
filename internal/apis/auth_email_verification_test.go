package apis_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tkahng/playground/internal/apis"
	"github.com/tkahng/playground/internal/auth"
	"github.com/tkahng/playground/internal/core"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/database/repository"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/test"
	"github.com/tkahng/playground/internal/tools/mailer"
)

func TestApi_RequestVerification(t *testing.T) {
	// t.Parallel()
	test.SkipIfShort(t)
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := SetupApi(t, ctx, db)
		testMailer := ExtractTestMailer(t, testApi.App)

		tests := []ApiScenario{
			{
				Name:           "Test request verification fail",
				Method:         http.MethodPost,
				URL:            "/auth/request-verification",
				ExpectedStatus: http.StatusConflict,
				TestAppFactory: func(t testing.TB) *TestApi {
					return testApi
				},
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario) {
					userInfo := CreateUserWithOptions(
						t,
						testApi.App,
						UserWithPassword("Password123!"),
						UserWithEmail("test2@example.com"),
						UserWithVerified(time.Now()),
					)
					header := createTokenHeader(t, testApi.App, userInfo.User.Email)
					scenario.Headers = append(scenario.Headers, header)
				},
				ExpectedContent: []string{
					"Email already verified",
				},
			},
			{
				Name:           "Test request verification success",
				Method:         http.MethodPost,
				URL:            "/auth/request-verification",
				ExpectedStatus: http.StatusNoContent,
				TestAppFactory: func(t testing.TB) *TestApi {
					return testApi
				},
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario) {
					userInfo := CreateUserWithOptions(
						t,
						testApi.App,
						UserWithPassword("Password123!"),
						UserWithEmail("test1@example.com"),
					)
					header := createTokenHeader(t, testApi.App, userInfo.User.Email)
					scenario.Headers = append(scenario.Headers, header)

				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario, res *httptest.ResponseRecorder) {
					if err := app.JobManager().PollOnce(context.Background()); err != nil {
						t.Fatalf("Error polling job manager: %v", err)
					}
					var message *mailer.Message
					if len(testMailer.Messages) > 0 {
						message = testMailer.Messages[0]
					} else {
						t.Fatalf("No message found for user")
					}
					token, err := test.GetLinkParam(message.Body, "token")
					if err != nil {
						t.Fatalf("Error getting token from email: %v", err)
					}
					if token == "" {
						t.Fatalf("No token found in email. Body: %s", message.Body)
					}
					err = app.Token().CheckToken(ctx, token, models.TokenTypesVerificationToken)
					if err != nil {
						t.Fatalf("Error checking token email: %v", err)
					}
				},
			},
		}
		for _, tt := range tests {
			tt.Test(t)
		}
	})
}

func TestApi_VerifyEmail(t *testing.T) {
	// t.Parallel()
	test.SkipIfShort(t)
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := SetupApi(t, ctx, db)
		testMailer := ExtractTestMailer(t, testApi.App)
		// var customerStore *stores.CustomerStoreDecorator
		// if m, ok := testApi.App.Adapter().Customer().(*stores.CustomerStoreDecorator); ok {
		// 	customerStore = m
		// } else {
		// 	t.Fatal("mailer is not a TestMailer")
		// }
		tests := []ApiScenario{
			{
				Name:           "Test confirm email verification success",
				Method:         http.MethodPost,
				URL:            "/auth/confirm-verification",
				ExpectedStatus: http.StatusNoContent,
				TestAppFactory: func(t testing.TB) *TestApi {
					return testApi
				},
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario) {
					userInfo, err := app.Auth2().Signup(ctx, &auth.SignupInput{
						Email:    "test@example.com",
						Password: "Password123!",
					})
					if err != nil {
						t.Fatalf("Error signing up user: %v", err)
					}
					if err := app.JobManager().PollOnce(context.Background()); err != nil {
						t.Fatalf("Error polling job manager: %v", err)
					}
					var message *mailer.Message
					if len(testMailer.Messages) > 0 {
						message = testMailer.Messages[0]
					} else {
						t.Fatalf("No message found for user")
					}
					token, err := test.GetLinkParam(message.Body, "token")
					if err != nil {
						t.Fatalf("Error getting token from email: %v", err)
					}
					if token == "" {
						t.Fatalf("No token found in email. Body: %s", message.Body)
					}
					header := createTokenHeader(t, testApi.App, userInfo.User.Email)
					scenario.Headers = append(scenario.Headers, header)
					dto := apis.EmailVerificationPostInput{
						Token: token,
					}
					data, err := json.Marshal(dto)
					if err != nil {
						t.Errorf("Error marshalling input: %v", err)
					}
					scenario.Body = strings.NewReader(string(data))
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario, res *httptest.ResponseRecorder) {
					dbx := app.Db()
					customerCount := repository.MustCountAllCtx(t, ctx, repository.StripeCustomer, dbx, nil)
					assert.Equal(t, 1, int(customerCount))
				},
			},
		}
		for _, tt := range tests {
			tt.Test(t)
		}
	})
}
