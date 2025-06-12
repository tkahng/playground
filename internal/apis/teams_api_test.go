package apis_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/danielgtaylor/huma/v2/humatest"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/tkahng/authgo/internal/apis"
	"github.com/tkahng/authgo/internal/conf"
	"github.com/tkahng/authgo/internal/core"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/test"
	"github.com/tkahng/authgo/internal/tools/utils"
)

func createTeamAndMember(app core.App, user *shared.User, teamName string) (*shared.TeamInfoModel, error) {

	team, err := app.Adapter().TeamGroup().CreateTeam(context.Background(), teamName, strings.TrimSpace(teamName))
	if err != nil {
		return nil, err
	}
	member, err := app.Adapter().TeamMember().CreateTeamMember(context.Background(), team.ID, user.ID, models.TeamMemberRoleOwner, true)
	if err != nil {
		return nil, err
	}
	return &shared.TeamInfoModel{
		Team: *team,
		User: models.User{
			ID:              user.ID,
			Name:            user.Name,
			EmailVerifiedAt: user.EmailVerifiedAt,
		},
		Member: *member,
	}, nil
}
func createVerifiedUser(app core.App) (*shared.UserInfo, error) {
	nw := time.Now()
	user, err := app.Adapter().User().CreateUser(context.Background(), &models.User{
		Email:           "authenticated@example.com",
		EmailVerifiedAt: &nw,
	})
	if err != nil {
		return nil, err
	}
	_, err = app.Adapter().UserAccount().CreateUserAccount(context.Background(), &models.UserAccount{
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
func createUnverifiedUser(app *core.BaseAppDecorator) (*shared.UserInfo, error) {
	user, err := app.Adapter().User().CreateUser(context.Background(), &models.User{
		Email: "authenticated@example.com",
	})
	if err != nil {
		return nil, err
	}
	_, err = app.Adapter().UserAccount().CreateUserAccount(context.Background(), &models.UserAccount{
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
		_, err = app.Adapter().TeamGroup().CreateTeam(context.Background(), "test team",
			"public")
		if err != nil {
			t.Errorf("Error creating team: %v", err)
			return
		}
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

func TestGetTeam_unauthorized(t *testing.T) {
	test.DbSetup()
	test.WithTx(t, func(ctx context.Context, db database.Dbx) {

		cfg := conf.ZeroEnvConfig()
		app := core.NewDecorator(ctx, cfg, db)
		appApi := apis.NewApi(app)
		_, api := humatest.New(t)
		apis.AddRoutes(api, appApi)

		t.Run("Unauthorized access", func(t *testing.T) {
			resp := api.Get("/teams/"+uuid.NewString(), "")
			if resp.Code == 200 {
				t.Fatalf("Unexpected response: %s", resp.Body.String())
			}
		})
	},
	)
}

func TestGetTeam_invalidID(t *testing.T) {
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

		teamIdString := uuid.NewString()
		tokensVerifiedTokens, err := app.Auth().CreateAuthTokensFromEmail(context.Background(), user.User.Email)
		if err != nil {
			t.Errorf("Error creating auth tokens: %v", err)
			return
		}
		VerifiedHeader := fmt.Sprintf("Authorization: Bearer %s", tokensVerifiedTokens.Tokens.AccessToken)

		resp := api.Get("/teams/"+teamIdString+"23", VerifiedHeader)
		if resp.Code == 200 {
			t.Fatalf("Unexpected response: %s", resp.Body.String())
		}
		assert.Equal(t, 400, resp.Code)
		assert.Contains(t, resp.Body.String(), "invalid UUID format")
	},
	)

}

func TestGetTeam_success(t *testing.T) {
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
		team, err := createTeamAndMember(app, &user.User, "test team")
		if err != nil {
			t.Errorf("Error creating user: %v", err)
			return
		}
		teamIdString := team.Team.ID.String()
		// team, err :=
		tokensVerifiedTokens, err := app.Auth().CreateAuthTokensFromEmail(context.Background(), user.User.Email)
		if err != nil {
			t.Errorf("Error creating auth tokens: %v", err)
			return
		}
		VerifiedHeader := fmt.Sprintf("Authorization: Bearer %s", tokensVerifiedTokens.Tokens.AccessToken)
		resp := api.Get("/teams/"+teamIdString, VerifiedHeader)
		if resp.Code != 200 {
			t.Fatalf("Unexpected response: %s", resp.Body.String())
		}
	},
	)

}
func TestCreateTeam_SuccessfulCreation(t *testing.T) {
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
			Name             string
			ctxUserInfo      *shared.UserInfo
			createTeamErr    error
			createTeamResult *shared.TeamInfoModel
			expectedErr      error
			expectedOutput   *apis.TeamOutput
			header           string
			body             *apis.CreateTeamInput
		}{
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

func TestCreateTeam_emailNotVerified(t *testing.T) {
	test.DbSetup()
	test.WithTx(t, func(ctx context.Context, db database.Dbx) {

		cfg := conf.ZeroEnvConfig()
		app := core.NewDecorator(ctx, cfg, db)
		appApi := apis.NewApi(app)
		_, api := humatest.New(t)
		apis.AddRoutes(api, appApi)
		user, err := createUnverifiedUser(app)
		if err != nil {
			t.Errorf("Error creating user: %v", err)
			return
		}
		// create
		tokensVerifiedTokens, err := app.Auth().CreateAuthTokensFromEmail(ctx, user.User.Email)
		if err != nil {
			t.Errorf("Error creating auth tokens: %v", err)
			return
		}
		VerifiedHeader := fmt.Sprintf("Authorization: Bearer %s", tokensVerifiedTokens.Tokens.AccessToken)
		resp := api.Post("/teams", VerifiedHeader, &apis.CreateTeamInput{
			Name: "test team",
			Slug: "test-team",
		})
		if resp.Code == 200 {
			t.Fatalf("Unexpected response: %s", resp.Body.String())
		}
		assert.Equal(t, 401, resp.Code)
		assert.Contains(t, resp.Body.String(), "email not verified")
	},
	)
}

// test team update api when not owner and fail
func TestUpdateTeam_failedNotOwner(t *testing.T) {
	test.DbSetup()
	test.WithTx(t, func(ctx context.Context, db database.Dbx) {
		cfg := conf.ZeroEnvConfig()
		app := core.NewDecorator(ctx, cfg, db)

		appApi := apis.NewApi(app)
		_, api := humatest.New(t)
		apis.AddRoutes(api, appApi)
		user1, err := app.Adapter().User().CreateUser(
			ctx,
			&models.User{
				Email: "user1@example",
			},
		)
		if err != nil {
			t.Errorf("Error creating user: %v", err)
			return
		}

		member1, err := app.Team().CreateTeam(
			ctx,
			"test team",
			"test-team",
			user1.ID,
		)
		if err != nil {
			t.Errorf("Error creating user: %v", err)
			return
		}
		user2, err := app.Adapter().User().CreateUser(
			ctx,
			&models.User{
				Email: "user2@example",
			},
		)
		if err != nil {
			t.Errorf("Error creating user: %v", err)
			return
		}
		member2, err := app.Adapter().TeamMember().CreateTeamMember(
			ctx,
			member1.Team.ID,
			user2.ID,
			models.TeamMemberRoleMember,
			false,
		)
		if member2 == nil {
			t.Errorf("Error creating user: %v", err)
			return
		}
		// create
		tokensVerifiedTokens, err := app.Auth().CreateAuthTokensFromEmail(ctx, user2.Email)
		if err != nil {
			t.Errorf("Error creating auth tokens: %v", err)
			return
		}
		VerifiedHeader := fmt.Sprintf("Authorization: Bearer %s", tokensVerifiedTokens.Tokens.AccessToken)
		resp := api.Put("/teams/"+member1.Team.ID.String(), VerifiedHeader, &apis.UpdateTeamInput{
			TeamID: member1.Team.ID.String(),
			Body: struct {
				Name string "json:\"name\" required:\"true\""
			}{Name: "test team"},
		})
		if resp.Code == 200 {
			t.Fatalf("Unexpected response: %s", resp.Body.String())
		}
		assert.Equal(t, 403, resp.Code)
		assert.Contains(t, resp.Body.String(), "You do not have the required team member role")
	})
}

func TestUpdateTeam_successOwner(t *testing.T) {
	test.DbSetup()
	test.WithTx(t, func(ctx context.Context, db database.Dbx) {
		cfg := conf.ZeroEnvConfig()
		app := core.NewDecorator(ctx, cfg, db)

		appApi := apis.NewApi(app)
		_, api := humatest.New(t)
		apis.AddRoutes(api, appApi)
		user1, err := app.Adapter().User().CreateUser(
			ctx,
			&models.User{
				Email: "user1@example",
			},
		)
		if err != nil {
			t.Errorf("Error creating user: %v", err)
			return
		}

		member1, err := app.Team().CreateTeam(
			ctx,
			"test team",
			"test-team",
			user1.ID,
		)
		if err != nil {
			t.Errorf("Error creating user: %v", err)
			return
		}

		// create
		tokensVerifiedTokens, err := app.Auth().CreateAuthTokensFromEmail(ctx, user1.Email)
		if err != nil {
			t.Errorf("Error creating auth tokens: %v", err)
			return
		}
		VerifiedHeader := fmt.Sprintf("Authorization: Bearer %s", tokensVerifiedTokens.Tokens.AccessToken)
		resp := api.Put("/teams/"+member1.Team.ID.String(), VerifiedHeader, apis.UpdateTeamInput{
			TeamID: member1.Team.ID.String(),
			Body: struct {
				Name string "json:\"name\" required:\"true\""
			}{Name: "test team"},
		}.Body)
		if resp.Code != 200 {
			t.Fatalf("Unexpected response: %s", resp.Body.String())
		}
	})
}

func TestDeleteTeam_successOwner(t *testing.T) {
	test.DbSetup()
	test.WithTx(t, func(ctx context.Context, db database.Dbx) {
		cfg := conf.ZeroEnvConfig()
		app := core.NewDecorator(ctx, cfg, db)

		appApi := apis.NewApi(app)
		_, api := humatest.New(t)
		apis.AddRoutes(api, appApi)
		user1, err := app.Adapter().User().CreateUser(
			ctx,
			&models.User{
				Email: "user1@example",
			},
		)
		if err != nil {
			t.Errorf("Error creating user: %v", err)
			return
		}

		member1, err := app.Team().CreateTeam(
			ctx,
			"test team",
			"test-team",
			user1.ID,
		)
		if err != nil {
			t.Errorf("Error creating user: %v", err)
			return
		}

		// create
		tokensVerifiedTokens, err := app.Auth().CreateAuthTokensFromEmail(ctx, user1.Email)
		if err != nil {
			t.Errorf("Error creating auth tokens: %v", err)
			return
		}
		VerifiedHeader := fmt.Sprintf("Authorization: Bearer %s", tokensVerifiedTokens.Tokens.AccessToken)
		resp := api.Delete("/teams/"+member1.Team.ID.String(), VerifiedHeader)
		if resp.Code != 200 {
			t.Fatalf("Unexpected response: %s", resp.Body.String())
		}
	})
}
func TestDeleteTeam_failNonOwner(t *testing.T) {
	test.DbSetup()
	test.WithTx(t, func(ctx context.Context, db database.Dbx) {
		cfg := conf.ZeroEnvConfig()
		app := core.NewDecorator(ctx, cfg, db)

		appApi := apis.NewApi(app)
		_, api := humatest.New(t)
		apis.AddRoutes(api, appApi)
		user1, err := app.Adapter().User().CreateUser(
			ctx,
			&models.User{
				Email: "user1@example",
			},
		)
		if err != nil {
			t.Errorf("Error creating user: %v", err)
			return
		}

		member1, err := app.Team().CreateTeam(
			ctx,
			"test team",
			"test-team",
			user1.ID,
		)
		if err != nil {
			t.Errorf("Error creating user: %v", err)
			return
		}
		user2, err := app.Adapter().User().CreateUser(
			ctx,
			&models.User{
				Email: "user2@example",
			},
		)
		if err != nil {
			t.Errorf("Error creating user: %v", err)
			return
		}
		member2, err := app.Adapter().TeamMember().CreateTeamMember(
			ctx,
			member1.Team.ID,
			user2.ID,
			models.TeamMemberRoleMember,
			false,
		)
		if member2 == nil {
			t.Errorf("Error creating user: %v", err)
			return
		}
		// create
		tokensVerifiedTokens, err := app.Auth().CreateAuthTokensFromEmail(ctx, user2.Email)
		if err != nil {
			t.Errorf("Error creating auth tokens: %v", err)
			return
		}
		VerifiedHeader := fmt.Sprintf("Authorization: Bearer %s", tokensVerifiedTokens.Tokens.AccessToken)
		resp := api.Delete("/teams/"+member1.Team.ID.String(), VerifiedHeader)
		if resp.Code != 403 {
			t.Fatalf("Unexpected response: %s", resp.Body.String())
		}
		if !strings.Contains(resp.Body.String(), "You do not have the required team member role") {
			t.Fatalf("Unexpected response: %s", resp.Body.String())
		}
	})
}

func TestGetActiveTeamMember_success(t *testing.T) {
	test.DbSetup()
	test.WithTx(t, func(ctx context.Context, db database.Dbx) {
		cfg := conf.ZeroEnvConfig()
		app := core.NewDecorator(ctx, cfg, db)

		appApi := apis.NewApi(app)
		_, api := humatest.New(t)
		apis.AddRoutes(api, appApi)
		user1, err := app.Adapter().User().CreateUser(
			ctx,
			&models.User{
				Email: "user1@example",
			},
		)
		if err != nil {
			t.Errorf("Error creating user: %v", err)
			return
		}

		member1, err := app.Team().CreateTeam(
			ctx,
			"test team",
			"test-team",
			user1.ID,
		)
		if err != nil {
			t.Errorf("Error creating user: %v", err)
			return
		}
		if member1 == nil {
			t.Errorf("Error creating user: %v", err)
			return
		}

		tokensVerifiedTokens, err := app.Auth().CreateAuthTokensFromEmail(ctx, user1.Email)
		if err != nil {
			t.Errorf("Error creating auth tokens: %v", err)
			return
		}
		VerifiedHeader := fmt.Sprintf("Authorization: Bearer %s", tokensVerifiedTokens.Tokens.AccessToken)
		resp := api.Get("/team-members/active", VerifiedHeader)
		if resp.Code != 200 {
			t.Fatalf("Unexpected response: %s", resp.Body.String())
		}
		obj, err := utils.UnmarshalJSON[models.TeamMember](resp.Body.Bytes())
		if err != nil {
			t.Fatalf("error marshaling response: %v", err)
		}
		if obj.ID != member1.Member.ID {
			t.Fatalf("wrong member id. expected: %v, got: %v", member1.Member.ID, obj.ID)
		}
	})
}

func TestGetActiveTeamMember_nomember(t *testing.T) {
	test.DbSetup()
	test.WithTx(t, func(ctx context.Context, db database.Dbx) {
		cfg := conf.ZeroEnvConfig()
		app := core.NewDecorator(ctx, cfg, db)

		appApi := apis.NewApi(app)
		_, api := humatest.New(t)
		apis.AddRoutes(api, appApi)
		user1, err := app.Adapter().User().CreateUser(
			ctx,
			&models.User{
				Email: "user1@example",
			},
		)
		if err != nil {
			t.Errorf("Error creating user: %v", err)
			return
		}

		tokensVerifiedTokens, err := app.Auth().CreateAuthTokensFromEmail(ctx, user1.Email)
		if err != nil {
			t.Errorf("Error creating auth tokens: %v", err)
			return
		}
		VerifiedHeader := fmt.Sprintf("Authorization: Bearer %s", tokensVerifiedTokens.Tokens.AccessToken)
		resp := api.Get("/team-members/active", VerifiedHeader)
		if resp.Code != 404 {
			t.Fatalf("Unexpected response: %s", resp.Body.String())
		}
	})
}

func TestGetUserTeamMembers_basic(t *testing.T) {
	test.DbSetup()
	test.WithTx(t, func(ctx context.Context, db database.Dbx) {
		cfg := conf.ZeroEnvConfig()
		app := core.NewDecorator(ctx, cfg, db)

		appApi := apis.NewApi(app)
		_, api := humatest.New(t)
		apis.AddRoutes(api, appApi)
		user1, err := app.Adapter().User().CreateUser(
			ctx,
			&models.User{
				Email: "user1@example",
			},
		)
		if err != nil {
			t.Errorf("Error creating user: %v", err)
			return
		}

		member1, err := app.Team().CreateTeam(
			ctx,
			"test team",
			"test-team",
			user1.ID,
		)
		if err != nil {
			t.Errorf("Error creating user: %v", err)
			return
		}
		if member1 == nil {
			t.Errorf("Error creating user: %v", err)
			return
		}

		tokensVerifiedTokens, err := app.Auth().CreateAuthTokensFromEmail(ctx, user1.Email)
		if err != nil {
			t.Errorf("Error creating auth tokens: %v", err)
			return
		}
		VerifiedHeader := fmt.Sprintf("Authorization: Bearer %s", tokensVerifiedTokens.Tokens.AccessToken)
		resp := api.Get("/team-members", VerifiedHeader)
		if resp.Code != 200 {
			t.Fatalf("Unexpected response: %s", resp.Body.String())
		}
		obj, err := utils.UnmarshalJSON[shared.PaginatedResponse[*shared.TeamMember]](resp.Body.Bytes())
		if err != nil {
			t.Fatalf("error marshaling response: %v", err)
		}
		if len(obj.Data) == 0 || obj.Data[0].ID != member1.Member.ID {
			t.Fatalf("wrong member id. expected: %v, got: %v", member1.Member.ID, obj.Data[0].ID)
		}
	})
}

func TestGetUserTeamMembers_sortbyname(t *testing.T) {
	test.DbSetup()
	test.WithTx(t, func(ctx context.Context, db database.Dbx) {
		cfg := conf.ZeroEnvConfig()
		app := core.NewDecorator(ctx, cfg, db)

		appApi := apis.NewApi(app)
		_, api := humatest.New(t)
		apis.AddRoutes(api, appApi)
		user1, err := app.Adapter().User().CreateUser(
			ctx,
			&models.User{
				Email: "user1@example",
			},
		)
		if err != nil {
			t.Errorf("Error creating user: %v", err)
			return
		}

		member1, err := app.Team().CreateTeam(
			ctx,
			"test team a",
			"test-team a",
			user1.ID,
		)
		if err != nil {
			t.Errorf("Error creating user: %v", err)
			return
		}
		if member1 == nil {
			t.Errorf("Error creating user: %v", err)
			return
		}
		member2, err := app.Team().CreateTeam(
			ctx,
			"test team b",
			"test-team b",
			user1.ID,
		)
		if err != nil {
			t.Errorf("Error creating user: %v", err)
			return
		}
		if member2 == nil {
			t.Errorf("Error creating user: %v", err)
			return
		}

		tokensVerifiedTokens, err := app.Auth().CreateAuthTokensFromEmail(ctx, user1.Email)
		if err != nil {
			t.Errorf("Error creating auth tokens: %v", err)
			return
		}
		VerifiedHeader := fmt.Sprintf("Authorization: Bearer %s", tokensVerifiedTokens.Tokens.AccessToken)
		resp := api.Get("/team-members?sort_by=team.name&sort_order=asc", VerifiedHeader)
		if resp.Code != 200 {
			t.Fatalf("Unexpected response: %s", resp.Body.String())
		}
		obj, err := utils.UnmarshalJSON[shared.PaginatedResponse[*shared.TeamMember]](resp.Body.Bytes())
		if err != nil {
			t.Fatalf("error marshaling response: %v", err)
		}
		if len(obj.Data) == 0 || obj.Data[0].ID != member1.Member.ID {
			t.Fatalf("wrong member id. expected: %v, got: %v", member1.Member.ID, obj.Data[0].ID)
		}
	})
}
func TestGetUserTeamMembers_sortbyname2(t *testing.T) {
	test.DbSetup()
	test.WithTx(t, func(ctx context.Context, db database.Dbx) {
		cfg := conf.ZeroEnvConfig()
		app := core.NewDecorator(ctx, cfg, db)

		appApi := apis.NewApi(app)
		_, api := humatest.New(t)
		apis.AddRoutes(api, appApi)
		user1, err := app.Adapter().User().CreateUser(
			ctx,
			&models.User{
				Email: "user1@example",
			},
		)
		if err != nil {
			t.Errorf("Error creating user: %v", err)
			return
		}

		member1, err := app.Team().CreateTeam(
			ctx,
			"test team a",
			"test-team a",
			user1.ID,
		)
		if err != nil {
			t.Errorf("Error creating user: %v", err)
			return
		}
		if member1 == nil {
			t.Errorf("Error creating user: %v", err)
			return
		}
		member2, err := app.Team().CreateTeam(
			ctx,
			"test team b",
			"test-team b",
			user1.ID,
		)
		if err != nil {
			t.Errorf("Error creating user: %v", err)
			return
		}
		if member2 == nil {
			t.Errorf("Error creating user: %v", err)
			return
		}

		tokensVerifiedTokens, err := app.Auth().CreateAuthTokensFromEmail(ctx, user1.Email)
		if err != nil {
			t.Errorf("Error creating auth tokens: %v", err)
			return
		}
		VerifiedHeader := fmt.Sprintf("Authorization: Bearer %s", tokensVerifiedTokens.Tokens.AccessToken)
		resp := api.Get("/team-members?sort_by=team.name&sort_order=desc", VerifiedHeader)
		if resp.Code != 200 {
			t.Fatalf("Unexpected response: %s", resp.Body.String())
		}
		obj, err := utils.UnmarshalJSON[shared.PaginatedResponse[*shared.TeamMember]](resp.Body.Bytes())
		if err != nil {
			t.Fatalf("error marshaling response: %v", err)
		}
		if len(obj.Data) == 0 || obj.Data[0].ID != member2.Member.ID {
			t.Fatalf("wrong member id. expected: %v, got: %v", member2.Member.ID, obj.Data[0].ID)
		}
	})
}
