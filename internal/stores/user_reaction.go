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

type ReactionByCountry struct {
	Country        string `json:"country"`
	TotalReactions int64  `json:"total_reactions"`
}

type UserReactionStore interface {
	CreateUserReaction(ctx context.Context, input *models.UserReaction) (*models.UserReaction, error)
	CountUserReactions(ctx context.Context, filter *UserReactionFilter) (int64, error)
	CountByCountry(ctx context.Context, filter *UserReactionFilter) ([]*ReactionByCountry, error)
}

type DbUserReactionStore struct {
	db database.Dbx
}

// CountByCountry implements UserReactionStore.
func (d *DbUserReactionStore) CountByCountry(ctx context.Context, filter *UserReactionFilter) ([]*ReactionByCountry, error) {
	limit, _ := filter.LimitOffset()
	const query = `
	SELECT country, COUNT(*) AS total_reactions
	FROM public.user_reactions
	WHERE country IS NOT NULL
	GROUP BY country
	ORDER BY total_reactions DESC
	LIMIT $1;
`
	res, err := database.QueryAll[*ReactionByCountry](ctx, d.db, query, limit)
	if err != nil {
		return nil, err
	}
	return res, nil
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
	CountByCountryFunc     func(ctx context.Context, filter *UserReactionFilter) ([]*ReactionByCountry, error)
}

// CountByCountry implements UserReactionStore.
func (d *DbUserReactionStoreDectorator) CountByCountry(ctx context.Context, filter *UserReactionFilter) ([]*ReactionByCountry, error) {
	if d.CountByCountryFunc != nil {
		return d.CountByCountryFunc(ctx, filter)
	}
	if d.delegate == nil {
		return nil, errors.New("delegate for CountByCountry in UserReactionStore is nil")
	}
	return d.delegate.CountByCountry(ctx, filter)
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
