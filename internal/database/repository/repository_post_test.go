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
	"github.com/tkahng/playground/internal/tools/types"
)

type PostScenario[T any] struct {
	// Name is the name of the scenario
	Name string
	// Dbx is the database connection
	Dbx database.Dbx
	// Repo is the repository to be tested
	Repo Repository[T]
	// Args is the arguments to be passed to the repository
	Args []T
	// ArgsFunc returns the arguments to be passed to the repository. this is for arguments that need some computation.
	ArgsFunc func(t testing.TB, ctx context.Context, scenario *PostScenario[T]) []T
	// SetupFunc is the function to setup the test
	SetupFunc func(t testing.TB, ctx context.Context, scenario *PostScenario[T])
	// TestFunc is the function to verify the post result
	TestFunc func(t testing.TB, ctx context.Context, args, res *T)
	// ㅉantErr indicates whether the test should expect an error
	WantErr bool
	// CauseErr enables manual transaction failure for testing
	CauseErr error
}

// PostTestScenarioFunc runs a single test scenario
func PostTestScenarioFunc[T any](t testing.TB, ctx context.Context, scenario *PostScenario[T]) {
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
	txRes, err := repo.Post(ctx, dbx, args)
	if err != nil {
		if scenario.WantErr {
			return
		}
		t.Fatal(err)
	}
	for i, arg := range args {
		r := txRes[i]
		scenario.TestFunc(t, ctx, &arg, r)
	}
}
func TestRepositoryPost_TeamInvitation(t *testing.T) {
	// t.Parallel()
	scenarios := []*PostScenario[models.TeamInvitation]{
		{
			Name: "creating 10 unique team invitations from 1 team.",
			ArgsFunc: func(t testing.TB, ctx context.Context, scenario *PostScenario[models.TeamInvitation]) []models.TeamInvitation {
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
				return invitationArgs
			},
			Repo:      TeamInvitation,
			SetupFunc: func(t testing.TB, ctx context.Context, scenario *PostScenario[models.TeamInvitation]) {},
			TestFunc: func(t testing.TB, ctx context.Context, arg, res *models.TeamInvitation) {
				t.Helper()
				// check name. string pointer.
				assert.Equal(t, arg.InviterMemberID, res.InviterMemberID, "InviterMemberID should be equal.")
				assert.Equal(t, arg.TeamID, res.TeamID, "team should be equal.")
				assert.Equal(t, arg.Role, res.Role, "role should be equal.")
				assert.Equal(t, arg.Status, res.Status, "status should be equal.")
				assert.Equal(t, arg.Email, res.Email, "email should be equal.")
				assert.Equal(t, arg.Token, res.Token, "token should be equal.")
				assert.True(t, arg.ExpiresAt.Equal(res.ExpiresAt), "expiresAt should be true.")
				assert.NotEqual(t, arg.ID, res.ID, "ID should not be equal since it is generated on db side, arg will have zero value uuid.UUID")
			},
		},
	}
	for _, scenario := range scenarios {
		t.Run(scenario.Name, func(t *testing.T) {
			database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
				scenario.Dbx = db
				PostTestScenarioFunc(t, ctx, scenario)
			})
		})
	}
}

