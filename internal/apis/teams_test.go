//go:build integration

package apis_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/tkahng/playground/internal/apis"
	"github.com/tkahng/playground/internal/core"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/database/repository"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/stores"
	"github.com/tkahng/playground/internal/test"
	apphttp "github.com/tkahng/playground/internal/tools/http"
	"github.com/tkahng/playground/internal/tools/types"
	"github.com/tkahng/playground/internal/tools/utils"
)

func TestGetGreeting(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		api := testApi.TestApi

		resp := api.Get("/")
		if !strings.Contains(resp.Body.String(), "public") {
			t.Fatalf("Unexpected response: %s", resp.Body.String())
		}
	})
}

func TestTeamSlug(t *testing.T) {
	// t.Parallel()
	test.SkipIfShort(t)
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		api := testApi.TestApi
		user := core.CreateUserWithOptions(t, testApi.App, core.UserWithVerified(time.Now()))
		_ = core.CreateTeamAndMemberWithOptions(t, testApi.App, &user.User, core.TeamWithName("public"))
		tokensVerifiedTokens := core.CreateTokenHeader(t, testApi.App, user.User.Email)

		resp := api.Post("/teams/check-slug", tokensVerifiedTokens, struct {
			Slug string `json:"slug" required:"true"`
		}{
			Slug: "public",
		},
		)
		type body struct {
			Exists bool `json:"exists"`
		}
		res, err := utils.UnmarshalJSON[body](resp.Body.Bytes())
		assert.NoError(t, err)
		assert.False(t, res.Exists)
		assert.Equal(t, 200, resp.Code)
	})
}
func TestTeamSlug_FailRegex(t *testing.T) {
	// t.Parallel()
	test.SkipIfShort(t)
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		api := testApi.TestApi
		user := core.CreateUserWithOptions(t, testApi.App, core.UserWithVerified(time.Now()))
		_ = core.CreateTeamAndMemberWithOptions(t, testApi.App, &user.User, core.TeamWithName("public"))
		tokensVerifiedTokens := core.CreateTokenHeader(t, testApi.App, user.User.Email)

		resp := api.Post("/teams/check-slug", tokensVerifiedTokens, struct {
			Slug string `json:"slug" required:"true"`
		}{
			Slug: "_^&sdas9=",
		},
		)

		res, err := utils.UnmarshalJSON[apphttp.ErrorModel](resp.Body.Bytes())
		assert.NoError(t, err)
		assert.Equal(t, 400, res.Status)
	})
}
func TestTeamSlug_TooShort(t *testing.T) {
	// t.Parallel()
	test.SkipIfShort(t)
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		api := testApi.TestApi
		user := core.CreateUserWithOptions(t, testApi.App, core.UserWithVerified(time.Now()))
		_ = core.CreateTeamAndMemberWithOptions(t, testApi.App, &user.User, core.TeamWithName("public"))
		tokensVerifiedTokens := core.CreateTokenHeader(t, testApi.App, user.User.Email)

		resp := api.Post("/teams/check-slug", tokensVerifiedTokens, struct {
			Slug string `json:"slug" required:"true"`
		}{
			Slug: "9=",
		},
		)

		res, err := utils.UnmarshalJSON[apphttp.ErrorModel](resp.Body.Bytes())
		assert.NoError(t, err)
		assert.Equal(t, 422, res.Status)
	})
}

func TestGetTeam_unauthorized(t *testing.T) {
	// t.Parallel()
	test.SkipIfShort(t)
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		api := testApi.TestApi
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
	// t.Parallel()
	test.SkipIfShort(t)
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		api := testApi.TestApi
		user := core.CreateUserWithOptions(t, testApi.App, core.UserWithVerified(time.Now()))
		teamIdString := uuid.NewString()
		tokensVerifiedTokens := core.CreateTokenHeader(t, testApi.App, user.User.Email)

		resp := api.Get("/teams/"+teamIdString+"23", tokensVerifiedTokens)
		if resp.Code == 200 {
			t.Fatalf("Unexpected response: %s", resp.Body.String())
		}
		assert.Equal(t, 400, resp.Code)
		assert.Contains(t, resp.Body.String(), "invalid UUID format")
	},
	)
}

