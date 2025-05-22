package services

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	stripe "github.com/stripe/stripe-go/v82"
	"github.com/tkahng/authgo/internal/conf"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/types"
)

type mockPaymentStore struct{ mock.Mock }

// CreateProductRoles implements PaymentStore.
func (m *mockPaymentStore) CreateProductRoles(ctx context.Context, productId string, roleIds ...uuid.UUID) error {
	args := m.Called(ctx, productId, roleIds)
	return args.Error(0)
}

// CountCustomers implements PaymentStore.
func (m *mockPaymentStore) CountCustomers(ctx context.Context, filter *shared.StripeCustomerListFilter) (int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(int64), args.Error(1)
}

// CountPrices implements PaymentStore.
func (m *mockPaymentStore) CountPrices(ctx context.Context, filter *shared.StripePriceListFilter) (int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(int64), args.Error(1)
}

// CountSubscriptions implements PaymentStore.
func (m *mockPaymentStore) CountSubscriptions(ctx context.Context, filter *shared.StripeSubscriptionListFilter) (int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(int64), args.Error(1)
}

// ListCustomers implements PaymentStore.
func (m *mockPaymentStore) ListCustomers(ctx context.Context, input *shared.StripeCustomerListParams) ([]*models.StripeCustomer, error) {
	args := m.Called(ctx, input)
	if args.Get(0) != nil {
		return args.Get(0).([]*models.StripeCustomer), args.Error(1)
	}
	return nil, args.Error(1)
}

// ListSubscriptions implements PaymentStore.
func (m *mockPaymentStore) ListSubscriptions(ctx context.Context, input *shared.StripeSubscriptionListParams) ([]*models.StripeSubscription, error) {
	args := m.Called(ctx, input)
	if args.Get(0) != nil {
		return args.Get(0).([]*models.StripeSubscription), args.Error(1)
	}
	return nil, args.Error(1)
}

// LoadProductRoles implements PaymentStore.
func (m *mockPaymentStore) LoadProductRoles(ctx context.Context, productIds ...string) ([][]*models.Role, error) {
	args := m.Called(ctx, productIds)
	if args.Get(0) != nil {
		return args.Get(0).([][]*models.Role), args.Error(1)
	}
	return nil, args.Error(1)
}

// CountProducts implements PaymentStore.
func (m *mockPaymentStore) CountProducts(ctx context.Context, filter *shared.StripeProductListFilter) (int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(int64), args.Error(1)
}

// LoadProductPrices implements PaymentStore.
func (m *mockPaymentStore) LoadProductPrices(ctx context.Context, where *map[string]any, productIds ...string) ([][]*models.StripePrice, error) {
	args := m.Called(ctx, where, productIds)
	if args.Get(0) != nil {
		return args.Get(0).([][]*models.StripePrice), args.Error(1)
	}
	return nil, args.Error(1)
}

// UpsertCustomerStripeId implements PaymentStore.
func (m *mockPaymentStore) UpsertCustomerStripeId(ctx context.Context, customer *models.StripeCustomer) error {
	args := m.Called(ctx, customer)
	return args.Error(0)
}

// CreateCustomer implements PaymentStore.
func (m *mockPaymentStore) CreateCustomer(ctx context.Context, customer *models.StripeCustomer) (*models.StripeCustomer, error) {
	args := m.Called(ctx, customer)
	if args.Get(0) != nil {
		return args.Get(0).(*models.StripeCustomer), args.Error(1)
	}
	return nil, args.Error(1)
}

// FindCustomer implements PaymentStore.
func (m *mockPaymentStore) FindCustomer(ctx context.Context, customer *models.StripeCustomer) (*models.StripeCustomer, error) {
	args := m.Called(ctx, customer)
	if args.Get(0) != nil {
		return args.Get(0).(*models.StripeCustomer), args.Error(1)
	}
	return nil, args.Error(1)
}

