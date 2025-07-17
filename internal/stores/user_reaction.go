package stores

import (
	"context"
	"errors"

	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/repository"
)

type UserReactionStore interface {
	CreateUserReaction(ctx context.Context, input *models.UserReaction) (*models.UserReaction, error)
}

type DbUserReactionStore struct {
	db database.Dbx
}

func NewDbUserReactionStore(db database.Dbx) *DbUserReactionStore {
	return &DbUserReactionStore{
		db: db,
	}
}

// CreateUserReaction implements UserReactionStore.
func (d *DbUserReactionStore) CreateUserReaction(ctx context.Context, input *models.UserReaction) (*models.UserReaction, error) {
	return repository.UserReaction.PostOne(ctx, d.db, input)
}

var _ UserReactionStore = &DbUserReactionStore{}

type DbUserReactionStoreDectorator struct {
	delegate               UserReactionStore
	CreateUserReactionFunc func(ctx context.Context, input *models.UserReaction) (*models.UserReaction, error)
}

// CreateUserReaction implements UserReactionStore.
func (d *DbUserReactionStoreDectorator) CreateUserReaction(ctx context.Context, input *models.UserReaction) (*models.UserReaction, error) {
	if d.CreateUserReactionFunc != nil {
		return d.CreateUserReactionFunc(ctx, input)
	}
	if d.delegate == nil {
		return nil, errors.New("delegate for CreateUserReaction in UserReactionStore is nil")
	}
	return d.delegate.CreateUserReaction(ctx, input)
}

var _ UserReactionStore = &DbUserReactionStoreDectorator{}
