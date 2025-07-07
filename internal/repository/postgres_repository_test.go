package repository

import (
	"context"
	"testing"

	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/test"
	"github.com/tkahng/authgo/internal/tools/types"
)

type RepositoryTestCase[T any] struct {
	TestGet    func(t *testing.T, c RepositoryTest[T])
	TestPost   func(t *testing.T, c RepositoryTest[T])
	TestPut    func(t *testing.T)
	TestDelete func(t *testing.T)
}
type TestContext struct {
	Ctx context.Context
	Db  database.Dbx
}

type PostTest[T any] struct {
	Setup func(t *testing.T, ctx TestContext)
	Args  func(t *testing.T, ctx TestContext) []T
	Check func(t *testing.T, ctx TestContext, got, want T)
}
type GetParams struct {
	Where  *map[string]any
	Sort   *map[string]string
	Limit  int
	Offset int
}
type RepositoryTest[T any] struct {
	Repo        Repository[T]
	CompareFunc func(got, want T) bool
	SetupGet    func(t *testing.T, ctx TestContext) GetParams
	SetupPost   func(t *testing.T, ctx TestContext) []T
	SetupPut    func(t *testing.T, ctx TestContext) []T
	SetupDelete func(t *testing.T, ctx TestContext) []T
	ArgsFunc    func(t *testing.T, ctx TestContext) []T
	WhereFunc   func(t *testing.T, ctx TestContext) map[string]any
}

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