// CreateProductPermissions implements PaymentStore.
func (m *mockPaymentStore) CreateProductPermissions(ctx context.Context, productId string, permissionIds ...uuid.UUID) error {
	args := m.Called(ctx, productId, permissionIds)
	return args.Error(0)
}

// FindLatestActiveSubscriptionWithPriceByCustomerId implements PaymentStore.
func (m *mockPaymentStore) FindLatestActiveSubscriptionWithPriceByCustomerId(ctx context.Context, customerId string) (*models.SubscriptionWithPrice, error) {
	args := m.Called(ctx, customerId)
	if args.Get(0) != nil {
		return args.Get(0).(*models.SubscriptionWithPrice), args.Error(1)
	}
	return nil, args.Error(1)
}

// FindPermissionByName implements PaymentStore.
func (m *mockPaymentStore) FindPermissionByName(ctx context.Context, name string) (*models.Permission, error) {
	args := m.Called(ctx, name)
	if args.Get(0) != nil {
		return args.Get(0).(*models.Permission), args.Error(1)
	}
	return nil, args.Error(1)
}

// FindProductByStripeId implements PaymentStore.
func (m *mockPaymentStore) FindProductByStripeId(ctx context.Context, productId string) (*models.StripeProduct, error) {
	args := m.Called(ctx, productId)
	if args.Get(0) != nil {
		return args.Get(0).(*models.StripeProduct), args.Error(1)
	}
	return nil, args.Error(1)
}

// FindSubscriptionWithPriceById implements PaymentStore.
func (m *mockPaymentStore) FindSubscriptionWithPriceById(ctx context.Context, stripeId string) (*models.SubscriptionWithPrice, error) {
	args := m.Called(ctx, stripeId)
	if args.Get(0) != nil {
		return args.Get(0).(*models.SubscriptionWithPrice), args.Error(1)
	}
	return nil, args.Error(1)
}

// FindTeamByStripeCustomerId implements PaymentStore.
func (m *mockPaymentStore) FindTeamByStripeCustomerId(ctx context.Context, stripeCustomerId string) (*models.Team, error) {
	args := m.Called(ctx, stripeCustomerId)
	if args.Get(0) != nil {
		return args.Get(0).(*models.Team), args.Error(1)
	}
	return nil, args.Error(1)
}

// FindValidPriceById implements PaymentStore.
func (m *mockPaymentStore) FindValidPriceById(ctx context.Context, priceId string) (*models.StripePrice, error) {
	args := m.Called(ctx, priceId)
	if args.Get(0) != nil {
		return args.Get(0).(*models.StripePrice), args.Error(1)
	}
	return nil, args.Error(1)
}

// IsFirstSubscription implements PaymentStore.
func (m *mockPaymentStore) IsFirstSubscription(ctx context.Context, customerId string) (bool, error) {
	args := m.Called(ctx, customerId)
	return args.Get(0).(bool), args.Error(1)
}

// ListPrices implements PaymentStore.
func (m *mockPaymentStore) ListPrices(ctx context.Context, input *shared.StripePriceListParams) ([]*models.StripePrice, error) {
	args := m.Called(ctx, input)
	if args.Get(0) != nil {
		return args.Get(0).([]*models.StripePrice), args.Error(1)
	}
	return nil, args.Error(1)
}

// ListProducts implements PaymentStore.
func (m *mockPaymentStore) ListProducts(ctx context.Context, input *shared.StripeProductListParams) ([]*models.StripeProduct, error) {
	args := m.Called(ctx, input)
	if args.Get(0) != nil {
		return args.Get(0).([]*models.StripeProduct), args.Error(1)
	}
	return nil, args.Error(1)
}

// UpsertPrice implements PaymentStore.
func (m *mockPaymentStore) UpsertPrice(ctx context.Context, price *models.StripePrice) error {
	args := m.Called(ctx, price)
	return args.Error(0)
}

// UpsertPriceFromStripe implements PaymentStore.
func (m *mockPaymentStore) UpsertPriceFromStripe(ctx context.Context, price *stripe.Price) error {
	args := m.Called(ctx, price)
	return args.Error(0)
}

