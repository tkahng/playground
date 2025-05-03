package seeders

import (
	"context"
	"database/sql"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jaswdr/faker/v2"
	"github.com/stephenafamo/scan"
	"github.com/tkahng/authgo/internal/crud/crudModels"
	"github.com/tkahng/authgo/internal/crud/crudrepo"
)

type Queryer interface {
	Query(ctx context.Context, sql string, arguments ...any) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	QueryContext(ctx context.Context, query string, args ...any) (scan.Rows, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

func CreateUserFromEmails(ctx context.Context, dbx Queryer, emails ...string) ([]*crudModels.User, error) {
	var users []crudModels.User
	for _, emails := range emails {
		users = append(users, crudModels.User{Email: emails})
	}

	res, err := crudrepo.User.Post(ctx, dbx, users)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func CreateUsers(ctx context.Context, dbx Queryer, count int) ([]*crudModels.User, error) {
	faker := faker.New().Internet()
	var users []crudModels.User
	for i := 0; i < count; i++ {
		user := crudModels.User{
			Email: faker.Email(),
		}
		users = append(users, user)
	}
	res, err := crudrepo.User.Post(ctx, dbx, users)
	if err != nil {
		return nil, err
	}
	return res, nil
}
