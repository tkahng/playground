package resource

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"reflect"
	"slices"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/repository"
	"github.com/tkahng/authgo/internal/tools/utils"
)

var (
	ResourceNotImplementedError = errors.New("resource not implemented")
)

type Resource[Model any, Key comparable, Filter any] interface {
	Delete(ctx context.Context, id Key) error
	Update(ctx context.Context, model *Model) (*Model, error)
	Create(ctx context.Context, model *Model) (*Model, error)
	Count(ctx context.Context, filter *Filter) (int64, error)
	Find(ctx context.Context, filter *Filter) ([]*Model, error)
	FindOne(ctx context.Context, filter *Filter) (*Model, error)
	FindByID(ctx context.Context, id Key) (*Model, error)
	WithTx(tx database.Dbx) Resource[Model, Key, Filter]
}

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
		slog.Info("sort called with nil filter, returning nil")
		return nil // return nil if no filter is provided
	}
	if p.sortFn != nil {
		slog.Info("sortFn is not nil, calling sortFn with filter", "filter", filter)
		return p.sortFn(filter)
	}
	if sortable, ok := any(filter).(Sortable); ok {
		slog.Info("filter implements Sortable, calling Sort method", "filter", filter)
		sortBy, sortOrder := sortable.Sort()
		if sortBy != "" && slices.Contains(p.repository.Builder().ColumnNames(), utils.Quote(sortBy)) {
			slog.Info("valid sort by field found", "sortBy", sortBy, "sortOrder", sortOrder)
			return &map[string]string{
				sortBy: sortOrder,
			}
		} else {
			slog.Info("sort by field not found in repository columns", "sortBy", sortBy, "sortOrder", sortOrder, "columns", p.repository.Builder().ColumnNames())
		}
	} else {
		slog.Info("filter does not implement Sortable, returning nil")
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
func (p *RepositoryResource[M, K, F]) FindOne(ctx context.Context, filter *F) (*M, error) {
	where := p.filter(filter)
	data, err := p.repository.GetOne(ctx, p.db, where)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// FindByID implements Resource.
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

type QueryResource[Model any, Key comparable, Filter any] struct {
	db           database.Dbx
	builder      *repository.SQLBuilder[Model]
	filterFn     func(qs sq.SelectBuilder, filter *Filter) sq.SelectBuilder
	sortFn       func(qs sq.SelectBuilder, filter *Filter) sq.SelectBuilder
	paginationFn func(qs sq.SelectBuilder, filter *Filter) sq.SelectBuilder
}

func (p *QueryResource[M, K, F]) filter(qs sq.SelectBuilder, filter *F) sq.SelectBuilder {
	if filter == nil {
		return qs // return the original query if no filter is provided
	}
	if p.filterFn != nil {
		qs = p.filterFn(qs, filter)
		return qs
	}
	return qs
}

func (p *QueryResource[M, K, F]) pagination(qs sq.SelectBuilder, filter *F) sq.SelectBuilder {
	if filter == nil {
		return qs
	}
	if p.paginationFn != nil {
		qs = p.paginationFn(qs, filter)
		return qs
	} else if paginable, ok := any(filter).(Paginable); ok {
		limit, offset := paginable.Pagination()
		qs = qs.Limit(uint64(limit)).Offset(uint64(offset))
		return qs
	}
	return qs
}

func (p *QueryResource[M, K, F]) sort(qs sq.SelectBuilder, filter *F) sq.SelectBuilder {
	if filter == nil {
		return qs // return the original query if no filter is provided
	}
	if p.sortFn != nil {
		qs = p.sortFn(qs, filter)
		return qs
	} else if sortable, ok := any(filter).(Sortable); ok {
		sortby, sortOrder := sortable.Sort()
		if sortby != "" && slices.Contains(p.builder.ColumnNames(), utils.Quote(sortby)) {
			qs = qs.OrderBy(p.builder.Identifier(sortby) + " " + strings.ToUpper(sortOrder))
			return qs
		}
	}
	return qs
}

// Count implements Resource.
func (s *QueryResource[Model, Key, Filter]) Count(ctx context.Context, filter *Filter) (int64, error) {
	qs := sq.Select("COUNT(" + s.builder.Table() + ".*)").
		From(s.builder.Table())

	qs = s.filter(qs, filter)

	count, err := database.QueryWithBuilder[database.CountOutput](ctx, s.db, qs.PlaceholderFormat(sq.Dollar))
	if err != nil {
		return 0, fmt.Errorf("error counting models: %w", err)
	}
	if len(count) == 0 {
		return 0, errors.New("no rows found")
	}
	return count[0].Count, nil
}

// Create implements Resource.
func (s *QueryResource[Model, Key, Filter]) Create(ctx context.Context, model *Model) (*Model, error) {
	_value := reflect.ValueOf(*model)
	_type := reflect.TypeOf(*model)
	var fieldsArray []string
	var valuesArray []interface{}
	for _, field := range s.builder.Fields() {
		if field.Name == s.builder.IdColumnName() {
			if gen := s.builder.Generator(); gen != nil {
				id, err := gen(_type.Field(field.Idx), nil)
				if err != nil {
					return nil, fmt.Errorf("error generating primary key for field %s: %w", field.Name, err)
				}
				fieldsArray = append(fieldsArray, s.builder.Identifier(field.Name))
				valuesArray = append(valuesArray, id)
			}
			if s.builder.InsertID() {
				continue
			}
			if _field := _value.Field(field.Idx); !_field.IsValid() || _field.IsZero() {
				continue
			} else {
				fieldsArray = append(fieldsArray, s.builder.Identifier(field.Name))
				valuesArray = append(valuesArray, _field.Interface())
			}
		}
		_field := _value.Field(field.Idx)
		if _field.IsValid() && !_field.IsZero() {
			fieldsArray = append(fieldsArray, s.builder.Identifier(field.Name))
			valuesArray = append(valuesArray, _field.Interface())
		}
	}
	qs := sq.Insert(s.builder.Table()).
		Columns(fieldsArray...).
		Values(valuesArray...).
		Suffix(fmt.Sprintf("RETURNING %s", s.builder.FieldString("")))
	res, err := database.QueryWithBuilder[*Model](ctx, s.db, qs.PlaceholderFormat(sq.Dollar))
	if err != nil {
		return nil, fmt.Errorf("error creating model: %w", err)
	}
	if len(res) == 0 {
		return nil, errors.New("no rows inserted")
	}
	return res[0], nil
}

// Delete implements Resource.
func (s *QueryResource[Model, Key, Filter]) Delete(ctx context.Context, id Key) error {
	qs := sq.Delete(s.builder.Table()).Where(sq.Eq{s.builder.IdColumnName(): id})
	err := database.ExecWithBuilder(ctx, s.db, qs.PlaceholderFormat(sq.Dollar))
	return err
}

// Find implements Resource.
func (s *QueryResource[Model, Key, Filter]) Find(ctx context.Context, filter *Filter) ([]*Model, error) {
	qs := sq.Select(s.builder.ColumnNamesTablePrefix()...).
		From(s.builder.Table())

	// Apply filters, sorting, and pagination
	qs = s.filter(qs, filter)
	qs = s.sort(qs, filter)
	qs = s.pagination(qs, filter)

	res, err := database.QueryWithBuilder[*Model](ctx, s.db, qs.PlaceholderFormat(sq.Dollar))
	if err != nil {
		return nil, fmt.Errorf("error finding models: %w", err)
	}
	return res, nil
}

// FindOne implements Resource.
func (s *QueryResource[Model, Key, Filter]) FindOne(ctx context.Context, filter *Filter) (*Model, error) {
	qs := sq.Select(s.builder.ColumnNamesTablePrefix()...).
		From(s.builder.Table())
	// Apply filters, sorting, and pagination
	qs = s.filter(qs, filter).Limit(1)
	res, err := database.QueryWithBuilder[*Model](ctx, s.db, qs.PlaceholderFormat(sq.Dollar))
	if err != nil {
		return nil, fmt.Errorf("error finding one model: %w", err)
	}
	if len(res) == 0 {
		return nil, nil
	}
	return res[0], nil
}

// FindByID implements Resource.
func (s *QueryResource[Model, Key, Filter]) FindByID(ctx context.Context, id Key) (*Model, error) {
	qs := sq.Select(s.builder.ColumnNamesTablePrefix()...).
		From(s.builder.Table()).Where(sq.Eq{s.builder.IdColumnName(): id}).Limit(1)
	res, err := database.QueryWithBuilder[*Model](ctx, s.db, qs.PlaceholderFormat(sq.Dollar))
	if err != nil {
		return nil, fmt.Errorf("error finding model by ID: %w", err)
	}
	if len(res) == 0 {
		return nil, nil
	}
	return res[0], nil
}

// Update implements Resource.
func (s *QueryResource[Model, Key, Filter]) Update(ctx context.Context, model *Model) (*Model, error) {
	if model == nil {
		return nil, nil
	}
	_value := reflect.ValueOf(*model)
	qs := sq.Update(s.builder.Table())
	for _, field := range s.builder.Fields() {
		if field.Name == s.builder.IdColumnName() {
			_field := _value.Field(field.Idx)
			qs = qs.Where(sq.Eq{s.builder.IdColumnName(): _field.Interface()})
		} else {
			_field := _value.Field(field.Idx)
			qs = qs.Set(field.Name, _field.Interface())
		}
	}
	qs = qs.Suffix(fmt.Sprintf("RETURNING %s", s.builder.FieldString("")))

	res, err := database.QueryWithBuilder[*Model](ctx, s.db, qs.PlaceholderFormat(sq.Dollar))
	if err != nil {
		return nil, fmt.Errorf("error updating model: %w", err)
	}
	if len(res) == 0 {
		return nil, errors.New("no rows updated")
	}
	return res[0], nil
}

// WithTx implements Resource.
func (s *QueryResource[Model, Key, Filter]) WithTx(tx database.Dbx) Resource[Model, Key, Filter] {
	return &QueryResource[Model, Key, Filter]{
		db:           tx,
		filterFn:     s.filterFn,
		sortFn:       s.sortFn,
		paginationFn: s.paginationFn,
		builder:      s.builder,
	}
}

func NewQueryResource[Model any, Key comparable, Filter any](
	db database.Dbx,
	builder *repository.SQLBuilder[Model],
	filterFn func(qs sq.SelectBuilder, filter *Filter) sq.SelectBuilder,
	sortFn func(qs sq.SelectBuilder, filter *Filter) sq.SelectBuilder,
	paginationFn func(qs sq.SelectBuilder, filter *Filter) sq.SelectBuilder,
) *QueryResource[Model, Key, Filter] {
	return &QueryResource[Model, Key, Filter]{
		db:           db,
		builder:      builder,
		filterFn:     filterFn,
		sortFn:       sortFn,
		paginationFn: paginationFn,
	}
}

var _ Resource[any, any, any] = (*QueryResource[any, any, any])(nil)

type Sortable interface {
	Sort() (sortBy, sortOrder string)
}
type DefaultFilter interface {
	Sortable
	Paginable
}
type SortParams struct {
	SortBy    string `query:"sort_by,omitempty" required:"false"`
	SortOrder string `query:"sort_order,omitempty" required:"false" enum:"asc,desc"`
}

func (s *SortParams) Sort() (sortBy, sortOrder string) {
	if s == nil {
		return "", "" // default values
	}
	if s.SortBy == "" {
		s.SortBy = "created_at" // default sort by
	}
	if s.SortOrder == "" {
		s.SortOrder = "desc" // default sort order
	}
	return s.SortBy, s.SortOrder
}

type PaginatedInput struct {
	Page    int64 `query:"page,omitempty" minimum:"0" required:"false"`
	PerPage int64 `query:"per_page,omitempty" default:"10" minimum:"1" maximum:"100" required:"false"`
}

type Paginable interface {
	Pagination() (limit, offset int)
}

func (p *PaginatedInput) Pagination() (limit, offset int) {
	if p == nil {
		return 10, 0 // default values
	}
	if p.PerPage <= 0 {
		p.PerPage = 10 // default value
	}
	if p.Page < 0 {
		p.Page = 0 // default value
	}
	return int(p.PerPage), int(p.Page) * int(p.PerPage)
}
