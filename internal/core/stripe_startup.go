package core

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/tkahng/authgo/internal/conf"
	"github.com/tkahng/authgo/internal/db"
	"github.com/tkahng/authgo/internal/queries"
	"github.com/tkahng/authgo/internal/repository"
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

func (srv *StripeService) SyncRoles(ctx context.Context, dbx db.Dbx) error {
	var err error
	for productId, role := range shared.StripeRoleMap {
		err = srv.SyncProductRole(ctx, dbx, productId, role)
	}
	return err
}

func (srv *StripeService) SyncProductRole(ctx context.Context, dbx db.Dbx, productId string, roleName string) error {
	product, err := queries.FindProductByStripeId(ctx, dbx, productId)
	if err != nil {
		return err
	}
	if product == nil {
		return errors.New("product not found")
	}
	role, err := repository.Role.GetOne(
		ctx,
		dbx,
		&map[string]any{
			"name": map[string]any{
				"_eq": roleName,
			},
		},
	)
	if err != nil {
		return err
	}
	if role == nil {
		return errors.New("role not found")
	}
	return queries.CreateProductRoles(ctx, dbx, product.ID, role.ID)
}

func (srv *StripeService) UpsertPriceProductFromStripe(ctx context.Context, dbx db.Dbx) error {
	if err := srv.FindAndUpsertAllProducts(ctx, dbx); err != nil {
		fmt.Println(err)
		return err
	}
	if err := srv.FindAndUpsertAllPrices(ctx, dbx); err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func (srv *StripeService) FindAndUpsertAllProducts(ctx context.Context, dbx db.Dbx) error {
	products, err := srv.client.FindAllProducts()
	if err != nil {
		srv.logger.Error("error finding all products", "error", err)
		return err
	}
	for _, product := range products {
		err = queries.UpsertProductFromStripe(ctx, dbx, product)
		if err != nil {
			srv.logger.Error("error upserting product", "product", product.ID, "error", err)
			continue
		}
	}
	return nil
}

func (srv *StripeService) FindAndUpsertAllPrices(ctx context.Context, dbx db.Dbx) error {
	prices, err := srv.client.FindAllPrices()
	if err != nil {
		srv.logger.Error("error finding all prices", "error", err)
		return err
	}
	for _, price := range prices {
		err = queries.UpsertPriceFromStripe(ctx, dbx, price)
		if err != nil {
			srv.logger.Error("error upserting price", "price", price.ID, "error", err)
			continue
		}
	}
	return nil
}