// UpsertProduct implements PaymentStore.
func (m *mockPaymentStore) UpsertProduct(ctx context.Context, product *models.StripeProduct) error {
	args := m.Called(ctx, product)
	return args.Error(0)
}

// UpsertProductFromStripe implements PaymentStore.
func (m *mockPaymentStore) UpsertProductFromStripe(ctx context.Context, product *stripe.Product) error {
	args := m.Called(ctx, product)
	return args.Error(0)
}

// UpsertSubscription implements PaymentStore.
func (m *mockPaymentStore) UpsertSubscription(ctx context.Context, sub *models.StripeSubscription) error {
	args := m.Called(ctx, sub)
	return args.Error(0)
}

// UpsertSubscriptionFromStripe implements PaymentStore.
func (m *mockPaymentStore) UpsertSubscriptionFromStripe(ctx context.Context, sub *stripe.Subscription) error {
	args := m.Called(ctx, sub)
	return args.Error(0)
}

var _ PaymentStore = (*mockPaymentStore)(nil)

type mockPaymentClient struct{ mock.Mock }

// Config implements PaymentClient.
func (m *mockPaymentClient) Config() *conf.StripeConfig {
	args := m.Called()
	if args.Get(0) != nil {
		return args.Get(0).(*conf.StripeConfig)
	}
	return nil
}

// CreateBillingPortalSession implements PaymentClient.
func (m *mockPaymentClient) CreateBillingPortalSession(customerId string, configurationId string) (*stripe.BillingPortalSession, error) {
	args := m.Called(customerId, configurationId)
	if args.Get(0) != nil {
		return args.Get(0).(*stripe.BillingPortalSession), args.Error(1)
	}
	return nil, args.Error(1)
}

// CreateCheckoutSession implements PaymentClient.
func (m *mockPaymentClient) CreateCheckoutSession(customerId string, priceId string, quantity int64, trialDays *int64) (*stripe.CheckoutSession, error) {
	args := m.Called(customerId, priceId, quantity, trialDays)
	if args.Get(0) != nil {
		return args.Get(0).(*stripe.CheckoutSession), args.Error(1)
	}
	return nil, args.Error(1)
}

// CreateCustomer implements PaymentClient.
func (m *mockPaymentClient) CreateCustomer(email string, name *string) (*stripe.Customer, error) {
	args := m.Called(email, name)
	if args.Get(0) != nil {
		return args.Get(0).(*stripe.Customer), args.Error(1)
	}
	return nil, args.Error(1)
}

// CreatePortalConfiguration implements PaymentClient.
func (m *mockPaymentClient) CreatePortalConfiguration(input ...*stripe.BillingPortalConfigurationFeaturesSubscriptionUpdateProductParams) (string, error) {
	args := m.Called(input)
	return args.String(0), args.Error(1)
}

// FindAllPrices implements PaymentClient.
func (m *mockPaymentClient) FindAllPrices() ([]*stripe.Price, error) {
	args := m.Called()
	if args.Get(0) != nil {
		return args.Get(0).([]*stripe.Price), args.Error(1)
	}
	return nil, args.Error(1)
}

// FindAllProducts implements PaymentClient.
func (m *mockPaymentClient) FindAllProducts() ([]*stripe.Product, error) {
	args := m.Called()
	if args.Get(0) != nil {
		return args.Get(0).([]*stripe.Product), args.Error(1)
	}
	return nil, args.Error(1)
}

// FindCheckoutSessionByStripeId implements PaymentClient.
func (m *mockPaymentClient) FindCheckoutSessionByStripeId(stripeId string) (*stripe.CheckoutSession, error) {
	args := m.Called(stripeId)
	if args.Get(0) != nil {
		return args.Get(0).(*stripe.CheckoutSession), args.Error(1)
	}
	return nil, args.Error(1)
}

