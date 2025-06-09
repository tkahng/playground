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
	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/repository"
	"github.com/tkahng/authgo/internal/tools/utils"
)

type ResourceService[Model any, Key comparable, Filter any] interface {
	Delete(ctx context.Context, dbx database.Dbx, id Key) error
	// DeleteMany(ctx context.Context, filter *Filter) (int64, error)
	Update(ctx context.Context, dbx database.Dbx, model *Model) (*Model, error)
	Create(ctx context.Context, dbx database.Dbx, model *Model) (*Model, error)
	Count(ctx context.Context, dbx database.Dbx, filter *Filter) (int64, error)
	Find(ctx context.Context, dbx database.Dbx, filter *Filter) ([]*Model, error)
	FindOne(ctx context.Context, dbx database.Dbx, filter *Filter) (*Model, error)
	FindByID(ctx context.Context, dbx database.Dbx, id Key) (*Model, error)
}

var _ ResourceService[any, any, any] = (*RepositoryResourceService[any, any, any])(nil)

type RepositoryResourceService[M any, K comparable, F any] struct {
	repository   *repository.PostgresRepository[M]
	filterFn     func(filter *F) *map[string]any
	sortFn       func(filter *F) *map[string]string
	paginationFn func(filter *F) (limit, offset int)
}

func NewRepositoryResourceService[M any, K comparable, F any](
	repository *repository.PostgresRepository[M],
	filterFn func(filter *F) *map[string]any,
	sortFn func(filter *F) *map[string]string,
	paginationFn func(filter *F) (limit, offset int),
) *RepositoryResourceService[M, K, F] {
	if repository == nil {
		panic("repository cannot be nil")
	}
	if filterFn == nil {
		panic("filterFn cannot be nil")
	}
	return &RepositoryResourceService[M, K, F]{
		repository:   repository,
		filterFn:     filterFn,
		sortFn:       sortFn,
		paginationFn: paginationFn,
	}
}

func (p *RepositoryResourceService[M, K, F]) filter(filter *F) *map[string]any {
	where := new(map[string]any)
	if filter == nil {
		return where // return empty map if no filter is provided
	}
	if p.filterFn != nil {
		where = p.filterFn(filter)
	}
	return where
}

