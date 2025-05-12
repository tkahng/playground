package seeders

import (
	"context"
	"errors"

	"github.com/jaswdr/faker/v2"
	"github.com/tkahng/authgo/internal/db"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/queries"
	"github.com/tkahng/authgo/internal/repository"
)

func CreateUserFromEmails(ctx context.Context, dbx db.Dbx, emails ...string) ([]*models.User, error) {
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

func CreateUsers(ctx context.Context, dbx db.Dbx, count int) ([]*models.User, error) {
	faker := faker.New().Internet()
	var users []models.User
	for i := 0; i < count; i++ {
		user := models.User{
			Email: faker.Email(),
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

func CreateUserWithAccountAndRole(ctx context.Context, dbx db.Dbx, count int, provider models.Providers, roleName string, faker faker.Internet) ([]*models.User, error) {
	role, err := queries.FindOrCreateRole(ctx, dbx, roleName)
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
