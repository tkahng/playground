package apis_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/tkahng/playground/internal/apis"
	"github.com/tkahng/playground/internal/core"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/test"
	"github.com/tkahng/playground/internal/tools/mailer"
)

func TestApi_SignUp(t *testing.T) {
	test.Parallel(t)
	test.SkipIfShort(t)
	test.WithTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := SetupApi(t, ctx, db)
		// testApi.App.RunBackgroundProcesses(ctx)
		var testMailer *mailer.TestMailer
		if m, ok := testApi.App.Mailer().(*mailer.TestMailer); ok {
			testMailer = m
		} else {
			t.Fatal("mailer is not a TestMailer")
		}

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
					testMailer.Wg = &sync.WaitGroup{}
					testMailer.Wg.Add(1)
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, res *httptest.ResponseRecorder) {
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
