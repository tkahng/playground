package services

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	stripe "github.com/stripe/stripe-go/v82"

	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/stores"

	"github.com/tkahng/authgo/internal/tools/types"
)

func TestStripeService_CreateTeamCustomer(t *testing.T) {
	ctx := context.Background()
	team := &models.Team{ID: uuid.New(), Name: "Test Team"}
	user := &models.User{ID: uuid.New(), Email: "user@example.com"}
	customer := &stripe.Customer{ID: "cus_123", Email: user.Email}
	created := &models.StripeCustomer{ID: customer.ID, Email: customer.Email, Name: &team.Name, TeamID: types.Pointer(team.ID), CustomerType: models.StripeCustomerTypeTeam}

	t.Run("success", func(t *testing.T) {
		// store := new(MockPaymentStore)
		adapter := stores.NewAdapterDecorators()
		client := NewTestPaymentClient()
		service := &StripeService{client: client, adapter: adapter}

		// client.On("CreateCustomer", user.Email, &team.Name).Return(customer, nil)
		client.CreateCustomerFunc = func(email string, name *string) (*stripe.Customer, error) {
			return customer, nil
		}
		adapter.CustomerFunc.CreateCustomerFunc = func(ctx context.Context, customer *models.StripeCustomer) (*models.StripeCustomer, error) {
			return created, nil
		}
		// store.On("CreateCustomer", ctx, mock.AnythingOfType("*models.StripeCustomer")).Return(created, nil)
		result, err := service.CreateTeamCustomer(ctx, team, user)
		assert.NoError(t, err)
		assert.Equal(t, created, result)
	})

	t.Run("client error", func(t *testing.T) {
		// store := new(MockPaymentStore)
		client := NewTestPaymentClient()
		service := &StripeService{client: client}
		// client.On("CreateCustomer", user.Email, &team.Name).Return(nil, errors.New("stripe error"))
		client.CreateCustomerFunc = func(email string, name *string) (*stripe.Customer, error) {
			return nil, errors.New("stripe error")
		}
		// store.On("CreateCustomer", ctx, mock.AnythingOfType("*models.StripeCustomer")).Return(nil, errors.New("stripe error"))
		result, err := service.CreateTeamCustomer(ctx, team, user)
		assert.Error(t, err)
		assert.Nil(t, result)

	})

}

func TestStripeService_CreateUserCustomer(t *testing.T) {
	ctx := context.Background()
	user := &models.User{ID: uuid.New(), Email: "user@example.com", Name: types.Pointer("User Name")}
	customer := &stripe.Customer{ID: "cus_456", Email: user.Email}
	created := &models.StripeCustomer{ID: customer.ID, Email: customer.Email, Name: user.Name, UserID: types.Pointer(user.ID), CustomerType: models.StripeCustomerTypeUser}

	t.Run("success", func(t *testing.T) {
		adapter := stores.NewAdapterDecorators()
		// store := new(MockPaymentStore)
		client := NewTestPaymentClient()
		service := &StripeService{client: client, adapter: adapter}
		// client.On("CreateCustomer", user.Email, user.Name).Return(customer, nil)
		client.CreateCustomerFunc = func(email string, name *string) (*stripe.Customer, error) {
			return customer, nil
		}
		adapter.CustomerFunc.CreateCustomerFunc = func(ctx context.Context, customer *models.StripeCustomer) (*models.StripeCustomer, error) {
			return created, nil
		}
		result, err := service.CreateUserCustomer(ctx, user)
		assert.NoError(t, err)
		assert.Equal(t, created, result)

	})

	t.Run("client error", func(t *testing.T) {
		adapter := stores.NewAdapterDecorators()
		// store := new(MockPaymentStore)
		client := NewTestPaymentClient()
		service := &StripeService{client: client, adapter: adapter}
		// client.On("CreateCustomer", user.Email, user.Name).Return(nil, errors.New("stripe error"))
		client.CreateCustomerFunc = func(email string, name *string) (*stripe.Customer, error) {
			return nil, errors.New("stripe error")
		}
		// adapter.CustomerFunc.CreateCustomerFunc = func(ctx context.Context, customer *models.StripeCustomer) (*models.StripeCustomer, error) {
		// 	return nil, errors.New("stripe error")
		// }
		result, err := service.CreateUserCustomer(ctx, user)
		assert.Error(t, err)
		assert.Nil(t, result)

	})

}

