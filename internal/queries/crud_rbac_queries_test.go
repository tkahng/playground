package queries_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/tkahng/authgo/internal/db"
	"github.com/tkahng/authgo/internal/queries"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/test"
)

func TestListPermissions(t *testing.T) {
	ctx, dbx := test.DbSetup()
	dbx.RunInTransaction(ctx, func(tx db.Dbx) error {
		err := queries.EnsureRoleAndPermissions(
			ctx,
			tx,
			shared.PermissionNameAdmin,
			shared.PermissionNameAdmin,
			shared.PermissionNameBasic,
		)
		if err != nil {
			t.Fatalf("failed to ensure role and permissions: %v", err)
		}
		type args struct {
			ctx   context.Context
			db    db.Dbx
			input *shared.PermissionsListParams
		}
		tests := []struct {
			name      string
			args      args
			wantCount int
			wantErr   bool
		}{
			{
				name: "Test Case 1",
				args: args{
					ctx: ctx,
					db:  tx,
					input: &shared.PermissionsListParams{
						PermissionsListFilter: shared.PermissionsListFilter{
							// Q: "super",
							Names: []string{
								shared.PermissionNameAdmin,
							},
						},
						PaginatedInput: shared.PaginatedInput{
							Page:    0,
							PerPage: 20,
						},
					},
				},
				wantCount: 1,
				wantErr:   false,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := queries.ListPermissions(tt.args.ctx, tt.args.db, tt.args.input)
				if (err != nil) != tt.wantErr {
					t.Errorf("ListPermissions() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(len(got), tt.wantCount) {
					t.Errorf("ListPermissions() = %v, want %v", len(got), tt.wantCount)
				}
			})
		}
		return test.EndTestErr
	})
}