// FindOrCreateCustomer implements PaymentClient.
func (m *mockPaymentClient) FindOrCreateCustomer(email string, name *string) (*stripe.Customer, error) {
	args := m.Called(email, name)
	if args.Get(0) != nil {
		return args.Get(0).(*stripe.Customer), args.Error(1)
	}
	return nil, args.Error(1)
}

// FindSubscriptionByStripeId implements PaymentClient.
func (m *mockPaymentClient) FindSubscriptionByStripeId(stripeId string) (*stripe.Subscription, error) {
	args := m.Called(stripeId)
	if args.Get(0) != nil {
		return args.Get(0).(*stripe.Subscription), args.Error(1)
	}
	return nil, args.Error(1)
}

// UpdateCustomer implements PaymentClient.
func (m *mockPaymentClient) UpdateCustomer(customerId string, params *stripe.CustomerParams) (*stripe.Customer, error) {
	args := m.Called(customerId, params)
	if args.Get(0) != nil {
		return args.Get(0).(*stripe.Customer), args.Error(1)
	}
	return nil, args.Error(1)
}

// UpdateItemQuantity implements PaymentClient.
func (m *mockPaymentClient) UpdateItemQuantity(itemId string, priceId string, count int64) (*stripe.SubscriptionItem, error) {
	args := m.Called(itemId, priceId, count)
	if args.Get(0) != nil {
		return args.Get(0).(*stripe.SubscriptionItem), args.Error(1)
	}
	return nil, args.Error(1)
}

var _ PaymentClient = (*mockPaymentClient)(nil)

