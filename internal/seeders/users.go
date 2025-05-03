package seeders

import (
	"context"
	"database/sql"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jaswdr/faker/v2"
	"github.com/stephenafamo/scan"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/repository"
)

type Queryer interface {
	Query(ctx context.Context, sql string, arguments ...any) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	QueryContext(ctx context.Context, query string, args ...any) (scan.Rows, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

func CreateUserFromEmails(ctx context.Context, dbx Queryer, emails ...string) ([]*models.User, error) {
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

func CreateUsers(ctx context.Context, dbx Queryer, count int) ([]*models.User, error) {
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