func TestGetTeam_success(t *testing.T) {
	// t.Parallel()
	test.SkipIfShort(t)
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		app := testApi.App
		api := testApi.TestApi
		user := core.CreateUserWithOptions(t, testApi.App, core.UserWithVerified(time.Now()))
		team := core.CreateTeamAndMemberWithOptions(t, app, &user.User, core.TeamWithName("test team"))
		teamIdString := team.Team.ID.String()
		// team, err :=
		tokensVerifiedTokens, err := app.Auth().GenerateAuthTokens(context.Background(), user.User.Email)
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

func TestCreateTeam_Failed(t *testing.T) {
	t.Run("failed: unknown error during db customer creation", func(t *testing.T) {
		database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
			appApi := apis.SetupApi(t, ctx, db)
			adapter := core.ExtractAdapterDecorator(t, appApi.App)
			var customerStore *stores.CustomerStoreDecorator
			if m, ok := adapter.Customer().(*stores.CustomerStoreDecorator); ok {
				customerStore = m
			} else {
				t.Fatal("mailer is not a TestMailer")
			}
			user := core.CreateUserWithOptions(t, appApi.App, core.UserWithVerified(time.Now()))
			customerStore.CreateCustomerFunc = func(ctx context.Context, customer *models.StripeCustomer) (*models.StripeCustomer, error) {
				return nil, errors.New("unknown error")
			}
			tokensVerifiedTokens, err := appApi.App.Auth().GenerateAuthTokens(ctx, user.User.Email)
			if err != nil {
				t.Errorf("Error creating auth tokens: %v", err)
				return
			}
			VerifiedHeader := fmt.Sprintf("Authorization: Bearer %s", tokensVerifiedTokens.Tokens.AccessToken)

			resp := appApi.TestApi.Post("/teams", VerifiedHeader, &apis.CreateTeamInput{
				Name: "test team",
			})
			assert.Equal(t, 500, resp.Code, "expected 500 status code")
			assert.Contains(t, resp.Body.String(), "unknown error")
			teamCount := repository.MustCountAllCtx(t, ctx, repository.Team, db, nil)
			assert.Equal(t, 0, int(teamCount))
			customerCount := repository.MustCountAllCtx(t, ctx, repository.StripeCustomer, db, nil)
			assert.Equal(t, 1, int(customerCount))
			teamMembers := repository.MustFindWithOptionsCtx(t, ctx, repository.TeamMember, db)
			assert.Equal(t, 0, len(teamMembers))
		})
	})
	t.Run("failed: emailNotVerified", func(t *testing.T) {
		database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
			appApi := apis.SetupApi(t, ctx, db)
			user := core.CreateUserWithOptions(t, appApi.App)
			tokensVerifiedTokens, err := appApi.App.Auth().GenerateAuthTokens(ctx, user.User.Email)
			if err != nil {
				t.Errorf("Error creating auth tokens: %v", err)
				return
			}
			VerifiedHeader := fmt.Sprintf("Authorization: Bearer %s", tokensVerifiedTokens.Tokens.AccessToken)

			resp := appApi.TestApi.Post("/teams", VerifiedHeader, &apis.CreateTeamInput{
				Name: "test team",
			})
			assert.Equal(t, 401, resp.Code, "expected 401 status code")
			assert.Contains(t, resp.Body.String(), "email not verified", "expected error message containing 'email not verified'")
			teamCount := repository.MustCountAllCtx(t, ctx, repository.Team, db, nil)
			assert.Equal(t, 0, int(teamCount))
			customerCount := repository.MustCountAllCtx(t, ctx, repository.StripeCustomer, db, nil)
			assert.Equal(t, 0, int(customerCount))
			teamMembers := repository.MustFindWithOptionsCtx(t, ctx, repository.TeamMember, db)
			assert.Equal(t, 0, len(teamMembers))
		})
	})
}
func TestCreateTeam_Success_OptionalSlug_TeamNameIsUrlSafeAndUnique(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		appApi := apis.SetupApi(t, ctx, db)
		for i := range 10 {
			scenario := apis.ApiScenario{
				Name:           fmt.Sprintf("test create team success optional slug %d", i),
				Method:         http.MethodPost,
				URL:            "/teams",
				ExpectedStatus: http.StatusOK,
				TestAppFactory: func(t testing.TB) *apis.TestApi {
					return appApi
				},
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
					userEmail := fmt.Sprintf("%d@gmail.com", i)
					teamName := fmt.Sprintf("team%d", i)
					user := core.CreateUserWithOptions(t, app, core.UserWithEmail(userEmail), core.UserWithVerifiedNow())
					token, _ := core.CreateAccessHeaderAndRefreshToken(t, app, user.User.Email)
					scenario.Headers = []string{token}
					scenario.Body = strings.NewReader(`{"name":` + fmt.Sprintf(`"%s"`, teamName) + `}`)
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
					var body apis.TeamWithMember
					err := json.NewDecoder(res.Body).Decode(&body)
					if err != nil {
						t.Errorf("Error decoding response: %v", err)
					}
					slug := fmt.Sprintf("team%d", i)
					if body.Team.Slug != slug {
						t.Errorf("Expected slug %s, got %s", slug, body.Team.Slug)
					}
				},
			}
			scenario.Test(t)
		}
	})
}

