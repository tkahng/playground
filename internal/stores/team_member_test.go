package stores_test

import (
	"context"
	"errors"
	"log/slog"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/repository"
	"github.com/tkahng/playground/internal/stores"
	"github.com/tkahng/playground/internal/test"
)

func TestTeamStore_UpdateTeamMember(t *testing.T) {
	test.DbSetup()
	test.WithTx(t, func(ctx context.Context, db database.Dbx) {
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
	test.Parallel(t)
	test.SkipIfShort(t)
	test.WithTx(t, func(ctx context.Context, db database.Dbx) {
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
	test.Parallel(t)
	test.SkipIfShort(t)
	ctx, dbx := test.DbSetup()
	_ = dbx.RunInTx(func(dbxx database.Dbx) error {
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
		member, err := adapter.TeamMember().CreateTeamMember(ctx, team.ID, userID, models.TeamMemberRoleMember, true)
		if err != nil {
			t.Fatalf("CreateTeamMember() error = %v", err)
		}
		if member.TeamID != team.ID || member.UserID == nil || *member.UserID != userID {
			t.Errorf("CreateTeamMember() = %v, want teamID %v and userID %v", member, team.ID, userID)
		}
		return errors.New("rollback")
	})
}

func TestFindTeamMembersByUserID(t *testing.T) {
	test.Parallel(t)
	test.SkipIfShort(t)
	ctx, dbx := test.DbSetup()
	_ = dbx.RunInTx(func(dbxx database.Dbx) error {
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
		_, err = teamStore.CreateTeamMember(ctx, team.ID, userID, models.TeamMemberRoleMember, true)
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
		return errors.New("rollback")
	})
}

func TestFindLatestTeamMemberByUserID(t *testing.T) {
	test.Parallel(t)
	test.SkipIfShort(t)
	ctx, dbx := test.DbSetup()
	_ = dbx.RunInTx(func(dbxx database.Dbx) error {
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
		teamMember1, err := adapter.TeamMember().CreateTeamMember(ctx, team1.ID, userID, models.TeamMemberRoleMember, true)
		if err != nil {
			t.Fatalf("CreateTeamMember() error = %v", err)
		}
		teamMember2, err := adapter.TeamMember().CreateTeamMember(ctx, team2.ID, userID, models.TeamMemberRoleMember, true)
		if err != nil {
			t.Fatalf("CreateTeamMember() error = %v", err)
		}
		time.Sleep(time.Millisecond * 10)
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
		time.Sleep(time.Millisecond * 10)
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
		return errors.New("rollback")
	})
}

func TestUpdateTeamMemberUpdatedAt(t *testing.T) {
	test.Parallel(t)
	test.SkipIfShort(t)
	ctx, dbx := test.DbSetup()
	t.Cleanup(func() {
		_, err := repository.TeamMember.Delete(ctx, dbx, nil)
		if err != nil {
			slog.ErrorContext(ctx, "Error deleting team members", slog.Any("error", err))
		}
		_, err = repository.Team.Delete(ctx, dbx, nil)
		if err != nil {
			slog.ErrorContext(ctx, "Error deleting teams", slog.Any("error", err))
		}
		_, err = repository.User.Delete(ctx, dbx, nil)
		if err != nil {
			slog.ErrorContext(ctx, "Error deleting users", slog.Any("error", err))
		}
	})

	_ = dbx.RunInTx(func(dbxx database.Dbx) error {
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
		member, err := adapter.TeamMember().CreateTeamMember(ctx, team.ID, user.ID, models.TeamMemberRoleMember, true)
		if err != nil {
			t.Fatalf("CreateTeamMember() error = %v", err)
		}
		// Capture the original updated_at
		original := member.CreatedAt

		// Sleep to ensure updated_at will be different
		time.Sleep(time.Second * 1)

		err = adapter.TeamMember().UpdateTeamMemberSelectedAt(ctx, team.ID, user.ID)
		if err != nil {
			t.Fatalf("UpdateTeamMemberUpdatedAt() error = %v", err)
		}
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
		return errors.New("rollback")
	})
}
func TestUpdateTeamMemberSelectedAt(t *testing.T) {
	test.Parallel(t)
	test.SkipIfShort(t)
	ctx, dbx := test.DbSetup()
	_ = dbx.RunInTx(func(dbxx database.Dbx) error {
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
		member, err := teamStore.CreateTeamMember(ctx, team.ID, user.ID, models.TeamMemberRoleMember, true)
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
		return errors.New("rollback")
	})
}

func TestDbTeamMemberStore_LoadTeamMembersByUserAndTeamIds(t *testing.T) {
	test.DbSetup()
	test.WithTx(t, func(ctx context.Context, db database.Dbx) {
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
			_, err = adapter.TeamMember().CreateTeamMember(ctx, team.ID, user1.ID, models.TeamMemberRoleMember, true)
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

func TestDbTeamMemberStore_FindTeamMembers(t *testing.T) {
	test.Parallel(t)
	test.SkipIfShort(t)
	test.WithTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewStorageAdapter(db)
		user := CreateUser(adapter, ctx, "alpha@example.com")
		user2 := CreateUser(adapter, ctx, "beta@example.com")
		team := CreateTeam(adapter, ctx, "Team1")
		teamMember := CreateTeamMember(adapter, ctx, team, user, models.TeamMemberRoleMember, true)
		teamMember.User = user
		teamMember2 := CreateTeamMember(adapter, ctx, team, user2, models.TeamMemberRoleMember, true)
		teamMember2.User = user2
		type fields struct {
			db database.Dbx
		}
		type args struct {
			ctx    context.Context
			filter *stores.TeamMemberFilter
		}
		tests := []struct {
			name    string
			fields  fields
			args    args
			want    []*models.TeamMember
			wantErr bool
		}{
			{
				name: "find team members alpha",
				fields: fields{
					db: db,
				},
				args: args{
					ctx:    ctx,
					filter: &stores.TeamMemberFilter{Q: "alpha", TeamIds: []uuid.UUID{team.ID}},
				},
				want:    []*models.TeamMember{teamMember},
				wantErr: false,
			},
			{
				name: "find team members beta",
				fields: fields{
					db: db,
				},
				args: args{
					ctx:    ctx,
					filter: &stores.TeamMemberFilter{Q: "beta", TeamIds: []uuid.UUID{team.ID}, PaginatedInput: stores.PaginatedInput{Page: 0, PerPage: 10}},
				},
				want:    []*models.TeamMember{teamMember2},
				wantErr: false,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {

				got, err := adapter.TeamMember().FindTeamMembers(tt.args.ctx, tt.args.filter)
				if (err != nil) != tt.wantErr {
					t.Errorf("DbTeamMemberStore.FindTeamMembers() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if len(got) > 0 {
					userIds := make([]uuid.UUID, len(got))
					for idx, member := range got {
						if member == nil {
							continue
						}
						if member.UserID == nil {
							continue
						}
						userIds[idx] = *member.UserID
					}
					users, err := adapter.User().LoadUsersByUserIds(ctx, userIds...)
					if err != nil {
						t.Fatalf("LoadUsersByUserIds() error = %v", err)
					}
					for idx := range userIds {
						member := got[idx]
						if member == nil {
							continue
						}
						user := users[idx]
						if user == nil {
							continue
						}
						member.User = user
					}

				}
				if len(got) != len(tt.want) {
					t.Errorf("DbTeamMemberStore.FindTeamMembers() = %v, want %v", len(got), len(tt.want))
				}
				for idx, item := range got {
					if item.User.Email != tt.want[idx].User.Email {
						t.Errorf("DbTeamMemberStore.FindTeamMembers() = %v, want %v", item.ID, tt.want[idx].ID)
					}
				}
			})
		}
	})
}
