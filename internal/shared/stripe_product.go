package shared

import (
	"time"

	crudModels "github.com/tkahng/authgo/internal/models"
)

type Product struct {
	ID          string            `db:"id,pk" json:"id"`
	Active      bool              `db:"active" json:"active"`
	Name        string            `db:"name" json:"name"`
	Description *string           `db:"description" json:"description"`
	Image       *string           `db:"image" json:"image"`
	Metadata    map[string]string `db:"metadata" json:"metadata"`
	CreatedAt   time.Time         `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time         `db:"updated_at" json:"updated_at"`
}

func FromModelProduct(product *crudModels.StripeProduct) *Product {
	return &Product{
		ID:          product.ID,
		Active:      product.Active,
		Name:        product.Name,
		Description: product.Description,
		Image:       product.Image,
		Metadata:    product.Metadata,
		CreatedAt:   product.CreatedAt,
		UpdatedAt:   product.UpdatedAt,
	}
}

type StripeProductWitPermission struct {
	*Product
	Permissions []*Permission `db:"permissions" json:"permissions,omitempty" required:"false"`
	Prices      []*Price      `db:"prices" json:"prices,omitempty" required:"false"`
}

type StripeProductListFilter struct {
	Q      string       `query:"q,omitempty" required:"false"`
	Ids    []string     `query:"ids,omitempty" required:"false" minimum:"1" maximum:"100"`
	Active ActiveStatus `query:"active,omitempty" required:"false" enum:"active,inactive"`
}

type StripeProductListParams struct {
	PaginatedInput
	StripeProductListFilter
	SortParams
	StripeProductExpand
	PriceActive ActiveStatus `query:"price_active,omitempty" required:"false" enum:"active,inactive"`
}

type StripeProductExpand struct {
	Expand []string `query:"expand,omitempty" required:"false" minimum:"1" maximum:"100" uniqueItems:"true" enum:"prices,permissions"`
}

type StripeProductGetParams struct {
	ProductID string `path:"product-id" json:"product_id" required:"true"`
	StripeProductExpand
}
