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
)

type RoleListFilter struct {
	PaginatedInput
	SortParams
	Q         string      `query:"q,omitempty" required:"false"`
	Ids       []uuid.UUID `query:"ids,omitempty" required:"false" minimum:"1" maximum:"100" format:"uuid"`
	Names     []string    `query:"names,omitempty" required:"false" minimum:"1" maximum:"100"`
	UserId    uuid.UUID   `query:"user_id,omitempty" required:"false" format:"uuid"`
	Reverse   string      `query:"reverse,omitempty" required:"false" doc:"When true, it will return the roles that do not match the filter criteria" enum:"user,product"`
	ProductId string      `query:"product_id,omitempty" required:"false"`
	Expand    []string    `query:"expand,omitempty" required:"false" minimum:"1" maximum:"100" enum:"users,permissions"`
}

type DbRbacStore struct {
	db database.Dbx
}

func (p *DbRbacStore) WithTx(tx database.Dbx) *DbRbacStore {
	return &DbRbacStore{
		db: tx,
	}
}

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

type CreateRoleDto struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}

func (p *DbRbacStore) CreateRole(ctx context.Context, role *CreateRoleDto) (*models.Role, error) {
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

type UpdateRoleDto struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}

func (p *DbRbacStore) UpdateRole(ctx context.Context, id uuid.UUID, roledto *UpdateRoleDto) error {
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
		role, err = p.CreateRole(ctx, &CreateRoleDto{Name: roleName})
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

func (p *DbRbacStore) CountRoles(ctx context.Context, filter *RoleListFilter) (int64, error) {
	q := squirrel.Select("COUNT(roles.*)").From("roles")

	q = p.filter(q, filter)

	data, err := database.QueryWithBuilder[database.CountOutput](ctx, p.db, q.PlaceholderFormat(squirrel.Dollar))
	if err != nil {
		return 0, err
	}
	if len(data) == 0 {
		return 0, nil
	}

	return data[0].Count, nil
}
func (p *DbRbacStore) ListRoles(ctx context.Context, input *RoleListFilter) ([]*models.Role, error) {
	q := squirrel.Select("roles.*").From("roles")

	q = p.filter(q, input)
	q = queryPagination(q, input)
	if input.SortBy != "" && input.SortOrder != "" {
		q = q.OrderBy(input.SortBy + " " + strings.ToUpper(input.SortOrder))
	}
	data, err := database.QueryWithBuilder[*models.Role](ctx, p.db, q.PlaceholderFormat(squirrel.Dollar))
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (p *DbRbacStore) filter(q squirrel.SelectBuilder, filter *RoleListFilter) squirrel.SelectBuilder {
	var sq = q
	if filter == nil {
		return sq
	}
	if filter.Q != "" {
		sq = sq.Where(
			squirrel.Or{
				squirrel.ILike{"name": "%" + filter.Q + "%"},
				squirrel.ILike{"description": "%" + filter.Q + "%"},
			})
	}
	if len(filter.Names) > 0 {
		sq = sq.Where(
			squirrel.Eq{
				"name": filter.Names,
			},
		)
	}
	if len(filter.Ids) > 0 {
		sq = sq.Where(
			squirrel.Eq{
				"id": filter.Ids,
			},
		)
	}
	if filter.UserId != uuid.Nil {
		if filter.Reverse == "user" {
			sq = sq.LeftJoin("user_roles"+" on "+"roles.id"+" = "+"user_roles"+"."+"role_id"+" and "+"user_roles"+"."+"user_id"+" = ?", filter.UserId)
			sq = sq.Where("user_roles.role_id is null")
		} else {
			sq = sq.Join("user_roles on roles.id = user_roles.role_id").Where(squirrel.Eq{"user_roles.user_id": filter.UserId})
		}
	}
	return sq
}

type DbRbacStoreInterface interface { // size=16 (0x10)
	AssignUserRoles(ctx context.Context, userId uuid.UUID, roleNames ...string) error
	CountNotUserPermissionSource(ctx context.Context, userId uuid.UUID) (int64, error)
	CountPermissions(ctx context.Context, filter *PermissionFilter) (int64, error)
	CountRoles(ctx context.Context, filter *RoleListFilter) (int64, error)
	CountUserPermissionSource(ctx context.Context, userId uuid.UUID) (int64, error)
	CreatePermission(ctx context.Context, name string, description *string) (*models.Permission, error)
	CreateProductPermissions(ctx context.Context, productId string, permissionIds ...uuid.UUID) error
	CreateProductRoles(ctx context.Context, productId string, roleIds ...uuid.UUID) error
	CreateRole(ctx context.Context, role *CreateRoleDto) (*models.Role, error)
	CreateRolePermissions(ctx context.Context, roleId uuid.UUID, permissionIds ...uuid.UUID) error
	CreateUserPermissions(ctx context.Context, userId uuid.UUID, permissionIds ...uuid.UUID) error
	CreateUserRoles(ctx context.Context, userId uuid.UUID, roleIds ...uuid.UUID) error
	DeletePermission(ctx context.Context, id uuid.UUID) error
	DeleteProductRoles(ctx context.Context, productId string, roleIds ...uuid.UUID) error
	DeleteProductPermissions(ctx context.Context, productId string, permissionIds ...uuid.UUID) error
	DeleteRole(ctx context.Context, id uuid.UUID) error
	DeleteRolePermissions(ctx context.Context, roleId uuid.UUID, permissionIds ...uuid.UUID) error
	DeleteUserRole(ctx context.Context, userId uuid.UUID, roleId uuid.UUID) error
	EnsureRoleAndPermissions(ctx context.Context, roleName string, permissionNames ...string) error
	FindOrCreatePermission(ctx context.Context, permissionName string) (*models.Permission, error)
	FindOrCreateRole(ctx context.Context, roleName string) (*models.Role, error)
	FindPermission(ctx context.Context, filter *PermissionFilter) (*models.Permission, error)
	FindPermissionById(ctx context.Context, id uuid.UUID) (*models.Permission, error)
	FindPermissionByName(ctx context.Context, name string) (*models.Permission, error)
	FindPermissionsByIds(ctx context.Context, params []uuid.UUID) ([]*models.Permission, error)
	FindRoleById(ctx context.Context, id uuid.UUID) (*models.Role, error)
	FindRoleByName(ctx context.Context, name string) (*models.Role, error)
	FindRolesByIds(ctx context.Context, params []uuid.UUID) ([]*models.Role, error)
	GetUserRoles(ctx context.Context, userIds ...uuid.UUID) ([][]*models.Role, error)
	ListPermissions(ctx context.Context, input *PermissionFilter) ([]*models.Permission, error)
	ListRoles(ctx context.Context, input *RoleListFilter) ([]*models.Role, error)
	ListUserNotPermissionsSource(ctx context.Context, userId uuid.UUID, limit int64, offset int64) ([]*models.PermissionSource, error)
	ListUserPermissionsSource(ctx context.Context, userId uuid.UUID, limit int64, offset int64) ([]*models.PermissionSource, error)
	LoadProductPermissions(ctx context.Context, productIds ...string) ([][]*models.Permission, error)
	LoadRolePermissions(ctx context.Context, roleIds ...uuid.UUID) ([][]*models.Permission, error)
	UpdatePermission(ctx context.Context, id uuid.UUID, roledto *UpdatePermissionDto) error
	UpdateRole(ctx context.Context, id uuid.UUID, roledto *UpdateRoleDto) error
}

var _ DbRbacStoreInterface = (*DbRbacStore)(nil)
