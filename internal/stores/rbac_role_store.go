package stores

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/repository"
	"github.com/tkahng/authgo/internal/services"
	"github.com/tkahng/authgo/internal/shared"
)

type DbRbacStore struct {
	db database.Dbx
}

type RBACStore struct {
	*DbRbacStore
}

var _ services.RBACStore = &DbRbacStore{}

func NewDbRBACStore(db database.Dbx) *DbRbacStore {
	return &DbRbacStore{
		db: db,
	}
}

func (s *DbRbacStore) FindRolesByIds(ctx context.Context, params []uuid.UUID) ([]*models.Role, error) {
	if len(params) == 0 {
		return nil, nil
	}
	newIds := make([]string, len(params))
	for i, id := range params {
		newIds[i] = id.String()
	}
	return repository.Role.Get(
		ctx,
		s.db,
		&map[string]any{
			models.RoleTable.ID: map[string]any{
				"_in": newIds,
			},
		},
		&map[string]string{
			models.RoleTable.Name: "asc",
		},
		nil,
		nil,
	)
}
func (a *DbRbacStore) FindRoleById(ctx context.Context, id uuid.UUID) (*models.Role, error) {
	return repository.Role.GetOne(ctx, a.db, &map[string]any{
		models.RoleTable.ID: map[string]any{
			"_eq": id,
		},
	})
}

func (a *DbRbacStore) FindRoleByName(ctx context.Context, name string) (*models.Role, error) {
	return repository.Role.GetOne(
		ctx,
		a.db,
		&map[string]any{
			models.RoleTable.Name: map[string]any{
				"_eq": name,
			},
		},
	)
}

func (p *DbRbacStore) CreateRole(ctx context.Context, role *shared.CreateRoleDto) (*models.Role, error) {
	if role == nil {
		return nil, fmt.Errorf("role is nil")
	}
	data, err := repository.Role.PostOne(ctx, p.db, &models.Role{
		Name:        role.Name,
		Description: role.Description,
	})
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (p *DbRbacStore) UpdateRole(ctx context.Context, id uuid.UUID, roledto *shared.UpdateRoleDto) error {
	role, err := repository.Role.GetOne(
		ctx,
		p.db,
		&map[string]any{
			models.RoleTable.ID: map[string]any{
				"_eq": id,
			},
		},
	)
	if err != nil {
		return err
	}
	if role == nil {
		return nil
	}
	role.Name = roledto.Name
	role.Description = roledto.Description
	_, err = repository.Role.PutOne(ctx, p.db, role)
	if err != nil {
		return err
	}
	return nil
}

func (p *DbRbacStore) DeleteRole(ctx context.Context, id uuid.UUID) error {
	_, err := repository.Role.Delete(
		ctx,
		p.db,
		&map[string]any{
			models.RoleTable.ID: map[string]any{
				"_eq": id,
			},
		},
	)
	if err != nil {
		return err
	}
	return nil
}

func (p *DbRbacStore) FindOrCreateRole(ctx context.Context, roleName string) (*models.Role, error) {
	role, err := repository.Role.GetOne(
		ctx,
		p.db,
		&map[string]any{
			models.RoleTable.Name: map[string]any{
				"_eq": roleName,
			},
		},
	)
	if err != nil {
		return nil, err
	}
	if role == nil {
		role, err = p.CreateRole(ctx, &shared.CreateRoleDto{Name: roleName})
		if err != nil {
			return nil, err
		}
	}
	return role, nil
}

func (p *DbRbacStore) EnsureRoleAndPermissions(ctx context.Context, roleName string, permissionNames ...string) error {
	// find superuser role
	role, err := p.FindOrCreateRole(ctx, roleName)
	if err != nil {
		return err
	}
	for _, permissionName := range permissionNames {
		perm, err := p.FindOrCreatePermission(ctx, permissionName)
		if err != nil {
			slog.ErrorContext(ctx, "error finding or creating permission", "name", permissionName, "error", err)
			continue
		}
		if perm == nil {
			continue
		}
		err = p.CreateRolePermissions(ctx, role.ID, perm.ID)
		if err != nil && !database.IsUniqConstraintErr(err) {
			log.Println(err)
		}
	}
	return nil
}

func (p *DbRbacStore) CountRoles(ctx context.Context, filter *shared.RoleListFilter) (int64, error) {
	q := squirrel.Select("COUNT(roles.*)").From("roles")

	q = ListRolesFilterFuncQuery(q, filter)

	data, err := database.QueryWithBuilder[database.CountOutput](ctx, p.db, q.PlaceholderFormat(squirrel.Dollar))
	if err != nil {
		return 0, err
	}
	if len(data) == 0 {
		return 0, nil
	}

	return data[0].Count, nil
}
func (p *DbRbacStore) ListRoles(ctx context.Context, input *shared.RolesListParams) ([]*models.Role, error) {
	q := squirrel.Select("roles.*").From("roles")
	filter := input.RoleListFilter
	pageInput := &input.PaginatedInput

	// q = ViewApplyPagination(q, pageInput)
	q = ListRolesFilterFuncQuery(q, &filter)
	q = database.Paginate(q, pageInput)
	if input.SortBy != "" && input.SortOrder != "" {
		q = q.OrderBy(input.SortBy + " " + strings.ToUpper(input.SortOrder))
	}
	data, err := database.QueryWithBuilder[*models.Role](ctx, p.db, q.PlaceholderFormat(squirrel.Dollar))
	if err != nil {
		return nil, err
	}
	return data, nil
}

func ListRolesFilterFuncQuery(sq squirrel.SelectBuilder, filter *shared.RoleListFilter) squirrel.SelectBuilder {
	// where := make(map[string]any)
	if filter == nil {
		return sq
	}
	if filter.Q != "" {
		sq = sq.Where(
			squirrel.Or{
				squirrel.ILike{"name": "%" + filter.Q + "%"},
				squirrel.ILike{"description": "%" + filter.Q + "%"},
			},
		)

	}
	if len(filter.Names) > 0 {
		sq = sq.Where(squirrel.Eq{"name": filter.Names})
	}
	if len(filter.Ids) > 0 {
		sq = sq.Where(squirrel.Eq{"id": filter.Ids})
	}

	if filter.UserId != "" {

		if filter.Reverse == "user" {
			sq = sq.LeftJoin(
				"user_roles"+" on "+"roles.id"+" = "+"user_roles"+"."+"role_id"+" and "+"user_roles"+"."+"user_id"+" = ?",
				filter.UserId,
			)
			sq = sq.Where("user_roles.role_id is null")

		} else {
			sq = sq.Join("user_roles on roles.id = user_roles.role_id").
				Where(squirrel.Eq{"user_roles.user_id": filter.UserId})

		}
	}
	return sq
}
