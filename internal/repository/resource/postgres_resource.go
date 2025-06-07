package resource

import (
	"context"

	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/repository"
)

type RepositoryResource[M any, K comparable, F any] struct {
	db           database.Dbx
	repository   *repository.PostgresRepository[M]
	filterFn     func(filter *F) *map[string]any
	sortFn       func(filter *F) *map[string]string
	paginationFn func(filter *F) (limit, offset int)
}

func NewRepositoryResource[M any, K comparable, F any](
	db database.Dbx,
	repository *repository.PostgresRepository[M],
	filterFn func(filter *F) *map[string]any,
	sortFn func(filter *F) *map[string]string,
	paginationFn func(filter *F) (limit, offset int),
) *RepositoryResource[M, K, F] {
	return &RepositoryResource[M, K, F]{
		db:           db,
		repository:   repository,
		filterFn:     filterFn,
		sortFn:       sortFn,
		paginationFn: paginationFn,
	}
}

func (p *RepositoryResource[M, K, F]) filter(filter *F) *map[string]any {
	where := new(map[string]any)
	if filter == nil {
		return where // return empty map if no filter is provided
	}
	if p.filterFn != nil {
		where = p.filterFn(filter)
	}
	return where
}

func (p *RepositoryResource[M, K, F]) pagination(filter *F) (limit, offset int) {
	if filter == nil {
		return 10, 0 // default values
	}
	if p.paginationFn != nil {
		return p.paginationFn(filter)
	}
	if paginable, ok := any(filter).(Paginable); ok {
		return paginable.Pagination()
	}
	return 10, 0 // default values
}

func (p *RepositoryResource[M, K, F]) sort(filter *F) *map[string]string {
	if filter == nil {
		return nil // return nil if no filter is provided
	}
	if p.sortFn != nil {
		return p.sortFn(filter)
	}
	if sortable, ok := any(filter).(Sortable); ok {
		sortBy, sortOrder := sortable.Sort()
		return &map[string]string{
			sortBy: sortOrder,
		}
	}
	return nil // default no sorting
}

// Count implements Resource.
func (p *RepositoryResource[M, K, F]) Count(ctx context.Context, filter *F) (int64, error) {
	where := p.filter(filter)
	return p.repository.Count(ctx, p.db, where)
}

// Create implements Resource.
func (p *RepositoryResource[M, K, F]) Create(ctx context.Context, model *M) (*M, error) {
	if model == nil {
		return nil, nil
	}
	result, err := p.repository.PostOne(ctx, p.db, model)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (p *RepositoryResource[M, K, F]) idWhere(id K) *map[string]any {
	where := map[string]any{
		p.repository.Builder().IdColumnName(): map[string]any{
			"_eq": id,
		},
	}
	return &where
}

// Delete implements Resource.
func (p *RepositoryResource[M, K, F]) Delete(ctx context.Context, id K) error {
	_, err := p.repository.Delete(ctx, p.db, p.idWhere(id))
	if err != nil {
		return err
	}
	return nil
}

// Find implements Resource.
func (p *RepositoryResource[M, K, F]) Find(ctx context.Context, filter *F) ([]*M, error) {
	// if filter == nil {
	// 	return nil, nil
	// }
	where := p.filter(filter)
	sort := p.sort(filter)
	limit, offset := p.pagination(filter)

	data, err := p.repository.Get(
		ctx,
		p.db,
		where,
		sort,
		&limit,
		&offset,
	)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// FindOne implements Resource.
func (p *RepositoryResource[M, K, F]) FindByID(ctx context.Context, id K) (*M, error) {
	where := p.idWhere(id)
	data, err := p.repository.GetOne(ctx, p.db, where)
	if err != nil {
		return nil, err
	}
	if data == nil {
		return nil, nil // or return an error if not found
	}
	return data, nil
}

// Update implements Resource.
func (p *RepositoryResource[M, K, F]) Update(ctx context.Context, model *M) (*M, error) {
	if model == nil {
		return nil, nil
	}
	result, err := p.repository.PutOne(ctx, p.db, model)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// WithTx implements Resource.
func (p *RepositoryResource[M, K, F]) WithTx(tx database.Dbx) Resource[M, K, F] {
	return &RepositoryResource[M, K, F]{
		db:           tx,
		repository:   p.repository,
		filterFn:     p.filterFn,
		sortFn:       p.sortFn,
		paginationFn: p.paginationFn,
	}
}
