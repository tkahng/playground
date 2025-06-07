package stores

import (
	"context"
	"slices"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/stripe/stripe-go/v82"
	"github.com/tkahng/authgo/internal/crudrepo"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/shared"
)

func (s *DbStripeStore) ListProducts(ctx context.Context, input *shared.StripeProductListParams) ([]*models.StripeProduct, error) {
	q := squirrel.Select("stripe_products.*").
		From("stripe_products")
	filter := input.StripeProductListFilter
	pageInput := &input.PaginatedInput

	q = database.Paginate(q, pageInput)
	q = listProductFilterFuncQuery(q, &filter)
	q = listProductOrderByQuery(q, input)
	data, err := database.QueryWithBuilder[*models.StripeProduct](ctx, s.db, q.PlaceholderFormat(squirrel.Dollar))
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (s *DbStripeStore) CountProducts(ctx context.Context, filter *shared.StripeProductListFilter) (int64, error) {
	q := squirrel.Select("COUNT(stripe_products.*)").
		From("stripe_products")

	q = listProductFilterFuncQuery(q, filter)
	data, err := database.QueryWithBuilder[database.CountOutput](ctx, s.db, q.PlaceholderFormat(squirrel.Dollar))

	if err != nil {
		return 0, err
	}
	if len(data) == 0 {
		return 0, nil
	}

	return data[0].Count, nil
}

func (s *DbStripeStore) FindProductById(ctx context.Context, productId string) (*models.StripeProduct, error) {
	data, err := crudrepo.StripeProduct.GetOne(
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
func (s *DbStripeStore) UpsertProduct(ctx context.Context, product *models.StripeProduct) error {
	var dbx database.Dbx = s.db
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
	return database.ExecWithBuilder(ctx, dbx, q.PlaceholderFormat(squirrel.Dollar))
}

// UpsertProductFromStripe implements PaymentStore.
func (s *DbStripeStore) UpsertProductFromStripe(ctx context.Context, product *stripe.Product) error {
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

func listProductOrderByQuery(q squirrel.SelectBuilder, input *shared.StripeProductListParams) squirrel.SelectBuilder {
	if input == nil {
		return q
	}
	if input.SortParams.SortBy == "" {
		q = q.OrderBy("metadata->'index'" + " " + strings.ToUpper(input.SortOrder))
	}
	if input.SortBy == MetadataIndexName {
		q = q.OrderBy("metadata->'index'" + " " + strings.ToUpper(input.SortOrder))
	} else if slices.Contains(models.StripeProductTable.Columns, input.SortBy) {
		q = q.OrderBy(input.SortBy + " " + strings.ToUpper(input.SortOrder))
	}
	return q
}

func listProductFilterFuncQuery(q squirrel.SelectBuilder, filter *shared.StripeProductListFilter) squirrel.SelectBuilder {
	if filter == nil {
		return q
	}
	if filter.Active != "" {
		if filter.Active == shared.Active {
			q = q.Where("active = ?", true)
		}
		if filter.Active == shared.Inactive {
			q = q.Where("active = ?", false)
		}
	}
	if len(filter.Ids) > 0 {
		q = q.Where("id in (?)", filter.Ids)
	}
	return q
}

func SelectStripeProductColumns(qs squirrel.SelectBuilder, prefix string) squirrel.SelectBuilder {
	qs = qs.Column(models.StripeProductTablePrefix.ID + " AS " + Quote(WithPrefix(prefix, models.StripeProductTable.ID))).
		Column(models.StripeProductTablePrefix.Name + " AS " + Quote(WithPrefix(prefix, models.StripeProductTable.Name))).
		Column(models.StripeProductTablePrefix.Description + " AS " + Quote(WithPrefix(prefix, models.StripeProductTable.Description))).
		Column(models.StripeProductTablePrefix.Active + " AS " + Quote(WithPrefix(prefix, models.StripeProductTable.Active))).
		Column(models.StripeProductTablePrefix.Image + " AS " + Quote(WithPrefix(prefix, models.StripeProductTable.Image))).
		Column(models.StripeProductTablePrefix.Metadata + " AS " + Quote(WithPrefix(prefix, models.StripeProductTable.Metadata))).
		Column(models.StripeProductTablePrefix.CreatedAt + " AS " + Quote(WithPrefix(prefix, models.StripeProductTable.CreatedAt))).
		Column(models.StripeProductTablePrefix.UpdatedAt + " AS " + Quote(WithPrefix(prefix, models.StripeProductTable.UpdatedAt)))

	return qs
}
