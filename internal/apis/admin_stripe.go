package apis

import (
	"context"
	"net/http"
	"slices"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/crud/crudModels"
	"github.com/tkahng/authgo/internal/db/models"
	"github.com/tkahng/authgo/internal/queries"
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
	count, err := queries.CountSubscriptions(ctx, db, &input.StripeSubscriptionListFilter)
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
	var roles map[string][]*crudModels.Role = make(map[string][]*crudModels.Role)
	var prices map[string][]*crudModels.StripePrice = make(map[string][]*crudModels.StripePrice)
	var productIds []string
	for _, p := range products {
		productIds = append(productIds, p.ID)
	}
	if slices.Contains(input.Expand, "prices") {
		data, err := queries.LoadProductPrices(ctx, db, productIds...)
		if err != nil {
			return nil, err
		}
		for _, d := range data {
			prices[d.Key] = d.Data
		}
	}
	if slices.Contains(input.Expand, "roles") {
		data, err := queries.LoadProductRoles(ctx, db, productIds...)
		if err != nil {
			return nil, err
		}
		for _, d := range data {
			roles[d.Key] = d.Data
		}
	}
	prods := mapper.Map(products, func(p *crudModels.StripeProduct) *shared.StripeProductWithData {
		spwd := &shared.StripeProductWithData{
			Product: shared.FromCrudProduct(p),
		}

		if data, ok := prices[p.ID]; ok {
			// If the product has prices, we map them to the shared model
			// and include them in the response.
			spwd.Prices = mapper.Map(data, shared.FromCrudModel)
		}
		if data, ok := roles[p.ID]; ok {

			// If the product has prices and we are expanding prices,
			spwd.Roles = mapper.Map(data, shared.FromCrudRole)
		}
		return spwd
	})
	count, err := queries.CountProducts(ctx, db, &input.StripeProductListFilter)
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
	product, err := queries.FindProductByStripeId(ctx, db, input.ProductID)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, nil
	}
	var prices []*crudModels.StripePrice
	var roles []*crudModels.Role
	if slices.Contains(input.Expand, "prices") {
		pr, err := queries.ListPrices(ctx, db, &shared.StripePriceListParams{
			StripePriceListFilter: shared.StripePriceListFilter{
				ProductIds: []string{input.ProductID},
			},
			PaginatedInput: shared.PaginatedInput{
				PerPage: 100,
			},
		})
		if err != nil {
			return nil, err
		}
		prices = pr
	}
	if slices.Contains(input.Expand, "roles") {
		rl, err := queries.ListRoles(ctx, db, &shared.RolesListParams{
			PaginatedInput: shared.PaginatedInput{
				PerPage: 100,
			},
			RoleListFilter: shared.RoleListFilter{
				ProductId: input.ProductID,
			},
		})
		if err != nil {
			return nil, err
		}
		roles = rl
	}
	spwd := &shared.StripeProductWithData{
		Product: shared.FromCrudProduct(product),
		Roles:   mapper.Map(roles, shared.FromCrudRole),
		Prices:  mapper.Map(prices, shared.FromCrudModel),
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

	role, err := queries.FindRoleById(ctx, db, roleId)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, huma.Error404NotFound("Role not found")
	}

	_, err = models.ProductRoles.Delete(
		models.DeleteWhere.ProductRoles.RoleID.EQ(role.ID),
		models.DeleteWhere.ProductRoles.ProductID.EQ(id),
	).Exec(ctx, db)
	if err != nil {
		return nil, err
	}
	return nil, nil
}
