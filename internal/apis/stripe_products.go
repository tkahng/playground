package apis

import (
	"context"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/stores"
	"github.com/tkahng/playground/internal/tools/mapper"
	"github.com/tkahng/playground/internal/tools/types"
)

type StripeProduct struct {
	_           struct{}          `db:"stripe_products" json:"-"`
	ID          string            `db:"id" json:"id"`
	Active      bool              `db:"active" json:"active"`
	Name        string            `db:"name" json:"name"`
	Description *string           `db:"description" json:"description"`
	Image       *string           `db:"image" json:"image"`
	Metadata    map[string]string `db:"metadata" json:"metadata"`
	CreatedAt   time.Time         `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time         `db:"updated_at" json:"updated_at"`
	Prices      []*StripePrice    `db:"prices" src:"id" dest:"product_id" table:"stripe_prices" json:"prices,omitempty"`
	Roles       []*Role           `db:"roles" src:"id" dest:"product_id" table:"roles" through:"product_roles,role_id,id" json:"roles,omitempty"`
	Permissions []*Permission     `db:"permissions" src:"id" dest:"product_id" table:"permissions" through:"product_permissions,permission_id,id" json:"permissions,omitempty"`
}

func fromModelProduct(product *models.StripeProduct) *StripeProduct {
	if product == nil {
		return nil
	}
	return &StripeProduct{
		ID:          product.ID,
		Active:      product.Active,
		Name:        product.Name,
		Description: product.Description,
		Image:       product.Image,
		Metadata:    product.Metadata,
		CreatedAt:   product.CreatedAt,
		UpdatedAt:   product.UpdatedAt,
		Prices:      mapper.Map(product.Prices, fromModelPrice),
		Permissions: mapper.Map(product.Permissions, fromModelPermission),
		Roles:       mapper.Map(product.Roles, fromModelRole),
	}
}

type StripeProductListParams struct {
	PaginatedInput
	SortParams
	StripeProductExpand
	Q            string                                        `query:"q,omitempty" required:"false"`
	Ids          []string                                      `query:"ids,omitempty" required:"false" minimum:"1" maximum:"100"`
	Active       types.OptionalParam[bool]                     `query:"active,omitempty" required:"false"`
	MetadataType types.OptionalParam[models.StripeProductType] `query:"metadata_type,omitempty" required:"false" enum:"subscription,points"`
}

type StripeProductExpand struct {
	Expand []string `query:"expand,omitempty" required:"false" minimum:"1" maximum:"100" uniqueItems:"true" enum:"prices,permissions"`
}

type StripeProductGetParams struct {
	ProductID string `path:"product-id" json:"product_id" required:"true"`
	StripeProductExpand
}

func (a *Api) bindStripeProductsWithPrices(api huma.API) {
	huma.Register(
		api,
		huma.Operation{
			OperationID: "stripe-products-with-prices",
			Method:      http.MethodGet,
			Path:        "/stripe/products",
			Summary:     "stripe-products-with-prices",
			Description: "stripe-products-with-prices",
			Tags:        []string{"Stripe", "Products"},
			Errors:      []int{http.StatusInternalServerError, http.StatusBadRequest},
		},
		a.StripeProductsWithPrices,
	)
}

type StripeProductsWithPricesInput struct {
	PaginatedInput
	SortParams
	MetadataType types.OptionalParam[models.StripeProductType] `query:"metadata_type,omitempty" required:"false" enum:"subscription,points"`
}

func (api *Api) StripeProductsWithPrices(ctx context.Context, input *StripeProductsWithPricesInput) (*ApiPaginatedOutput[*StripeProduct], error) {
	filter := &stores.StripeProductFilter{}
	filter.Page = input.Page
	filter.PerPage = input.PerPage
	filter.Active.IsSet = true
	filter.Active.Value = true
	filter.SortBy = input.SortBy
	filter.SortOrder = input.SortOrder
	filter.MetadataType = input.MetadataType

	products, err := api.App().Adapter().Product().ListProducts(ctx, filter)
	if err != nil {
		return nil, err
	}
	ids := []string{}
	for _, u := range products {
		ids = append(ids, u.ID)
	}
	if len(ids) > 0 {
		prices, err := api.App().Adapter().Price().LoadPricesByProductIds(ctx, ids...)
		if err != nil {
			return nil, err
		}
		for i, product := range products {
			if price := prices[i]; len(price) > 0 {
				product.Prices = price
			}
		}
	}

	count, err := api.App().Adapter().Product().CountProducts(ctx, filter)
	if err != nil {
		return nil, err
	}

	return &ApiPaginatedOutput[*StripeProduct]{Body: ApiPaginatedResponse[*StripeProduct]{
		Data: mapper.Map(products, fromModelProduct),
		Meta: ApiGenerateMeta(&input.PaginatedInput, count),
	}}, nil
}