func TestStripeService_FindCustomerByTeam(t *testing.T) {
	ctx := context.Background()
	teamId := uuid.New()
	customer := &models.StripeCustomer{ID: "cus_789", TeamID: types.Pointer(teamId)}

	t.Run("success", func(t *testing.T) {
		adapter := stores.NewAdapterDecorators()
		client := NewTestPaymentClient()
		service := NewPaymentService(client, adapter)
		adapter.CustomerFunc.FindCustomerFunc = func(ctx context.Context, filter *stores.StripeCustomerFilter) (*models.StripeCustomer, error) {
			return customer, nil
		}
		// store.On("FindCustomer", ctx, mock.AnythingOfType("*models.StripeCustomer")).Return(customer, nil)
		result, err := service.FindCustomerByTeam(ctx, teamId)
		assert.NoError(t, err)
		assert.Equal(t, customer, result)
	})

	t.Run("store error", func(t *testing.T) {
		adapter := stores.NewAdapterDecorators()
		// store := new(MockPaymentStore)
		client := NewTestPaymentClient()
		service := NewPaymentService(client, adapter)
		adapter.CustomerFunc.FindCustomerFunc = func(ctx context.Context, filter *stores.StripeCustomerFilter) (*models.StripeCustomer, error) {
			return nil, errors.New("db error")
		}
		// store.On("FindCustomer", ctx, mock.AnythingOfType("*models.StripeCustomer")).Return(nil, errors.New("db error"))
		result, err := service.FindCustomerByTeam(ctx, teamId)
		assert.Error(t, err)
		assert.Nil(t, result)
		// store.AssertExpectations(t)
	})
}

