package apis

import (
	"context"
	"slices"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/crudrepo"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/mapper"
	"github.com/tkahng/authgo/internal/tools/utils"
)

func (api *Api) AdminStripeSubscriptions(ctx context.Context,
	input *shared.StripeSubscriptionListParams,
) (*shared.PaginatedOutput[*shared.SubscriptionWithData], error) {
	subscriptions, err := api.app.Payment().Store().ListSubscriptions(ctx, input)
	if err != nil {
		return nil, err
	}

	count, err := api.app.Payment().Store().CountSubscriptions(ctx, &input.StripeSubscriptionListFilter)
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

func (api *Api) AdminStripeSubscriptionsGet(ctx context.Context,
	input *shared.StripeSubscriptionGetParams,
) (*struct{ Body *shared.SubscriptionWithData }, error) {
	if input == nil || input.SubscriptionID == "" {
		return nil, huma.Error400BadRequest("subscription_id is required")
	}
	subscription, err := api.app.Payment().Store().FindSubscriptionWithPriceById(ctx, input.SubscriptionID)
	if err != nil {
		return nil, err
	}
	if subscription == nil {
		return nil, nil
	}
	sub := shared.FromCrudToSubWithUserAndPrice(subscription)
	return &struct{ Body *shared.SubscriptionWithData }{Body: sub}, nil
}

func (api *Api) AdminStripeProducts(ctx context.Context,
	input *shared.StripeProductListParams,
) (*shared.PaginatedOutput[*shared.StripeProductWithData], error) {

	products, err := api.app.Payment().Store().ListProducts(ctx, input)
	if err != nil {
		return nil, err
	}
	productIds := mapper.Map(products, func(p *models.StripeProduct) string {
		return p.ID
	})
	if slices.Contains(input.Expand, "prices") {
		data, err := api.app.Payment().Store().LoadProductPrices(ctx, nil, productIds...)
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
		data, err := api.app.Payment().Store().LoadProductRoles(ctx, productIds...)
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
	count, err := api.app.Payment().Store().CountProducts(ctx, &input.StripeProductListFilter)
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

func (api *Api) AdminStripeProductsGet(ctx context.Context,
	input *shared.StripeProductGetParams,
) (*struct{ Body *shared.StripeProductWithData }, error) {

	if input == nil || input.ProductID == "" {
		return nil, huma.Error400BadRequest("product_id is required")
	}
	product, err := api.app.Payment().Store().FindProductByStripeId(ctx, input.ProductID)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, nil
	}

	if slices.Contains(input.Expand, "prices") {
		prices, err := api.app.Payment().Store().LoadProductPrices(ctx, nil, input.ProductID)
		if err != nil {
			return nil, err
		}
		if len(prices) > 0 {
			product.Prices = prices[0]
		}
	}
	if slices.Contains(input.Expand, "roles") {
		roles, err := api.app.Payment().Store().LoadProductRoles(ctx, input.ProductID)
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

func (api *Api) AdminStripeProductsRolesCreate(ctx context.Context, input *struct {
	ProductID string `path:"product-id" required:"true"`
	Body      RoleIdsInput
}) (*struct{}, error) {

	id := input.ProductID
	user, err := api.app.Payment().Store().FindProductByStripeId(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, huma.Error404NotFound("Product not found")
	}

	roleIds := utils.ParseValidUUIDs(input.Body.RolesIds)

	err = api.app.Payment().Store().CreateProductRoles(ctx, user.ID, roleIds...)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (api *Api) AdminStripeProductsRolesDelete(ctx context.Context, input *struct {
	ProductID string `path:"product-id" required:"true"`
	RoleID    string `path:"role-id" required:"true" format:"uuid"`
}) (*struct{}, error) {
	db := api.app.Db()
	id := input.ProductID
	product, err := api.app.Payment().Store().FindProductByStripeId(ctx, id)
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

	role, err := crudrepo.Role.GetOne(ctx, db, &map[string]any{
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
	_, err = crudrepo.ProductRole.DeleteReturn(
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
