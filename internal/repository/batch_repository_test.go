package repository_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stephenafamo/bob"
	"github.com/tkahng/authgo/internal/db/models"
	"github.com/tkahng/authgo/internal/db/models/factory"
	"github.com/tkahng/authgo/internal/repository"
	"github.com/tkahng/authgo/internal/test"
)

var (
	f = factory.New()
)

func TestExecWrapper(t *testing.T) {
	ctx, db, pl := test.DbSetup()
	t.Cleanup(func() {
		repository.TruncateModels(ctx, db)
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

	err = repository.ErrorWrapper(ctx, db, true,
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
	ctx, db, pl := test.DbSetup()
	t.Cleanup(func() {
		repository.TruncateModels(ctx, db)
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
	permissions_count := repository.DefaultCountWrapper(ctx, db, models.Permissions.Query().Count)
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
	err = func(ctx context.Context, db bob.DB) error {
		var err error
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			t.Fatal("error creating transaction", err)
		}
		defer tx.Rollback()

		err = repository.ErrorWrapper(ctx, tx, true,
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
		tx.Commit()
		return err
	}(ctx, db)

	if err == nil {
		t.Log("expected error", err)
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
