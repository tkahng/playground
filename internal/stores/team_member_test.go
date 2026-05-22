package stores_test

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/populator"
	"github.com/tkahng/playground/internal/stores"
	"github.com/tkahng/playground/internal/test"
)

func TestTeamStore_UpdateTeamMember(t *testing.T) {
	t.Parallel()
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewStorageAdapter(db)
		user, err := adapter.User().CreateUser(ctx, &models.User{
			Email: "testuser@example.com",
		})
		if err != nil {
			t.Fatalf("CreateUser() error = %v", err)
		}
		if user == nil {
			t.Fatal("CreateUser() returned nil user")
		}
		team1, err := adapter.TeamMember().CreateTeamMemberFromUserAndSlug(ctx, user, "TestTeam", models.TeamMemberRoleOwner)
		if err != nil {
			t.Fatalf("CreateTeamMemberFromUserAndSlug() error = %v", err)
		}
		if team1 == nil {
			t.Fatal("CreateTeamMemberFromUserAndSlug() returned nil team member")
		}
		type args struct {
			ctx    context.Context
			member *models.TeamMember
		}
		tests := []struct {
			name    string
			args    args
			want    *models.TeamMember
			wantErr bool
		}{
			{
				name: "update team member",
				args: args{
					ctx: ctx,
					member: &models.TeamMember{
						ID:               team1.ID,
						TeamID:           team1.TeamID,
						UserID:           team1.UserID,
						Role:             models.TeamMemberRoleMember,
						Active:           true,
						HasBillingAccess: true,
						LastSelectedAt:   team1.LastSelectedAt,
					},
				},
				want: &models.TeamMember{
					ID:               team1.ID,
					TeamID:           team1.TeamID,
					UserID:           team1.UserID,
					Role:             models.TeamMemberRoleMember,
					Active:           true,
					HasBillingAccess: true,
					LastSelectedAt:   team1.LastSelectedAt,
				},
				wantErr: false,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := adapter.TeamMember().UpdateTeamMember(tt.args.ctx, tt.args.member)
				if (err != nil) != tt.wantErr {
					t.Errorf("PostgresTeamStore.UpdateTeamMember() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got.Role, tt.want.Role) {
					t.Errorf("PostgresTeamStore.UpdateTeamMember() = %v, want %v", got.Role, tt.want.Role)
				}
			})
		}
	})
}

func TestTeamStore_CountTeamMembers(t *testing.T) {
	t.Parallel()
	test.SkipIfShort(t)
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		type fields struct {
			db database.Dbx
		}
		type args struct {
			ctx    context.Context
			teamId uuid.UUID
		}
		tests := []struct {
			name    string
			fields  fields
			args    args
			want    int64
			wantErr bool
		}{
			// TODO: Add test cases.
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				s := stores.NewStorageAdapter(tt.fields.db)
				got, err := s.TeamMember().CountTeamMembers(tt.args.ctx, &stores.TeamMemberFilter{
					TeamIds: []uuid.UUID{tt.args.teamId},
				})
				if (err != nil) != tt.wantErr {
					t.Errorf("PostgresTeamStore.CountTeamMembers() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if got != tt.want {
					t.Errorf("PostgresTeamStore.CountTeamMembers() = %v, want %v", got, tt.want)
				}
			})
		}
	})
}

func TestCreateTeamMember(t *testing.T) {
	t.Parallel()
	test.SkipIfShort(t)
	database.WithNewTestTx(t, func(ctx context.Context, dbxx database.Dbx) {
		adapter := stores.NewStorageAdapter(dbxx)
		teamStore := adapter.TeamGroup()
		userStore := adapter.User()
		team, err := teamStore.CreateTeam(ctx, "TeamWithMember", "team-with-member-slug")
		if err != nil {
			t.Fatalf("CreateTeam() error = %v", err)
		}
		user, err := userStore.CreateUser(ctx, &models.User{
			Email: "testuser@example.com",
		})
		if err != nil {
			t.Fatalf("CreateUser() error = %v", err)
		}
		userID := user.ID
		member, err := adapter.TeamMember().CreateTeamMember(ctx, &models.TeamMember{
			TeamID:           team.ID,
			UserID:           &userID,
			Role:             models.TeamMemberRoleMember,
			Active:           true,
			HasBillingAccess: true,
		})
		if err != nil {
			t.Fatalf("CreateTeamMember() error = %v", err)
		}
		if member.TeamID != team.ID || member.UserID == nil || *member.UserID != userID {
			t.Errorf("CreateTeamMember() = %v, want teamID %v and userID %v", member, team.ID, userID)
		}
	})
}

