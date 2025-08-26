package apis_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/tkahng/playground/internal/apis"
	"github.com/tkahng/playground/internal/core"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/test"
	"github.com/tkahng/playground/internal/tools/mailer"
)

func TestApi_SignIn(t *testing.T) {
	test.Parallel(t)
	test.SkipIfShort(t)
	test.WithTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := SetupApi(t, ctx, db)
		var testMailer *mailer.TestMailer
		if m, ok := testApi.App.Mailer().(*mailer.TestMailer); ok {
			testMailer = m
		} else {
			t.Fatal("mailer is not a TestMailer")
		}

		tests := []ApiScenario{
			{
				Name:           "Test signin success",
				Method:         http.MethodPost,
				URL:            "/auth/signin",
				ExpectedStatus: http.StatusOK,
				TestAppFactory: func(t testing.TB) *TestApi {
					return testApi
				},
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario) {
					testMailer.Wg = &sync.WaitGroup{}
					testMailer.Wg.Add(1)
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, res *httptest.ResponseRecorder) {
					testMailer.Wg.Wait()
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
