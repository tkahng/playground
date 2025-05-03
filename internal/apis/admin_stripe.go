package apis

import (
	"context"
	"net/http"
	"slices"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/db/models"
	"github.com/tkahng/authgo/internal/queries"
	"github.com/tkahng/authgo/internal/repository"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/mapper"
)

func (api *Api) AdminStripeSubscriptionsOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "admin-stripe-subscriptions",
		Method:      http.MethodGet,
		Path:        path,
		Summary:     "Admin stripe subscriptions",
		Description: "List of stripe subscriptions",
		Tags:        []string{"Admin", "Subscription", "Stripe"},
		Errors:      []int{http.StatusNotFound},
		Security: []map[string][]string{
			{shared.BearerAuthSecurityKey: {}},
		},
	}
}

func (api *Api) AdminStripeSubscriptions(ctx context.Context,
	input *shared.StripeSubscriptionListParams,
) (*shared.PaginatedOutput[*shared.SubscriptionWithData], error) {
	db := api.app.Db()
	subscriptions, err := queries.ListSubscriptions(ctx, db, input)
	if err != nil {
		return nil, err
	}

	count, err := queries.CountSubscriptions(ctx, db, &input.StripeSubscriptionListFilter)
	if err != nil {
		return nil, err
	}
	return &shared.PaginatedOutput[*shared.SubscriptionWithData]{
		Body: shared.PaginatedResponse[*shared.SubscriptionWithData]{
			Data: mapper.Map(subscriptions, func(sub *models.StripeSubscription) *shared.SubscriptionWithData {
				return &shared.SubscriptionWithData{
					Subscription: shared.FromCrudSubscription(sub),
				}
			}),
			Meta: shared.GenerateMeta(input.PaginatedInput, count),
		},
	}, nil

}

func (api *Api) AdminStripeSubscriptionsGetOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "admin-stripe-subscription-get",
		Method:      http.MethodGet,
		Path:        path,
		Summary:     "Admin stripe subscription get",
		Description: "Get a stripe subscription by ID",
		Tags:        []string{"Admin", "Subscription", "Stripe"},
		Errors:      []int{http.StatusNotFound, http.StatusBadRequest},
		Security: []map[string][]string{
			{shared.BearerAuthSecurityKey: {}},
		},
	}
}

func (api *Api) AdminStripeSubscriptionsGet(ctx context.Context,
	input *shared.StripeSubscriptionGetParams,
) (*struct{ Body *shared.SubscriptionWithData }, error) {
	db := api.app.Db()
	if input == nil || input.SubscriptionID == "" {
		return nil, huma.Error400BadRequest("subscription_id is required")
	}
	subscription, err := queries.FindSubscriptionWithPriceById(ctx, db, input.SubscriptionID)
	if err != nil {
		return nil, err
	}
	if subscription == nil {
		return nil, nil
	}
	sub := shared.FromCrudToSubWithUserAndPrice(subscription)
	return &struct{ Body *shared.SubscriptionWithData }{Body: sub}, nil
}

func (api *Api) AdminStripeProductsOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "admin-stripe-products",
		Method:      http.MethodGet,
		Path:        path,
		Summary:     "Admin stripe products",
		Description: "List of stripe products",
		Tags:        []string{"Admin", "Product", "Stripe"},
		Errors:      []int{http.StatusNotFound},
		Security: []map[string][]string{
			{shared.BearerAuthSecurityKey: {}},
		},
	}
}
func (api *Api) AdminStripeProducts(ctx context.Context,
	input *shared.StripeProductListParams,
) (*shared.PaginatedOutput[*shared.StripeProductWithData], error) {
	db := api.app.Db()
	products, err := queries.ListProducts(ctx, db, input)
	if err != nil {
		return nil, err
	}
	productIds := mapper.Map(products, func(p *models.StripeProduct) string {
		return p.ID
	})
	if slices.Contains(input.Expand, "prices") {
		data, err := queries.LoadeProductPrices(ctx, db, nil, productIds...)
		if err != nil {
			return nil, err
		}
		for idx, products := range products {
			prices := data[idx]
			if len(prices) > 0 {
				products.Prices = prices
			}
		}
	}
	if slices.Contains(input.Expand, "roles") {
		data, err := queries.LoadProductRoles(ctx, db, productIds...)
		if err != nil {
			return nil, err
		}
		for idx, products := range products {
			roles := data[idx]
			if len(roles) > 0 {
				products.Roles = roles
			}
		}
	}
	count, err := queries.CountProducts(ctx, db, &input.StripeProductListFilter)
	if err != nil {
		return nil, err
	}
	return &shared.PaginatedOutput[*shared.StripeProductWithData]{
		Body: shared.PaginatedResponse[*shared.StripeProductWithData]{
			Data: mapper.Map(products, func(p *models.StripeProduct) *shared.StripeProductWithData {
				return &shared.StripeProductWithData{
					Product: shared.FromCrudProduct(p),
					Roles: mapper.Map(p.Roles, func(r *models.Role) *shared.Role {
						return shared.FromCrudRole(r)
					}),
					Prices: mapper.Map(p.Prices, func(p *models.StripePrice) *shared.Price {
						return shared.FromCrudPrice(p)
					}),
				}
			}),
			Meta: shared.GenerateMeta(input.PaginatedInput, count),
		},
	}, nil
}

func (api *Api) AdminStripeProductsGetOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "admin-stripe-product-get",
		Method:      http.MethodGet,
		Path:        path,
		Summary:     "Admin stripe product get",
		Description: "Get a stripe product by ID",
		Tags:        []string{"Admin", "Product", "Stripe"},
		Errors:      []int{http.StatusNotFound, http.StatusBadRequest},
		Security: []map[string][]string{
			{shared.BearerAuthSecurityKey: {}},
		},
	}
}
func (api *Api) AdminStripeProductsGet(ctx context.Context,
	input *shared.StripeProductGetParams,
) (*struct{ Body *shared.StripeProductWithData }, error) {
	db := api.app.Db()
	if input == nil || input.ProductID == "" {
		return nil, huma.Error400BadRequest("product_id is required")
	}
	product, err := queries.FindProductByStripeId(ctx, db, input.ProductID)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, nil
	}

	if slices.Contains(input.Expand, "prices") {
		prices, err := queries.LoadeProductPrices(ctx, db, nil, input.ProductID)
		if err != nil {
			return nil, err
		}
		if len(prices) > 0 {
			product.Prices = prices[0]
		}
	}
	if slices.Contains(input.Expand, "roles") {
		roles, err := queries.LoadProductRoles(ctx, db, input.ProductID)
		if err != nil {
			return nil, err
		}
		if len(roles) > 0 {
			product.Roles = roles[0]
		}
	}
	return &struct{ Body *shared.StripeProductWithData }{
		Body: &shared.StripeProductWithData{
			Product: shared.FromCrudProduct(product),
			Roles:   mapper.Map(product.Roles, shared.FromCrudRole),
			Prices:  mapper.Map(product.Prices, shared.FromCrudPrice),
		},
	}, nil
}

func (api *Api) AdminStripeProductsRolesCreateOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "admin-create-product-roles",
		Method:      http.MethodPost,
		Path:        path,
		Summary:     "Create product roles",
		Description: "Create product roles",
		Tags:        []string{"Admin", "Roles", "Product", "Stripe"},
		Errors:      []int{http.StatusNotFound},
		Security: []map[string][]string{
			{shared.BearerAuthSecurityKey: {}},
		},
	}
}

func (api *Api) AdminStripeProductsRolesCreate(ctx context.Context, input *struct {
	ProductID string `path:"product-id" required:"true"`
	Body      RoleIdsInput
}) (*struct{}, error) {
	db := api.app.Db()
	id := input.ProductID
	user, err := queries.FindProductByStripeId(ctx, db, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, huma.Error404NotFound("Product not found")
	}

	roleIds := queries.ParseUUIDs(input.Body.RolesIds)

	err = queries.CreateProductRoles(ctx, db, user.ID, roleIds...)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (api *Api) AdminStripeProductsRolesDeleteOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "admin-delete-product-roles",
		Method:      http.MethodDelete,
		Path:        path,
		Summary:     "Delete product roles",
		Description: "Delete product roles",
		Tags:        []string{"Admin", "Roles", "Product", "Stripe"},
		Errors:      []int{http.StatusNotFound},
		Security: []map[string][]string{
			{shared.BearerAuthSecurityKey: {}},
		},
	}
}
func (api *Api) AdminStripeProductsRolesDelete(ctx context.Context, input *struct {
	ProductID string `path:"product-id" required:"true"`
	RoleID    string `path:"role-id" required:"true" format:"uuid"`
}) (*struct{}, error) {
	db := api.app.Db()
	id := input.ProductID
	product, err := queries.FindProductByStripeId(ctx, db, id)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, huma.Error404NotFound("Product not found")
	}

	roleId, err := uuid.Parse(input.RoleID)
	if err != nil {
		return nil, huma.Error400BadRequest("role_id is not a valid UUID")
	}

	role, err := repository.Role.GetOne(ctx, db, &map[string]any{
		"id": map[string]any{
			"_eq": roleId.String(),
		},
	})
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, huma.Error404NotFound("Role not found")
	}
	_, err = repository.ProductRole.DeleteReturn(
		ctx,
		db,
		&map[string]any{
			"product_id": map[string]any{
				"_eq": id,
			},
			"role_id": map[string]any{
				"_eq": role.ID.String(),
			},
		},
	)
	if err != nil {
		return nil, err
	}
	return nil, nil
}
