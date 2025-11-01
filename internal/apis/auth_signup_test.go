package apis_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tkahng/playground/internal/apis"
	"github.com/tkahng/playground/internal/core"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/test"
	"github.com/tkahng/playground/internal/tools/mailer"
)

func TestApi_SignUp(t *testing.T) {
	// t.Parallel()
	test.SkipIfShort(t)
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := SetupApi(t, ctx, db)
		testMailer := core.ExtractTestMailer(t, testApi.App)

		tests := []ApiScenario{
			{
				Name:           "Test signup success",
				Method:         http.MethodPost,
				URL:            "/auth/signup",
				ExpectedStatus: http.StatusOK,
				TestAppFactory: func(t testing.TB) *TestApi {
					return testApi
				},
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario) {
					dto := apis.SignupInput{
						Email:    "test@example.com",
						Password: "Password123!",
					}
					data, err := json.Marshal(dto)
					if err != nil {
						t.Errorf("Error marshalling input: %v", err)
					}
					scenario.Body = strings.NewReader(string(data))
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario, res *httptest.ResponseRecorder) {
					if err := app.JobManager().PollOnce(context.Background()); err != nil {
						t.Fatalf("Error polling job manager: %v", err)
					}
					// testMailer.Wg.Wait()
					var body apis.ApiOutput[*apis.ApiUserInfoTokens]
					err := json.NewDecoder(res.Body).Decode(&body)
					if err != nil {
						t.Errorf("Error decoding response: %v", err)
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

				},
			},
		}
		for _, tt := range tests {
			tt.Test(t)
		}
	})
}

func TestApi_SignUp_ExistingUsers(t *testing.T) {
	// t.Parallel()
	test.SkipIfShort(t)
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := SetupApi(t, ctx, db)
		// testMailer := ExtractTestMailer(t, testApi.App)
		tests := []ApiScenario{
			{
				Name:           "Test signup fail for existing user",
				Method:         http.MethodPost,
				URL:            "/auth/signup",
				ExpectedStatus: http.StatusConflict,
				TestAppFactory: func(t testing.TB) *TestApi {
					return testApi
				},
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario) {
					existingUser := CreateUserWithOptions(
						t,
						testApi.App,
						UserWithPassword("Password123!"),
						UserWithEmail("existing1@example.com"),
					)
					assert.NotNil(t, existingUser, "User should not be nil")
					dto := apis.SignupInput{
						Email:    existingUser.User.Email,
						Password: "Password123!",
					}
					data, err := json.Marshal(dto)
					if err != nil {
						t.Errorf("Error marshalling input: %v", err)
					}
					scenario.Body = strings.NewReader(string(data))
				},
				ExpectedContent: []string{
					"user already exists",
				},
			},
		}
		for _, tt := range tests {
			tt.Test(t)
		}
	})
}
