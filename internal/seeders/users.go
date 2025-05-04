package seeders

import (
	"context"

	"github.com/jaswdr/faker/v2"
	"github.com/tkahng/authgo/internal/db"
	"github.com/tkahng/authgo/internal/models"
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
