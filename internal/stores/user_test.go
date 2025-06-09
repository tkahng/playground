package stores_test

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/stores"
	"github.com/tkahng/authgo/internal/test"
)

func TestUserStore_CRUD(t *testing.T) {
	test.Short(t)
	ctx, dbx := test.DbSetup()
	_ = dbx.RunInTx( func(dbxx database.Dbx) error {
		store := stores.NewDbUserStore(dbxx)

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
		found, err := store.FindUser(ctx, &models.User{
			Email: email,
		})
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
		deleted, err := store.FindUser(ctx, &models.User{
			Email: email,
		})
		if err != nil {
			t.Errorf("FindUserByEmail() after delete error = %v", err)
		}
		if deleted != nil {
			t.Errorf("User should be deleted, got = %v", deleted)
		}

		return errors.New("rollback")
	})
}

func TestUserStore_LoadUsersByUserIds(t *testing.T) {
	test.Short(t)
	ctx, dbx := test.DbSetup()
	_ = dbx.RunInTx( func(dbxx database.Dbx) error {
		store := stores.NewDbUserStore(dbxx)
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

func TestUserStore_FindUserById(t *testing.T) {
	test.Short(t)
	ctx, dbx := test.DbSetup()
	_ = dbx.RunInTx( func(dbxx database.Dbx) error {
		p := stores.NewDbUserStore(dbxx)
		type fields struct {
			db database.Dbx
		}
		type args struct {
			ctx    context.Context
			userId uuid.UUID
		}
		tests := []struct {
			name    string
			fields  fields
			args    args
			want    *models.User
			wantErr bool
		}{}
		for i := range 10 {
			user, err := p.CreateUser(
				ctx,
				&models.User{Email: fmt.Sprintf("testuser%d@example.com", i)},
			)
			if err != nil {
				t.Fatalf("CreateUser() error = %v", err)
			}
			tests = append(tests, struct {
				name    string
				fields  fields
				args    args
				want    *models.User
				wantErr bool
			}{
				name: fmt.Sprintf("FindUserByID-%s", user.ID.String()),
				fields: fields{
					db: dbxx,
				},
				args: args{
					ctx:    ctx,
					userId: user.ID,
				},
				want:    user,
				wantErr: false,
			})
		}

		tests = append(tests, struct {
			name    string
			fields  fields
			args    args
			want    *models.User
			wantErr bool
		}{
			name: "NotFound",
			fields: fields{
				db: dbxx,
			},
			args: args{
				ctx:    ctx,
				userId: uuid.New(),
			},
			want:    nil,
			wantErr: false,
		})

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				p := stores.NewDbUserStore(tt.fields.db)
				got, err := p.FindUserByID(tt.args.ctx, tt.args.userId)
				if (err != nil) != tt.wantErr {
					t.Errorf("PostgresUserStore.FindUserByID() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("PostgresUserStore.FindUserByID() = %v, want %v", got, tt.want)
				}
			})
		}

		return errors.New("rollback")
	})
}
