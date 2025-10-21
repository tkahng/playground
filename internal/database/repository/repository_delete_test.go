package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/test"
	"github.com/tkahng/playground/internal/tools/types"
)

type DeleteScenario[T any] struct {
	// Name is the name of the scenario
	Name string
	// Dbx is the database connection
	Dbx database.Dbx
	// Repo is the repository to be tested
	Repo Repository[T]
	// Args is the arguments to be passed to the repository
	Args *map[string]any
	// ArgsFunc returns the arguments to be passed to the repository. this is for arguments that need some computation.
	ArgsFunc func(t testing.TB, ctx context.Context, scenario *DeleteScenario[T]) *map[string]any
	// SetupFunc is the function to setup the test
	SetupFunc func(t testing.TB, ctx context.Context, scenario *DeleteScenario[T])
	// TestFunc is the function to verify the Delete result
	TestFunc func(t testing.TB, ctx context.Context, scenario *DeleteScenario[T], res int64)
	// ㅉantErr indicates whether the test should expect an error
	WantErr bool
	// CauseErr enables manual transaction failure for testing
	CauseErr error
}

// DeleteTestScenarioFunc runs a single test scenario
func DeleteTestScenarioFunc[T any](t testing.TB, ctx context.Context, scenario *DeleteScenario[T]) {
	t.Helper()
	if scenario.SetupFunc != nil {
		scenario.SetupFunc(t, ctx, scenario)
	}
	dbx := scenario.Dbx
	repo := scenario.Repo
	args := scenario.Args
	if scenario.ArgsFunc != nil {
		args = scenario.ArgsFunc(t, ctx, scenario)
	}
	var txRes, err = repo.Delete(ctx, dbx, args)
	if err != nil {
		t.Fatal(err)
	}
	scenario.TestFunc(t, ctx, scenario, txRes)
}
func TestRepositoryDelete_User(t *testing.T) {
	// t.Parallel()
	scenarios := []*DeleteScenario[models.User]{
		{
			Name: "10 users, delete all by without where",
			ArgsFunc: func(t testing.TB, ctx context.Context, scenario *DeleteScenario[models.User]) *map[string]any {
				var args []models.User
				for i := range 10 {
					args = append(args, models.User{
						Name:  types.Pointer("Name:" + fmt.Sprint(i)),
						Email: fmt.Sprint(i) + "@email.com",
					})
				}
				users := MustCreateManyCtx(t, ctx, User, scenario.Dbx, args)
				assert.Len(t, users, 10)
				return nil
			},
			Repo:      User,
			SetupFunc: func(t testing.TB, ctx context.Context, scenario *DeleteScenario[models.User]) {},
			TestFunc: func(t testing.TB, ctx context.Context, scenario *DeleteScenario[models.User], res int64) {
				t.Helper()
				count := MustCountAllCtx(t, ctx, User, scenario.Dbx, &map[string]any{})
				assert.Equal(t, int64(10), res)
				assert.Equal(t, int64(0), count)
			},
		},
		{
			Name: "10 users, randomly set email_verified_at, delete all verified",
			ArgsFunc: func(t testing.TB, ctx context.Context, scenario *DeleteScenario[models.User]) *map[string]any {
				var args []models.User
				selector := test.NewRandomeSelector(nil, types.Pointer(time.Now()))
				for i := range 10 {
					args = append(args, models.User{
						Name:            types.Pointer("Name:" + fmt.Sprint(i)),
						Email:           fmt.Sprint(i) + "@email.com",
						EmailVerifiedAt: selector.Select(),
					})
				}
				_ = MustCreateManyCtx(t, ctx, User, scenario.Dbx, args)
				return &map[string]any{
					"email_verified_at": map[string]any{
						"_isnotnull": nil,
					},
				}
			},
			Repo:      User,
			SetupFunc: func(t testing.TB, ctx context.Context, scenario *DeleteScenario[models.User]) {},
			TestFunc: func(t testing.TB, ctx context.Context, scenario *DeleteScenario[models.User], res int64) {
				t.Helper()
				// find all remaining users
				remaining := MustFindWithOptionsCtx(t, ctx, User, scenario.Dbx)
				for _, user := range remaining {
					// they should not be verified.
					assert.Nil(t, user.EmailVerifiedAt)
				}

				assert.Equal(t, int64(10), res+int64(len(remaining)))
			},
		},
	}
	for _, scenario := range scenarios {
		t.Run(scenario.Name, func(t *testing.T) {
			database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
				scenario.Dbx = db
				DeleteTestScenarioFunc(t, ctx, scenario)
			})
		})
	}
}
func TestRepositoryDelete_UserAccount(t *testing.T) {
	// t.Parallel()
	scenarios := []*DeleteScenario[models.UserAccount]{
		{
			Name: "10 user accounts, delete one by userId and provider",
			ArgsFunc: func(t testing.TB, ctx context.Context, scenario *DeleteScenario[models.UserAccount]) *map[string]any {
				dbx := scenario.Dbx
				var userArgs []models.User
				for i := range 10 {
					userArgs = append(userArgs, models.User{
						Name:  types.Pointer("Name:" + fmt.Sprint(i)),
						Email: fmt.Sprint(i) + "@email.com",
					})
				}
				users := MustCreateManyCtx(t, ctx, User, dbx, userArgs)
				var userAccountArgs []models.UserAccount
				for i := range 10 {
					user := users[i]
					userAccountArgs = append(userAccountArgs, models.UserAccount{
						UserID:            user.ID,
						Provider:          models.ProvidersCredentials,
						ProviderAccountID: user.Email,
						Type:              models.ProviderTypeCredentials,
					})
				}
				accounts := MustCreateManyCtx(t, ctx, UserAccount, dbx, userAccountArgs)
				selected := test.NewRandomeSelector(accounts...).Select()
				return &map[string]any{
					"user_id": map[string]any{
						"_eq": selected.UserID,
					},
					"provider": map[string]any{
						"_eq": selected.Provider,
					},
				}
			},
			Repo:      UserAccount,
			SetupFunc: func(t testing.TB, ctx context.Context, scenario *DeleteScenario[models.UserAccount]) {},
			TestFunc: func(t testing.TB, ctx context.Context, scenario *DeleteScenario[models.UserAccount], res int64) {
				t.Helper()
				// find user without accounts
				userCount := MustCountAllCtx(t, ctx, User, scenario.Dbx, &map[string]any{})
				userAccountCount := MustCountAllCtx(t, ctx, UserAccount, scenario.Dbx, &map[string]any{})
				assert.Equal(t, int64(1), res)
				assert.Equal(t, int64(10), userCount)
				assert.Equal(t, int64(9), userAccountCount)
			},
		},
		{
			Name: "10 user accounts, delete all by without where, users remain",
			ArgsFunc: func(t testing.TB, ctx context.Context, scenario *DeleteScenario[models.UserAccount]) *map[string]any {
				dbx := scenario.Dbx
				var userArgs []models.User
				for i := range 10 {

					userArgs = append(userArgs, models.User{
						Name:  types.Pointer("Name:" + fmt.Sprint(i)),
						Email: fmt.Sprint(i) + "@email.com",
					})
				}
				users := MustCreateManyCtx(t, ctx, User, dbx, userArgs)
				var userAccountArgs []models.UserAccount
				for i := range 10 {
					user := users[i]
					userAccountArgs = append(userAccountArgs, models.UserAccount{
						UserID:            user.ID,
						Provider:          models.ProvidersCredentials,
						ProviderAccountID: user.Email,
						Type:              models.ProviderTypeCredentials,
					})
				}
				_ = MustCreateManyCtx(t, ctx, UserAccount, dbx, userAccountArgs)
				return nil
			},
			Repo:      UserAccount,
			SetupFunc: func(t testing.TB, ctx context.Context, scenario *DeleteScenario[models.UserAccount]) {},
			TestFunc: func(t testing.TB, ctx context.Context, scenario *DeleteScenario[models.UserAccount], res int64) {
				t.Helper()
				userCount := MustCountAllCtx(t, ctx, User, scenario.Dbx, &map[string]any{})
				userAccountCount := MustCountAllCtx(t, ctx, UserAccount, scenario.Dbx, &map[string]any{})
				assert.Equal(t, int64(10), res)
				assert.Equal(t, int64(10), userCount)
				assert.Equal(t, int64(0), userAccountCount)
			},
		},
	}
	for _, scenario := range scenarios {
		t.Run(scenario.Name, func(t *testing.T) {
			database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
				scenario.Dbx = db
				DeleteTestScenarioFunc(t, ctx, scenario)
			})
		})
	}
}

