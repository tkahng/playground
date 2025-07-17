package stores

import (
	"context"
	"errors"

	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/repository"
)

type UserReactionFilter struct {
	PaginatedInput
	SortParams
}

type UserReactionStore interface {
	CreateUserReaction(ctx context.Context, input *models.UserReaction) (*models.UserReaction, error)
	CountUserReactions(ctx context.Context, filter *UserReactionFilter) (int64, error)
}

type DbUserReactionStore struct {
	db database.Dbx
}

// CountUserReactions implements UserReactionStore.
func (d *DbUserReactionStore) CountUserReactions(ctx context.Context, filter *UserReactionFilter) (int64, error) {
	where := d.filter(filter)
	return repository.UserReaction.Count(ctx, d.db, where)
}

func (d *DbUserReactionStore) filter(_ *UserReactionFilter) *map[string]any {
	return nil
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
	CountUserReactionsFunc func(ctx context.Context, filter *UserReactionFilter) (int64, error)
}

// CountUserReactions implements UserReactionStore.
func (d *DbUserReactionStoreDectorator) CountUserReactions(ctx context.Context, filter *UserReactionFilter) (int64, error) {
	if d.CountUserReactionsFunc != nil {
		return d.CountUserReactionsFunc(ctx, filter)
	}
	if d.delegate == nil {
		return 0, errors.New("delegate for CountUserReactions in UserReactionStore is nil")
	}
	return d.delegate.CountUserReactions(ctx, filter)
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
