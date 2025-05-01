package repository

import (
	"context"

	"github.com/tkahng/authgo/internal/db/models"
	"github.com/tkahng/authgo/internal/shared"
)

// ListSessions implements AdminCrudActions.
func ListSessions(ctx context.Context, db Queryer, input *shared.PaginatedInput) ([]*models.UserSession, error) {
	panic("unimplemented")
}

// CountSessions implements AdminCrudActions.
func CountSessions(ctx context.Context, db Queryer) (int64, error) {
	q := models.UserSessions.Query()
	return CountExec(ctx, db, q)
}

// CountTokens implements AdminCrudActions.
func CountTokens(ctx context.Context, db Queryer) (int64, error) {
	q := models.Tokens.Query()
	return CountExec(ctx, db, q)
}

// ListTokens implements AdminCrudActions.
func ListTokens(ctx context.Context, db Queryer, input *shared.PaginatedInput) ([]*models.Token, error) {
	panic("unimplemented")
}