func TestRepositoryDelete_Team(t *testing.T) {
	// t.Parallel()
	scenarios := []*DeleteScenario[models.Team]{
		{
			Name: "10 teams, delete all by without where",
			ArgsFunc: func(t testing.TB, ctx context.Context, scenario *DeleteScenario[models.Team]) *map[string]any {
				var teamArgs []models.Team
				for i := range 10 {
					teamArgs = append(teamArgs, models.Team{
						Name: "name:" + fmt.Sprint(i),
						Slug: "slug:" + fmt.Sprint(i),
					})
				}
				_ = MustCreateManyCtx(t, ctx, scenario.Repo, scenario.Dbx, teamArgs)
				return nil
			},
			Repo:      Team,
			SetupFunc: func(t testing.TB, ctx context.Context, scenario *DeleteScenario[models.Team]) {},
			TestFunc: func(t testing.TB, ctx context.Context, scenario *DeleteScenario[models.Team], res int64) {
				t.Helper()
				count := MustCountAllCtx(t, ctx, scenario.Repo, scenario.Dbx, &map[string]any{})
				assert.Equal(t, int64(10), res)
				assert.Equal(t, int64(0), count)
			},
		},
		{
			Name: "10 teams, delete 1 by id",
			ArgsFunc: func(t testing.TB, ctx context.Context, scenario *DeleteScenario[models.Team]) *map[string]any {
				var teamArgs []models.Team
				for i := range 10 {
					teamArgs = append(teamArgs, models.Team{
						Name: "name:" + fmt.Sprint(i),
						Slug: "slug:" + fmt.Sprint(i),
					})
				}
				teams := MustCreateManyCtx(t, ctx, scenario.Repo, scenario.Dbx, teamArgs)
				selected := test.NewRandomeSelector(teams...).Select()
				return &map[string]any{
					"id": map[string]any{
						"_eq": selected.ID,
					},
				}
			},
			Repo:      Team,
			SetupFunc: func(t testing.TB, ctx context.Context, scenario *DeleteScenario[models.Team]) {},
			TestFunc: func(t testing.TB, ctx context.Context, scenario *DeleteScenario[models.Team], res int64) {
				t.Helper()
				count := MustCountAllCtx(t, ctx, scenario.Repo, scenario.Dbx, &map[string]any{})
				assert.Equal(t, int64(1), res)
				assert.Equal(t, int64(9), count)
			},
		},
	}
	for _, scenario := range scenarios {
		t.Run(scenario.Name, func(t *testing.T) {
			database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
				scenario.Dbx = db
				DeleteTestScenarioFunc(t, ctx, scenario)
			})
		})
	}
}
func TestRepositoryDelete_TeamMember(t *testing.T) {
	// t.Parallel()
	scenarios := []*DeleteScenario[models.TeamMember]{
		{
			Name: "10 team members, delete 1 by user_id",
			ArgsFunc: func(t testing.TB, ctx context.Context, scenario *DeleteScenario[models.TeamMember]) *map[string]any {
				dbx := scenario.Dbx
				var userArgs []models.User
				for i := range 10 {

					userArgs = append(userArgs, models.User{
						Name:  types.Pointer("Name:" + fmt.Sprint(i)),
						Email: fmt.Sprint(i) + "@email.com",
					})
				}
				users := MustCreateManyCtx(t, ctx, User, dbx, userArgs)
				var teamArgs []models.Team
				for i := range 10 {

					teamArgs = append(teamArgs, models.Team{
						Name: "name:" + fmt.Sprint(i),
						Slug: "slug:" + fmt.Sprint(i),
					})
				}
				teams := MustCreateManyCtx(t, ctx, Team, dbx, teamArgs)
				var teamMemberArgs []models.TeamMember

				for i := range 10 {

					teamMemberArgs = append(teamMemberArgs, models.TeamMember{
						TeamID: teams[i].ID,
						UserID: &users[i].ID,
						Active: true,
						Role:   models.TeamMemberRoleMember,
					})
				}

				members := MustCreateManyCtx(t, ctx, TeamMember, dbx, teamMemberArgs)

				selected := test.NewRandomeSelector(members...).Select()
				return &map[string]any{
					"user_id": map[string]any{
						"_eq": selected.UserID,
					},
				}
			},
			Repo:      TeamMember,
			SetupFunc: func(t testing.TB, ctx context.Context, scenario *DeleteScenario[models.TeamMember]) {},
			TestFunc: func(t testing.TB, ctx context.Context, scenario *DeleteScenario[models.TeamMember], res int64) {
				t.Helper()
				userCount := MustCountAllCtx(t, ctx, User, scenario.Dbx, &map[string]any{})
				teamCount := MustCountAllCtx(t, ctx, Team, scenario.Dbx, &map[string]any{})
				teamMemberCount := MustCountAllCtx(t, ctx, TeamMember, scenario.Dbx, &map[string]any{})
				assert.Equal(t, int64(1), res)
				assert.Equal(t, int64(9), teamMemberCount)
				assert.Equal(t, int64(10), teamCount)
				assert.Equal(t, int64(10), userCount)
			},
		},
		{
			Name: "10 team members, delete all by without where, users and teams remain",
			ArgsFunc: func(t testing.TB, ctx context.Context, scenario *DeleteScenario[models.TeamMember]) *map[string]any {
				dbx := scenario.Dbx
				var userArgs []models.User
				for i := range 10 {

					userArgs = append(userArgs, models.User{
						Name:  types.Pointer("Name:" + fmt.Sprint(i)),
						Email: fmt.Sprint(i) + "@email.com",
					})
				}
				users := MustCreateManyCtx(t, ctx, User, dbx, userArgs)
				var teamArgs []models.Team
				for i := range 10 {

					teamArgs = append(teamArgs, models.Team{
						Name: "name:" + fmt.Sprint(i),
						Slug: "slug:" + fmt.Sprint(i),
					})
				}
				teams := MustCreateManyCtx(t, ctx, Team, dbx, teamArgs)
				var teamMemberArgs []models.TeamMember

				for i := range 10 {

					teamMemberArgs = append(teamMemberArgs, models.TeamMember{
						TeamID: teams[i].ID,
						UserID: &users[i].ID,
						Active: true,
						Role:   models.TeamMemberRoleMember,
					})
				}

				_ = MustCreateManyCtx(t, ctx, TeamMember, dbx, teamMemberArgs)
				return nil
			},
			Repo:      TeamMember,
			SetupFunc: func(t testing.TB, ctx context.Context, scenario *DeleteScenario[models.TeamMember]) {},
			TestFunc: func(t testing.TB, ctx context.Context, scenario *DeleteScenario[models.TeamMember], res int64) {
				t.Helper()
				userCount := MustCountAllCtx(t, ctx, User, scenario.Dbx, &map[string]any{})
				teamCount := MustCountAllCtx(t, ctx, Team, scenario.Dbx, &map[string]any{})
				teamMemberCount := MustCountAllCtx(t, ctx, TeamMember, scenario.Dbx, &map[string]any{})
				assert.Equal(t, int64(10), res)
				assert.Equal(t, int64(0), teamMemberCount)
				assert.Equal(t, int64(10), teamCount)
				assert.Equal(t, int64(10), userCount)
			},
		},
	}
	for _, scenario := range scenarios {
		t.Run(scenario.Name, func(t *testing.T) {
			database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
				scenario.Dbx = db
				DeleteTestScenarioFunc(t, ctx, scenario)
			})
		})
	}
}
func TestRepositoryDelete_TeamInvitation(t *testing.T) {
	// t.Parallel()
	scenarios := []*DeleteScenario[models.TeamInvitation]{
		{
			Name: "creating 10 team invitations from 1 team, 1 user and 1 owner. delete 1 by token",
			ArgsFunc: func(t testing.TB, ctx context.Context, scenario *DeleteScenario[models.TeamInvitation]) *map[string]any {
				dbx := scenario.Dbx
				teams := MustCreateOneCtx(t, ctx, Team, dbx, &models.Team{
					Name: "name:" + fmt.Sprint(1),
					Slug: "slug:" + fmt.Sprint(1),
				})
				user := MustCreateOneCtx(t, ctx, User, dbx, &models.User{
					Name:  types.Pointer("Name:" + fmt.Sprint(1)),
					Email: fmt.Sprint(1) + "@email.com",
				})
				owner := MustCreateOneCtx(t, ctx, TeamMember, dbx, &models.TeamMember{
					UserID: &user.ID,
					TeamID: teams.ID,
					Active: true,
					Role:   models.TeamMemberRoleOwner,
				})

				var invitationArgs []models.TeamInvitation

				for i := range 10 {
					invitationArgs = append(invitationArgs, models.TeamInvitation{
						TeamID:          teams.ID,
						InviterMemberID: owner.ID,
						Email:           "inviteuser" + fmt.Sprint(i) + "@email.com",
						Role:            models.TeamMemberRoleMember,
						Token:           uuid.NewString(),
						Status:          models.TeamInvitationStatusPending,
						ExpiresAt:       time.Now().Add(time.Hour * 7),
					})
				}

				invitations := MustCreateManyCtx(t, ctx, TeamInvitation, dbx, invitationArgs)
				selected := test.NewRandomeSelector(invitations...).Select()
				return &map[string]any{
					"token": map[string]any{
						"_eq": selected.Token,
					},
				}
			},
			Repo:      TeamInvitation,
			SetupFunc: func(t testing.TB, ctx context.Context, scenario *DeleteScenario[models.TeamInvitation]) {},
			TestFunc: func(t testing.TB, ctx context.Context, scenario *DeleteScenario[models.TeamInvitation], res int64) {
				t.Helper()
				invitationCOunt := MustCountAllCtx(t, ctx, TeamInvitation, scenario.Dbx, &map[string]any{})
				teamCount := MustCountAllCtx(t, ctx, Team, scenario.Dbx, &map[string]any{})
				teamMemberCount := MustCountAllCtx(t, ctx, TeamMember, scenario.Dbx, &map[string]any{})
				userCount := MustCountAllCtx(t, ctx, User, scenario.Dbx, &map[string]any{})
				assert.Equal(t, int64(1), res)
				assert.Equal(t, int64(9), invitationCOunt)
				assert.Equal(t, int64(1), teamCount)
				assert.Equal(t, int64(1), teamMemberCount)
				assert.Equal(t, int64(1), userCount)
			},
		},
		{
			Name: "creating 10 team invitations from 1 team, 1 user and 1 owner. delete all by without where",
			ArgsFunc: func(t testing.TB, ctx context.Context, scenario *DeleteScenario[models.TeamInvitation]) *map[string]any {
				dbx := scenario.Dbx
				teams := MustCreateOneCtx(t, ctx, Team, dbx, &models.Team{
					Name: "name:" + fmt.Sprint(1),
					Slug: "slug:" + fmt.Sprint(1),
				})
				user := MustCreateOneCtx(t, ctx, User, dbx, &models.User{
					Name:  types.Pointer("Name:" + fmt.Sprint(1)),
					Email: fmt.Sprint(1) + "@email.com",
				})
				owner := MustCreateOneCtx(t, ctx, TeamMember, dbx, &models.TeamMember{
					UserID: &user.ID,
					TeamID: teams.ID,
					Active: true,
					Role:   models.TeamMemberRoleOwner,
				})

				var invitationArgs []models.TeamInvitation

				for i := range 10 {
					invitationArgs = append(invitationArgs, models.TeamInvitation{
						TeamID:          teams.ID,
						InviterMemberID: owner.ID,
						Email:           "inviteuser" + fmt.Sprint(i) + "@email.com",
						Role:            models.TeamMemberRoleMember,
						Token:           uuid.NewString(),
						Status:          models.TeamInvitationStatusPending,
						ExpiresAt:       time.Now().Add(time.Hour * 7),
					})
				}

				_ = MustCreateManyCtx(t, ctx, TeamInvitation, dbx, invitationArgs)
				return nil
			},
			Repo:      TeamInvitation,
			SetupFunc: func(t testing.TB, ctx context.Context, scenario *DeleteScenario[models.TeamInvitation]) {},
			TestFunc: func(t testing.TB, ctx context.Context, scenario *DeleteScenario[models.TeamInvitation], res int64) {
				t.Helper()
				invitationCOunt := MustCountAllCtx(t, ctx, TeamInvitation, scenario.Dbx, &map[string]any{})
				teamCount := MustCountAllCtx(t, ctx, Team, scenario.Dbx, &map[string]any{})
				teamMemberCount := MustCountAllCtx(t, ctx, TeamMember, scenario.Dbx, &map[string]any{})
				userCount := MustCountAllCtx(t, ctx, User, scenario.Dbx, &map[string]any{})
				assert.Equal(t, int64(10), res)
				assert.Equal(t, int64(0), invitationCOunt)
				assert.Equal(t, int64(1), teamCount)
				assert.Equal(t, int64(1), teamMemberCount)
				assert.Equal(t, int64(1), userCount)
			},
		},
	}
	for _, scenario := range scenarios {
		t.Run(scenario.Name, func(t *testing.T) {
			database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
				scenario.Dbx = db
				DeleteTestScenarioFunc(t, ctx, scenario)
			})
		})
	}
}
