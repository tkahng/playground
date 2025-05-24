package apis_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/danielgtaylor/huma/v2/humatest"
	"github.com/tkahng/authgo/internal/apis"
	"github.com/tkahng/authgo/internal/conf"
	"github.com/tkahng/authgo/internal/core"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/test"
)

func createVerifiedUser(app core.App) (*shared.UserInfo, error) {
	nw := time.Now()
	user, err := app.User().Store().CreateUser(context.Background(), &models.User{
		Email:           "authenticated@example.com",
		EmailVerifiedAt: &nw,
	})
	if err != nil {
		return nil, err
	}
	_, err = app.UserAccount().Store().CreateUserAccount(context.Background(), &models.UserAccount{
		UserID:            user.ID,
		Provider:          models.ProvidersGoogle,
		Type:              "oauth",
		ProviderAccountID: "google-123",
	})
	if err != nil {
		return nil, err
	}
	return &shared.UserInfo{
		User: *shared.FromUserModel(user),
	}, nil
}

// func createTeam(app core.App, user *shared.User) *shared.TeamInfo {
// 	return must(app.Team().Store().CreateTeam(context.Background(), &models.Team{
// 		Name:  "test team",
// 		Owner: user.ID,
// 	})
// }

// func TestIndexGet(t *testing.T) {
// 	test.DbSetup()
// 	test.WithTx(t, func(ctx context.Context, db database.Dbx) {
// 		cfg := conf.ZeroEnvConfig()
// 		app := core.NewDecorator(ctx, cfg, db)
// 		appApi := apis.NewApi(app)
// 		_, api := humatest.New(t)
// 		apis.AddRoutes(api, appApi)

//			_, err := api.Get("/team", header)
//			assert.Nil(t, err)
//		})
//	}
func TestGetGreeting(t *testing.T) {
	_, api := humatest.New(t)
	cfg := conf.ZeroEnvConfig()
	ctx, db := test.DbSetup()
	app := core.NewDecorator(ctx, cfg, db)
	appApi := apis.NewApi(app)
	apis.AddRoutes(api, appApi)

	resp := api.Get("/")
	if !strings.Contains(resp.Body.String(), "public") {
		t.Fatalf("Unexpected response: %s", resp.Body.String())
	}
}

func TestTeamSlug(t *testing.T) {
	test.DbSetup()
	test.WithTx(t, func(ctx context.Context, db database.Dbx) {
		_, api := humatest.New(t)
		cfg := conf.ZeroEnvConfig()
		app := core.NewDecorator(ctx, cfg, db)
		appApi := apis.NewApi(app)
		apis.AddRoutes(api, appApi)
		user, err := createVerifiedUser(app)
		if err != nil {
			t.Errorf("Error creating user: %v", err)
			return
		}
		tokensVerifiedTokens, err := app.Auth().CreateAuthTokensFromEmail(context.Background(), user.User.Email)
		if err != nil {
			t.Errorf("Error creating auth tokens: %v", err)
			return
		}
		_, err = app.Team().Store().CreateTeam(context.Background(), "test team",
			"public")
		VerifiedHeader := fmt.Sprintf("Authorization: Bearer %s", tokensVerifiedTokens.Tokens.AccessToken)
		resp := api.Post("/teams/check-slug", VerifiedHeader, struct {
			Slug string `json:"slug" required:"true"`
		}{
			Slug: "public",
		},
		)
		if resp.Code != 200 {
			t.Fatalf("Unexpected response: %s", resp.Body.String())
		}
		resp2 := api.Post("/teams/check-slug", VerifiedHeader, struct {
			Slug string `json:"slug" required:"true"`
		}{
			Slug: "baba",
		},
		)
		if !strings.Contains(resp2.Body.String(), "true") {
			t.Fatalf("Unexpected response: %s", resp2.Body.String())
		}
	})
}

func TestCreateTeam(t *testing.T) {
	test.DbSetup()

	test.WithTx(t, func(ctx context.Context, db database.Dbx) {
		cfg := conf.ZeroEnvConfig()
		app := core.NewDecorator(ctx, cfg, db)
		appApi := apis.NewApi(app)
		_, api := humatest.New(t)
		apis.AddRoutes(api, appApi)
		user, err := createVerifiedUser(app)
		if err != nil {
			t.Errorf("Error creating user: %v", err)
			return
		}
		tokensVerifiedTokens, err := app.Auth().CreateAuthTokensFromEmail(context.Background(), user.User.Email)
		if err != nil {
			t.Errorf("Error creating auth tokens: %v", err)
			return
		}
		VerifiedHeader := fmt.Sprintf("Authorization: Bearer %s", tokensVerifiedTokens.Tokens.AccessToken)
		sdasd := struct {
			name             string
			ctxUserInfo      *shared.UserInfo
			createTeamErr    error
			createTeamResult *shared.TeamInfo
			expectedErr      error
			expectedOutput   *apis.TeamOutput
			header           string
			body             *apis.CreateTeamInput
		}{
			name:   "successful team creation",
			header: VerifiedHeader,
			body: &apis.CreateTeamInput{
				Name: "test team",
				Slug: "test-team",
			},
		}

		resp := api.Post("/teams", sdasd.header, sdasd.body)
		if resp.Code != 200 {
			t.Errorf("Api.GetStripeSubscriptions() = %v, want %v", resp.Code, 200)
		}

	})
}
