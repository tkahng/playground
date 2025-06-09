package stores_test

import (
	"context"
	"errors"
	"log/slog"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/repository"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/stores"
	"github.com/tkahng/authgo/internal/test"
)

func TestTeamStore_UpdateTeamMember(t *testing.T) {
	type fields struct {
		db database.Dbx
	}
	type args struct {
		ctx    context.Context
		member *models.TeamMember
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *models.TeamMember
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := stores.NewDbTeamStore(tt.fields.db)
			got, err := s.UpdateTeamMember(tt.args.ctx, tt.args.member)
			if (err != nil) != tt.wantErr {
				t.Errorf("PostgresTeamStore.UpdateTeamMember() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PostgresTeamStore.UpdateTeamMember() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTeamStore_CountTeamMembers(t *testing.T) {
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
			s := stores.NewDbTeamStore(tt.fields.db)
			got, err := s.CountTeamMembers(tt.args.ctx, tt.args.teamId)
			if (err != nil) != tt.wantErr {
				t.Errorf("PostgresTeamStore.CountTeamMembers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("PostgresTeamStore.CountTeamMembers() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCreateTeamMember(t *testing.T) {
	test.Short(t)
	ctx, dbx := test.DbSetup()
	_ = dbx.RunInTx(func(dbxx database.Dbx) error {
		teamStore := stores.NewDbTeamStore(dbxx)
		userStore := stores.NewDbUserStore(dbxx)
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
		member, err := teamStore.CreateTeamMember(ctx, team.ID, userID, models.TeamMemberRoleMember, true)
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
	test.Short(t)
	ctx, dbx := test.DbSetup()
	_ = dbx.RunInTx(func(dbxx database.Dbx) error {
		teamStore := stores.NewDbTeamStore(dbxx)
		userStore := stores.NewDbUserStore(dbxx)
		team, err := teamStore.CreateTeam(ctx, "TeamForMembers", "team-for-members-slug")
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
		members, err := teamStore.FindTeamMembersByUserID(ctx, userID, &shared.TeamMemberListInput{})
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
	test.Short(t)
	ctx, dbx := test.DbSetup()
	_ = dbx.RunInTx(func(dbxx database.Dbx) error {
		teamStore := stores.NewDbTeamStore(dbxx)
		userStore := stores.NewDbUserStore(dbxx)
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
		teamMember1, err := teamStore.CreateTeamMember(ctx, team1.ID, userID, models.TeamMemberRoleMember, true)
		if err != nil {
			t.Fatalf("CreateTeamMember() error = %v", err)
		}
		teamMember2, err := teamStore.CreateTeamMember(ctx, team2.ID, userID, models.TeamMemberRoleMember, true)
		if err != nil {
			t.Fatalf("CreateTeamMember() error = %v", err)
		}
		time.Sleep(time.Millisecond * 10)
		err = teamStore.UpdateTeamMemberSelectedAt(ctx, teamMember1.TeamID, userID)
		if err != nil {
			t.Fatalf("UpdateTeamMemberUpdatedAt() error = %v", err)
		}
		latest, err := teamStore.FindLatestTeamMemberByUserID(ctx, userID)
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
		err = teamStore.UpdateTeamMemberSelectedAt(ctx, teamMember1.TeamID, userID)
		if err != nil {
			t.Fatalf("UpdateTeamMemberUpdatedAt() error = %v", err)
		}
		latest, err = teamStore.FindLatestTeamMemberByUserID(ctx, userID)
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
	test.Short(t)
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
		teamStore := stores.NewDbTeamStore(dbxx)
		userStore := stores.NewDbUserStore(dbxx)

		team, err := teamStore.CreateTeam(ctx, "UpdateMemberTeam", "update-member-team-slug")
		if err != nil {
			t.Fatalf("CreateTeam() error = %v", err)
		}
		user, err := userStore.CreateUser(ctx, &models.User{
			Email: "updatemember@example.com",
		})
		if err != nil {
			t.Fatalf("CreateUser() error = %v", err)
		}
		member, err := teamStore.CreateTeamMember(ctx, team.ID, user.ID, models.TeamMemberRoleMember, true)
		if err != nil {
			t.Fatalf("CreateTeamMember() error = %v", err)
		}
		// Capture the original updated_at
		original := member.CreatedAt

		// Sleep to ensure updated_at will be different
		time.Sleep(time.Second * 1)

		err = teamStore.UpdateTeamMemberSelectedAt(ctx, team.ID, user.ID)
		if err != nil {
			t.Fatalf("UpdateTeamMemberUpdatedAt() error = %v", err)
		}
		if err != nil {
			t.Fatalf("UpdateTeamMemberUpdatedAt() error = %v", err)
		}

		// Fetch the member again to check updated_at
		updated, err := teamStore.FindTeamMemberByTeamAndUserId(ctx, team.ID, user.ID)
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
	test.Short(t)
	ctx, dbx := test.DbSetup()
	_ = dbx.RunInTx(func(dbxx database.Dbx) error {
		teamStore := stores.NewDbTeamStore(dbxx)
		userStore := stores.NewDbUserStore(dbxx)

		// Create team and user
		team, err := teamStore.CreateTeam(ctx, "SelectedAtTeam", "selected-at-team-slug")
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
		updated, err := teamStore.FindTeamMemberByTeamAndUserId(ctx, team.ID, user.ID)
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