func TestFindTeamMembersByUserID(t *testing.T) {
	t.Parallel()
	test.SkipIfShort(t)
	database.WithNewTestTx(t, func(ctx context.Context, dbxx database.Dbx) {
		adapter := stores.NewStorageAdapter(dbxx)
		teamStore := adapter.TeamMember()
		userStore := adapter.User()
		team, err := adapter.TeamGroup().CreateTeam(ctx, "TeamForMembers", "team-for-members-slug")
		if err != nil {
			t.Fatalf("CreateTeam() error = %v", err)
		}
		user, err := userStore.CreateUser(ctx, &models.User{
			Email: "tkahng@gmail.com",
		})
		if err != nil {
			t.Fatalf("CreateUser() error = %v", err)
		}
		userID := user.ID
		_, err = teamStore.CreateTeamMember(ctx, &models.TeamMember{
			TeamID:           team.ID,
			UserID:           &userID,
			Role:             models.TeamMemberRoleMember,
			Active:           true,
			HasBillingAccess: true,
		})
		if err != nil {
			t.Fatalf("CreateTeamMember() error = %v", err)
		}
		members, err := teamStore.FindTeamMembersByUserID(ctx, userID, &stores.TeamMemberListInput{})
		if err != nil {
			t.Fatalf("FindTeamMembersByUserID() error = %v", err)
		}
		if len(members) == 0 || *members[0].UserID != userID {
			t.Errorf("FindTeamMembersByUserID() = %v, want userID %v", members, userID)
		}
	})
}

func TestFindLatestTeamMemberByUserID(t *testing.T) {
	t.Parallel()
	test.SkipIfShort(t)
	database.WithNewTestTx(t, func(ctx context.Context, dbxx database.Dbx) {
		adapter := stores.NewStorageAdapter(dbxx)
		teamStore := adapter.TeamGroup()
		userStore := adapter.User()
		team1, err := teamStore.CreateTeam(ctx, "team1", "team1-slug")
		if err != nil {
			t.Fatalf("CreateTeam() error = %v", err)
		}
		team2, err := teamStore.CreateTeam(ctx, "team2", "team2-slug")
		if err != nil {
			t.Fatalf("CreateTeam() error = %v", err)
		}
		user, err := userStore.CreateUser(ctx, &models.User{
			Email: "testuser@example.com",
		})
		if err != nil {
			t.Fatalf("CreateUser() error = %v", err)
		}
		userID := user.ID
		teamMember1, err := adapter.TeamMember().CreateTeamMember(ctx, &models.TeamMember{
			TeamID:           team1.ID,
			UserID:           &userID,
			Role:             models.TeamMemberRoleMember,
			Active:           true,
			HasBillingAccess: true,
		})
		if err != nil {
			t.Fatalf("CreateTeamMember() error = %v", err)
		}
		teamMember2, err := adapter.TeamMember().CreateTeamMember(ctx, &models.TeamMember{
			TeamID:           team2.ID,
			UserID:           &userID,
			Role:             models.TeamMemberRoleMember,
			Active:           true,
			HasBillingAccess: true,
		})
		if err != nil {
			t.Fatalf("CreateTeamMember() error = %v", err)
		}
		time.Sleep(time.Millisecond)
		err = adapter.TeamMember().UpdateTeamMemberSelectedAt(ctx, teamMember1.TeamID, userID)
		if err != nil {
			t.Fatalf("UpdateTeamMemberUpdatedAt() error = %v", err)
		}
		latest, err := adapter.TeamMember().FindLatestTeamMemberByUserID(ctx, userID)
		if err != nil {
			t.Fatalf("FindLatestTeamMemberByUserID() error = %v", err)
		}
		if latest == nil || latest.UserID == nil || *latest.UserID != userID {
			t.Errorf("FindLatestTeamMemberByUserID() = %v, want userID %v", latest, userID)
		}
		if latest.ID != teamMember1.ID {
			t.Errorf("FindLatestTeamMemberByUserID() = %v, want teamMember1 ID %v", latest.ID, teamMember1.ID)
		}
		time.Sleep(time.Millisecond)
		err = adapter.TeamMember().UpdateTeamMemberSelectedAt(ctx, teamMember1.TeamID, userID)
		if err != nil {
			t.Fatalf("UpdateTeamMemberUpdatedAt() error = %v", err)
		}
		latest, err = adapter.TeamMember().FindLatestTeamMemberByUserID(ctx, userID)
		if err != nil {
			t.Fatalf("FindLatestTeamMemberByUserID() error = %v", err)
		}
		if latest == nil || latest.UserID == nil || *latest.UserID != userID {
			t.Errorf("FindLatestTeamMemberByUserID() = %v, want userID %v", latest, userID)
		}
		if latest.ID != teamMember1.ID {
			t.Errorf("FindLatestTeamMemberByUserID() = %v, want teamMember2 ID %v", latest.ID, teamMember2.ID)
		}
	})
}