func TestCreateTeam_Success_SlugConflict_NumericSuffix(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		appApi := apis.SetupApi(t, ctx, db)
		for i := range 10 {
			scenario := apis.ApiScenario{
				Name:           fmt.Sprintf("test create team slug conflict uses numeric suffix %d", i),
				Method:         http.MethodPost,
				URL:            "/teams",
				ExpectedStatus: http.StatusOK,
				TestAppFactory: func(t testing.TB) *apis.TestApi {
					return appApi
				},
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
					userEmail := fmt.Sprintf("%d@gmail.com", i)
					otherUserEmail := fmt.Sprintf("%d+other@gmail.com", i)
					teamName := fmt.Sprintf("team%d", i)
					user := core.CreateUserWithOptions(t, app, core.UserWithEmail(userEmail), core.UserWithVerifiedNow())
					otherUser := core.CreateUserWithOptions(t, app, core.UserWithEmail(otherUserEmail), core.UserWithVerifiedNow())
					team := core.CreateTeamAndMemberWithOptions(t, app, &otherUser.User, core.TeamWithName(teamName), core.TeamWithSlug(teamName))
					if team == nil {
						t.Fatalf("Error creating team")
					}
					token, _ := core.CreateAccessHeaderAndRefreshToken(t, app, user.User.Email)
					scenario.Headers = []string{token}
					scenario.Body = strings.NewReader(`{"name":` + fmt.Sprintf(`"%s"`, teamName) + `}`)
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
					var body apis.TeamWithMember
					err := json.NewDecoder(res.Body).Decode(&body)
					if err != nil {
						t.Errorf("Error decoding response: %v", err)
					}
					baseSlug := fmt.Sprintf("team%d", i)
					expectedSlug := fmt.Sprintf("team%d-1", i)
					if body.Team.Slug == baseSlug {
						t.Errorf("slug should not equal taken base slug %s", baseSlug)
					}
					if body.Team.Slug != expectedSlug {
						t.Errorf("expected numeric suffix slug %s, got %s", expectedSlug, body.Team.Slug)
					}
				},
			}
			scenario.Test(t)
		}
	})
}
func TestCreateTeam_Success(t *testing.T) {
	t.Run("success: create team slug generated from name", func(t *testing.T) {
		database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
			appApi := apis.SetupApi(t, ctx, db)
			user := core.CreateUserWithOptions(t, appApi.App, core.UserWithVerified(time.Now()))
			tokensVerifiedTokens, err := appApi.App.Auth().GenerateAuthTokens(ctx, user.User.Email)
			if err != nil {
				t.Errorf("Error creating auth tokens: %v", err)
				return
			}
			VerifiedHeader := fmt.Sprintf("Authorization: Bearer %s", tokensVerifiedTokens.Tokens.AccessToken)

			resp := appApi.TestApi.Post("/teams", VerifiedHeader, &apis.CreateTeamInput{
				Name: "My Test Team",
			})
			if resp.Code != 200 {
				t.Errorf("unexpected status: %v body: %s", resp.Code, resp.Body.String())
			}
			teamRes, err := utils.UnmarshalJSON[apis.Team](resp.Body.Bytes())
			assert.NoError(t, err)
			assert.Equal(t, "my-test-team", teamRes.Slug)

			team := repository.MustFindOneCtx(t, ctx, repository.Team, db, nil)
			assert.NotNil(t, team)
			customer, customerErr := appApi.App.Payment().FindCustomerByTeamId(ctx, teamRes.ID)
			assert.NoError(t, customerErr)
			assert.NotNil(t, customer)
			teamMembers := repository.MustFindWithOptionsCtx(t, ctx, repository.TeamMember, db)
			assert.Equal(t, 1, len(teamMembers))
			assert.Equal(t, models.TeamMemberRoleOwner, teamMembers[0].Role)
		})
	})
}

