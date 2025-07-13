package apis

import (
	"time"

	"github.com/tkahng/playground/internal/models"
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

func FromModelProduct(product *models.StripeProduct) *StripeProduct {
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
		Prices:      mapper.Map(product.Prices, FromModelPrice),
		Permissions: mapper.Map(product.Permissions, FromModelPermission),
		Roles:       mapper.Map(product.Roles, FromModelRole),
	}
}

type StripeProductListParams struct {
	PaginatedInput
	SortParams
	StripeProductExpand
	Q      string                    `query:"q,omitempty" required:"false"`
	Ids    []string                  `query:"ids,omitempty" required:"false" minimum:"1" maximum:"100"`
	Active types.OptionalParam[bool] `query:"active,omitempty" required:"false"`
}

type StripeProductExpand struct {
	Expand []string `query:"expand,omitempty" required:"false" minimum:"1" maximum:"100" uniqueItems:"true" enum:"prices,permissions"`
}

type StripeProductGetParams struct {
	ProductID string `path:"product-id" json:"product_id" required:"true"`
	StripeProductExpand
}
