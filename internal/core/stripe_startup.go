package core

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/stephenafamo/bob"
	"github.com/tkahng/authgo/internal/conf"
	"github.com/tkahng/authgo/internal/queries"
	"github.com/tkahng/authgo/internal/shared"
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
func NewStripeServiceFromConf(conf conf.StripeConfig) *StripeService {
	return &StripeService{client: payment.NewStripeClient(conf), logger: slog.Default()}
}
func NewStripeService(client *payment.StripeClient) *StripeService {
	return &StripeService{client: client, logger: slog.Default()}
}

func (srv *StripeService) SyncRoles(ctx context.Context, exec bob.Executor) error {
	var err error
	for productId, role := range shared.StripeRoleMap {
		err = srv.SyncProductRole(ctx, exec, productId, role)
	}
	return err
}

func (srv *StripeService) SyncProductRole(ctx context.Context, exec bob.Executor, productId string, roleName string) error {
	product, err := queries.FindProductByStripeId(ctx, exec, productId)
	if err != nil {
		return err
	}
	if product == nil {
		return errors.New("product not found")
	}
	role, err := queries.FindRoleByName(ctx, exec, roleName)
	if err != nil {
		return err
	}
	if role == nil {
		return errors.New("role not found")
	}
	return product.AttachRoles(ctx, exec, role)
}

func (srv *StripeService) UpsertPriceProductFromStripe(ctx context.Context, exec bob.Executor) error {
	if err := srv.FindAndUpsertAllProducts(ctx, exec); err != nil {
		fmt.Println(err)
		return err
	}
	if err := srv.FindAndUpsertAllPrices(ctx, exec); err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func (srv *StripeService) FindAndUpsertAllProducts(ctx context.Context, exec bob.Executor) error {
	products, err := srv.client.FindAllProducts()
	if err != nil {
		srv.logger.Error("error finding all products", "error", err)
		return err
	}
	for _, product := range products {
		err = queries.UpsertProductFromStripe(ctx, exec, product)
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
		err = queries.UpsertPriceFromStripe(ctx, exec, price)
		if err != nil {
			srv.logger.Error("error upserting price", "price", price.ID, "error", err)
			continue
		}
	}
	return nil
}
