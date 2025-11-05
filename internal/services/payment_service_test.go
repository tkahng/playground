package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stripe/stripe-go/v82"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/services"
	"github.com/tkahng/playground/internal/stores"
	"github.com/tkahng/playground/internal/stores/testutils"
	"github.com/tkahng/playground/internal/tools/types"
)

func TestStripeService_CreateTeamCustomer(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
			// init
			adpt := stores.NewDbAdapterDecorators(db)
			client := services.NewMockPaymentClient()
			service := services.NewPaymentService(client, adpt)
			// setup
			userInfo1 := testutils.CreateUserWithOptions(t, adpt)
			teamInfo1 := testutils.CreateTeamAndMemberWithOptions(t, adpt, &userInfo1.User)
			team1 := &teamInfo1.Team
			user1 := &userInfo1.User
			// test
			result, err := service.CreateTeamCustomer(ctx, team1, user1)
			// assert
			assert.NoError(t, err)
			if cus := client.GetCustomerByFunc(func(c *stripe.Customer) bool {
				if c.Email == user1.Email && c.Name == team1.Name {
					return true
				}
				return false
			}); cus != nil {
				assert.Equal(t, cus.ID, result.ID, "customer id should be the same")
				assert.Equal(t, &cus.Name, result.Name, "customer name should be the same")
				assert.Equal(t, cus.Email, result.Email, "customer email should be the same")
				assert.Equal(t, cus.Metadata["team_id"], team1.ID.String(), "customer metadata should be the same")
				assert.Equal(t, cus.Metadata["customer_type"], string(models.StripeCustomerTypeTeam), "customer metadata should be the same")
			} else {
				t.Fatal("customer not found")
			}
		})
	})
	t.Run("client error", func(t *testing.T) {
		database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
			// init
			adpt := stores.NewDbAdapterDecorators(db)
			client := services.NewMockPaymentClient()
			service := services.NewPaymentService(client, adpt)
			// setup
			userInfo1 := testutils.CreateUserWithOptions(t, adpt)
			teamInfo1 := testutils.CreateTeamAndMemberWithOptions(t, adpt, &userInfo1.User)
			team1 := &teamInfo1.Team
			user1 := &userInfo1.User

			client.CreateCustomerFunc = func(email string, name *string, metadata *map[string]string) (*stripe.Customer, error) {
				return nil, errors.New("stripe error")
			}
			// test
			result, err := service.CreateTeamCustomer(ctx, team1, user1)
			// assert
			assert.Error(t, err)
			assert.Nil(t, result)
		})
	})
}

func TestStripeService_CreateUserCustomer(t *testing.T) {
	// ctx := context.Background()
	// user := &models.User{ID: uuid.New(), Email: "user@example.com", Name: types.Pointer("User Name")}
	// customer := &stripe.Customer{ID: "cus_456", Email: user.Email}
	// created := &models.StripeCustomer{ID: customer.ID, Email: customer.Email, Name: user.Name, UserID: types.Pointer(user.ID), CustomerType: models.StripeCustomerTypeUser}

	t.Run("success", func(t *testing.T) {
		database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
			// init
			adapter := stores.NewDbAdapterDecorators(db)
			client := services.NewMockPaymentClient()
			service := services.NewPaymentService(client, adapter)
			// setup
			userInfo1 := testutils.CreateUserWithOptions(t, adapter)
			user1 := &userInfo1.User
			// test
			result, err := service.CreateUserCustomer(ctx, user1)
			// assert
			assert.NoError(t, err)
			if cus, ok := client.CustomerByEmail[user1.Email]; ok {
				assert.Equal(t, cus.ID, result.ID, "customer id should be the same")
				assert.Equal(t, &cus.Name, result.Name, "customer name should be the same")
				assert.Equal(t, cus.Email, result.Email, "customer email should be the same")
				assert.Equal(t, cus.Metadata["user_id"], user1.ID.String(), "customer metadata should be the same")
				assert.Equal(t, cus.Metadata["customer_type"], string(models.StripeCustomerTypeUser), "customer metadata should be the same")
			} else {
				assert.Fail(t, "customer should be created")
			}
		})

	})

	t.Run("client error", func(t *testing.T) {
		database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
			// init
			adapter := stores.NewDbAdapterDecorators(db)
			client := services.NewMockPaymentClient()
			service := services.NewPaymentService(client, adapter)
			// setup
			userInfo1 := testutils.CreateUserWithOptions(t, adapter)
			user1 := &userInfo1.User

			client.CreateCustomerFunc = func(email string, name *string, metadata *map[string]string) (*stripe.Customer, error) {
				return nil, errors.New("stripe error")
			}
			// test
			result, err := service.CreateUserCustomer(ctx, user1)
			// assert
			assert.Error(t, err)
			assert.Nil(t, result)
		})

	})

}

