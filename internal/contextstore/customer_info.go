package contextstore

import (
	"context"

	"github.com/tkahng/authgo/internal/models"
)

const (
	ContextKeyCurrentCustomer ContextKey = "current_customer"
)

func SetContextCurrentCustomer(ctx context.Context, customer *models.StripeCustomer) context.Context {
	return context.WithValue(ctx, ContextKeyCurrentCustomer, customer)
}
func GetContextCurrentCustomer(ctx context.Context) *models.StripeCustomer {
	if customer, ok := ctx.Value(ContextKeyCurrentCustomer).(*models.StripeCustomer); ok {
		return customer
	} else {
		return nil
	}
}