func TestUpdateTeam_failedNotOwner(t *testing.T) {
	// t.Parallel()
	test.SkipIfShort(t)
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		app := testApi.App
		api := testApi.TestApi
		user1 := core.CreateUserWithOptions(
			t,
			testApi.App,
			core.UserWithVerified(time.Now()),
			core.UserWithEmail("user1@example"),
		)
		team := core.CreateTeamAndMemberWithOptions(t, app, &user1.User)
		user2 := core.CreateUserWithOptions(
			t,
			testApi.App,
			core.UserWithVerified(time.Now()),
			core.UserWithEmail("user2@example"),
		)

		_ = core.CreateTeamMemberWithOptions(
			t,
			app,
			team.Team.ID,
			user2.User.ID,
			core.TeamWithRole(models.TeamMemberRoleMember),
			core.TeamWithBilling(false),
		)
		// create
		VerifiedHeader := core.CreateTokenHeader(t, app, user2.User.Email)
		resp := api.Put("/teams/"+team.Team.ID.String(), VerifiedHeader, &apis.UpdateTeamInput{
			TeamID: team.Team.ID.String(),
			Body: apis.UpdateTeamDto{
				Name: "test team",
				Slug: "test-team",
			},
		})
		if resp.Code == 200 {
			t.Fatalf("Unexpected response: %v", resp.Code)
		}
		if resp.Code != 403 {
			t.Fatalf("Unexpected response: %v", resp.Code)
		}
		if !strings.Contains(resp.Body.String(), "Forbidden") {
			t.Fatalf("Unexpected response: %s", resp.Body.String())
		}
	})
}

func TestUpdateTeam_successOwner(t *testing.T) {
	// t.Parallel()
	test.SkipIfShort(t)
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		app := testApi.App
		api := testApi.TestApi
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

		member1, err := app.Team().CreateTeamWithOwner(
			ctx,
			"test team",
			user1.ID,
		)
		if err != nil {
			t.Errorf("Error creating user: %v", err)
			return
		}

		// create
		tokensVerifiedTokens, err := app.Auth().GenerateAuthTokens(ctx, user1.Email)
		if err != nil {
			t.Errorf("Error creating auth tokens: %v", err)
			return
		}
		VerifiedHeader := fmt.Sprintf("Authorization: Bearer %s", tokensVerifiedTokens.Tokens.AccessToken)
		resp := api.Put("/teams/"+member1.Team.ID.String(), VerifiedHeader, apis.UpdateTeamInput{
			TeamID: member1.Team.ID.String(),
			Body: apis.UpdateTeamDto{
				Name: "test team",
				Slug: "test-team",
			},
		}.Body)
		if resp.Code != 200 {
			t.Fatalf("Unexpected response: %s", resp.Body.String())
		}
	})
}

func TestDeleteTeam_successOwner(t *testing.T) {
	// t.Parallel()
	test.SkipIfShort(t)
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		app := testApi.App
		api := testApi.TestApi
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

		member1, err := app.Team().CreateTeamWithOwner(
			ctx,
			"test team",
			user1.ID,
		)
		if err != nil {
			t.Errorf("Error creating user: %v", err)
			return
		}

		// create
		tokensVerifiedTokens, err := app.Auth().GenerateAuthTokens(ctx, user1.Email)
		if err != nil {
			t.Errorf("Error creating auth tokens: %v", err)
			return
		}
		VerifiedHeader := fmt.Sprintf("Authorization: Bearer %s", tokensVerifiedTokens.Tokens.AccessToken)
		resp := api.Delete("/teams/"+member1.Team.ID.String(), VerifiedHeader)
		fmt.Println("resp", resp.Body.String())
		if resp.Code != 200 {
			t.Fatalf("Unexpected response: %s", resp.Body.String())
		}
	})
}
func TestDeleteTeam_failNonOwner(t *testing.T) {
	// t.Parallel()
	test.SkipIfShort(t)
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		app := testApi.App
		api := testApi.TestApi

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

		member1, err := app.Team().CreateTeamWithOwner(
			ctx,
			"test team",
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
		member2, err := app.Adapter().TeamMember().CreateTeamMember(ctx, &models.TeamMember{
			TeamID:           member1.Team.ID,
			UserID:           types.Pointer(user2.ID),
			Role:             models.TeamMemberRoleMember,
			HasBillingAccess: false,
			Active:           true,
		})
		if member2 == nil {
			t.Errorf("Error creating user: %v", err)
			return
		}
		// create
		tokensVerifiedTokens, err := app.Auth().GenerateAuthTokens(ctx, user2.Email)
		if err != nil {
			t.Errorf("Error creating auth tokens: %v", err)
			return
		}
		VerifiedHeader := fmt.Sprintf("Authorization: Bearer %s", tokensVerifiedTokens.Tokens.AccessToken)
		resp := api.Delete("/teams/"+member1.Team.ID.String(), VerifiedHeader)
		if resp.Code != 403 {
			t.Fatalf("Unexpected response: %s", resp.Body.String())
		}
		if !strings.Contains(resp.Body.String(), "Forbidden") {
			t.Fatalf("Unexpected response: %s", resp.Body.String())
		}
	})
}
