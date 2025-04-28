package apis

import (
	"context"
	"net/http"
	"slices"

	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/authgo/internal/db/models"
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
	subscriptions, err := repository.ListSubscriptions(ctx, db, input)
	if err != nil {
		return nil, err
	}
	if slices.Contains(input.Expand, "user") {
		err = subscriptions.LoadStripeSubscriptionUser(ctx, db)
		if err != nil {
			return nil, err
		}
	}
	if slices.Contains(input.Expand, "price") {
		if slices.Contains(input.Expand, "product") {
			err = subscriptions.LoadStripeSubscriptionPriceStripePrice(ctx, db,
				models.PreloadStripePriceProductStripeProduct(),
			)
			if err != nil {
				return nil, err
			}
		} else {
			err = subscriptions.LoadStripeSubscriptionPriceStripePrice(ctx, db)
			if err != nil {
				return nil, err
			}
		}
	}
	subs := mapper.Map(subscriptions, func(sub *models.StripeSubscription) *shared.SubscriptionWithData {
		ss := &shared.SubscriptionWithData{
			Subscription: shared.ModelToSubscription(sub),
		}
		if sub.R.User != nil {
			ss.SubscriptionUser = shared.ToUser(sub.R.User)
		}
		if sub.R.PriceStripePrice != nil {
			ss.Price = &shared.StripePricesWithProduct{
				Price: shared.ModelToPrice(sub.R.PriceStripePrice),
			}
			if sub.R.PriceStripePrice.R.ProductStripeProduct != nil {
				ss.Price.Product = shared.ModelToProduct(sub.R.PriceStripePrice.R.ProductStripeProduct)
			}
		}
		return ss
	})
	count, err := repository.CountSubscriptions(ctx, db, &input.StripeSubscriptionListFilter)
	if err != nil {
		return nil, err
	}
	return &shared.PaginatedOutput[*shared.SubscriptionWithData]{
		Body: shared.PaginatedResponse[*shared.SubscriptionWithData]{
			Data: subs,
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
	subscription, err := repository.FindSubscriptionById(ctx, db, input.SubscriptionID)
	if err != nil {
		return nil, err
	}
	if subscription == nil {
		return nil, huma.Error404NotFound("subscription not found")
	}
	if slices.Contains(input.Expand, "user") {
		err = subscription.LoadStripeSubscriptionUser(ctx, db)
		if err != nil {
			return nil, err
		}
	}
	if slices.Contains(input.Expand, "price") {
		if slices.Contains(input.Expand, "product") {

			err = subscription.LoadStripeSubscriptionPriceStripePrice(ctx, db,
				models.PreloadStripePriceProductStripeProduct(),
			)
			if err != nil {
				return nil, err
			}
		} else {
			err = subscription.LoadStripeSubscriptionPriceStripePrice(ctx, db)
			if err != nil {
				return nil, err
			}
		}
	}
	ss := &shared.SubscriptionWithData{
		Subscription: shared.ModelToSubscription(subscription),
	}
	if subscription.R.User != nil {
		ss.SubscriptionUser = shared.ToUser(subscription.R.User)
	}
	if subscription.R.PriceStripePrice != nil {
		ss.Price = &shared.StripePricesWithProduct{
			Price: shared.ModelToPrice(subscription.R.PriceStripePrice),
		}
		if subscription.R.PriceStripePrice.R.ProductStripeProduct != nil {
			ss.Price.Product = shared.ModelToProduct(subscription.R.PriceStripePrice.R.ProductStripeProduct)
		}
	}
	return &struct{ Body *shared.SubscriptionWithData }{Body: ss}, nil
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
	products, err := repository.ListProducts(ctx, db, input)
	if err != nil {
		return nil, err
	}
	if slices.Contains(input.Expand, "prices") {
		err = products.LoadStripeProductProductStripePrices(ctx, db)
		if err != nil {
			return nil, err
		}
	}
	if slices.Contains(input.Expand, "roles") {
		err = products.LoadStripeProductRoles(ctx, db)
		if err != nil {
			return nil, err
		}
	}
	prods := mapper.Map(products, func(p *models.StripeProduct) *shared.StripeProductWithData {
		spwd := &shared.StripeProductWithData{
			Product: shared.ModelToProduct(p),
		}
		if p.R.ProductStripePrices != nil {
			// If the product has prices, we map them to the shared model
			// and include them in the response.
			spwd.Prices = mapper.Map(p.R.ProductStripePrices, shared.ModelToPrice)
		}
		if p.R.Roles != nil {
			// If the product has prices and we are expanding prices,
			spwd.Roles = mapper.Map(p.R.Roles, shared.ToRole)
		}
		return spwd
	})
	count, err := repository.CountProducts(ctx, db, &input.StripeProductListFilter)
	if err != nil {
		return nil, err
	}
	return &shared.PaginatedOutput[*shared.StripeProductWithData]{
		Body: shared.PaginatedResponse[*shared.StripeProductWithData]{
			Data: prods,
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
	product, err := repository.FindProductByStripeId(ctx, db, input.ProductID)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, huma.Error404NotFound("product not found")
	}
	if slices.Contains(input.Expand, "prices") {
		err = product.LoadStripeProductProductStripePrices(ctx, db)
		if err != nil {
			return nil, err
		}
	}
	if slices.Contains(input.Expand, "roles") {
		err = product.LoadStripeProductRoles(ctx, db)
		if err != nil {
			return nil, err
		}
	}
	spwd := &shared.StripeProductWithData{
		Product: shared.ModelToProduct(product),
	}
	if product.R.ProductStripePrices != nil {
		spwd.Prices = mapper.Map(product.R.ProductStripePrices, shared.ModelToPrice)
	}
	if product.R.Roles != nil {
		spwd.Roles = mapper.Map(product.R.Roles, shared.ToRole)
	}
	return &struct{ Body *shared.StripeProductWithData }{
		Body: spwd,
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
	user, err := repository.FindProductByStripeId(ctx, db, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, huma.Error404NotFound("Product not found")
	}

	roleIds := repository.ParseUUIDs(input.Body.RolesIds)

	roles, err := repository.FindRolesByIds(ctx, db, roleIds)
	if err != nil {
		return nil, err
	}

	err = user.AttachRoles(ctx, db, roles...)
	if err != nil {
		return nil, err
	}
	return nil, nil
}
