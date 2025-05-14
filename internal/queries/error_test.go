package queries_test

import (
	"errors"
	"testing"

	"github.com/tkahng/authgo/internal/db"
	"github.com/tkahng/authgo/internal/queries"
	"github.com/tkahng/authgo/internal/test"
)

func TestIsUniqConstraintErr(t *testing.T) {
	test.Short(t)
ctx, dbx := test.DbSetup()
	dbx.RunInTransaction(ctx, func(dbxx db.Dbx) error {
		// _, err := queries.CreateUser(ctx, dbxx, &shared.AuthenticationInput{
		// 	Email: "test@example.com",
		// })
		// if err != nil {
		// 	return err
		// }
		// _, err = queries.CreateUser(ctx, dbxx, &shared.AuthenticationInput{
		// 	Email: "test@example.com",
		// })
		type args struct {
			err error
		}
		tests := []struct {
			name string
			args args
			want bool
		}{
			{
				name: "Unique constraint error",
				args: args{
					err: errors.New("duplicate key value violates unique constraint \"users_email_unique\" (SQLSTATE 23505)"),
				},
				want: true,
			},
			{
				name: "Not unique constraint error",
				args: args{
					err: errors.New("duplicate key value violates unique constraint \"users_email_unique\" (SQLSTATE 23505)"),
				},
				want: true,
			},
			{
				name: "Other error",
				args: args{
					err: errors.New("some other error"),
				},
				want: false,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if got := queries.IsUniqConstraintErr(tt.args.err); got != tt.want {
					t.Errorf("IsUniqConstraintErr() = %v, want %v", got, tt.want)
				}
			})
		}
		return test.EndTestErr
	})
}