func TestUpdateTeamMemberUpdatedAt(t *testing.T) {
	t.Parallel()
	test.SkipIfShort(t)
	database.WithNewTestTx(t, func(ctx context.Context, dbxx database.Dbx) {
		adapter := stores.NewStorageAdapter(dbxx)

		team, err := adapter.TeamGroup().CreateTeam(ctx, "UpdateMemberTeam", "update-member-team-slug")
		if err != nil {
			t.Fatalf("CreateTeam() error = %v", err)
		}
		user, err := adapter.User().CreateUser(ctx, &models.User{
			Email: "updatemember@example.com",
		})
		if err != nil {
			t.Fatalf("CreateUser() error = %v", err)
		}
		member, err := adapter.TeamMember().CreateTeamMember(ctx, &models.TeamMember{
			TeamID:           team.ID,
			UserID:           &user.ID,
			Role:             models.TeamMemberRoleMember,
			Active:           true,
			HasBillingAccess: true,
		})
		if err != nil {
			t.Fatalf("CreateTeamMember() error = %v", err)
		}
		// Capture the original updated_at
		original := member.CreatedAt

		err = adapter.TeamMember().UpdateTeamMemberSelectedAt(ctx, team.ID, user.ID)
		if err != nil {
			t.Fatalf("UpdateTeamMemberUpdatedAt() error = %v", err)
		}

		// Fetch the member again to check updated_at
		updated, err := adapter.TeamMember().FindTeamMember(ctx, &stores.TeamMemberFilter{
			TeamIds: []uuid.UUID{team.ID},
			UserIds: []uuid.UUID{user.ID},
		})
		if err != nil {
			t.Fatalf("GetOne() error = %v", err)
		}
		if updated == nil {
			t.Fatalf("Updated member not found")
		}
		if !updated.UpdatedAt.After(original) {
			t.Errorf(
				"UpdateTeamMemberUpdatedAt() did not update updated_at: before=%v after=%v",
				original,
				updated.UpdatedAt,
			)
		}
	})
}
func TestUpdateTeamMemberSelectedAt(t *testing.T) {
	t.Parallel()
	test.SkipIfShort(t)
	database.WithNewTestTx(t, func(ctx context.Context, dbxx database.Dbx) {
		adapter := stores.NewStorageAdapter(dbxx)
		teamStore := adapter.TeamMember()
		userStore := adapter.User()

		// Create team and user
		team, err := adapter.TeamGroup().CreateTeam(ctx, "SelectedAtTeam", "selected-at-team-slug")
		if err != nil {
			t.Fatalf("CreateTeam() error = %v", err)
		}
		user, err := userStore.CreateUser(ctx, &models.User{
			Email: "selectedat@example.com",
		})
		if err != nil {
			t.Fatalf("CreateUser() error = %v", err)
		}
		member, err := teamStore.CreateTeamMember(ctx, &models.TeamMember{
			TeamID:           team.ID,
			UserID:           &user.ID,
			Role:             models.TeamMemberRoleMember,
			Active:           true,
			HasBillingAccess: true,
		})
		if err != nil {
			t.Fatalf("CreateTeamMember() error = %v", err)
		}
		original := member.LastSelectedAt
		// Call UpdateTeamMemberSelectedAt
		err = teamStore.UpdateTeamMemberSelectedAt(ctx, team.ID, user.ID)
		if err != nil {
			t.Fatalf("UpdateTeamMemberSelectedAt() error = %v", err)
		}

		// Fetch the member again and check last_selected_at
		updated, err := teamStore.FindTeamMember(ctx, &stores.TeamMemberFilter{
			TeamIds: []uuid.UUID{team.ID},
			UserIds: []uuid.UUID{user.ID},
		})
		if err != nil {
			t.Fatalf("FindTeamMemberByTeamAndUserId() error = %v", err)
		}
		if updated == nil {
			t.Fatalf("Updated member not found")
		}
		if updated.LastSelectedAt.IsZero() {
			t.Errorf("Expected LastSelectedAt to be set, got zero value")
		}
		// Should be within a reasonable time window (2s)
		if !updated.LastSelectedAt.After(original) {
			t.Errorf("LastSelectedAt not updated recently: %v", updated.LastSelectedAt)
		}
	})
}