func TestRepositoryPost_User(t *testing.T) {
	// t.Parallel()
	scenarios := []*PostScenario[models.User]{
		{
			Name: "creating 10 unique users from numbers",
			ArgsFunc: func(t testing.TB, ctx context.Context, scenario *PostScenario[models.User]) []models.User {
				var args []models.User
				for i := range 10 {

					args = append(args, models.User{
						Name:  types.Pointer("Name:" + fmt.Sprint(i)),
						Email: fmt.Sprint(i) + "@email.com",
					})
				}
				return args
			},
			Repo:      User,
			SetupFunc: func(t testing.TB, ctx context.Context, scenario *PostScenario[models.User]) {},
			TestFunc: func(t testing.TB, ctx context.Context, arg, res *models.User) {
				t.Helper()
				assert.Equal(t, arg.Name, res.Name, "Name should be the same")
				assert.Equal(t, arg.Email, res.Email, "Email should be the same")
				assert.NotEqual(t, arg.ID, res.ID, "ID should not be equal since it is generated on db side, arg will have zero value uuid.UUID")
			},
		},
	}
	for _, scenario := range scenarios {
		t.Run(scenario.Name, func(t *testing.T) {
			database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
				scenario.Dbx = db
				PostTestScenarioFunc(t, ctx, scenario)
			})
		})
	}
}
func TestRepositoryPost_UserAccount(t *testing.T) {
	// t.Parallel()
	scenarios := []*PostScenario[models.UserAccount]{
		{
			Name: "creating 10 unique users and their accounts from numbers",
			ArgsFunc: func(t testing.TB, ctx context.Context, scenario *PostScenario[models.UserAccount]) []models.UserAccount {
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
				return userAccountArgs
			},
			Repo:      UserAccount,
			SetupFunc: func(t testing.TB, ctx context.Context, scenario *PostScenario[models.UserAccount]) {},
			TestFunc: func(t testing.TB, ctx context.Context, arg, res *models.UserAccount) {
				t.Helper()
				// check name. string pointer.
				assert.Equal(t, arg.UserID, res.UserID, "user id should be equal.")
				assert.Equal(t, arg.Provider, res.Provider, "provider should be equal.")
				assert.Equal(t, arg.ProviderAccountID, res.ProviderAccountID, "provider account id should be equal.")
				assert.Equal(t, arg.Type, res.Type, "type should be equal.")
				assert.NotEqual(t, arg.ID, res.ID, "ID should not be equal since it is generated on db side, arg will have zero value uuid.UUID")
			},
		},
	}
	for _, scenario := range scenarios {
		t.Run(scenario.Name, func(t *testing.T) {
			database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
				scenario.Dbx = db
				PostTestScenarioFunc(t, ctx, scenario)
			})
		})
	}
}

func TestRepositoryPost_Team(t *testing.T) {
	// t.Parallel()
	scenarios := []*PostScenario[models.Team]{
		{
			Name: "creating 10 unique teams from numbers",
			ArgsFunc: func(t testing.TB, ctx context.Context, scenario *PostScenario[models.Team]) []models.Team {
				var teamArgs []models.Team

				for i := range 10 {

					teamArgs = append(teamArgs, models.Team{
						Name: "name:" + fmt.Sprint(i),
						Slug: "slug:" + fmt.Sprint(i),
					})
				}
				return teamArgs
			},
			Repo:      Team,
			SetupFunc: func(t testing.TB, ctx context.Context, scenario *PostScenario[models.Team]) {},
			TestFunc: func(t testing.TB, ctx context.Context, arg, res *models.Team) {
				t.Helper()
				// check name. string pointer.
				assert.Equal(t, arg.Name, res.Name, "name should be equal.")
				assert.Equal(t, arg.Slug, res.Slug, "slug should be equal.")
				assert.NotEqual(t, arg.ID, res.ID, "ID should not be equal since it is generated on db side, arg will have zero value uuid.UUID")
			},
		},
	}
	for _, scenario := range scenarios {
		t.Run(scenario.Name, func(t *testing.T) {
			database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
				scenario.Dbx = db
				PostTestScenarioFunc(t, ctx, scenario)
			})
		})
	}
}
func TestRepositoryPost_TeamMember(t *testing.T) {
	// t.Parallel()
	scenarios := []*PostScenario[models.TeamMember]{
		{
			Name: "creating 10 unique team members from numbers",
			ArgsFunc: func(t testing.TB, ctx context.Context, scenario *PostScenario[models.TeamMember]) []models.TeamMember {
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
				return teamMemberArgs
			},
			Repo:      TeamMember,
			SetupFunc: func(t testing.TB, ctx context.Context, scenario *PostScenario[models.TeamMember]) {},
			TestFunc: func(t testing.TB, ctx context.Context, arg, res *models.TeamMember) {
				t.Helper()
				// check name. string pointer.
				assert.Equal(t, arg.UserID, res.UserID, "user id should be equal.")
				assert.Equal(t, arg.TeamID, res.TeamID, "slug should be equal.")
				assert.Equal(t, arg.Role, res.Role, "role should be equal.")
				assert.NotEqual(t, arg.ID, res.ID, "ID should not be equal since it is generated on db side, arg will have zero value uuid.UUID")
			},
		},
	}
	for _, scenario := range scenarios {
		t.Run(scenario.Name, func(t *testing.T) {
			database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
				scenario.Dbx = db
				PostTestScenarioFunc(t, ctx, scenario)
			})
		})
	}
}
