package repository

import (
	"context"
	"testing"

	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/test"
	"github.com/tkahng/playground/internal/tools/types"
)

func UserCompareFunc(got, want models.User) bool {
	if got.Name != want.Name {
		return false
	}
	if got.Email != want.Email {
		return false
	}
	if got.Image != want.Image {
		return false
	}
	if got.EmailVerifiedAt != want.EmailVerifiedAt {
		return false
	}
	return true
}

func TestRepository(t *testing.T) {
	t.Run("test post", func(t *testing.T) {
		test.WithTx(t, func(ctx context.Context, db database.Dbx) {
			record := models.User{
				Name:  types.Pointer("Test User"),
				Email: "test@example.com",
			}

			users, err := User.Post(ctx, db, []models.User{record})
			if err != nil || users == nil {
				t.Fatalf("Failed to create user: %v", err)
			}
			if len(users) != 1 {
				t.Fatalf("Expected 1 user, got %d", len(users))
			}
			user := users[0]
			if UserCompareFunc(*user, record) {
				t.Fatalf("Expected user to be different, got %v", user)
			}
		})
	})
	t.Run("test post exec", func(t *testing.T) {
		test.WithTx(t, func(ctx context.Context, db database.Dbx) {
			record := models.User{
				Name:  types.Pointer("Test User"),
				Email: "test@example.com",
			}

			users, err := User.PostExec(ctx, db, []models.User{record})
			if err != nil || users != 1 {
				t.Fatalf("Failed to create user: %v", err)
			}
		})
	})
}