func (m *mockPaymentStore) FindLatestActiveSubscriptionByTeamId(ctx context.Context, teamId uuid.UUID) (*models.StripeSubscription, error) {
	args := m.Called(ctx, teamId)
	if args.Get(0) != nil {
		return args.Get(0).(*models.StripeSubscription), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *mockPaymentStore) CountTeamMembers(ctx context.Context, teamId uuid.UUID) (int64, error) {
	args := m.Called(ctx, teamId)
	return args.Get(0).(int64), args.Error(1)
}

func TestStripeService_VerifyAndUpdateTeamSubscriptionQuantity(t *testing.T) {
	ctx := context.Background()
	teamId := uuid.New()
	customer := &models.StripeCustomer{
		ID:           "stripe_customer",
		TeamID:       types.Pointer(teamId),
		CustomerType: models.StripeCustomerTypeTeam,
	}

	product := &models.StripeProduct{
		ID: "product1",
	}
	price := &models.StripePrice{
		ID:        "price1",
		ProductID: "product1",
	}
	sub := &models.StripeSubscription{
		ItemID:           "item1",
		PriceID:          "price1",
		Quantity:         2,
		StripeCustomerID: customer.ID,
	}
	subwithprice := &models.SubscriptionWithPrice{
		Price:        *price,
		Product:      *product,
		Subscription: *sub,
	}

	t.Run("updates quantity if different", func(t *testing.T) {
		store := new(mockPaymentStore)
		client := new(mockPaymentClient)
		service := &StripeService{paymentStore: store, client: client}
		store.On("FindCustomer", ctx, mock.Anything).Return(customer, nil)
		store.On("FindLatestActiveSubscriptionWithPriceByCustomerId", ctx, customer.ID).Return(subwithprice, nil)
		store.On("CountTeamMembers", ctx, teamId).Return(int64(3), nil)
		client.On("UpdateItemQuantity", sub.ItemID, sub.PriceID, int64(3)).Return(nil, nil)
		err := service.VerifyAndUpdateTeamSubscriptionQuantity(ctx, teamId)
		assert.NoError(t, err)
		store.AssertExpectations(t)
		client.AssertExpectations(t)
	})

	t.Run("no update if quantity matches", func(t *testing.T) {
		store := new(mockPaymentStore)
		client := new(mockPaymentClient)
		service := &StripeService{paymentStore: store, client: client}
		store.On("FindCustomer", ctx, mock.Anything).Return(customer, nil)
		store.On("FindLatestActiveSubscriptionWithPriceByCustomerId", ctx, customer.ID).Return(subwithprice, nil)
		store.On("CountTeamMembers", ctx, teamId).Return(int64(2), nil)
		err := service.VerifyAndUpdateTeamSubscriptionQuantity(ctx, teamId)
		assert.NoError(t, err)
		store.AssertExpectations(t)
	})

	t.Run("returns error if no subscription", func(t *testing.T) {
		store := new(mockPaymentStore)
		client := new(mockPaymentClient)
		service := &StripeService{paymentStore: store, client: client}
		store.On("FindCustomer", ctx, mock.Anything).Return(customer, nil)
		store.On("FindLatestActiveSubscriptionWithPriceByCustomerId", ctx, customer.ID).Return(nil, nil)
		err := service.VerifyAndUpdateTeamSubscriptionQuantity(ctx, teamId)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no subscription")
		store.AssertExpectations(t)
	})

	t.Run("returns error if store fails", func(t *testing.T) {
		store := new(mockPaymentStore)
		client := new(mockPaymentClient)
		service := &StripeService{paymentStore: store, client: client}
		store.On("FindCustomer", ctx, mock.Anything).Return(nil, errors.New("db error"))
		err := service.VerifyAndUpdateTeamSubscriptionQuantity(ctx, teamId)
		assert.Error(t, err)
		store.AssertExpectations(t)
	})
}

func TestStripeService_CreateTeamCustomer(t *testing.T) {
	ctx := context.Background()
	team := &models.Team{ID: uuid.New(), Name: "Test Team"}
	user := &models.User{ID: uuid.New(), Email: "user@example.com"}
	customer := &stripe.Customer{ID: "cus_123", Email: user.Email}
	created := &models.StripeCustomer{ID: customer.ID, Email: customer.Email, Name: &team.Name, TeamID: types.Pointer(team.ID), CustomerType: models.StripeCustomerTypeTeam}

	t.Run("success", func(t *testing.T) {
		store := new(mockPaymentStore)
		client := new(mockPaymentClient)
		service := &StripeService{paymentStore: store, client: client}
		client.On("CreateCustomer", user.Email, &team.Name).Return(customer, nil)
		store.On("CreateCustomer", ctx, mock.AnythingOfType("*models.StripeCustomer")).Return(created, nil)
		result, err := service.CreateTeamCustomer(ctx, team, user)
		assert.NoError(t, err)
		assert.Equal(t, created, result)
		client.AssertExpectations(t)
		store.AssertExpectations(t)
	})

	t.Run("client error", func(t *testing.T) {
		store := new(mockPaymentStore)
		client := new(mockPaymentClient)
		service := &StripeService{paymentStore: store, client: client}
		client.On("CreateCustomer", user.Email, &team.Name).Return(nil, errors.New("stripe error"))
		result, err := service.CreateTeamCustomer(ctx, team, user)
		assert.Error(t, err)
		assert.Nil(t, result)
		client.AssertExpectations(t)
	})

	t.Run("nil customer", func(t *testing.T) {
		store := new(mockPaymentStore)
		client := new(mockPaymentClient)
		service := &StripeService{paymentStore: store, client: client}
		client.On("CreateCustomer", user.Email, &team.Name).Return(nil, nil)
		result, err := service.CreateTeamCustomer(ctx, team, user)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no customer found")
		assert.Nil(t, result)
		client.AssertExpectations(t)
	})
}

func TestStripeService_CreateUserCustomer(t *testing.T) {
	ctx := context.Background()
	user := &models.User{ID: uuid.New(), Email: "user@example.com", Name: types.Pointer("User Name")}
	customer := &stripe.Customer{ID: "cus_456", Email: user.Email}
	created := &models.StripeCustomer{ID: customer.ID, Email: customer.Email, Name: user.Name, UserID: types.Pointer(user.ID), CustomerType: models.StripeCustomerTypeUser}

	t.Run("success", func(t *testing.T) {
		store := new(mockPaymentStore)
		client := new(mockPaymentClient)
		service := &StripeService{paymentStore: store, client: client}
		client.On("CreateCustomer", user.Email, user.Name).Return(customer, nil)
		store.On("CreateCustomer", ctx, mock.AnythingOfType("*models.StripeCustomer")).Return(created, nil)
		result, err := service.CreateUserCustomer(ctx, user)
		assert.NoError(t, err)
		assert.Equal(t, created, result)
		client.AssertExpectations(t)
		store.AssertExpectations(t)
	})

	t.Run("client error", func(t *testing.T) {
		store := new(mockPaymentStore)
		client := new(mockPaymentClient)
		service := &StripeService{paymentStore: store, client: client}
		client.On("CreateCustomer", user.Email, user.Name).Return(nil, errors.New("stripe error"))
		result, err := service.CreateUserCustomer(ctx, user)
		assert.Error(t, err)
		assert.Nil(t, result)
		client.AssertExpectations(t)
	})

	t.Run("nil customer", func(t *testing.T) {
		store := new(mockPaymentStore)
		client := new(mockPaymentClient)
		service := &StripeService{paymentStore: store, client: client}
		client.On("CreateCustomer", user.Email, user.Name).Return(nil, nil)
		result, err := service.CreateUserCustomer(ctx, user)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no customer found")
		assert.Nil(t, result)
		client.AssertExpectations(t)
	})
}

func TestStripeService_FindCustomerByTeam(t *testing.T) {
	ctx := context.Background()
	teamId := uuid.New()
	customer := &models.StripeCustomer{ID: "cus_789", TeamID: types.Pointer(teamId)}

	t.Run("success", func(t *testing.T) {
		store := new(mockPaymentStore)
		client := new(mockPaymentClient)
		service := &StripeService{paymentStore: store, client: client}
		store.On("FindCustomer", ctx, mock.AnythingOfType("*models.StripeCustomer")).Return(customer, nil)
		result, err := service.FindCustomerByTeam(ctx, teamId)
		assert.NoError(t, err)
		assert.Equal(t, customer, result)
		store.AssertExpectations(t)
	})

	t.Run("store error", func(t *testing.T) {
		store := new(mockPaymentStore)
		client := new(mockPaymentClient)
		service := &StripeService{paymentStore: store, client: client}
		store.On("FindCustomer", ctx, mock.AnythingOfType("*models.StripeCustomer")).Return(nil, errors.New("db error"))
		result, err := service.FindCustomerByTeam(ctx, teamId)
		assert.Error(t, err)
		assert.Nil(t, result)
		store.AssertExpectations(t)
	})
}

func TestStripeService_FindCustomerByUser(t *testing.T) {
	ctx := context.Background()
	userId := uuid.New()
	customer := &models.StripeCustomer{ID: "cus_101", UserID: types.Pointer(userId)}

	t.Run("success", func(t *testing.T) {
		store := new(mockPaymentStore)
		client := new(mockPaymentClient)
		service := &StripeService{paymentStore: store, client: client}
		store.On("FindCustomer", ctx, mock.AnythingOfType("*models.StripeCustomer")).Return(customer, nil)
		result, err := service.FindCustomerByUser(ctx, userId)
		assert.NoError(t, err)
		assert.Equal(t, customer, result)
		store.AssertExpectations(t)
	})

	t.Run("store error", func(t *testing.T) {
		store := new(mockPaymentStore)
		client := new(mockPaymentClient)
		service := &StripeService{paymentStore: store, client: client}
		store.On("FindCustomer", ctx, mock.AnythingOfType("*models.StripeCustomer")).Return(nil, errors.New("db error"))
		result, err := service.FindCustomerByUser(ctx, userId)
		assert.Error(t, err)
		assert.Nil(t, result)
		store.AssertExpectations(t)
	})
}