func TestDbTeamMemberStore_LoadTeamMembersByUserAndTeamIds(t *testing.T) {
	t.Parallel()
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewStorageAdapter(db)
		user1, err := adapter.User().CreateUser(ctx, &models.User{
			Email: "user1@example.com",
		})
		if err != nil {
			t.Fatalf("CreateUser() error = %v", err)
		}
		var teamInfo = [][]string{
			{"Team1", "team1-slug"},
			{"Team2", "team2-slug"},
			{"Team3", "team3-slug"},
		}
		var teamsMap = make(map[uuid.UUID]*models.Team)
		var teamsSlice []*models.Team
		var teamIds []uuid.UUID
		for _, info := range teamInfo {
			team, err := adapter.TeamGroup().CreateTeam(ctx, info[0], info[1])
			if err != nil {
				t.Fatalf("CreateTeam() error = %v", err)
			}
			teamsMap[team.ID] = team
			teamsSlice = append(teamsSlice, team)
			teamIds = append(teamIds, team.ID)
			_, err = adapter.TeamMember().CreateTeamMember(ctx, &models.TeamMember{
				TeamID:           team.ID,
				UserID:           &user1.ID,
				Role:             models.TeamMemberRoleMember,
				Active:           true,
				HasBillingAccess: true,
			})
			if err != nil {
				t.Fatalf("CreateTeamMember() error = %v", err)
			}
		}
		type fields struct {
			db database.Dbx
		}
		type args struct {
			ctx     context.Context
			userId  uuid.UUID
			teamIds []uuid.UUID
		}
		tests := []struct {
			name    string
			fields  fields
			args    args
			want    map[uuid.UUID]*models.TeamMember
			wantErr bool
		}{
			{
				name: "load team members by team",
				fields: fields{
					db: db,
				},
				args: args{
					userId:  user1.ID,
					teamIds: teamIds,
					ctx:     ctx,
				},

				wantErr: false,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				store := stores.NewDbTeamMemberStore(tt.fields.db)
				got, _ := store.LoadTeamMembersByUserAndTeamIds(tt.args.ctx, tt.args.userId, tt.args.teamIds...)
				for _, teamMember := range got {
					if team, ok := teamsMap[teamMember.TeamID]; ok {
						if teamMember.UserID == nil || *teamMember.UserID != user1.ID {
							t.Errorf("LoadTeamMembersByUserAndTeamIds() = %v, want userID %v for team %v", teamMember.UserID, user1.ID, teamMember.ID)
						}
						if teamMember.TeamID != team.ID {
							t.Errorf("LoadTeamMembersByUserAndTeamIds() = %v, want teamID %v for team %v", teamMember.TeamID, team.ID, teamMember.ID)
						}
					} else {
						t.Errorf("LoadTeamMembersByUserAndTeamIds() did not find member for team %v", teamMember.ID)
					}
				}
			})
		}
	})
}

