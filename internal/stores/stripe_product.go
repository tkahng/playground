package stores

import (
	"context"
	"slices"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/stripe/stripe-go/v82"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/repository"
	"github.com/tkahng/authgo/internal/tools/types"
	"github.com/tkahng/authgo/internal/tools/utils"
)

type DbProductStore struct {
	db database.Dbx
}

func NewDbProductStore(db database.Dbx) *DbProductStore {
	return &DbProductStore{
		db: db,
	}
}
func (s *DbProductStore) WithTx(tx database.Dbx) *DbProductStore {
	return &DbProductStore{
		db: tx,
	}
}

func (s *DbProductStore) FindProduct(ctx context.Context, filter *StripeProductFilter) (*models.StripeProduct, error) {
	q := squirrel.Select("stripe_products.*").
		From("stripe_products")

	q = s.listProductFilterFuncQuery(q, filter)
	data, err := database.QueryWithBuilder[*models.StripeProduct](ctx, s.db, q.Limit(1).PlaceholderFormat(squirrel.Dollar))
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, nil
	}
	return data[0], nil
}

func (s *DbProductStore) ListProducts(ctx context.Context, input *StripeProductFilter) ([]*models.StripeProduct, error) {
	q := squirrel.Select("stripe_products.*").
		From("stripe_products")

	q = queryPagination(q, input)
	q = s.listProductFilterFuncQuery(q, input)
	q = s.listProductOrderByQuery(q, input)
	data, err := database.QueryWithBuilder[*models.StripeProduct](ctx, s.db, q.PlaceholderFormat(squirrel.Dollar))
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (s *DbProductStore) CountProducts(ctx context.Context, filter *StripeProductFilter) (int64, error) {
	q := squirrel.Select("COUNT(stripe_products.*)").
		From("stripe_products")

	q = s.listProductFilterFuncQuery(q, filter)
	data, err := database.QueryWithBuilder[database.CountOutput](ctx, s.db, q.PlaceholderFormat(squirrel.Dollar))

	if err != nil {
		return 0, err
	}
	if len(data) == 0 {
		return 0, nil
	}

	return data[0].Count, nil
}

func (s *DbProductStore) FindProductById(ctx context.Context, productId string) (*models.StripeProduct, error) {
	data, err := repository.StripeProduct.GetOne(
		ctx,
		s.db,
		&map[string]any{
			models.StripeProductTable.ID: map[string]any{
				"_eq": productId,
			},
		},
	)
	return database.OptionalRow(data, err)
}
func (s *DbProductStore) UpsertProduct(ctx context.Context, product *models.StripeProduct) error {
	dbx := s.db
	q := squirrel.Insert("stripe_products").
		Columns(
			"id",
			"active",
			"name",
			"description",
			"image",
			"metadata",
		).
		Values(
			product.ID,
			product.Active,
			product.Name,
			product.Description,
			product.Image,
			product.Metadata,
		).Suffix(`ON CONFLICT (id) DO UPDATE SET 
						active = EXCLUDED.active, 
						name = EXCLUDED.name, 
						description = EXCLUDED.description, 
						image = EXCLUDED.image, 
						metadata = EXCLUDED.metadata
		`)
	_, err := database.ExecWithBuilder(ctx, dbx, q.PlaceholderFormat(squirrel.Dollar))
	return err
}

// UpsertProductFromStripe implements PaymentStore.
func (s *DbProductStore) UpsertProductFromStripe(ctx context.Context, product *stripe.Product) error {
	if product == nil {
		return nil
	}
	var image *string
	if len(product.Images) > 0 {
		image = &product.Images[0]
	}
	param := &models.StripeProduct{
		ID:          product.ID,
		Active:      product.Active,
		Name:        product.Name,
		Description: &product.Description,
		Image:       image,
		Metadata:    product.Metadata,
	}
	return s.UpsertProduct(ctx, param)
}

func (s *DbProductStore) listProductOrderByQuery(q squirrel.SelectBuilder, input *StripeProductFilter) squirrel.SelectBuilder {
	if input == nil {
		return q
	}
	if input.SortBy == "" {
		q = q.OrderBy("metadata->'index'" + " " + strings.ToUpper(input.SortOrder))
	}
	if input.SortBy == MetadataIndexName {
		q = q.OrderBy("metadata->'index'" + " " + strings.ToUpper(input.SortOrder))
	} else if slices.Contains(models.StripeProductTable.Columns, input.SortBy) {
		q = q.OrderBy(input.SortBy + " " + strings.ToUpper(input.SortOrder))
	}
	return q
}

func (s *DbProductStore) listProductFilterFuncQuery(q squirrel.SelectBuilder, filter *StripeProductFilter) squirrel.SelectBuilder {
	if filter == nil {
		return q
	}
	if filter.Active.IsSet {
		q = q.Where("active = ?", filter.Active.Value)
	}
	if len(filter.Ids) > 0 {
		q = q.Where("id in (?)", filter.Ids)
	}

	return q
}

func SelectStripeProductColumns(qs squirrel.SelectBuilder, prefix string) squirrel.SelectBuilder {
	qs = qs.Column(models.StripeProductTablePrefix.ID + " AS " + utils.Quote(utils.WithPrefix(prefix, models.StripeProductTable.ID))).
		Column(models.StripeProductTablePrefix.Name + " AS " + utils.Quote(utils.WithPrefix(prefix, models.StripeProductTable.Name))).
		Column(models.StripeProductTablePrefix.Description + " AS " + utils.Quote(utils.WithPrefix(prefix, models.StripeProductTable.Description))).
		Column(models.StripeProductTablePrefix.Active + " AS " + utils.Quote(utils.WithPrefix(prefix, models.StripeProductTable.Active))).
		Column(models.StripeProductTablePrefix.Image + " AS " + utils.Quote(utils.WithPrefix(prefix, models.StripeProductTable.Image))).
		Column(models.StripeProductTablePrefix.Metadata + " AS " + utils.Quote(utils.WithPrefix(prefix, models.StripeProductTable.Metadata))).
		Column(models.StripeProductTablePrefix.CreatedAt + " AS " + utils.Quote(utils.WithPrefix(prefix, models.StripeProductTable.CreatedAt))).
		Column(models.StripeProductTablePrefix.UpdatedAt + " AS " + utils.Quote(utils.WithPrefix(prefix, models.StripeProductTable.UpdatedAt)))

	return qs
}

type StripeProductFilter struct {
	PaginatedInput
	SortParams
	Q      string                    `query:"q,omitempty" required:"false"`
	Ids    []string                  `query:"ids,omitempty" required:"false" minimum:"1" maximum:"100" uniqueItems:"true"`
	Active types.OptionalParam[bool] `query:"active,omitempty" required:"false"`
	Expand []string                  `query:"expand,omitempty" required:"false" minimum:"1" maximum:"100" uniqueItems:"true" enum:"prices,permissions"`
}
type DbProductStoreInterface interface {
	ListProducts(ctx context.Context, input *StripeProductFilter) ([]*models.StripeProduct, error)
	CountProducts(ctx context.Context, filter *StripeProductFilter) (int64, error)
	FindProduct(ctx context.Context, filter *StripeProductFilter) (*models.StripeProduct, error)
	FindProductById(ctx context.Context, productId string) (*models.StripeProduct, error)
	UpsertProduct(ctx context.Context, product *models.StripeProduct) error
	UpsertProductFromStripe(ctx context.Context, product *stripe.Product) error
}
