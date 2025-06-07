package resource

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/Masterminds/squirrel"
	sq "github.com/Masterminds/squirrel"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/repository"
)

type QueryResource[Model any, Key comparable, Filter any] struct {
	db           database.Dbx
	builder      *repository.SQLBuilder[Model]
	filterFn     func(qs squirrel.SelectBuilder, filter *Filter) squirrel.SelectBuilder
	sortFn       func(qs squirrel.SelectBuilder, filter *Filter) squirrel.SelectBuilder
	paginationFn func(qs squirrel.SelectBuilder, filter *Filter) squirrel.SelectBuilder
}

func (p *QueryResource[M, K, F]) filter(qs squirrel.SelectBuilder, filter *F) squirrel.SelectBuilder {
	if filter == nil {
		return qs // return the original query if no filter is provided
	}
	if p.filterFn != nil {
		qs = p.filterFn(qs, filter)
		return qs
	}
	return qs
}

func (p *QueryResource[M, K, F]) pagination(qs squirrel.SelectBuilder, filter *F) squirrel.SelectBuilder {
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

func (p *QueryResource[M, K, F]) sort(qs squirrel.SelectBuilder, filter *F) squirrel.SelectBuilder {
	if filter == nil {
		return qs // return the original query if no filter is provided
	}
	if p.sortFn != nil {
		qs = p.sortFn(qs, filter)
		return qs
	} else if sortable, ok := any(filter).(Sortable); ok {
		sortby, sortOrder := sortable.Sort()
		qs = qs.OrderBy(p.builder.Identifier(sortby) + " " + strings.ToUpper(sortOrder))
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
			if s.builder.SkipIdInsert() {
				continue
			} else if gen := s.builder.Generator(); gen != nil {
				id, err := gen(_type.Field(field.Idx), nil)
				if err != nil {
					return nil, fmt.Errorf("error generating primary key for field %s: %w", field.Name, err)
				}
				fieldsArray = append(fieldsArray, s.builder.Identifier(field.Name))
				valuesArray = append(valuesArray, id)
			} else if _field := _value.Field(field.Idx); !_field.IsValid() || _field.IsZero() {
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

func NewSqResource[Model any, Key comparable, Filter any](
	db database.Dbx,
	filterFn func(qs squirrel.SelectBuilder, filter *Filter) squirrel.SelectBuilder,
	sortFn func(qs squirrel.SelectBuilder, filter *Filter) squirrel.SelectBuilder,
	paginationFn func(qs squirrel.SelectBuilder, filter *Filter) squirrel.SelectBuilder,
) *QueryResource[Model, Key, Filter] {
	return &QueryResource[Model, Key, Filter]{
		db:           db,
		filterFn:     filterFn,
		sortFn:       sortFn,
		paginationFn: paginationFn,
	}
}

var _ Resource[any, any, any] = (*QueryResource[any, any, any])(nil)
