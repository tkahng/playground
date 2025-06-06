package apis

import (
	"context"
	"slices"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/mapper"
	"github.com/tkahng/authgo/internal/tools/utils"
)

func (api *Api) AdminStripeSubscriptions(ctx context.Context,
	input *shared.StripeSubscriptionListParams,
) (*shared.PaginatedOutput[*shared.Subscription], error) {
	subscriptions, err := api.app.Payment().Store().ListSubscriptions(ctx, input)
	if err != nil {
		return nil, err
	}

	count, err := api.app.Payment().Store().CountSubscriptions(ctx, &input.StripeSubscriptionListFilter)
	if err != nil {
		return nil, err
	}
	return &shared.PaginatedOutput[*shared.Subscription]{
		Body: shared.PaginatedResponse[*shared.Subscription]{
			Data: mapper.Map(subscriptions, shared.FromModelSubscription),
			Meta: shared.GenerateMeta(&input.PaginatedInput, count),
		},
	}, nil
}

func (api *Api) AdminStripeSubscriptionsGet(ctx context.Context,
	input *shared.StripeSubscriptionGetParams,
) (*struct{ Body *shared.Subscription }, error) {
	if input == nil || input.SubscriptionID == "" {
		return nil, huma.Error400BadRequest("subscription_id is required")
	}
	subscriptions, err := api.app.Payment().Store().ListSubscriptions(ctx, &shared.StripeSubscriptionListParams{
		StripeSubscriptionListFilter: shared.StripeSubscriptionListFilter{
			Ids: []string{input.SubscriptionID},
		},
		PaginatedInput: shared.PaginatedInput{
			Page:    0,
			PerPage: 1,
		},
	})
	if err != nil {
		return nil, err
	}
	if len(subscriptions) == 0 {
		return nil, nil
	}
	return &struct{ Body *shared.Subscription }{Body: shared.FromModelSubscription(subscriptions[0])}, nil
}

func (api *Api) AdminStripeProducts(ctx context.Context,
	input *shared.StripeProductListParams,
) (*shared.PaginatedOutput[*shared.StripeProduct], error) {

	products, err := api.app.Payment().Store().ListProducts(ctx, input)
	if err != nil {
		return nil, err
	}
	productIds := mapper.Map(products, func(p *models.StripeProduct) string {
		return p.ID
	})
	if slices.Contains(input.Expand, "prices") {
		data, err := api.app.Payment().Store().LoadPricesByProductIds(ctx, productIds...)
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
	if slices.Contains(input.Expand, "permissions") {
		data, err := api.app.Rbac().Store().LoadProductPermissions(ctx, productIds...)
		if err != nil {
			return nil, err
		}
		for idx, products := range products {
			permissions := data[idx]
			if len(permissions) > 0 {
				products.Permissions = permissions
			}
		}
	}
	count, err := api.app.Payment().Store().CountProducts(ctx, &input.StripeProductListFilter)
	if err != nil {
		return nil, err
	}
	return &shared.PaginatedOutput[*shared.StripeProduct]{
		Body: shared.PaginatedResponse[*shared.StripeProduct]{
			Data: mapper.Map(products, shared.FromModelProduct),
			Meta: shared.GenerateMeta(&input.PaginatedInput, count),
		},
	}, nil
}

func (api *Api) AdminStripeProductsGet(ctx context.Context,
	input *shared.StripeProductGetParams,
) (*struct {
	Body *shared.StripeProduct
}, error) {

	if input == nil || input.ProductID == "" {
		return nil, huma.Error400BadRequest("product_id is required")
	}
	product, err := api.app.Payment().Store().FindProductById(ctx, input.ProductID)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, nil
	}

	if slices.Contains(input.Expand, "prices") {
		prices, err := api.app.Payment().Store().LoadPricesByProductIds(ctx, input.ProductID)
		if err != nil {
			return nil, err
		}
		if len(prices) > 0 {
			product.Prices = prices[0]
		}
	}
	if slices.Contains(input.Expand, "permissions") {
		roles, err := api.app.Payment().Store().LoadProductPermissions(ctx, input.ProductID)
		if err != nil {
			return nil, err
		}
		if len(roles) > 0 {
			product.Permissions = roles[0]
		}
	}
	return &struct {
		Body *shared.StripeProduct
	}{
		Body: shared.FromModelProduct(product),
	}, nil
}

func (api *Api) AdminStripeProductsRolesCreate(ctx context.Context, input *struct {
	ProductID string `path:"product-id" required:"true"`
	Body      RoleIdsInput
}) (*struct{}, error) {

	id := input.ProductID
	user, err := api.app.Payment().Store().FindProductById(ctx, id)
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
	id := input.ProductID
	product, err := api.app.Payment().Store().FindProductById(ctx, id)
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

	role, err := api.app.Rbac().Store().FindRoleById(ctx, roleId)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, huma.Error404NotFound("Role not found")
	}
	err = api.app.Rbac().Store().DeleteProductRoles(ctx, product.ID, role.ID)
	if err != nil {
		return nil, err
	}
	return nil, nil
}
