package apis_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	stripe "github.com/stripe/stripe-go/v82"
	"github.com/stretchr/testify/assert"
	"github.com/tkahng/playground/internal/apis"
	"github.com/tkahng/playground/internal/core"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/services"
	"github.com/tkahng/playground/internal/stores"
	"github.com/tkahng/playground/internal/test"
)

func TestLedgerApi_GetBalance_Unauthorized(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		scenario := &apis.ApiScenario{
			Name:            "unauthorized",
			Method:          http.MethodGet,
			URL:             "/ledger/balance",
			ExpectedStatus:  http.StatusUnauthorized,
			ExpectedContent: []string{"Unauthorized"},
			TestAppFactory:  func(t testing.TB) *apis.TestApi { return testApi },
		}
		scenario.Test(t)
	})
}

func TestLedgerApi_GetBalance_Authorized_NoWallet(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		scenario := &apis.ApiScenario{
			Name:           "authorized no wallet",
			Method:         http.MethodGet,
			URL:            "/ledger/balance",
			ExpectedStatus: http.StatusOK,
			TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				user := core.CreateUserWithOptions(t, app, core.UserWithEmail("balance_nowal@example.com"))
				scenario.Headers = []string{core.CreateTokenHeader(t, app, user.User.Email)}
			},
			AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
				type body struct {
					Balance          int64 `json:"balance"`
					AvailableBalance int64 `json:"available_balance"`
				}
				b := test.MustUnMarshal[body](t, res.Body.Bytes())
				assert.Equal(t, int64(0), b.Balance)
				assert.Equal(t, int64(0), b.AvailableBalance)
			},
		}
		scenario.Test(t)
	})
}

func TestLedgerApi_GetBalance_Authorized_WithFunds(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		scenario := &apis.ApiScenario{
			Name:           "authorized with funds",
			Method:         http.MethodGet,
			URL:            "/ledger/balance",
			ExpectedStatus: http.StatusOK,
			TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				user := core.CreateUserWithOptions(t, app, core.UserWithEmail("balance_funded@example.com"))
				scenario.Headers = []string{core.CreateTokenHeader(t, app, user.User.Email)}

				adapter := stores.NewDbAdapterDecorators(db)
				ledger := services.NewDbLedgerService(adapter)
				if err := services.FulfillPointsPurchase(ctx, adapter, ledger, services.PointsPurchaseFulfillInput{
					UserID:          user.User.ID,
					PointsAmount:    250,
					StripeSessionID: "cs_balance_funded",
				}); err != nil {
					t.Fatalf("FulfillPointsPurchase() error = %v", err)
				}
			},
			AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
				type body struct {
					Balance          int64 `json:"balance"`
					AvailableBalance int64 `json:"available_balance"`
				}
				b := test.MustUnMarshal[body](t, res.Body.Bytes())
				assert.Equal(t, int64(250), b.Balance)
				assert.Equal(t, int64(250), b.AvailableBalance)
			},
		}
		scenario.Test(t)
	})
}

func TestLedgerApi_GetTransactions_Unauthorized(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		scenario := &apis.ApiScenario{
			Name:            "unauthorized",
			Method:          http.MethodGet,
			URL:             "/ledger/transactions",
			ExpectedStatus:  http.StatusUnauthorized,
			ExpectedContent: []string{"Unauthorized"},
			TestAppFactory:  func(t testing.TB) *apis.TestApi { return testApi },
		}
		scenario.Test(t)
	})
}

func TestLedgerApi_GetTransactions_NoWallet(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		scenario := &apis.ApiScenario{
			Name:           "no wallet returns empty list",
			Method:         http.MethodGet,
			URL:            "/ledger/transactions",
			ExpectedStatus: http.StatusOK,
			TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				user := core.CreateUserWithOptions(t, app, core.UserWithEmail("txns_nowal@example.com"))
				scenario.Headers = []string{core.CreateTokenHeader(t, app, user.User.Email)}
			},
			AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
				result := test.MustUnMarshal[apis.ApiPaginatedResponse[any]](t, res.Body.Bytes())
				assert.Empty(t, result.Data)
			},
		}
		scenario.Test(t)
	})
}

func TestLedgerApi_GetTransactions_WithTransfers(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		scenario := &apis.ApiScenario{
			Name:           "returns transfers for wallet owner",
			Method:         http.MethodGet,
			URL:            "/ledger/transactions",
			ExpectedStatus: http.StatusOK,
			TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				user := core.CreateUserWithOptions(t, app, core.UserWithEmail("txns_funded@example.com"))
				scenario.Headers = []string{core.CreateTokenHeader(t, app, user.User.Email)}

				adapter := stores.NewDbAdapterDecorators(db)
				ledger := services.NewDbLedgerService(adapter)
				for _, sessionID := range []string{"cs_txns_1", "cs_txns_2"} {
					if err := services.FulfillPointsPurchase(ctx, adapter, ledger, services.PointsPurchaseFulfillInput{
						UserID:          user.User.ID,
						PointsAmount:    100,
						StripeSessionID: sessionID,
					}); err != nil {
						t.Fatalf("FulfillPointsPurchase() error = %v", err)
					}
				}
			},
			AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
				result := test.MustUnMarshal[apis.ApiPaginatedResponse[any]](t, res.Body.Bytes())
				assert.Equal(t, int64(2), result.Meta.Total)
				assert.Len(t, result.Data, 2)
			},
		}
		scenario.Test(t)
	})
}

