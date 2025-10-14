// nolint:exhaustruct
package repository

import (
	"context"
	"testing"

	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/test"
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
	test.WithTx(t, func(ctx context.Context, tx database.Dbx) {

	})
}
