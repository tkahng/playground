package apis

import (
	"context"
	"slices"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/stores"
	"github.com/tkahng/authgo/internal/tools/mapper"
	"github.com/tkahng/authgo/internal/tools/utils"
)

type StripeSubscriptionListFilter struct {
	Q       string                     `query:"q,omitempty" required:"false"`
	Ids     []string                   `query:"ids,omitempty" required:"false" minimum:"1" maximum:"100" format:"uuid"`
	UserIDs []string                   `query:"user_id,omitempty" required:"false" format:"uuid"`
	TeamIDs []string                   `query:"team_id,omitempty" required:"false" format:"uuid"`
	Status  []StripeSubscriptionStatus `query:"status,omitempty" required:"false" minimum:"1" maximum:"100" enum:"trialing,active,canceled,incomplete,incomplete_expired,past_due,unpaid,paused"`
}
type StripeSubscriptionListParams struct {
	PaginatedInput
	StripeSubscriptionListFilter
	SortParams
	Expand []string `query:"expand,omitempty" required:"false" minimum:"1" maximum:"100" enum:"user,price,product"`
}

func ToStripeSubscriptionListFilter(input *StripeSubscriptionListParams) (*stores.StripeSubscriptionListFilter, error) {
	filter := &stores.StripeSubscriptionListFilter{}
	filter.Page = input.Page
	filter.PerPage = input.PerPage
	filter.Status = mapper.Map(input.Status, func(s StripeSubscriptionStatus) models.StripeSubscriptionStatus {
		return models.StripeSubscriptionStatus(s)
	})
	filter.Ids = input.Ids
	filter.TeamIDs = utils.ParseValidUUIDs(input.TeamIDs...)
	filter.UserIDs = utils.ParseValidUUIDs(input.UserIDs...)
	filter.Expand = input.Expand

	return filter, nil
}

func (api *Api) AdminStripeSubscriptions(ctx context.Context,
	input *StripeSubscriptionListParams,
) (*ApiPaginatedOutput[*StripeSubscription], error) {
	filter, err := ToStripeSubscriptionListFilter(input)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid filter parameters", err)
	}

	subscriptions, err := api.App().Adapter().Subscription().ListSubscriptions(ctx, filter)
	if err != nil {
		return nil, err
	}

	count, err := api.App().Adapter().Subscription().CountSubscriptions(ctx, filter)
	if err != nil {
		return nil, err
	}
	return &ApiPaginatedOutput[*StripeSubscription]{
		Body: ApiPaginatedResponse[*StripeSubscription]{
			Data: mapper.Map(subscriptions, FromModelSubscription),
			Meta: ApiGenerateMeta(&input.PaginatedInput, count),
		},
	}, nil
}

func (api *Api) AdminStripeSubscriptionsGet(ctx context.Context,
	input *StripeSubscriptionGetParams,
) (*struct{ Body *StripeSubscription }, error) {
	if input == nil || input.SubscriptionID == "" {
		return nil, huma.Error400BadRequest("subscription_id is required")
	}
	subscriptions, err := api.App().Adapter().Subscription().ListSubscriptions(ctx, &stores.StripeSubscriptionListFilter{
		Ids: []string{input.SubscriptionID},
		PaginatedInput: stores.PaginatedInput{
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
	return &struct{ Body *StripeSubscription }{Body: FromModelSubscription(subscriptions[0])}, nil
}

func ToStripeProductListFilter(input *StripeProductListParams) (*stores.StripeProductFilter, error) {
	filter := &stores.StripeProductFilter{}
	filter.Page = input.Page
	filter.PerPage = input.PerPage
	filter.Ids = input.Ids
	filter.Active = input.Active
	filter.Q = input.Q
	return filter, nil
}

func (api *Api) AdminStripeProducts(ctx context.Context,
	input *StripeProductListParams,
) (*ApiPaginatedOutput[*StripeProduct], error) {
	filter, err := ToStripeProductListFilter(input)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid filter parameters", err)
	}
	products, err := api.App().Adapter().Product().ListProducts(ctx, filter)
	if err != nil {
		return nil, err
	}
	productIds := mapper.Map(products, func(p *models.StripeProduct) string {
		return p.ID
	})
	if slices.Contains(input.Expand, "prices") {
		data, err := api.App().Adapter().Price().LoadPricesByProductIds(ctx, productIds...)
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
		data, err := api.App().Adapter().Rbac().LoadProductPermissions(ctx, productIds...)
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
	count, err := api.App().Adapter().Product().CountProducts(ctx, filter)
	if err != nil {
		return nil, err
	}
	return &ApiPaginatedOutput[*StripeProduct]{
		Body: ApiPaginatedResponse[*StripeProduct]{
			Data: mapper.Map(products, FromModelProduct),
			Meta: ApiGenerateMeta(&input.PaginatedInput, count),
		},
	}, nil
}

func (api *Api) AdminStripeProductsGet(ctx context.Context,
	input *StripeProductGetParams,
) (*struct {
	Body *StripeProduct
}, error) {

	if input == nil || input.ProductID == "" {
		return nil, huma.Error400BadRequest("product_id is required")
	}
	product, err := api.App().Adapter().Product().FindProductById(ctx, input.ProductID)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, nil
	}

	if slices.Contains(input.Expand, "prices") {
		prices, err := api.App().Adapter().Price().LoadPricesByProductIds(ctx, input.ProductID)
		if err != nil {
			return nil, err
		}
		if len(prices) > 0 {
			product.Prices = prices[0]
		}
	}
	if slices.Contains(input.Expand, "permissions") {
		roles, err := api.App().Adapter().Rbac().LoadProductPermissions(ctx, input.ProductID)
		if err != nil {
			return nil, err
		}
		if len(roles) > 0 {
			product.Permissions = roles[0]
		}
	}
	return &struct {
		Body *StripeProduct
	}{
		Body: FromModelProduct(product),
	}, nil
}

func (api *Api) AdminStripeProductsPermissionsCreate(ctx context.Context, input *struct {
	ProductID string `path:"product-id" required:"true"`
	Body      PermissionIdsInput
}) (*struct{}, error) {

	id := input.ProductID
	user, err := api.App().Adapter().Product().FindProductById(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, huma.Error404NotFound("Product not found")
	}

	permissionIds := utils.ParseValidUUIDs(input.Body.PermissionIDs...)

	err = api.App().Adapter().Rbac().CreateProductPermissions(ctx, user.ID, permissionIds...)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (api *Api) AdminStripeProductsPermissionDelete(ctx context.Context, input *struct {
	ProductID    string `path:"product-id" required:"true"`
	PermissionID string `path:"permission-id" required:"true" format:"uuid"`
}) (*struct{}, error) {
	id := input.ProductID
	product, err := api.App().Adapter().Product().FindProductById(ctx, id)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, huma.Error404NotFound("Product not found")
	}

	permissionId, err := uuid.Parse(input.PermissionID)
	if err != nil {
		return nil, huma.Error400BadRequest("role_id is not a valid UUID")
	}

	permission, err := api.App().Adapter().Rbac().FindPermission(ctx, &stores.PermissionFilter{
		Ids: []uuid.UUID{permissionId},
	})
	if err != nil {
		return nil, err
	}
	if permission == nil {
		return nil, huma.Error404NotFound("Permission not found")
	}
	err = api.App().Adapter().Rbac().DeleteProductPermissions(ctx, product.ID, permission.ID)
	if err != nil {
		return nil, err
	}
	return nil, nil
}