func TestLedgerApi_CreateWallet_Unauthorized(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		scenario := &apis.ApiScenario{
			Name:            "unauthorized",
			Method:          http.MethodPost,
			URL:             "/ledger/wallet",
			ExpectedStatus:  http.StatusUnauthorized,
			ExpectedContent: []string{"Unauthorized"},
			TestAppFactory:  func(t testing.TB) *apis.TestApi { return testApi },
		}
		scenario.Test(t)
	})
}

func TestLedgerApi_CreateWallet_CreatesAndIsIdempotent(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		user := core.CreateUserWithOptions(t, testApi.App, core.UserWithEmail("wallet_create@example.com"))
		authHeader := core.CreateTokenHeader(t, testApi.App, user.User.Email)

		// First call — creates the wallet.
		first := &apis.ApiScenario{
			Name:            "creates wallet",
			Method:          http.MethodPost,
			URL:             "/ledger/wallet",
			ExpectedStatus:  http.StatusOK,
			ExpectedContent: []string{"ledger_code"},
			TestAppFactory:  func(t testing.TB) *apis.TestApi { return testApi },
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				scenario.Headers = []string{authHeader}
			},
		}
		first.Test(t)

		// Second call — idempotent, returns same wallet.
		second := &apis.ApiScenario{
			Name:            "idempotent second call",
			Method:          http.MethodPost,
			URL:             "/ledger/wallet",
			ExpectedStatus:  http.StatusOK,
			ExpectedContent: []string{"ledger_code"},
			TestAppFactory:  func(t testing.TB) *apis.TestApi { return testApi },
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				scenario.Headers = []string{authHeader}
			},
		}
		second.Test(t)
	})
}

func TestLedgerApi_CreatePointsCheckout_Unauthorized(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		scenario := &apis.ApiScenario{
			Name:            "unauthorized",
			Method:          http.MethodPost,
			URL:             "/ledger/points/checkout",
			Body:            strings.NewReader(`{"price_id":"price_abc"}`),
			ExpectedStatus:  http.StatusUnauthorized,
			ExpectedContent: []string{"Unauthorized"},
			TestAppFactory:  func(t testing.TB) *apis.TestApi { return testApi },
		}
		scenario.Test(t)
	})
}

func TestLedgerApi_CreatePointsCheckout_Authorized_ReturnsURL(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		scenario := &apis.ApiScenario{
			Name:           "returns checkout URL",
			Method:         http.MethodPost,
			URL:            "/ledger/points/checkout",
			Body:           strings.NewReader(`{"price_id":"price_points_100"}`),
			ExpectedStatus: http.StatusOK,
			TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				// Seed points product + price required by CreatePointsCheckoutSession.
				adapter := stores.NewDbAdapterDecorators(db)
				if err := adapter.Product().UpsertProduct(ctx, &models.StripeProduct{
					ID:       services.PointsProduct.ID,
					Name:     services.PointsProduct.Name,
					Active:   services.PointsProduct.Active,
					Metadata: map[string]string{},
				}); err != nil {
					t.Fatalf("UpsertProduct: %v", err)
				}
				if err := adapter.Price().UpsertPrice(ctx, &models.StripePrice{
					ID:        services.PointsPrice100.ID,
					ProductID: services.PointsProduct.ID,
					Active:    true,
					Type:      models.StripePricingTypeOneTime,
					Currency:  "usd",
					Metadata:  map[string]string{"points_amount": "100", models.StripeProductTypeMetadataKey: string(models.StripeProductTypePoints)},
				}); err != nil {
					t.Fatalf("UpsertPrice: %v", err)
				}

				// Configure the mock Stripe client to return a fake checkout URL.
				mockClient := core.ExtractTestPaymentClient(t, app)
				mockClient.CreatePointsCheckoutSessionFunc = func(customerID, userID string, pointsAmount int64, priceID string) (*stripe.CheckoutSession, error) {
					return &stripe.CheckoutSession{URL: "https://checkout.stripe.com/pay/cs_test_mock"}, nil
				}

				// Need a verified user so CreateUserCustomer is available.
				user := core.CreateUserWithOptions(t, app,
					core.UserWithEmail("checkout@example.com"),
					core.UserWithVerifiedNow(),
				)
				scenario.Headers = []string{core.CreateTokenHeader(t, app, user.User.Email)}
				scenario.URL = "/ledger/points/checkout"
				scenario.Body = strings.NewReader(`{"price_id":"` + services.PointsPrice100.ID + `"}`)
			},
			AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
				type body struct {
					URL string `json:"url"`
				}
				b := test.MustUnMarshal[body](t, res.Body.Bytes())
				assert.NotEmpty(t, b.URL)
			},
		}
		scenario.Test(t)
	})
}