func TestStripeService_FindCustomerByTeam(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		client := services.NewMockPaymentClient()
		service := services.NewPaymentService(client, adapter)

		userInfo := testutils.CreateUserWithOptions(t, adapter)
		teamInfo := testutils.CreateTeamAndMemberWithOptions(t, adapter, &userInfo.User)

		team := &teamInfo.Team
		user := &userInfo.User

		customer := services.TestHelperCreateTeamCustomer(t, service, team, user)
		teamId := teamInfo.Team.ID

		t.Run("success", func(t *testing.T) {

			// store.On("FindCustomer", ctx, mock.AnythingOfType("*models.StripeCustomer")).Return(customer, nil)
			result, err := service.FindCustomerByTeamId(ctx, teamId)
			assert.NoError(t, err)
			assert.Equal(t, customer, result)
		})

		t.Run("store error", func(t *testing.T) {

			adapter.CustomerFunc.FindCustomerFunc = func(ctx context.Context, filter *stores.StripeCustomerFilter) (*models.StripeCustomer, error) {
				return nil, errors.New("db error")
			}
			result, err := service.FindCustomerByTeamId(ctx, teamId)
			assert.Error(t, err)
			assert.Nil(t, result)
		})
	})
}

func TestStripeService_FindCustomerByUser(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		client := services.NewMockPaymentClient()
		service := services.NewPaymentService(client, adapter)

		userInfo := testutils.CreateUserWithOptions(t, adapter)

		user := &userInfo.User

		customer := services.TestHelperCreateUserCustomer(t, service, user)

		t.Run("success", func(t *testing.T) {

			// store.On("FindCustomer", ctx, mock.AnythingOfType("*models.StripeCustomer")).Return(customer, nil)
			result, err := service.FindCustomerByUserId(ctx, user.ID)
			assert.NoError(t, err)
			assert.Equal(t, customer, result)
		})

		t.Run("store error", func(t *testing.T) {

			adapter.CustomerFunc.FindCustomerFunc = func(ctx context.Context, filter *stores.StripeCustomerFilter) (*models.StripeCustomer, error) {
				return nil, errors.New("db error")
			}
			result, err := service.FindCustomerByUserId(ctx, user.ID)
			assert.Error(t, err)
			assert.Nil(t, result)
		})
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

		client := services.NewMockPaymentClient()
		service := services.NewPaymentService(client, adapter)
		adapter.CustomerFunc.FindCustomerFunc = func(ctx context.Context, filter *stores.StripeCustomerFilter) (*models.StripeCustomer, error) {
			return customer, nil
		}

		adapter.SubscriptionFunc.FindActiveSubscriptionByCustomerIdFunc = func(ctx context.Context, customerId string) (*models.StripeSubscription, error) {
			return sub, nil
		}

		adapter.TeamMemberFunc.CountTeamMembersFunc = func(ctx context.Context, filter *stores.TeamMemberFilter) (int64, error) {
			return int64(3), nil
		}

		// client.On("UpdateItemQuantity", sub.ItemID, sub.PriceID, int64(3)).Return(&stripe.SubscriptionItem{}, nil)
		client.UpdateItemQuantityFunc = func(itemID string, priceID string, quantity int64) (*stripe.SubscriptionItem, error) {
			return &stripe.SubscriptionItem{}, nil
		}
		err := service.VerifyAndUpdateTeamSubscriptionQuantity(ctx, teamId)
		assert.NoError(t, err)

	})

	t.Run("no update if quantity matches", func(t *testing.T) {
		adapter := stores.NewAdapterDecorators()

		client := services.NewMockPaymentClient()
		service := services.NewPaymentService(client, adapter)
		adapter.CustomerFunc.FindCustomerFunc = func(ctx context.Context, filter *stores.StripeCustomerFilter) (*models.StripeCustomer, error) {
			return customer, nil
		}

		adapter.SubscriptionFunc.FindActiveSubscriptionByCustomerIdFunc = func(ctx context.Context, customerId string) (*models.StripeSubscription, error) {
			return sub, nil
		}

		adapter.TeamMemberFunc.CountTeamMembersFunc = func(ctx context.Context, filter *stores.TeamMemberFilter) (int64, error) {
			return int64(2), nil
		}
		// store.On("CountTeamMembers", ctx, teamId).Return(int64(2), nil)
		err := service.VerifyAndUpdateTeamSubscriptionQuantity(ctx, teamId)
		assert.NoError(t, err)

	})

	t.Run("does not return error if no subscription", func(t *testing.T) {
		adapter := stores.NewAdapterDecorators()

		client := services.NewMockPaymentClient()
		service := services.NewPaymentService(client, adapter)
		adapter.CustomerFunc.FindCustomerFunc = func(ctx context.Context, filter *stores.StripeCustomerFilter) (*models.StripeCustomer, error) {
			return customer, nil
		}
		adapter.SubscriptionFunc.FindActiveSubscriptionByCustomerIdFunc = func(ctx context.Context, customerId string) (*models.StripeSubscription, error) {
			return nil, nil
		}
		err := service.VerifyAndUpdateTeamSubscriptionQuantity(ctx, teamId)
		assert.NoError(t, err)
	})

	t.Run("returns error if store fails", func(t *testing.T) {
		adapter := stores.NewAdapterDecorators()

		client := services.NewMockPaymentClient()
		service := services.NewPaymentService(client, adapter)
		adapter.CustomerFunc.FindCustomerFunc = func(ctx context.Context, filter *stores.StripeCustomerFilter) (*models.StripeCustomer, error) {
			return nil, errors.New("db error")
		}
		// store.On("FindCustomer", ctx, mock.Anything).Return(nil, errors.New("db error"))
		err := service.VerifyAndUpdateTeamSubscriptionQuantity(ctx, teamId)
		assert.Error(t, err)

	})

	t.Run("returns error if no customer", func(t *testing.T) {
		adapter := stores.NewAdapterDecorators()

		client := services.NewMockPaymentClient()
		service := services.NewPaymentService(client, adapter)
		adapter.CustomerFunc.FindCustomerFunc = func(ctx context.Context, filter *stores.StripeCustomerFilter) (*models.StripeCustomer, error) {
			return nil, nil
		}

		err := service.VerifyAndUpdateTeamSubscriptionQuantity(ctx, teamId)
		assert.NoError(t, err)

	})

	t.Run("returns error if CountTeamMembers fails", func(t *testing.T) {
		adapter := stores.NewAdapterDecorators()

		client := services.NewMockPaymentClient()
		service := services.NewPaymentService(client, adapter)
		adapter.CustomerFunc.FindCustomerFunc = func(ctx context.Context, filter *stores.StripeCustomerFilter) (*models.StripeCustomer, error) {
			return customer, nil
		}

		adapter.SubscriptionFunc.FindActiveSubscriptionByCustomerIdFunc = func(ctx context.Context, customerId string) (*models.StripeSubscription, error) {
			return sub, nil
		}

		adapter.TeamMemberFunc.CountTeamMembersFunc = func(ctx context.Context, filter *stores.TeamMemberFilter) (int64, error) {
			return 0, errors.New("count error")
		}

		err := service.VerifyAndUpdateTeamSubscriptionQuantity(ctx, teamId)
		assert.Error(t, err)

	})

	t.Run("returns nil if team member count is zero", func(t *testing.T) {
		adapter := stores.NewAdapterDecorators()

		client := services.NewMockPaymentClient()
		service := services.NewPaymentService(client, adapter)
		adapter.CustomerFunc.FindCustomerFunc = func(ctx context.Context, filter *stores.StripeCustomerFilter) (*models.StripeCustomer, error) {
			return customer, nil
		}

		adapter.SubscriptionFunc.FindActiveSubscriptionByCustomerIdFunc = func(ctx context.Context, customerId string) (*models.StripeSubscription, error) {
			return sub, nil
		}

		adapter.TeamMemberFunc.CountTeamMembersFunc = func(ctx context.Context, filter *stores.TeamMemberFilter) (int64, error) {
			return int64(0), nil
		}
		client.UpdateItemQuantityFunc = func(itemID string, priceID string, quantity int64) (*stripe.SubscriptionItem, error) {
			return nil, nil
		}

		err := service.VerifyAndUpdateTeamSubscriptionQuantity(ctx, teamId)
		assert.NoError(t, err)

	})

	t.Run("returns error if UpdateItemQuantity fails", func(t *testing.T) {
		adapter := stores.NewAdapterDecorators()

		client := services.NewMockPaymentClient()
		service := services.NewPaymentService(client, adapter)
		adapter.CustomerFunc.FindCustomerFunc = func(ctx context.Context, filter *stores.StripeCustomerFilter) (*models.StripeCustomer, error) {
			return customer, nil
		}
		adapter.SubscriptionFunc.FindActiveSubscriptionByCustomerIdFunc = func(ctx context.Context, customerId string) (*models.StripeSubscription, error) {
			return sub, nil
		}
		adapter.TeamMemberFunc.CountTeamMembersFunc = func(ctx context.Context, filter *stores.TeamMemberFilter) (int64, error) {
			return int64(3), nil
		}

		client.UpdateItemQuantityFunc = func(itemID string, priceID string, quantity int64) (*stripe.SubscriptionItem, error) {
			return nil, errors.New("update error")
		}
		err := service.VerifyAndUpdateTeamSubscriptionQuantity(ctx, teamId)
		assert.Error(t, err)

	})
}
