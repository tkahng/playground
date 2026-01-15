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
	"github.com/tkahng/playground/internal/tools/logger"
	"github.com/tkahng/playground/internal/tools/types"
)

type PutScenario[T any] struct {
	// Name is the name of the scenario
	Name string
	// Dbx is the database connection
	Dbx database.Dbx
	// Repo is the repository to be tested
	Repo Repository[T]
	// Args is the arguments to be passed to the repository
	Args []T
	// ArgsFunc returns the arguments to be passed to the repository. this is for arguments that need some computation.
	ArgsFunc func(t testing.TB, ctx context.Context, scenario *PutScenario[T]) []T
	// SetupFunc is the function to setup the test
	SetupFunc func(t testing.TB, ctx context.Context, scenario *PutScenario[T])
	// TestFunc is the function to verify the post result
	TestFunc func(t testing.TB, ctx context.Context, args, res *T)
	// ㅉantErr indicates whether the test should expect an error
	WantErr bool
	// CauseErr enables manual transaction failure for testing
	CauseErr error
}

// PutTestScenarioFunc runs a single test scenario
func PutTestScenarioFunc[T any](t testing.TB, ctx context.Context, scenario *PutScenario[T]) {
	t.Helper()
	if scenario.SetupFunc != nil {
		scenario.SetupFunc(t, ctx, scenario)
	}
	dbx := scenario.Dbx
	repo := scenario.Repo
	var args []T
	args = scenario.Args
	if scenario.ArgsFunc != nil {
		args = scenario.ArgsFunc(t, ctx, scenario)
	}
	var txRes, err = repo.Put(ctx, dbx, args)
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
func TestRepositoryPut_User(t *testing.T) {
	// t.Parallel()
	_ = logger.GetDefaultLogger()
	scenarios := []*PutScenario[models.User]{
		{
			Name: "creating 10 unique users from numbers, then updating them",
			ArgsFunc: func(t testing.TB, ctx context.Context, scenario *PutScenario[models.User]) []models.User {
				var args []models.User
				timeOpions := []*time.Time{
					types.Pointer(time.Now()),
					types.Pointer(time.Now().UTC().Add(time.Hour * -24)),
					nil,
				}
				emailVerifiedAtSelector := test.NewRandomeSelector(timeOpions...)
				for i := range 10 {
					args = append(args, models.User{
						Name:  types.Pointer("Name:" + fmt.Sprint(i)),
						Email: fmt.Sprint(i) + "@email.com",
					})
				}
				users := MustCreateManyCtx(t, ctx, User, scenario.Dbx, args)
				var userArgs []models.User
				for i := range 10 {
					user := users[i]
					user.Name = types.Pointer(*user.Name + " updated")
					user.Email = user.Email + " updated"
					user.EmailVerifiedAt = emailVerifiedAtSelector.Select()
					userArgs = append(userArgs, *user)
				}
				return userArgs
			},
			Repo:      User,
			SetupFunc: func(t testing.TB, ctx context.Context, scenario *PutScenario[models.User]) {},
			TestFunc: func(t testing.TB, ctx context.Context, arg, res *models.User) {
				assert.Equal(t, arg.Name, res.Name, "Name should be the same")
				assert.Equal(t, arg.Email, res.Email, "Email should be the same")
				assert.Equal(t, arg.ID, res.ID, "ID should be the same")
				if arg.EmailVerifiedAt != nil && res.EmailVerifiedAt != nil {
					assert.WithinDuration(t, *arg.EmailVerifiedAt, *res.EmailVerifiedAt, time.Second, "EmailVerifiedAt should be equal.")
				} else if arg.EmailVerifiedAt == nil && res.EmailVerifiedAt == nil {
					// do nothing since both are nil and should be the same
				} else {
					assert.Fail(t, "EmailVerifiedAt should be the same")
				}
				assert.True(t, arg.UpdatedAt.Before(res.UpdatedAt), "UpdatedAt should be the same")
				assert.True(t, arg.CreatedAt.Equal(res.CreatedAt), "CreatedAt should be the same")
			},
		},
	}
	for _, scenario := range scenarios {
		t.Run(scenario.Name, func(t *testing.T) {
			database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
				scenario.Dbx = db
				PutTestScenarioFunc(t, ctx, scenario)
			})
		})
	}
}
func TestRepositoryPut_UserAccount(t *testing.T) {
	// t.Parallel()
	scenarios := []*PutScenario[models.UserAccount]{
		{
			Name: "creating 10 unique users and their accounts from numbers, then updating them",
			ArgsFunc: func(t testing.TB, ctx context.Context, scenario *PutScenario[models.UserAccount]) []models.UserAccount {
				providerSelector := test.NewRandomeSelector(
					models.ProvidersCredentials,
					models.ProvidersGoogle,
					models.ProvidersGithub,
				)
				dbx := scenario.Dbx
				var userArgs []models.User
				for i := range 10 {
					userArgs = append(userArgs, models.User{
						Name:  types.Pointer("Name:" + fmt.Sprint(i)),
						Email: fmt.Sprint(i) + "@email.com",
					})
				}
				users := MustCreateManyCtx(t, ctx, User, dbx, userArgs)
				assert.Len(t, users, 10)
				var userAccountArgs []models.UserAccount
				for _, user := range users {
					provider := providerSelector.Select()
					switch provider {
					case models.ProvidersGoogle:
						userAccountArgs = append(userAccountArgs, models.UserAccount{
							UserID:            user.ID,
							Provider:          models.ProvidersGoogle,
							ProviderAccountID: user.Email,
							Type:              models.ProviderTypeOAuth,
							RefreshToken:      types.Pointer(uuid.NewString()),
							AccessToken:       types.Pointer(uuid.NewString()),
							IDToken:           types.Pointer(uuid.NewString()),
							ExpiresAt:         types.Pointer(time.Now().UTC().Add(time.Hour * 24).Unix()),
						})
					case models.ProvidersApple:
						userAccountArgs = append(userAccountArgs, models.UserAccount{
							UserID:            user.ID,
							Provider:          models.ProvidersApple,
							ProviderAccountID: user.Email,
							Type:              models.ProviderTypeOAuth,
							RefreshToken:      types.Pointer(uuid.NewString()),
							AccessToken:       types.Pointer(uuid.NewString()),
							IDToken:           types.Pointer(uuid.NewString()),
							ExpiresAt:         types.Pointer(time.Now().UTC().Add(time.Hour * 24).Unix()),
						})
					case models.ProvidersCredentials:
						userAccountArgs = append(userAccountArgs, models.UserAccount{
							UserID:            user.ID,
							Provider:          models.ProvidersCredentials,
							ProviderAccountID: user.Email,
							Type:              models.ProviderTypeCredentials,
							Password:          types.Pointer(uuid.NewString()),
						})
					default:
						userAccountArgs = append(userAccountArgs, models.UserAccount{
							UserID:            user.ID,
							Provider:          models.ProvidersCredentials,
							ProviderAccountID: user.Email,
							Type:              models.ProviderTypeCredentials,
							Password:          types.Pointer(uuid.NewString()),
						})
					}
				}
				assert.Len(t, userAccountArgs, 10)
				userAccounts := MustCreateManyCtx(t, ctx, UserAccount, dbx, userAccountArgs)
				assert.Len(t, userAccounts, 10)
				var userAccountPutArgs []models.UserAccount
				for i := range 10 {
					userAccount := *userAccounts[i]
					switch userAccount.Type {
					case models.ProviderTypeCredentials:
						userAccount.Password = types.Pointer(uuid.NewString())
						userAccountPutArgs = append(userAccountPutArgs, userAccount)
					case models.ProviderTypeOAuth:
						userAccount.RefreshToken = types.Pointer(uuid.NewString())
						userAccount.AccessToken = types.Pointer(uuid.NewString())
						userAccount.IDToken = types.Pointer(uuid.NewString())
						userAccount.ExpiresAt = types.Pointer(time.Now().UTC().Add(time.Hour * 24).Unix())
					}
				}
				return userAccountPutArgs
			},
			Repo:      UserAccount,
			SetupFunc: func(t testing.TB, ctx context.Context, scenario *PutScenario[models.UserAccount]) {},
			TestFunc: func(t testing.TB, ctx context.Context, arg, res *models.UserAccount) {
				t.Helper()
				// check name. string pointer.
				assert.Equal(t, arg.UserID, res.UserID, "user id should be equal.")
				assert.Equal(t, arg.Provider, res.Provider, "provider should be equal.")
				assert.Equal(t, arg.ProviderAccountID, res.ProviderAccountID, "provider account id should be equal.")
				assert.Equal(t, arg.Type, res.Type, "type should be equal.")
				assert.Equal(t, arg.ID, res.ID, "ID be equal")
				assert.True(t, arg.UpdatedAt.Before(res.UpdatedAt), "UpdatedAt should have increased")
				assert.True(t, arg.CreatedAt.Equal(res.CreatedAt), "CreatedAt should be the same")
			},
		},
	}
	for _, scenario := range scenarios {
		t.Run(scenario.Name, func(t *testing.T) {
			database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
				scenario.Dbx = db
				PutTestScenarioFunc(t, ctx, scenario)
			})
		})
	}
}

