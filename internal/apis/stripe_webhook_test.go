package apis_test

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stripe/stripe-go/v82"
	"github.com/tkahng/playground/internal/apis"
	"github.com/tkahng/playground/internal/core"
	"github.com/tkahng/playground/internal/database"
)

const testWebhookSecret = "whsec_test_secret_for_unit_tests"

// stripeAPIVersion must match the version expected by the stripe-go library.
const stripeAPIVersion = "2025-08-27.basil"

// signWebhook builds a Stripe-Signature header value using the real
// HMAC-SHA256 scheme that webhook.ConstructEvent verifies.
func signWebhook(t testing.TB, payload []byte, secret string) string {
	t.Helper()
	ts := time.Now().Unix()
	mac := hmac.New(sha256.New, []byte(secret))
	fmt.Fprintf(mac, "%d.%s", ts, payload)
	sig := hex.EncodeToString(mac.Sum(nil))
	return fmt.Sprintf("t=%d,v1=%s", ts, sig)
}

// stripeEvent encodes a minimal but valid Stripe event envelope.
func stripeEvent(t testing.TB, eventType stripe.EventType, data any) []byte {
	t.Helper()
	raw, err := json.Marshal(data)
	require.NoError(t, err)
	event := map[string]any{
		"id":          "evt_test_" + uuid.New().String(),
		"object":      "event",
		"api_version": stripeAPIVersion,
		"type":        string(eventType),
		"data":        map[string]any{"object": data, "raw": json.RawMessage(raw)},
	}
	b, err := json.Marshal(event)
	require.NoError(t, err)
	return b
}

func setupWebhookTestApi(t testing.TB, ctx context.Context, db database.Dbx) *apis.TestApi {
	t.Helper()
	testApi := apis.SetupApi(t, ctx, db)
	testApi.Cfg.Webhook = testWebhookSecret
	return testApi
}

// --- signature validation ---

func TestStripeWebhook_MissingSignature_Returns400(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := setupWebhookTestApi(t, ctx, db)

		sc := &apis.ApiScenario{
			Name:            "missing Stripe-Signature header returns 400",
			Method:          http.MethodPost,
			URL:             "/stripe/webhook",
			ExpectedStatus:  http.StatusBadRequest,
			ExpectedContent: []string{"Missing signature header"},
			TestAppFactory:  func(t testing.TB) *apis.TestApi { return testApi },
			BeforeTestFunc: func(t testing.TB, _ *core.BaseApp, sc *apis.ApiScenario) {
				sc.Body = bytes.NewReader([]byte(`{}`))
			},
		}
		sc.Test(t)
	})
}

func TestStripeWebhook_InvalidSignature_Returns400(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := setupWebhookTestApi(t, ctx, db)
		payload := []byte(`{"id":"evt_1","type":"product.created","data":{"object":{}}}`)

		sc := &apis.ApiScenario{
			Name:            "wrong signature returns 400",
			Method:          http.MethodPost,
			URL:             "/stripe/webhook",
			ExpectedStatus:  http.StatusBadRequest,
			ExpectedContent: []string{"webhook signature verification failed"},
			TestAppFactory:  func(t testing.TB) *apis.TestApi { return testApi },
			BeforeTestFunc: func(t testing.TB, _ *core.BaseApp, sc *apis.ApiScenario) {
				sc.Headers = []string{"Stripe-Signature: t=1,v1=invalidsig"}
				sc.Body = bytes.NewReader(payload)
			},
		}
		sc.Test(t)
	})
}

// --- unknown event type ---

func TestStripeWebhook_UnknownEventType_Returns204(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := setupWebhookTestApi(t, ctx, db)
		payload := stripeEvent(t, "unknown.event.type", map[string]any{})
		sig := signWebhook(t, payload, testWebhookSecret)

		sc := &apis.ApiScenario{
			Name:           "unknown event type is gracefully ignored",
			Method:         http.MethodPost,
			URL:            "/stripe/webhook",
			ExpectedStatus: http.StatusNoContent,
			TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
			BeforeTestFunc: func(t testing.TB, _ *core.BaseApp, sc *apis.ApiScenario) {
				sc.Headers = []string{"Stripe-Signature: " + sig}
				sc.Body = bytes.NewReader(payload)
			},
		}
		sc.Test(t)
	})
}

// --- checkout.session.completed / payment mode ---

// TestStripeWebhook_PurchaseType_Missing_Returns204 verifies that a
// payment-mode session with no purchase_type key in metadata is silently
// accepted — this is the ok-check fix.
func TestStripeWebhook_PurchaseType_Missing_Returns204(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := setupWebhookTestApi(t, ctx, db)

		cs := stripe.CheckoutSession{
			ID:       "cs_test_missing",
			Mode:     stripe.CheckoutSessionModePayment,
			Metadata: map[string]string{}, // no purchase_type key
		}
		payload := stripeEvent(t, stripe.EventTypeCheckoutSessionCompleted, cs)
		sig := signWebhook(t, payload, testWebhookSecret)

		sc := &apis.ApiScenario{
			Name:           "missing purchase_type key is silently accepted",
			Method:         http.MethodPost,
			URL:            "/stripe/webhook",
			ExpectedStatus: http.StatusNoContent,
			TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
			BeforeTestFunc: func(t testing.TB, _ *core.BaseApp, sc *apis.ApiScenario) {
				sc.Headers = []string{"Stripe-Signature: " + sig}
				sc.Body = bytes.NewReader(payload)
			},
		}
		sc.Test(t)
	})
}

