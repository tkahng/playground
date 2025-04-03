package core

import (
	"context"
	"log/slog"

	"github.com/stephenafamo/bob"
	"github.com/tkahng/authgo/internal/repository"
	"github.com/tkahng/authgo/internal/tools/payment"
)

type StripeService struct {
	logger *slog.Logger
	client *payment.StripeClient
}

func (srv *StripeService) Logger() *slog.Logger {
	return srv.logger
}

func (srv *StripeService) Client() *payment.StripeClient {
	return srv.client
}

func NewStripeService(client *payment.StripeClient) *StripeService {
	return &StripeService{client: client, logger: slog.Default()}
}
func (srv *StripeService) UpsertPriceProductFromStripe(ctx context.Context, exec bob.Executor) error {
	if err := srv.FindAndUpsertAllProducts(ctx, exec); err != nil {
		return err
	}
	if err := srv.FindAndUpsertAllPrices(ctx, exec); err != nil {
		return err
	}
	// if err := srv.FindAndUpsertAllCustomers(ctx, exec); err != nil {
	// 	return err
	// }
	// if err := srv.FindAndUpsertAllSubscriptions(ctx, exec); err != nil {
	// 	return err
	// }
	return nil
}

func (srv *StripeService) FindAndUpsertAllProducts(ctx context.Context, exec bob.Executor) error {
	products, err := srv.client.FindAllProducts()
	if err != nil {
		srv.logger.Error("error finding all products", "error", err)
		return err
	}
	for _, product := range products {
		err = repository.UpsertProductFromStripe(ctx, exec, product)
		if err != nil {
			srv.logger.Error("error upserting product", "product", product.ID, "error", err)
			continue
		}
	}
	return nil
}

func (srv *StripeService) FindAndUpsertAllPrices(ctx context.Context, exec bob.Executor) error {
	prices, err := srv.client.FindAllPrices()
	if err != nil {
		srv.logger.Error("error finding all prices", "error", err)
		return err
	}
	for _, price := range prices {
		err = repository.UpsertPriceFromStripe(ctx, exec, price)
		if err != nil {
			srv.logger.Error("error upserting price", "price", price.ID, "error", err)
			continue
		}
	}
	return nil
}
