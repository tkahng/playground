package stores_test

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/stores"
	"github.com/tkahng/authgo/internal/test"
)

func TestPostgresUserStore_CRUD(t *testing.T) {
	test.Short(t)
	ctx, dbx := test.DbSetup()
	dbx.RunInTransaction(ctx, func(dbxx database.Dbx) error {
		store := stores.NewPostgresUserStore(dbxx)

		// CreateUser
		email := "testuser@example.com"
		user, err := store.CreateUser(ctx, &models.User{
			Email: email,
			Name:  ptrString("Test User"),
		})
		if err != nil {
			t.Fatalf("CreateUser() error = %v", err)
		}
		if user.Email != email {
			t.Errorf("CreateUser() = %v, want email %v", user, email)
		}

		// FindUserByEmail
		found, err := store.FindUserByEmail(ctx, email)
		if err != nil || found == nil || found.ID != user.ID {
			t.Errorf("FindUserByEmail() = %v, err = %v", found, err)
		}

		// AssignUserRoles (no roles)
		err = store.AssignUserRoles(ctx, user.ID)
		if err != nil {
			t.Errorf("AssignUserRoles() with no roles error = %v", err)
		}

		// AssignUserRoles (with role)
		roleName := "basic"
		// Assume a role named "basic" exists in your DB for this test to pass
		err = store.AssignUserRoles(ctx, user.ID, roleName)
		if err != nil {
			t.Errorf("AssignUserRoles() error = %v", err)
		}

		// UpdateUser
		user.Name = ptrString("Updated Name")
		err = store.UpdateUser(ctx, user)
		if err != nil {
			t.Errorf("UpdateUser() error = %v", err)
		}

		// GetUserInfo
		info, err := store.GetUserInfo(ctx, email)
		if err != nil || info == nil || info.User.ID != user.ID {
			t.Errorf("GetUserInfo() = %v, err = %v", info, err)
		}
		if info != nil && info.User.Name != nil && *info.User.Name != "Updated Name" {
			t.Errorf("GetUserInfo() name = %v, want 'Updated Name'", info.User.Name)
		}

		// DeleteUser
		err = store.DeleteUser(ctx, user.ID)
		if err != nil {
			t.Errorf("DeleteUser() error = %v", err)
		}
		deleted, err := store.FindUserByEmail(ctx, email)
		if err != nil {
			t.Errorf("FindUserByEmail() after delete error = %v", err)
		}
		if deleted != nil {
			t.Errorf("User should be deleted, got = %v", deleted)
		}

		return errors.New("rollback")
	})
}

func TestPostgresUserStore_LoadUsersByUserIds(t *testing.T) {
	test.Short(t)
	ctx, dbx := test.DbSetup()
	dbx.RunInTransaction(ctx, func(dbxx database.Dbx) error {
		store := stores.NewPostgresUserStore(dbxx)
		user1, err := store.CreateUser(ctx, &models.User{Email: "loaduser1@example.com"})
		if err != nil {
			t.Fatalf("CreateUser() error = %v", err)
		}
		user2, err := store.CreateUser(ctx, &models.User{Email: "loaduser2@example.com"})
		if err != nil {
			t.Fatalf("CreateUser() error = %v", err)
		}
		ids := []uuid.UUID{user1.ID, user2.ID}
		users, err := store.LoadUsersByUserIds(ctx, ids...)
		if err != nil {
			t.Fatalf("LoadUsersByUserIds() error = %v", err)
		}
		if len(users) != 2 {
			t.Errorf("LoadUsersByUserIds() = %v, want 2 users", len(users))
		}
		if users[0] == nil || users[1] == nil {
			t.Errorf("Expected non-nil users, got: %v", users)
		}
		return errors.New("rollback")
	})
}

func ptrString(s string) *string {
	return &s
}