// user  | team
// user1@gmail.com | team3
// user2@gmail.com | team2
// user3@gmail.com | team1
func TestDbTeamMemberStore_FindTeamMembers(t *testing.T) {
	t.Parallel()
	test.SkipIfShort(t)
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewStorageAdapter(db)
		user1 := stores.CreateUser(adapter, ctx, "user1@example.com")
		user2 := stores.CreateUser(adapter, ctx, "user2@example.com")
		user3 := stores.CreateUser(adapter, ctx, "user3@example.com")
		team1 := stores.CreateTeam(adapter, ctx, "Team1")
		team2 := stores.CreateTeam(adapter, ctx, "Team2")
		team3 := stores.CreateTeam(adapter, ctx, "Team3")
		user1Team3Member := stores.CreateTeamMember(adapter, ctx, team1, user3, models.TeamMemberRoleOwner, true)
		user2Team2Member := stores.CreateTeamMember(adapter, ctx, team2, user2, models.TeamMemberRoleOwner, true)
		user3Team1Member := stores.CreateTeamMember(adapter, ctx, team3, user1, models.TeamMemberRoleOwner, true)

		user1Team3Member.User = user1
		user1Team3Member.Team = team3
		user2Team2Member.User = user2
		user2Team2Member.Team = team2
		user3Team1Member.User = user3
		user3Team1Member.Team = team1

		err := adapter.TeamMember().UpdateTeamMemberSelectedAt(ctx, user2Team2Member.TeamID, *user2Team2Member.UserID)
		assert.NoError(t, err)
		err = adapter.TeamMember().UpdateTeamMemberSelectedAt(ctx, user3Team1Member.TeamID, *user3Team1Member.UserID)
		assert.NoError(t, err)
		err = adapter.TeamMember().UpdateTeamMemberSelectedAt(ctx, user1Team3Member.TeamID, *user1Team3Member.UserID)
		assert.NoError(t, err)

		t.Run("sort by user email asc", func(t *testing.T) {
			members := FindAndPopulateTeamMembers(t, ctx, adapter, &stores.TeamMemberFilter{
				SortParams: stores.SortParams{
					SortBy:    "user.email",
					SortOrder: "ASC",
				},
			})
			test.TestSliceItemsOrderByFunc(t, members, func(a, b *models.TeamMember) bool {
				return a.User.Email < b.User.Email
			})
		})
		t.Run("sort by user email desc", func(t *testing.T) {
			members := FindAndPopulateTeamMembers(t, ctx, adapter, &stores.TeamMemberFilter{
				SortParams: stores.SortParams{
					SortBy:    "user.email",
					SortOrder: "DESC",
				},
			})
			test.TestSliceItemsOrderByFunc(t, members, func(a, b *models.TeamMember) bool {
				return a.User.Email > b.User.Email
			})
		})
		t.Run("sort by team name asc", func(t *testing.T) {
			members := FindAndPopulateTeamMembers(t, ctx, adapter, &stores.TeamMemberFilter{
				SortParams: stores.SortParams{
					SortBy:    "team.name",
					SortOrder: "ASC",
				},
			})
			test.TestSliceItemsOrderByFunc(t, members, func(a, b *models.TeamMember) bool {
				return a.Team.Name < b.Team.Name
			})
		})
		t.Run("sort by team name desc", func(t *testing.T) {
			members := FindAndPopulateTeamMembers(t, ctx, adapter, &stores.TeamMemberFilter{
				SortParams: stores.SortParams{
					SortBy:    "team.name",
					SortOrder: "DESC",
				},
			})
			test.TestSliceItemsOrderByFunc(t, members, func(a, b *models.TeamMember) bool {
				return a.Team.Name > b.Team.Name
			})
		})
		t.Run("sort by team member last selected at asc", func(t *testing.T) {
			members := FindAndPopulateTeamMembers(t, ctx, adapter, &stores.TeamMemberFilter{
				SortParams: stores.SortParams{
					SortBy:    "last_selected_at",
					SortOrder: "ASC",
				},
			})
			test.TestSliceItemsOrderByFunc(t, members, func(a, b *models.TeamMember) bool {
				return b.LastSelectedAt.After(a.LastSelectedAt)
			})
		})
		t.Run("sort by team member last selected at desc", func(t *testing.T) {
			members := FindAndPopulateTeamMembers(t, ctx, adapter, &stores.TeamMemberFilter{
				SortParams: stores.SortParams{
					SortBy:    "last_selected_at",
					SortOrder: "DESC",
				},
			})
			test.TestSliceItemsOrderByFunc(t, members, func(a, b *models.TeamMember) bool {
				return b.LastSelectedAt.Before(a.LastSelectedAt)
			})
		})
		t.Run("sort by team member last selected at desc", func(t *testing.T) {
			members := FindAndPopulateTeamMembers(t, ctx, adapter, &stores.TeamMemberFilter{
				SortParams: stores.SortParams{
					SortBy:    "last_selected_at",
					SortOrder: "DESC",
				},
			})
			test.TestSliceItemsOrderByFunc(t, members, func(a, b *models.TeamMember) bool {
				return b.LastSelectedAt.Before(a.LastSelectedAt)
			})
		})
	})
}

func FindAndPopulateTeamMembers(t *testing.T, ctx context.Context, adapter *stores.StorageAdapter, filter *stores.TeamMemberFilter) []*models.TeamMember {
	pop := populator.New(adapter)
	res, err := adapter.TeamMember().FindTeamMembers(ctx, filter)
	assert.NoError(t, err)
	for _, r := range res {
		err := populator.PopulateTeamMember(ctx, pop, r)
		assert.NoError(t, err)
	}
	return res
}
