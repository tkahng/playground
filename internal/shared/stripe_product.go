package shared

import (
	"time"

	"github.com/tkahng/authgo/internal/db/models"
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

func ModelToProduct(product *models.StripeProduct) *Product {
	return &Product{
		ID:          product.ID,
		Active:      product.Active,
		Name:        product.Name,
		Description: product.Description.Ptr(),
		Image:       product.Image.Ptr(),
		Metadata:    product.Metadata.Val,
		CreatedAt:   product.CreatedAt,
		UpdatedAt:   product.UpdatedAt,
	}
}

type StripeProductWithPrices struct {
	*Product
	Prices []*Price `db:"prices" json:"prices,omitempty" required:"false"`
}
