package contextstore

import (
	"context"

	"github.com/tkahng/playground/internal/models"
)

const (
	contextKeyCurrentCustomer contextKey = "current_customer"
)

func SetContextCurrentCustomer(ctx context.Context, customer *models.StripeCustomer) context.Context {
	return context.WithValue(ctx, contextKeyCurrentCustomer, customer)
}
func GetContextCurrentCustomer(ctx context.Context) *models.StripeCustomer {
	if customer, ok := ctx.Value(contextKeyCurrentCustomer).(*models.StripeCustomer); ok {
		return customer
	} else {
		return nil
	}
}
