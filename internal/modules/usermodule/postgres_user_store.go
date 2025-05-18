package usermodule

import (
	"context"

	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/models"

	"github.com/tkahng/authgo/internal/crudrepo"
	"github.com/tkahng/authgo/internal/shared"
)

type PostgresUserStore struct {
	db database.Dbx
}

// FindUserByEmail implements UserStore.
func (p *PostgresUserStore) FindUserByEmail(ctx context.Context, email string) (*models.User, error) {
	a, err := crudrepo.User.GetOne(
		ctx,
		p.db,
		&map[string]any{
			"email": map[string]any{
				"_eq": email,
			},
		},
	)
	return database.OptionalRow(a, err)
}

// CreateUser implements UserStore.
func (p *PostgresUserStore) CreateUser(ctx context.Context, params *shared.AuthenticationInput) (*models.User, error) {
	return crudrepo.User.PostOne(
		ctx,
		p.db,
		&models.User{
			Email:           params.Email,
			Name:            params.Name,
			Image:           params.AvatarUrl,
			EmailVerifiedAt: params.EmailVerifiedAt,
		},
	)
}

func NewPostgresUserStore(db database.Dbx) UserStore {
	return &PostgresUserStore{
		db: db,
	}
}
