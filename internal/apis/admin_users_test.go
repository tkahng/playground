package apis_test

import (
	"context"
	"testing"

	"github.com/tkahng/playground/internal/apis"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/test"
)

func TestApi_AdminUsersList(t *testing.T) {
	test.Parallel(t)
	test.SkipIfShort(t)
	test.WithTx(t, func(ctx context.Context, db database.Dbx) {
		// testApi := SetupApi(t, ctx, db)
		// api := testApi.TestApi
		type args struct {
			url  string
			args []any
		}
		tests := []struct {
			name    string
			args    args
			want    *apis.ApiPaginatedOutput[*apis.ApiUser]
			wantErr bool
		}{
			// TODO: Add test cases.
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				// got, err := api.Get(tt.args.url, tt.args.input)
				// if (err != nil) != tt.wantErr {
				// 	t.Errorf("Api.AdminUsers() error = %v, wantErr %v", err, tt.wantErr)
				// 	return
				// }
				// if !reflect.DeepEqual(got, tt.want) {
				// 	t.Errorf("Api.AdminUsers() = %v, want %v", got, tt.want)
				// }
			})
		}
	})
}

func TestApi_AdminUsersCreate(t *testing.T) {
	test.Parallel(t)
	test.SkipIfShort(t)
	test.WithTx(t, func(ctx context.Context, db database.Dbx) {
		// testApi := SetupApi(t, ctx, db)
		// api := testApi.TestApi
		// user, err := createAdminUser(testApi.App)
		// if err != nil {
		// 	t.Errorf("Error creating user: %v", err)
		// 	return
		// }
		// tokensVerifiedTokens, err := app.Auth().CreateAuthTokensFromEmail(context.Background(), user.User.Email)
		// if err != nil {
		// 	t.Errorf("Error creating auth tokens: %v", err)
		// 	return
		// }
		// VerifiedHeader := fmt.Sprintf("Authorization: Bearer %s", tokensVerifiedTokens.Tokens.AccessToken)
		// resp := api.Post("/admin/users", VerifiedHeader, &apis.CreateAdminUserInput{
		// 	Email: "user1@example",
		// })
		// if resp.Code == 200 {
		// 	t.Fatalf("Unexpected response: %s", resp.Body.String())
		// }
	})
}
