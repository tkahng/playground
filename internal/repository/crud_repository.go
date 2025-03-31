package repository

import (
	"context"

	"github.com/stephenafamo/bob"

	"github.com/tkahng/authgo/internal/db/models"
	"github.com/tkahng/authgo/internal/shared"
)

// ListUserAccounts implements AdminCrudActions.
func ListUserAccounts(ctx context.Context, db bob.DB, input *shared.PaginatedInput) ([]*models.UserAccount, error) {
	q := models.UserAccounts.Query()
	pageInput := input
	ViewApplyPagination(q, pageInput)
	data, err := q.All(ctx, db)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// CountUserAccounts implements AdminCrudActions.
func CountUserAccounts(ctx context.Context, db bob.DB) (int64, error) {
	q := models.UserAccounts.Query()
	return CountExec(ctx, db, q)
}

// ListSessions implements AdminCrudActions.
func ListSessions(ctx context.Context, db bob.DB, input *shared.PaginatedInput) ([]*models.UserSession, error) {
	panic("unimplemented")
}

// CountSessions implements AdminCrudActions.
func CountSessions(ctx context.Context, db bob.DB) (int64, error) {
	q := models.UserSessions.Query()
	return CountExec(ctx, db, q)
}

// ListPermissions implements AdminCrudActions.
func ListPermissions(ctx context.Context, db bob.DB, input *shared.PaginatedInput) ([]models.Permission, error) {
	panic("unimplemented")
}

// CountPermissions implements AdminCrudActions.
func CountPermissions(ctx context.Context, db bob.DB) (int64, error) {
	q := models.Permissions.Query()
	return CountExec(ctx, db, q)
}

// ListRoles implements AdminCrudActions.
func ListRoles(ctx context.Context, db bob.DB, input *shared.PaginatedInput) ([]models.Role, error) {
	panic("unimplemented")
}

// CountRoles implements AdminCrudActions.
func CountRoles(ctx context.Context, db bob.DB) (int64, error) {
	q := models.Roles.Query()
	return CountExec(ctx, db, q)
}

// CountTokens implements AdminCrudActions.
func CountTokens(ctx context.Context, db bob.DB) (int64, error) {
	q := models.Tokens.Query()
	return CountExec(ctx, db, q)
}

// ListTokens implements AdminCrudActions.
func ListTokens(ctx context.Context, db bob.DB, input *shared.PaginatedInput) ([]*models.Token, error) {
	panic("unimplemented")
}