// TestStripeWebhook_PurchaseType_WrongValue_Returns204 verifies that a
// non-"points" purchase_type is silently accepted (no error).
func TestStripeWebhook_PurchaseType_WrongValue_Returns204(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := setupWebhookTestApi(t, ctx, db)

		cs := stripe.CheckoutSession{
			ID:   "cs_test_wrong_type",
			Mode: stripe.CheckoutSessionModePayment,
			Metadata: map[string]string{
				"purchase_type": "credits", // not "points"
			},
		}
		payload := stripeEvent(t, stripe.EventTypeCheckoutSessionCompleted, cs)
		sig := signWebhook(t, payload, testWebhookSecret)

		sc := &apis.ApiScenario{
			Name:           "unrecognised purchase_type is silently accepted",
			Method:         http.MethodPost,
			URL:            "/stripe/webhook",
			ExpectedStatus: http.StatusNoContent,
			TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
			BeforeTestFunc: func(t testing.TB, _ *core.BaseApp, sc *apis.ApiScenario) {
				sc.Headers = []string{"Stripe-Signature: " + sig}
				sc.Body = bytes.NewReader(payload)
			},
		}
		sc.Test(t)
	})
}

// TestStripeWebhook_PointsPurchase_FulfillsWallet verifies the full
// purchase_type=points path: a valid checkout session credits the user's wallet.
func TestStripeWebhook_PointsPurchase_FulfillsWallet(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := setupWebhookTestApi(t, ctx, db)
		owner := core.CreateUserWithOptions(t, testApi.App, core.UserWithVerifiedNow())

		pointsAmount := int64(500)
		cs := stripe.CheckoutSession{
			ID:   "cs_test_points_" + uuid.New().String(),
			Mode: stripe.CheckoutSessionModePayment,
			Metadata: map[string]string{
				"purchase_type": "points",
				"user_id":       owner.User.ID.String(),
				"points_amount": strconv.FormatInt(pointsAmount, 10),
			},
		}
		payload := stripeEvent(t, stripe.EventTypeCheckoutSessionCompleted, cs)
		sig := signWebhook(t, payload, testWebhookSecret)

		sc := &apis.ApiScenario{
			Name:           "purchase_type=points fulfills the user wallet",
			Method:         http.MethodPost,
			URL:            "/stripe/webhook",
			ExpectedStatus: http.StatusNoContent,
			TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
			BeforeTestFunc: func(t testing.TB, _ *core.BaseApp, sc *apis.ApiScenario) {
				sc.Headers = []string{"Stripe-Signature: " + sig}
				sc.Body = bytes.NewReader(payload)
			},
			AfterTestFunc: func(t testing.TB, app *core.BaseApp, _ *apis.ApiScenario, _ *httptest.ResponseRecorder) {
				balance, err := app.Ledger().GetUserBalance(context.Background(), owner.User.ID)
				require.NoError(t, err)
				assert.Equal(t, pointsAmount, balance,
					"wallet balance should equal the purchased points amount")
			},
		}
		sc.Test(t)
	})
}

// TestStripeWebhook_PointsPurchase_Idempotent verifies that replaying the same
// checkout session does not double-credit the wallet.
func TestStripeWebhook_PointsPurchase_Idempotent(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := setupWebhookTestApi(t, ctx, db)
		owner := core.CreateUserWithOptions(t, testApi.App, core.UserWithVerifiedNow())

		pointsAmount := int64(100)
		cs := stripe.CheckoutSession{
			ID:   "cs_test_idempotent_" + uuid.New().String(),
			Mode: stripe.CheckoutSessionModePayment,
			Metadata: map[string]string{
				"purchase_type": "points",
				"user_id":       owner.User.ID.String(),
				"points_amount": strconv.FormatInt(pointsAmount, 10),
			},
		}

		send := func() {
			payload := stripeEvent(t, stripe.EventTypeCheckoutSessionCompleted, cs)
			sig := signWebhook(t, payload, testWebhookSecret)
			sc := &apis.ApiScenario{
				Name:           "replay",
				Method:         http.MethodPost,
				URL:            "/stripe/webhook",
				ExpectedStatus: http.StatusNoContent,
				TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
				BeforeTestFunc: func(t testing.TB, _ *core.BaseApp, sc *apis.ApiScenario) {
					sc.Headers = []string{"Stripe-Signature: " + sig}
					sc.Body = bytes.NewReader(payload)
				},
			}
			sc.Test(t)
		}

		send()
		send() // replay — must not double-credit

		balance, err := testApi.App.Ledger().GetUserBalance(context.Background(), owner.User.ID)
		require.NoError(t, err)
		assert.Equal(t, pointsAmount, balance,
			"replaying the same session must not double-credit the wallet")
	})
}