func TestRepositoryPut_Team(t *testing.T) {
	// t.Parallel()
	scenarios := []*PutScenario[models.Team]{
		{
			Name: "creating 10 unique teams from numbers",
			ArgsFunc: func(t testing.TB, ctx context.Context, scenario *PutScenario[models.Team]) []models.Team {
				var teamArgs []models.Team
				for i := range 10 {
					teamArgs = append(teamArgs, models.Team{
						Name: "name:" + fmt.Sprint(i),
						Slug: "slug:" + fmt.Sprint(i),
					})
				}
				teams := MustCreateManyCtx(t, ctx, Team, scenario.Dbx, teamArgs)

				var teamPutArgs []models.Team
				for i, team := range teams {
					team.Name = "new name:" + fmt.Sprint(i)
					team.Slug = "new slug:" + fmt.Sprint(i)
					teamPutArgs = append(teamPutArgs, *team)
				}
				return teamPutArgs
			},
			Repo:      Team,
			SetupFunc: func(t testing.TB, ctx context.Context, scenario *PutScenario[models.Team]) {},
			TestFunc: func(t testing.TB, ctx context.Context, arg, res *models.Team) {
				// check name. string pointer.
				assert.Equal(t, arg.Name, res.Name, "name should be equal.")
				assert.Equal(t, arg.Slug, res.Slug, "slug should be equal.")
				assert.Equal(t, arg.ID, res.ID, "ID should be equal.")
			},
		},
	}
	for _, scenario := range scenarios {
		t.Run(scenario.Name, func(t *testing.T) {
			database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
				scenario.Dbx = db
				PutTestScenarioFunc(t, ctx, scenario)
			})
		})
	}
}
func TestRepositoryPut_TeamMember(t *testing.T) {
	// t.Parallel()
	scenarios := []*PutScenario[models.TeamMember]{
		{
			Name: "creating 10 unique team members from numbers, then updating them",
			ArgsFunc: func(t testing.TB, ctx context.Context, scenario *PutScenario[models.TeamMember]) []models.TeamMember {
				dbx := scenario.Dbx
				var userArgs []models.User
				for i := range 10 {
					userArgs = append(userArgs, models.User{
						Name:  types.Pointer("Name:" + fmt.Sprint(i)),
						Email: fmt.Sprint(i) + "@email.com",
					})
				}
				users := MustCreateManyCtx(t, ctx, User, dbx, userArgs)
				if len(users) != 10 {
					t.Fatalf("expected 10 teams, got %d", len(users))
				}
				var teamArgs []models.Team
				for i := range 10 {
					teamArgs = append(teamArgs, models.Team{
						Name: "name:" + fmt.Sprint(i),
						Slug: "slug:" + fmt.Sprint(i),
					})
				}
				teams := MustCreateManyCtx(t, ctx, Team, dbx, teamArgs)
				if len(teams) != 10 {
					t.Fatalf("expected 10 teams, got %d", len(teams))
				}

				var teamMemberArgs []models.TeamMember
				for i := range 10 {
					teamMemberArgs = append(teamMemberArgs, models.TeamMember{
						TeamID: teams[i].ID,
						UserID: &users[i].ID,
						Active: true,
						Role:   models.TeamMemberRoleMember,
					})
				}
				teamMembers := MustCreateManyCtx(t, ctx, TeamMember, dbx, teamMemberArgs)

				updatedTeamMemberArgs := []models.TeamMember{}
				for i := range 10 {
					teamMember := *teamMembers[i]
					teamMember.Role = models.TeamMemberRoleGuest
					teamMember.Active = false
					updatedTeamMemberArgs = append(updatedTeamMemberArgs, teamMember)
				}
				return updatedTeamMemberArgs
			},
			Repo:      TeamMember,
			SetupFunc: func(t testing.TB, ctx context.Context, scenario *PutScenario[models.TeamMember]) {},
			TestFunc: func(t testing.TB, ctx context.Context, arg, res *models.TeamMember) {
				// check name. string pointer.
				assert.Equal(t, arg.UserID, res.UserID, "user id should be equal.")
				assert.Equal(t, arg.TeamID, res.TeamID, "slug should be equal.")
				assert.Equal(t, arg.Role, res.Role, "role should be equal.")
				assert.Equal(t, arg.ID, res.ID, "ID be equal")
			},
		},
	}
	for _, scenario := range scenarios {
		t.Run(scenario.Name, func(t *testing.T) {
			database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
				scenario.Dbx = db
				PutTestScenarioFunc(t, ctx, scenario)
			})
		})
	}
}
func TestRepositoryPut_TeamInvitation(t *testing.T) {
	// t.Parallel()
	_ = logger.GetDefaultLogger()
	scenarios := []*PutScenario[models.TeamInvitation]{
		{
			Name: "creating 10 unique team invitations from 1 team. then updating them",
			ArgsFunc: func(t testing.TB, ctx context.Context, scenario *PutScenario[models.TeamInvitation]) []models.TeamInvitation {
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
						ExpiresAt:       time.Now().UTC().Add(time.Hour * 7),
					})
				}
				invitations := MustCreateManyCtx(t, ctx, TeamInvitation, dbx, invitationArgs)

				updatedInvitationArgs := []models.TeamInvitation{}
				for i := range 10 {
					invitation := *invitations[i]
					invitation.Token = uuid.NewString()
					invitation.ExpiresAt = time.Now().UTC().Add(time.Hour * 7)
					invitation.Role = models.TeamMemberRoleGuest
					invitation.Status = models.TeamInvitationStatusAccepted
					updatedInvitationArgs = append(updatedInvitationArgs, invitation)
				}
				return updatedInvitationArgs
			},
			Repo:      TeamInvitation,
			SetupFunc: func(t testing.TB, ctx context.Context, scenario *PutScenario[models.TeamInvitation]) {},
			TestFunc: func(t testing.TB, ctx context.Context, arg, res *models.TeamInvitation) {
				t.Helper()
				// check name. string pointer.
				assert.Equal(t, arg.InviterMemberID, res.InviterMemberID, "InviterMemberID should be equal.")
				assert.Equal(t, arg.TeamID, res.TeamID, "team should be equal.")
				assert.Equal(t, arg.Role, res.Role, "role should be equal.")
				assert.Equal(t, arg.Status, res.Status, "status should be equal.")
				assert.Equal(t, arg.Email, res.Email, "email should be equal.")
				assert.Equal(t, arg.Token, res.Token, "token should be equal.")
				assert.WithinDuration(t, arg.ExpiresAt, res.ExpiresAt, time.Second, "expiresAt should be equal.")
				assert.Equal(t, arg.ID, res.ID, "ID should be equal.")
			},
		},
	}
	for _, scenario := range scenarios {
		t.Run(scenario.Name, func(t *testing.T) {
			database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
				scenario.Dbx = db
				PutTestScenarioFunc(t, ctx, scenario)
			})
		})
	}
}
