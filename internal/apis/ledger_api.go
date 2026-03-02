package apis

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/contextstore"
	"github.com/tkahng/playground/internal/core"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/shared"
	"github.com/tkahng/playground/internal/stores"
)

type LedgerBalanceOutput struct {
	Body struct {
		Balance          int64 `json:"balance"`
		AvailableBalance int64 `json:"available_balance"`
	}
}

func bindGetLedgerBalanceApi(api huma.API, app core.App) {
	huma.Register(
		api,
		huma.Operation{
			OperationID: "get-ledger-balance",
			Method:      http.MethodGet,
			Path:        "/ledger/balance",
			Summary:     "get points balance",
			Description: "Returns the settled balance and available balance (settled minus pending holds) for the current user.",
			Tags:        []string{"Ledger"},
			Errors:      []int{http.StatusUnauthorized},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
		},
		func(ctx context.Context, _ *struct{}) (*LedgerBalanceOutput, error) {
			user := contextstore.GetContextUserInfo(ctx)
			if user == nil {
				return nil, huma.Error401Unauthorized("unauthorized")
			}
			balance, err := app.Ledger().GetUserBalance(ctx, user.User.ID)
			if err != nil {
				return nil, err
			}
			available, err := app.Ledger().GetUserAvailableBalance(ctx, user.User.ID)
			if err != nil {
				return nil, err
			}
			out := &LedgerBalanceOutput{}
			out.Body.Balance = balance
			out.Body.AvailableBalance = available
			return out, nil
		},
	)
}

type LedgerTransferFilter struct {
	PaginatedInput
	SortParams
	TransferCodes []string `query:"transfer_codes,omitempty" required:"false"`
}

func bindGetLedgerTransactionsApi(api huma.API, app core.App) {
	huma.Register(
		api,
		huma.Operation{
			OperationID: "get-ledger-transactions",
			Method:      http.MethodGet,
			Path:        "/ledger/transactions",
			Summary:     "get points transaction history",
			Description: "Returns a paginated list of ledger transfers for the current user.",
			Tags:        []string{"Ledger"},
			Errors:      []int{http.StatusUnauthorized},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
		},
		func(ctx context.Context, input *LedgerTransferFilter) (*ApiPaginatedOutput[*models.LedgerTransfer], error) {
			user := contextstore.GetContextUserInfo(ctx)
			if user == nil {
				return nil, huma.Error401Unauthorized("unauthorized")
			}
			wallet, err := app.Ledger().GetOrCreateUserWallet(ctx, user.User.ID)
			if err != nil {
				return nil, err
			}
			filter := &stores.LedgerTransferFilter{
				AccountIds:    []uuid.UUID{wallet.ID},
				TransferCodes: input.TransferCodes,
				PaginatedInput: stores.PaginatedInput{
					Page:    input.Page,
					PerPage: input.PerPage,
				},
			}
			transfers, err := app.Ledger().FindTransfers(ctx, filter)
			if err != nil {
				return nil, err
			}
			count, err := app.Ledger().CountTransfers(ctx, filter)
			if err != nil {
				return nil, err
			}
			return &ApiPaginatedOutput[*models.LedgerTransfer]{
				Body: ApiPaginatedResponse[*models.LedgerTransfer]{
					Data: transfers,
					Meta: ApiGenerateMeta(&input.PaginatedInput, count),
				},
			}, nil
		},
	)
}

type PointsCheckoutInput struct {
	PriceID string `json:"price_id" required:"true" doc:"Stripe price ID for the points package."`
}

func bindCreatePointsCheckoutApi(api huma.API, app core.App) {
	huma.Register(
		api,
		huma.Operation{
			OperationID: "create-points-checkout",
			Method:      http.MethodPost,
			Path:        "/ledger/points/checkout",
			Summary:     "create points purchase checkout",
			Description: "Creates a Stripe Checkout URL for purchasing points. The points are credited to the user's wallet after the Stripe webhook confirms payment.",
			Tags:        []string{"Ledger"},
			Errors:      []int{http.StatusUnauthorized, http.StatusBadRequest},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
		},
		func(ctx context.Context, input *struct {
			Body PointsCheckoutInput
		}) (*struct {
			Body struct {
				URL string `json:"url"`
			}
		}, error) {
			user := contextstore.GetContextUserInfo(ctx)
			if user == nil {
				return nil, huma.Error401Unauthorized("unauthorized")
			}
			// Find or create the user's Stripe customer.
			stripeCustomer, err := app.Payment().FindCustomerByUserId(ctx, user.User.ID)
			if err != nil {
				return nil, err
			}
			if stripeCustomer == nil {
				stripeCustomer, err = app.Payment().CreateUserCustomer(ctx, &user.User)
				if err != nil {
					return nil, err
				}
			}
			url, err := app.Payment().CreatePointsCheckoutSession(ctx, user.User.ID, stripeCustomer.ID, input.Body.PriceID)
			if err != nil {
				return nil, err
			}
			return &struct {
				Body struct {
					URL string `json:"url"`
				}
			}{Body: struct {
				URL string `json:"url"`
			}{URL: url}}, nil
		},
	)
}

func bindLedgerApi(api huma.API, app core.App) {
	bindGetLedgerBalanceApi(api, app)
	bindGetLedgerTransactionsApi(api, app)
	bindCreatePointsCheckoutApi(api, app)
}
