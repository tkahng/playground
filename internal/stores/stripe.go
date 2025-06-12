package stores

import (
	"context"

	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/repository"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/mapper"
)

// func NewDbStripeStore(db database.Dbx) *DbStripeStore {
// 	return &DbStripeStore{
// 		db:                  db,
// 		DbCustomerStore:     NewDbCustomerStore(db),
// 		DbProductStore:      NewDbProductStore(db),
// 		DbSubscriptionStore: NewDbSubscriptionStore(db),
// 		DbPriceStore:        NewDbPriceStore(db),
// 	}
// }

type DbStripeStore struct {
	*DbCustomerStore
	*DbProductStore
	*DbSubscriptionStore
	*DbPriceStore
	db database.Dbx
}

func (s *DbStripeStore) WithTx(tx database.Dbx) *DbStripeStore {
	return &DbStripeStore{
		db:                  tx,
		DbCustomerStore:     s.DbCustomerStore.WithTx(tx),
		DbProductStore:      s.DbProductStore.WithTx(tx),
		DbSubscriptionStore: s.DbSubscriptionStore.WithTx(tx),
		DbPriceStore:        s.DbPriceStore.WithTx(tx),
	}
}

// func (s *DbStripeStore) LoadPricesByIds(ctx context.Context, priceIds ...string) ([]*models.StripePrice, error) {
// 	if len(priceIds) == 0 {
// 		return nil, nil
// 	}
// 	prices, err := repository.StripePrice.Get(
// 		ctx,
// 		s.db,
// 		&map[string]any{
// 			models.StripePriceTable.ID: map[string]any{
// 				"_in": priceIds,
// 			},
// 		},
// 		nil,
// 		nil,
// 		nil,
// 	)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return mapper.MapToPointer(prices, priceIds, func(t *models.StripePrice) string {
// 		if t == nil {
// 			return ""
// 		}
// 		return t.ID
// 	}), nil
// }

func (s *DbStripeStore) LoadPricesByProductIds(ctx context.Context, productIds ...string) ([][]*models.StripePrice, error) {

	prices, err := repository.StripePrice.Get(
		ctx,
		s.db,
		&map[string]any{
			models.StripePriceTable.ProductID: map[string]any{
				"_in": productIds,
			},
		},
		nil,
		nil,
		nil,
	)
	if err != nil {
		return nil, err
	}
	return mapper.MapToManyPointer(prices, productIds, func(t *models.StripePrice) string {
		return t.ProductID
	}), nil
}

func (s *DbStripeStore) LoadProductsByIds(ctx context.Context, productIds ...string) ([]*models.StripeProduct, error) {
	if len(productIds) == 0 {
		return nil, nil
	}
	products, err := repository.StripeProduct.Get(
		ctx,
		s.db,
		&map[string]any{
			models.StripeProductTable.ID: map[string]any{
				"_in": productIds,
			},
		},
		nil,
		nil,
		nil,
	)
	if err != nil {
		return nil, err
	}
	return mapper.MapToPointer(products, productIds, func(t *models.StripeProduct) string {
		if t == nil {
			return ""
		}
		return t.ID
	}), nil
}

func (s *DbStripeStore) LoadPricesWithProductByPriceIds(ctx context.Context, priceIds ...string) ([]*models.StripePrice, error) {
	if len(priceIds) == 0 {
		return nil, nil
	}
	prices, err := s.LoadPricesByIds(ctx, priceIds...)
	if err != nil {
		return nil, err
	}
	productIds := mapper.Map(prices, func(price *models.StripePrice) string {
		if price == nil || price.ProductID == "" {
			return ""
		}
		return price.ProductID
	})
	products, err := s.LoadProductsByIds(ctx, productIds...)
	if err != nil {
		return nil, err
	}
	for i, price := range prices {
		if price == nil {
			continue
		}
		product := products[i]
		if product == nil {
			continue
		}
		if product.ID != price.ProductID {
			continue
		}
		price.Product = product
	}
	return prices, nil
}

// func (s *DbStripeStore) LoadSubscriptionsPriceProduct(ctx context.Context, subscriptions ...*models.StripeSubscription) error {
// 	if len(subscriptions) == 0 {
// 		return nil
// 	}
// 	priceIds := mapper.Map(subscriptions, func(sub *models.StripeSubscription) string {
// 		if sub == nil || sub.PriceID == "" {
// 			return ""
// 		}
// 		return sub.PriceID
// 	})
// 	prices, err := s.LoadPricesWithProductByPriceIds(ctx, priceIds...)
// 	if err != nil {
// 		return err
// 	}
// 	for i, sub := range subscriptions {
// 		if sub == nil {
// 			continue
// 		}
// 		price := prices[i]
// 		if price == nil {
// 			continue
// 		}
// 		if price.ID != sub.PriceID {
// 			continue
// 		}
// 		sub.Price = price
// 	}
// 	return nil
// }

func (s *DbStripeStore) LoadSubscriptionsByIds(ctx context.Context, subscriptionIds ...string) ([]*models.StripeSubscription, error) {
	if len(subscriptionIds) == 0 {
		return nil, nil
	}
	where := map[string]any{
		models.StripeSubscriptionTable.ID: map[string]any{
			"_in": subscriptionIds,
		},
	}
	subscriptions, err := repository.StripeSubscription.Get(
		ctx,
		s.db,
		&where,
		nil,
		nil,
		nil,
	)
	if err != nil {
		return nil, err
	}
	return mapper.MapToPointer(subscriptions, subscriptionIds, func(t *models.StripeSubscription) string {
		if t == nil {
			return ""
		}
		return t.ID
	}), nil
}

// func WithPrefix(prefix, name string) string {
// 	if prefix == "" {
// 		return name
// 	}
// 	return fmt.Sprintf("%s.%s", prefix, name)
// }

// func Quote(name string) string {
// 	return fmt.Sprintf("\"%s\"", name)
// }

var (
	MetadataIndexName = "metadata.index"
)

func (s *DbStripeStore) LoadProductRoles(ctx context.Context, productIds ...string) ([][]*models.Role, error) {
	const (
		getProductRolesQuery = `
		SELECT rp.product_id as key,
			COALESCE(
					json_agg(
							jsonb_build_object(
									'id',
									p.id,
									'name',
									p.name,
									'description',
									p.description,
									'created_at',
									p.created_at,
									'updated_at',
									p.updated_at
							)
					) FILTER (
							WHERE p.id IS NOT NULL
					),
					'[]'
			) AS data
	FROM public.product_roles rp
			LEFT JOIN public.roles p ON p.id = rp.role_id
			WHERE rp.product_id	 = ANY (
					$1::text []
			)
	GROUP BY rp.product_id;`
	)
	data, err := database.QueryAll[shared.JoinedResult[*models.Role, string]](
		ctx,
		s.db,
		getProductRolesQuery,
		productIds,
	)
	if err != nil {
		return nil, err
	}
	return mapper.Map(mapper.MapTo(data, productIds, func(a shared.JoinedResult[*models.Role, string]) string {
		return a.Key
	}), func(a *shared.JoinedResult[*models.Role, string]) []*models.Role {
		if a == nil {
			return nil
		}
		return a.Data
	}), nil
}
