package queries_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stephenafamo/bob"
	"github.com/tkahng/authgo/internal/db"
	"github.com/tkahng/authgo/internal/db/models"
	"github.com/tkahng/authgo/internal/db/models/factory"
	"github.com/tkahng/authgo/internal/queries"
	"github.com/tkahng/authgo/internal/test"
)

var (
	f = factory.New()
)

func TestExecWrapper(t *testing.T) {
	ctx, db, pl := test.DbSetup()
	t.Cleanup(func() {
		queries.TruncateModels(ctx, db)
		pl.Close()
	})
	_, err := f.NewUser(factory.UserMods.WithNewRoles(1, factory.RoleMods.WithNewPermissions(1))).Create(ctx, db)
	if err != nil {
		t.Fatal("error creating users", err)
	}
	user_count, err := models.Users.Query().Count(ctx, db)
	if err != nil {
		t.Fatal("error counting users", err)
	}
	roles_count, err := models.Roles.Query().Count(ctx, db)
	if err != nil {
		t.Fatal("error counting roles", err)
	}
	permissions_count, err := models.Permissions.Query().Count(ctx, db)
	if err != nil {
		t.Fatal("error counting permissions", err)
	}

	if user_count != 1 {
		t.Fatalf("expected 1 user, got %d", user_count)
	}
	if roles_count != 1 {
		t.Fatalf("expected 1 role, got %d", roles_count)
	}
	if permissions_count != 1 {
		t.Fatalf("expected 1 permission, got %d", permissions_count)
	}

	err = queries.ErrorWrapper(ctx, db, true,
		models.Users.Delete().Exec,
		models.Roles.Delete().Exec,
		models.Permissions.Delete().Exec,
	)
	if err != nil {
		t.Fatal("error deleting users", err)
	}
	new_user_count, err := models.Users.Query().Count(ctx, db)
	if err != nil {
		t.Fatal("error counting users", err)
	}
	new_roles_count, err := models.Roles.Query().Count(ctx, db)
	if err != nil {
		t.Fatal("error counting roles", err)
	}
	new_permissions_count, err := models.Permissions.Query().Count(ctx, db)
	if err != nil {
		t.Fatal("error counting permissions", err)
	}
	if new_user_count != 0 {
		t.Fatalf("expected 1 user, got %d", new_user_count)
	}
	if new_roles_count != 0 {
		t.Fatalf("expected 1 role, got %d", new_roles_count)
	}
	if new_permissions_count != 0 {
		t.Fatalf("expected 1 permission, got %d", new_permissions_count)
	}
}

func TestExecWrapperTransaction(t *testing.T) {
	ctx, dbx, pl := test.DbSetup()
	t.Cleanup(func() {
		queries.TruncateModels(ctx, dbx)
		pl.Close()
	})
	_, err := f.NewUser(factory.UserMods.WithNewRoles(1, factory.RoleMods.WithNewPermissions(1))).Create(ctx, dbx)
	if err != nil {
		t.Fatal("error creating users", err)
	}
	user_count, err := models.Users.Query().Count(ctx, dbx)
	if err != nil {
		t.Fatal("error counting users", err)
	}
	roles_count, err := models.Roles.Query().Count(ctx, dbx)
	if err != nil {
		t.Fatal("error counting roles", err)
	}
	permissions_count := queries.DefaultCountWrapper(ctx, dbx, models.Permissions.Query().Count)
	if err != nil {
		t.Fatal("error counting permissions", err)
	}
	if user_count != 1 {
		t.Fatalf("expected 1 user, got %d", user_count)
	}
	if roles_count != 1 {
		t.Fatalf("expected 1 role, got %d", roles_count)
	}
	if permissions_count != 1 {
		t.Fatalf("expected 1 permission, got %d", permissions_count)
	}
	err = func(ctx context.Context, plpool *pgxpool.Pool) error {
		var err error
		tx, err := plpool.Begin(ctx)
		if err != nil {
			t.Fatal("error creating transaction", err)
		}
		defer tx.Rollback(ctx)
		db := db.NewQueries(tx)
		err = queries.ErrorWrapper(ctx, db, true,
			models.Users.Delete().Exec,
			models.Roles.Delete().Exec,
			func(ctx context.Context, exec bob.Executor) (int64, error) {
				return 0, errors.New("error")
			},
			models.Permissions.Delete().Exec,
		)
		if err != nil {
			return err
		}
		tx.Commit(ctx)
		return err
	}(ctx, pl)

	if err == nil {
		t.Log("expected error", err)
	}

	new_user_count, err := models.Users.Query().Count(ctx, dbx)
	if err != nil {
		t.Fatal("error counting users", err)
	}
	new_roles_count, err := models.Roles.Query().Count(ctx, dbx)
	if err != nil {
		t.Fatal("error counting roles", err)
	}
	new_permissions_count, err := models.Permissions.Query().Count(ctx, dbx)
	if err != nil {
		t.Fatal("error counting permissions", err)
	}
	if new_user_count != 1 {
		t.Fatalf("expected 1 user, got %d", new_user_count)
	}
	if new_roles_count != 1 {
		t.Fatalf("expected 1 role, got %d", new_roles_count)
	}
	if new_permissions_count != 1 {
		t.Fatalf("expected 1 permission, got %d", new_permissions_count)
	}
}