func (p *RepositoryResourceService[M, K, F]) pagination(filter *F) (limit, offset int) {
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

func (p *RepositoryResourceService[M, K, F]) sort(filter *F) *map[string]string {
	if filter == nil {
		return nil // return nil if no filter is provided
	}
	if p.sortFn != nil {
		return p.sortFn(filter)
	}
	if sortable, ok := any(filter).(Sortable); ok {
		sortBy, sortOrder := sortable.Sort()
		if sortBy != "" && slices.Contains(p.repository.Builder().ColumnNames(), utils.Quote(sortBy)) {
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
func (p *RepositoryResourceService[M, K, F]) Count(ctx context.Context, dbx database.Dbx, filter *F) (int64, error) {
	where := p.filter(filter)
	return p.repository.Count(ctx, dbx, where)
}

// Create implements Resource.
func (p *RepositoryResourceService[M, K, F]) Create(ctx context.Context, dbx database.Dbx, model *M) (*M, error) {
	if model == nil {
		return nil, nil
	}
	result, err := p.repository.PostOne(ctx, dbx, model)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (p *RepositoryResourceService[M, K, F]) idWhere(id K) *map[string]any {
	where := map[string]any{
		p.repository.Builder().IdColumnName(): map[string]any{
			"_eq": id,
		},
	}
	return &where
}

// Delete implements Resource.
func (p *RepositoryResourceService[M, K, F]) Delete(ctx context.Context, dbx database.Dbx, id K) error {
	count, err := p.repository.Delete(ctx, dbx, p.idWhere(id))
	if err != nil {
		return err
	}
	if count == 0 {
		return errors.New("no rows deleted")
	}
	return nil
}

// Find implements Resource.
func (p *RepositoryResourceService[M, K, F]) Find(ctx context.Context, dbx database.Dbx, filter *F) ([]*M, error) {
	where := p.filter(filter)
	sort := p.sort(filter)
	limit, offset := p.pagination(filter)

	data, err := p.repository.Get(
		ctx,
		dbx,
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
func (p *RepositoryResourceService[M, K, F]) FindOne(ctx context.Context, dbx database.Dbx, filter *F) (*M, error) {
	where := p.filter(filter)
	data, err := p.repository.GetOne(ctx, dbx, where)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// FindByID implements Resource.
func (p *RepositoryResourceService[M, K, F]) FindByID(ctx context.Context, dbx database.Dbx, id K) (*M, error) {
	where := p.idWhere(id)
	data, err := p.repository.GetOne(ctx, dbx, where)
	if err != nil {
		return nil, err
	}
	if data == nil {
		return nil, nil // or return an error if not found
	}
	return data, nil
}

// Update implements Resource.
func (p *RepositoryResourceService[M, K, F]) Update(ctx context.Context, dbx database.Dbx, model *M) (*M, error) {
	if model == nil {
		return nil, nil
	}
	result, err := p.repository.PutOne(ctx, dbx, model)
	if err != nil {
		return nil, err
	}
	return result, nil
}

type QueryResourceService[Model any, Key comparable, Filter any] struct {
	builder      *repository.SQLBuilder[Model]
	filterFn     func(qs sq.SelectBuilder, filter *Filter) sq.SelectBuilder
	sortFn       func(qs sq.SelectBuilder, filter *Filter) sq.SelectBuilder
	paginationFn func(qs sq.SelectBuilder, filter *Filter) sq.SelectBuilder
}

func NewQueryResourceService[Model any, Key comparable, Filter any](
	builder *repository.SQLBuilder[Model],
	filterFn func(qs sq.SelectBuilder, filter *Filter) sq.SelectBuilder,
	sortFn func(qs sq.SelectBuilder, filter *Filter) sq.SelectBuilder,
	paginationFn func(qs sq.SelectBuilder, filter *Filter) sq.SelectBuilder,
) *QueryResourceService[Model, Key, Filter] {
	if builder == nil {
		panic("builder cannot be nil")
	}
	if filterFn == nil {
		panic("filterFn cannot be nil")
	}
	return &QueryResourceService[Model, Key, Filter]{
		builder:      builder,
		filterFn:     filterFn,
		sortFn:       sortFn,
		paginationFn: paginationFn,
	}
}

var _ ResourceService[any, any, any] = (*QueryResourceService[any, any, any])(nil)

func (p *QueryResourceService[M, K, F]) filter(qs sq.SelectBuilder, filter *F) sq.SelectBuilder {
	if filter == nil {
		return qs // return the original query if no filter is provided
	}
	if p.filterFn != nil {
		qs := p.filterFn(qs, filter)
		return qs
	}
	return qs
}

func (p *QueryResourceService[M, K, F]) pagination(qs sq.SelectBuilder, filter *F) sq.SelectBuilder {
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

func (p *QueryResourceService[M, K, F]) sort(qs sq.SelectBuilder, filter *F) sq.SelectBuilder {
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
func (s *QueryResourceService[Model, Key, Filter]) Count(ctx context.Context, dbx database.Dbx, filter *Filter) (int64, error) {
	qs := sq.Select("COUNT(" + s.builder.Table() + ".*)").
		From(s.builder.Table())

	qs = s.filter(qs, filter)

	count, err := database.QueryWithBuilder[database.CountOutput](ctx, dbx, qs.PlaceholderFormat(sq.Dollar))
	if err != nil {
		return 0, fmt.Errorf("error counting models: %w", err)
	}
	if len(count) == 0 {
		return 0, errors.New("no rows found")
	}
	return count[0].Count, nil
}

// Create implements Resource.
func (s *QueryResourceService[Model, Key, Filter]) Create(ctx context.Context, dbx database.Dbx, model *Model) (*Model, error) {
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
	res, err := database.QueryWithBuilder[*Model](ctx, dbx, qs.PlaceholderFormat(sq.Dollar))
	if err != nil {
		return nil, fmt.Errorf("error creating model: %w", err)
	}
	if len(res) == 0 {
		return nil, errors.New("no rows inserted")
	}
	return res[0], nil
}

// Delete implements Resource.
func (s *QueryResourceService[Model, Key, Filter]) Delete(ctx context.Context, dbx database.Dbx, id Key) error {
	qs := sq.Delete(s.builder.Table()).
		Where(sq.Eq{s.builder.IdColumnName(): id})
	count, err := database.ExecWithBuilder(ctx, dbx, qs.PlaceholderFormat(sq.Dollar))
	if err != nil {
		return fmt.Errorf("error deleting model: %w", err)
	}
	if count == 0 {
		return errors.New("no rows deleted")
	}
	return nil
}

// Default filter implementation if no custom filter function is provided
// Find implements Resource.
func (s *QueryResourceService[Model, Key, Filter]) Find(ctx context.Context, dbx database.Dbx, filter *Filter) ([]*Model, error) {
	qs := sq.Select(s.builder.ColumnNamesTablePrefix()...).
		From(s.builder.Table())

	// Apply filters, sorting, and pagination
	qs = s.filter(qs, filter)
	qs = s.sort(qs, filter)
	qs = s.pagination(qs, filter)

	res, err := database.QueryWithBuilder[*Model](ctx, dbx, qs.PlaceholderFormat(sq.Dollar))
	if err != nil {
		return nil, fmt.Errorf("error finding models: %w", err)
	}
	return res, nil
}

// FindOne implements Resource.
func (s *QueryResourceService[Model, Key, Filter]) FindOne(ctx context.Context, dbx database.Dbx, filter *Filter) (*Model, error) {
	qs := sq.Select(s.builder.ColumnNamesTablePrefix()...).
		From(s.builder.Table())
	// Apply filters, sorting, and pagination
	qs = s.filter(qs, filter).Limit(1)
	res, err := database.QueryWithBuilder[*Model](ctx, dbx, qs.PlaceholderFormat(sq.Dollar))
	if err != nil {
		return nil, fmt.Errorf("error finding one model: %w", err)
	}
	if len(res) == 0 {
		return nil, nil
	}
	return res[0], nil
}

// FindByID implements Resource.
func (s *QueryResourceService[Model, Key, Filter]) FindByID(ctx context.Context, dbx database.Dbx, id Key) (*Model, error) {
	qs := sq.Select(s.builder.ColumnNamesTablePrefix()...).
		From(s.builder.Table()).Where(sq.Eq{s.builder.IdColumnName(): id}).Limit(1)
	res, err := database.QueryWithBuilder[*Model](ctx, dbx, qs.PlaceholderFormat(sq.Dollar))
	if err != nil {
		return nil, fmt.Errorf("error finding model by ID: %w", err)
	}
	if len(res) == 0 {
		return nil, nil
	}
	return res[0], nil
}

// Update implements Resource.
func (s *QueryResourceService[Model, Key, Filter]) Update(ctx context.Context, dbx database.Dbx, model *Model) (*Model, error) {
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

	res, err := database.QueryWithBuilder[*Model](ctx, dbx, qs.PlaceholderFormat(sq.Dollar))
	if err != nil {
		return nil, fmt.Errorf("error updating model: %w", err)
	}
	if len(res) == 0 {
		return nil, errors.New("no rows updated")
	}
	return res[0], nil
}

var (
	User *RepositoryResourceService[models.User, uuid.UUID, UserFilter] = NewRepositoryResourceService[models.User, uuid.UUID](
		repository.User,
		func(filter *UserFilter) *map[string]any {
			where := make(map[string]any)
			if filter == nil {
				return &where // return empty map if no filter is provided
			}

			if filter.EmailVerified.IsSet {
				emailverified := filter.EmailVerified.Value
				if emailverified {
					where[models.UserTable.EmailVerifiedAt] = map[string]any{
						repository.IsNotNull: nil,
					}
				} else {
					where[models.UserTable.EmailVerifiedAt] = map[string]any{
						repository.IsNull: nil,
					}
				}
			}
			if len(filter.Emails) > 0 {
				where["email"] = map[string]any{
					"_in": filter.Emails,
				}
			}
			if len(filter.Ids) > 0 {
				where["id"] = map[string]any{
					"_in": filter.Ids,
				}
			}
			if len(filter.Providers) > 0 {
				where["accounts"] = map[string]any{
					"provider": map[string]any{
						"_in": filter.Providers,
					},
				}
			}
			if len(filter.RoleIds) > 0 {
				where["roles"] = map[string]any{
					"id": map[string]any{
						"_in": filter.RoleIds,
					},
				}
			}
			if filter.Q != "" {
				where["_or"] = []map[string]any{
					{
						"email": map[string]any{
							"_ilike": "%" + filter.Q + "%",
						},
					},
					{
						"name": map[string]any{
							"_ilike": "%" + filter.Q + "%",
						},
					},
				}
			}
			if len(where) == 0 {
				return nil
			}
			return &where
		},
		nil,
		nil,
	)
	UserAccount *RepositoryResourceService[models.UserAccount, uuid.UUID, UserAccountFilter] = NewRepositoryResourceService[models.UserAccount, uuid.UUID](
		repository.UserAccount,
		func(filter *UserAccountFilter) *map[string]any {
			where := make(map[string]any)
			if filter == nil {
				return &where
			}
			if len(filter.Providers) > 0 {
				where[models.UserAccountTable.Provider] = map[string]any{
					"_in": filter.Providers,
				}
			}
			if len(filter.ProviderTypes) > 0 {
				where[models.UserAccountTable.Type] = map[string]any{
					"_in": filter.ProviderTypes,
				}
			}
			if len(filter.Ids) > 0 {
				where[models.UserAccountTable.ID] = map[string]any{
					"_in": filter.Ids,
				}
			}
			if len(filter.UserIds) > 0 {
				where[models.UserAccountTable.UserID] = map[string]any{
					"_in": filter.UserIds,
				}
			}
			return &where
		},
		nil,
		nil,
	)

	Token *RepositoryResourceService[models.Token, uuid.UUID, TokenFilter] = NewRepositoryResourceService[models.Token, uuid.UUID](
		repository.Token,
		func(filter *TokenFilter) *map[string]any {
			if filter == nil {
				return nil
			}
			where := make(map[string]any)
			if len(filter.UserIds) > 0 {
				where["user_id"] = map[string]any{"_in": filter.UserIds}
			}
			if len(filter.Ids) > 0 {
				where["id"] = map[string]any{"_in": filter.Ids}
			}
			if len(filter.Types) > 0 {
				where["type"] = map[string]any{"_in": filter.Types}
			}
			if len(filter.Identifiers) > 0 {
				where["identifier"] = map[string]any{"_in": filter.Identifiers}
			}
			if len(filter.Tokens) > 0 {
				where["token"] = map[string]any{"_in": filter.Tokens}
			}
			if filter.ExpiresAfter.IsSet {
				where["expires"] = map[string]any{"_gte": filter.ExpiresAfter.Value}
			}
			if filter.ExpiresBefore.IsSet {
				where["expires"] = map[string]any{"_lte": filter.ExpiresBefore.Value}
			}
			return &where
		},
		nil,
		nil,
	)
)
