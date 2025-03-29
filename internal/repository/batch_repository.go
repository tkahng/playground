package repository

import (
	"context"

	"github.com/stephenafamo/bob"
	"github.com/tkahng/authgo/internal/db/models"
)

func TruncateModels(ctx context.Context, db bob.Executor) error {
	return ErrorWrapper(ctx, db, false,
		models.Roles.Delete().Exec,
		models.Permissions.Delete().Exec,
		models.Tokens.Delete().Exec,
		models.UserSessions.Delete().Exec,
		models.UserAccounts.Delete().Exec,
		models.Users.Delete().Exec,
	)
}
