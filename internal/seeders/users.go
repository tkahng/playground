package seeders

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jaswdr/faker/v2"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/repository"
	"github.com/tkahng/playground/internal/stores"
	"github.com/tkahng/playground/internal/tools/types"
)

func CreateUserFromEmails(ctx context.Context, dbx database.Dbx, emails ...string) ([]*models.User, error) {
	var users []models.User
	for _, emails := range emails {
		users = append(users, models.User{Email: emails})
	}

	res, err := repository.User.Post(ctx, dbx, users)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func CreateUsers(ctx context.Context, dbx database.Dbx, count int) ([]*models.User, error) {
	internet := faker.New().Internet()
	var users []models.User
	for i := 0; i < count; i++ {
		user := models.User{
			Email: internet.Email(),
		}
		users = append(users, user)
	}
	res, err := repository.User.Post(ctx, dbx, users)
	if err != nil {
		return nil, err
	}
	return res, nil
}

type CreateUserDto struct {
	Email    string
	Provider models.Providers
}

func CreateUserWithAccountAndRole(ctx context.Context, dbx database.Dbx, count int, provider models.Providers, roleName string, faker faker.Internet) ([]*models.User, error) {
	rbacStore := stores.NewDbRBACStore(dbx)
	role, err := rbacStore.FindOrCreateRole(ctx, roleName)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, errors.New("role not found")
	}
	// Create users
	var usersdto []models.User
	for range count {
		user := models.User{
			Email: faker.Email(),
		}
		usersdto = append(usersdto, user)
	}
	users, err := repository.User.Post(ctx, dbx, usersdto)
	if err != nil {
		return nil, err
	}
	var accountsDto []models.UserAccount
	var userRoles []models.UserRole
	for _, user := range users {
		var providertype models.ProviderTypes
		switch provider {
		case models.ProvidersGoogle:
			providertype = models.ProviderTypeOAuth
		case models.ProvidersGithub:
			providertype = models.ProviderTypeOAuth
		case models.ProvidersCredentials:
			providertype = models.ProviderTypeCredentials
		default:
			providertype = models.ProviderTypeOAuth
		}
		account := models.UserAccount{
			Provider:          provider,
			Type:              models.ProviderTypes(providertype),
			UserID:            user.ID,
			ProviderAccountID: user.ID.String(),
			ExpiresAt:         nil,
		}
		accountsDto = append(accountsDto, account)
		userRole := models.UserRole{
			UserID: user.ID,
			RoleID: role.ID,
		}
		userRoles = append(userRoles, userRole)
	}

	_, err = repository.UserAccount.Post(ctx, dbx, accountsDto)
	if err != nil {
		return nil, err
	}
	_, err = repository.UserRole.Post(ctx, dbx, userRoles)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func CreateStripeProductPrices(ctx context.Context, dbx database.Dbx, count int) ([]*models.StripeProduct, error) {
	var products []models.StripeProduct
	for range count {
		uid := uuid.NewString()
		product := models.StripeProduct{
			ID:       uid,
			Name:     uid,
			Active:   true,
			Metadata: map[string]string{"key": "value"},
		}
		products = append(products, product)
	}
	res, err := repository.StripeProduct.Post(ctx, dbx, products)
	if err != nil {
		return nil, err
	}
	var prices []models.StripePrice
	for _, product := range res {
		price := models.StripePrice{
			ID:         uuid.NewString(),
			ProductID:  product.ID,
			UnitAmount: types.Pointer(int64(1000)),
			Currency:   "usd",
			Active:     true,
			Type:       models.StripePricingTypeRecurring,
			Interval:   types.Pointer(models.StripePricingPlanIntervalDay),
			Metadata:   map[string]string{"key": "value"},
		}
		prices = append(prices, price)
	}
	newPrices, err := repository.StripePrice.Post(ctx, dbx, prices)
	if err != nil {
		return nil, err
	}
	for _, prod := range res {
		for _, price := range newPrices {
			if prod.ID == price.ProductID {
				prod.Prices = append(prod.Prices, price)
			}
		}
	}

	return res, nil
}