func TestStripeService_FindCustomerByUser(t *testing.T) {
	ctx := context.Background()
	userId := uuid.New()
	customer := &models.StripeCustomer{ID: "cus_101", UserID: types.Pointer(userId)}

	t.Run("success", func(t *testing.T) {
		adapter := stores.NewAdapterDecorators()
		// store := new(MockPaymentStore)
		client := NewTestPaymentClient()
		service := NewPaymentService(client, adapter)
		adapter.CustomerFunc.FindCustomerFunc = func(ctx context.Context, filter *stores.StripeCustomerFilter) (*models.StripeCustomer, error) {
			return customer, nil
		}
		result, err := service.FindCustomerByUser(ctx, userId)
		assert.NoError(t, err)
		assert.Equal(t, customer, result)
		// store.AssertExpectations(t)
	})

	t.Run("store error", func(t *testing.T) {
		adapter := stores.NewAdapterDecorators()
		// store := new(MockPaymentStore)
		client := NewTestPaymentClient()
		service := NewPaymentService(client, adapter)
		adapter.CustomerFunc.FindCustomerFunc = func(ctx context.Context, filter *stores.StripeCustomerFilter) (*models.StripeCustomer, error) {
			return nil, errors.New("db error")
		}
		result, err := service.FindCustomerByUser(ctx, userId)
		assert.Error(t, err)
		assert.Nil(t, result)
		// store.AssertExpectations(t)
	})
}
func TestStripeService_VerifyAndUpdateTeamSubscriptionQuantity(t *testing.T) {
	ctx := context.Background()
	teamId := uuid.New()
	customer := &models.StripeCustomer{
		ID:           "cus_test",
		TeamID:       types.Pointer(teamId),
		CustomerType: models.StripeCustomerTypeTeam,
	}
	product := &models.StripeProduct{ID: "prod_123"}
	price := &models.StripePrice{ID: "price_123", ProductID: "prod_123"}
	price.Product = product
	sub := &models.StripeSubscription{
		ItemID:           "item_123",
		PriceID:          "price_123",
		Quantity:         2,
		StripeCustomerID: customer.ID,
	}
	sub.Price = price

	t.Run("updates quantity if different", func(t *testing.T) {
		adapter := stores.NewAdapterDecorators()
		// store := new(MockPaymentStore)
		client := NewTestPaymentClient()
		service := NewPaymentService(client, adapter)
		adapter.CustomerFunc.FindCustomerFunc = func(ctx context.Context, filter *stores.StripeCustomerFilter) (*models.StripeCustomer, error) {
			return customer, nil
		}
		// store.On("FindCustomer", ctx, mock.Anything).Return(customer, nil)
		adapter.SubscriptionFunc.FindActiveSubscriptionByCustomerIdFunc = func(ctx context.Context, customerId string) (*models.StripeSubscription, error) {
			return sub, nil
		}
		// store.On("FindActiveSubscriptionByCustomerId", ctx, customer.ID).Return(sub, nil)
		adapter.TeamMemberFunc.CountTeamMembersFunc = func(ctx context.Context, filter *stores.TeamMemberFilter) (int64, error) {
			return int64(3), nil
		}

		// client.On("UpdateItemQuantity", sub.ItemID, sub.PriceID, int64(3)).Return(&stripe.SubscriptionItem{}, nil)
		client.UpdateItemQuantityFunc = func(itemID string, priceID string, quantity int64) (*stripe.SubscriptionItem, error) {
			return &stripe.SubscriptionItem{}, nil
		}
		err := service.VerifyAndUpdateTeamSubscriptionQuantity(ctx, teamId)
		assert.NoError(t, err)
		// store.AssertExpectations(t)

	})

	t.Run("no update if quantity matches", func(t *testing.T) {
		adapter := stores.NewAdapterDecorators()
		// store := new(MockPaymentStore)
		client := NewTestPaymentClient()
		service := &StripeService{client: client, adapter: adapter}
		adapter.CustomerFunc.FindCustomerFunc = func(ctx context.Context, filter *stores.StripeCustomerFilter) (*models.StripeCustomer, error) {
			return customer, nil
		}

		// store.On("FindCustomer", ctx, mock.Anything).Return(customer, nil)
		adapter.SubscriptionFunc.FindActiveSubscriptionByCustomerIdFunc = func(ctx context.Context, customerId string) (*models.StripeSubscription, error) {
			return sub, nil
		}
		// store.On("FindActiveSubscriptionByCustomerId", ctx, customer.ID).Return(sub, nil)
		adapter.TeamMemberFunc.CountTeamMembersFunc = func(ctx context.Context, filter *stores.TeamMemberFilter) (int64, error) {
			return int64(2), nil
		}
		// store.On("CountTeamMembers", ctx, teamId).Return(int64(2), nil)
		err := service.VerifyAndUpdateTeamSubscriptionQuantity(ctx, teamId)
		assert.NoError(t, err)
		// store.AssertExpectations(t)
	})

	t.Run("returns error if no subscription", func(t *testing.T) {
		adapter := stores.NewAdapterDecorators()
		// store := new(MockPaymentStore)
		client := NewTestPaymentClient()
		service := &StripeService{client: client, adapter: adapter}
		adapter.CustomerFunc.FindCustomerFunc = func(ctx context.Context, filter *stores.StripeCustomerFilter) (*models.StripeCustomer, error) {
			return customer, nil
		}
		// store.On("FindCustomer", ctx, mock.Anything).Return(customer, nil)
		adapter.SubscriptionFunc.FindActiveSubscriptionByCustomerIdFunc = func(ctx context.Context, customerId string) (*models.StripeSubscription, error) {
			return nil, nil
		}
		// store.On("FindActiveSubscriptionByCustomerId", ctx, customer.ID).Return(nil, nil)
		err := service.VerifyAndUpdateTeamSubscriptionQuantity(ctx, teamId)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no subscription")
		// store.AssertExpectations(t)
	})

	t.Run("returns error if store fails", func(t *testing.T) {
		adapter := stores.NewAdapterDecorators()
		// store := new(MockPaymentStore)
		client := NewTestPaymentClient()
		service := &StripeService{client: client, adapter: adapter}
		adapter.CustomerFunc.FindCustomerFunc = func(ctx context.Context, filter *stores.StripeCustomerFilter) (*models.StripeCustomer, error) {
			return nil, errors.New("db error")
		}
		// store.On("FindCustomer", ctx, mock.Anything).Return(nil, errors.New("db error"))
		err := service.VerifyAndUpdateTeamSubscriptionQuantity(ctx, teamId)
		assert.Error(t, err)
		// store.AssertExpectations(t)
	})

	t.Run("returns error if no customer", func(t *testing.T) {
		adapter := stores.NewAdapterDecorators()
		// store := new(MockPaymentStore)
		client := NewTestPaymentClient()
		service := &StripeService{client: client, adapter: adapter}
		adapter.CustomerFunc.FindCustomerFunc = func(ctx context.Context, filter *stores.StripeCustomerFilter) (*models.StripeCustomer, error) {
			return nil, nil
		}
		// store.On("FindCustomer", ctx, mock.Anything).Return(nil, nil)
		err := service.VerifyAndUpdateTeamSubscriptionQuantity(ctx, teamId)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no stripe customer id")
		// store.AssertExpectations(t)
	})

	t.Run("returns error if CountTeamMembers fails", func(t *testing.T) {
		adapter := stores.NewAdapterDecorators()
		// store := new(MockPaymentStore)
		client := NewTestPaymentClient()
		service := &StripeService{client: client, adapter: adapter}
		adapter.CustomerFunc.FindCustomerFunc = func(ctx context.Context, filter *stores.StripeCustomerFilter) (*models.StripeCustomer, error) {
			return customer, nil
		}
		// store.On("FindCustomer", ctx, mock.Anything).Return(customer, nil)
		adapter.SubscriptionFunc.FindActiveSubscriptionByCustomerIdFunc = func(ctx context.Context, customerId string) (*models.StripeSubscription, error) {
			return sub, nil
		}
		//
		// store.On("FindActiveSubscriptionByCustomerId", ctx, customer.ID).Return(sub, nil)
		adapter.TeamMemberFunc.CountTeamMembersFunc = func(ctx context.Context, filter *stores.TeamMemberFilter) (int64, error) {
			return 0, errors.New("count error")
		}
		// store.On("CountTeamMembers", ctx, teamId).Return(int64(0), errors.New("count error"))
		err := service.VerifyAndUpdateTeamSubscriptionQuantity(ctx, teamId)
		assert.Error(t, err)
		// store.AssertExpectations(t)
	})

	t.Run("returns nil if team member count is zero", func(t *testing.T) {
		adapter := stores.NewAdapterDecorators()
		// store := new(MockPaymentStore)
		client := NewTestPaymentClient()
		service := &StripeService{client: client, adapter: adapter}
		adapter.CustomerFunc.FindCustomerFunc = func(ctx context.Context, filter *stores.StripeCustomerFilter) (*models.StripeCustomer, error) {
			return customer, nil
		}
		// store.On("FindCustomer", ctx, mock.Anything).Return(customer, nil)
		adapter.SubscriptionFunc.FindActiveSubscriptionByCustomerIdFunc = func(ctx context.Context, customerId string) (*models.StripeSubscription, error) {
			return sub, nil
		}
		// store.On("FindActiveSubscriptionByCustomerId", ctx, customer.ID).Return(sub, nil)
		adapter.TeamMemberFunc.CountTeamMembersFunc = func(ctx context.Context, filter *stores.TeamMemberFilter) (int64, error) {
			return int64(0), nil
		}
		// store.On("CountTeamMembers", ctx, teamId).Return(int64(0), nil)
		// store.On("FindCustomer", ctx, mock.Anything).Return(customer, nil)
		// store.On("FindActiveSubscriptionByCustomerId", ctx, customer.ID).Return(sub, nil)
		// store.On("CountTeamMembers", ctx, teamId).Return(int64(0), nil)
		err := service.VerifyAndUpdateTeamSubscriptionQuantity(ctx, teamId)
		assert.NoError(t, err)
		// store.AssertExpectations(t)
	})

	t.Run("returns error if UpdateItemQuantity fails", func(t *testing.T) {
		adapter := stores.NewAdapterDecorators()
		// store := new(MockPaymentStore)
		client := NewTestPaymentClient()
		service := &StripeService{client: client, adapter: adapter}
		adapter.CustomerFunc.FindCustomerFunc = func(ctx context.Context, filter *stores.StripeCustomerFilter) (*models.StripeCustomer, error) {
			return customer, nil
		}
		adapter.SubscriptionFunc.FindActiveSubscriptionByCustomerIdFunc = func(ctx context.Context, customerId string) (*models.StripeSubscription, error) {
			return sub, nil
		}
		adapter.TeamMemberFunc.CountTeamMembersFunc = func(ctx context.Context, filter *stores.TeamMemberFilter) (int64, error) {
			return int64(3), nil
		}

		// store.On("FindCustomer", ctx, mock.Anything).Return(customer, nil)
		// store.On("FindActiveSubscriptionByCustomerId", ctx, customer.ID).Return(sub, nil)
		// store.On("CountTeamMembers", ctx, teamId).Return(int64(3), nil)
		// client.On("UpdateItemQuantity", sub.ItemID, sub.PriceID, int64(3)).Return(nil, errors.New("update error"))
		client.UpdateItemQuantityFunc = func(itemID string, priceID string, quantity int64) (*stripe.SubscriptionItem, error) {
			return nil, errors.New("update error")
		}
		err := service.VerifyAndUpdateTeamSubscriptionQuantity(ctx, teamId)
		assert.Error(t, err)

		// store.AssertExpectations(t)
	})
}
