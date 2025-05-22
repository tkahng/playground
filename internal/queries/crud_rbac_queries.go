package queries

var (
	PermissionColumnNames = []string{"id", "name", "description", "created_at", "updated_at"}
)

// ListPermissions implements AdminCrudActions.

type CountOutput struct {
	Count int64
}

// ListRoles implements AdminCrudActions.

// CountRoles implements AdminCrudActions.
