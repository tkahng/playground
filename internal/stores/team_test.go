package stores_test

import (
	"context"
	"errors"

	"testing"

	"github.com/google/uuid"

	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/stores"
	"github.com/tkahng/authgo/internal/test"
	"github.com/tkahng/authgo/internal/tools/types"
)

func TestCreateTeam(t *testing.T) {
	test.Short(t)
	ctx, dbx := test.DbSetup()
	_ = dbx.RunInTx(func(dbxx database.Dbx) error {
		adapter := stores.NewStorageAdapter(dbxx)
		teamStore := adapter.TeamGroup()
		team, err := teamStore.CreateTeam(ctx, "Test Team", "test-team-slug")
		if err != nil {
			t.Fatalf("CreateTeam() error = %v", err)
		}
		if team == nil || team.Name != "Test Team" {
			t.Errorf("CreateTeam() = %v, want name 'Test Team'", team)
		}
		return errors.New("rollback")
	})
}

func TestUpdateTeam(t *testing.T) {
	test.Short(t)
	ctx, dbx := test.DbSetup()
	_ = dbx.RunInTx(func(dbxx database.Dbx) error {
		adapter := stores.NewStorageAdapter(dbxx)
		teamStore := adapter.TeamGroup()
		team, err := teamStore.CreateTeam(ctx, "Old Name", "old-name-slug")
		if err != nil {
			t.Fatalf("CreateTeam() error = %v", err)
		}
		newName := "Updated Name"
		updated, err := teamStore.UpdateTeam(ctx, team.ID, newName)
		if err != nil {
			t.Fatalf("UpdateTeam() error = %v", err)
		}
		if updated.Name != newName {
			t.Errorf("UpdateTeam() = %v, want name %v", updated, newName)
		}
		return errors.New("rollback")
	})
}

func TestDeleteTeam(t *testing.T) {
	test.Short(t)
	ctx, dbx := test.DbSetup()
	_ = dbx.RunInTx(func(dbxx database.Dbx) error {
		adapter := stores.NewStorageAdapter(dbxx)
		teamStore := adapter.TeamGroup() // Create a team to delete
		team, err := teamStore.CreateTeam(ctx, "ToDelete", "to-delete-slug")
		if err != nil {
			t.Fatalf("CreateTeam() error = %v", err)
		}
		err = teamStore.DeleteTeam(ctx, team.ID)
		if err != nil {
			t.Errorf("DeleteTeam() error = %v", err)
		}
		return errors.New("rollback")
	})
}

func TestFindTeamByID(t *testing.T) {
	test.Short(t)
	ctx, dbx := test.DbSetup()
	_ = dbx.RunInTx(func(dbxx database.Dbx) error {
		adapter := stores.NewStorageAdapter(dbxx)
		teamStore := adapter.TeamGroup()
		team, err := teamStore.CreateTeam(ctx, "FindMe", "find-me-slug")
		if err != nil {
			t.Fatalf("CreateTeam() error = %v", err)
		}
		found, err := teamStore.FindTeamByID(ctx, team.ID)
		if err != nil {
			t.Fatalf("FindTeamByID() error = %v", err)
		}
		if found == nil || found.ID != team.ID {
			t.Errorf("FindTeamByID() = %v, want %v", found, team.ID)
		}
		return errors.New("rollback")
	})
}

func TestTeamStore_CheckTeamSlug(t *testing.T) {
	type fields struct {
		db database.Dbx
	}
	type args struct {
		ctx  context.Context
		slug string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := stores.NewStorageAdapter(tt.fields.db)
			got, err := adapter.TeamGroup().CheckTeamSlug(tt.args.ctx, tt.args.slug)
			if (err != nil) != tt.wantErr {
				t.Errorf("PostgresTeamStore.CheckTeamSlug() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("PostgresTeamStore.CheckTeamSlug() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTeamStore_FindTeamByStripeCustomerId(t *testing.T) {
	test.Short(t)
	ctx, dbx := test.DbSetup()
	_ = dbx.RunInTx(func(dbxx database.Dbx) error {
		adapter := stores.NewStorageAdapter(dbxx)
		stripeID := "cus_test_123"
		team, err := adapter.TeamGroup().CreateTeam(ctx, "StripeTeam", "stripe-team-slug")
		if err != nil {
			t.Fatalf("CreateTeam() error = %v", err)
		}
		customer, err := adapter.Customer().CreateCustomer(ctx, &models.StripeCustomer{
			ID:           stripeID,
			TeamID:       types.Pointer(team.ID),
			CustomerType: models.StripeCustomerTypeTeam,
		})
		if err != nil {
			t.Fatalf("CreateCustomer() error = %v", err)
		}
		if customer == nil || customer.ID != stripeID {
			t.Errorf("CreateCustomer() = %v, want ID %v", customer, stripeID)
		}
		found, err := adapter.TeamGroup().FindTeam(ctx, &stores.TeamFilter{
			CustomerIds: []string{stripeID},
		})
		if err != nil {
			t.Fatalf("FindTeamByStripeCustomerId() error = %v", err)
		}
		if found == nil || found.ID != team.ID {
			t.Errorf("FindTeamByStripeCustomerId() = %v, want %v", found, team.ID)
		}
		return errors.New("rollback")
	})
}

func TestTeamStore_ListTeams(t *testing.T) {
	test.Short(t)
	ctx, dbx := test.DbSetup()
	_ = dbx.RunInTx(func(dbxx database.Dbx) error {
		adapter := stores.NewStorageAdapter(dbxx)
		teamStore := adapter.TeamGroup()
		teamMemberStore := adapter.TeamMember()
		userStore := adapter.User()

		// Create users
		user1, err := userStore.CreateUser(ctx, &models.User{Email: "user1@example.com"})
		if err != nil {
			t.Fatalf("CreateUser() error = %v", err)
		}
		user2, err := userStore.CreateUser(ctx, &models.User{Email: "user2@example.com"})
		if err != nil {
			t.Fatalf("CreateUser() error = %v", err)
		}

		// Create teams
		teamA, err := teamStore.CreateTeam(ctx, "Alpha", "alpha-slug")
		if err != nil {
			t.Fatalf("CreateTeam() error = %v", err)
		}
		teamB, err := teamStore.CreateTeam(ctx, "Beta", "beta-slug")
		if err != nil {
			t.Fatalf("CreateTeam() error = %v", err)
		}
		teamC, err := teamStore.CreateTeam(ctx, "Gamma", "gamma-slug")
		if err != nil {
			t.Fatalf("CreateTeam() error = %v", err)
		}

		// Add members
		_, err = teamMemberStore.CreateTeamMember(ctx, teamA.ID, user1.ID, models.TeamMemberRoleMember, true)
		if err != nil {
			t.Fatalf("CreateTeamMember() error = %v", err)
		}
		_, err = teamMemberStore.CreateTeamMember(ctx, teamB.ID, user1.ID, models.TeamMemberRoleMember, true)
		if err != nil {
			t.Fatalf("CreateTeamMember() error = %v", err)
		}
		_, err = teamMemberStore.CreateTeamMember(ctx, teamC.ID, user2.ID, models.TeamMemberRoleMember, true)
		if err != nil {
			t.Fatalf("CreateTeamMember() error = %v", err)
		}

		// Test: List all teams
		teams, err := teamStore.ListTeams(ctx, nil)
		if err != nil {
			t.Fatalf("ListTeams() error = %v", err)
		}
		if len(teams) < 3 {
			t.Errorf("ListTeams() got %d, want at least 3", len(teams))
		}

		// Test: List teams for user1
		params := &stores.TeamFilter{
			UserIds: []uuid.UUID{user1.ID},
		}
		user1Teams, err := teamStore.ListTeams(ctx, params)
		if err != nil {
			t.Fatalf("ListTeams(user1) error = %v", err)
		}
		if len(user1Teams) != 2 {
			t.Errorf("ListTeams(user1) got %d, want 2", len(user1Teams))
		}
		found := map[uuid.UUID]bool{}
		for _, tm := range user1Teams {
			found[tm.ID] = true
		}
		if !found[teamA.ID] || !found[teamB.ID] {
			t.Errorf("ListTeams(user1) missing expected teams: %v", user1Teams)
		}

		// Test: List teams with search query
		paramsQ := &stores.TeamFilter{
			Q: "Alpha",
		}
		alphaTeams, err := teamStore.ListTeams(ctx, paramsQ)
		if err != nil {
			t.Fatalf("ListTeams(Q) error = %v", err)
		}
		if len(alphaTeams) == 0 || alphaTeams[0].Name != "Alpha" {
			t.Errorf("ListTeams(Q) got %v, want Alpha", alphaTeams)
		}

		// Test: List teams with sort order
		paramsSort := &stores.TeamFilter{
			SortParams: stores.SortParams{SortBy: "name", SortOrder: "desc"},
		}
		sortedTeams, err := teamStore.ListTeams(ctx, paramsSort)
		if err != nil {
			t.Fatalf("ListTeams(sort) error = %v", err)
		}
		if len(sortedTeams) < 2 {
			t.Errorf("ListTeams(sort) got %d, want at least 2", len(sortedTeams))
		}
		if len(sortedTeams) >= 2 && sortedTeams[0].Name < sortedTeams[1].Name {
			t.Errorf("ListTeams(sort) not sorted desc: %v", []string{sortedTeams[0].Name, sortedTeams[1].Name})
		}

		// Test: Pagination
		paramsPag := &stores.TeamFilter{}
		paramsPag.PerPage = 2
		paramsPag.Page = 0
		pagedTeams, err := teamStore.ListTeams(ctx, paramsPag)
		if err != nil {
			t.Fatalf("ListTeams(paginate) error = %v", err)
		}
		if len(pagedTeams) != 2 {
			t.Errorf("ListTeams(paginate) got %d, want 2", len(pagedTeams))
		}

		return errors.New("rollback")
	})
}
func TestTeamStore_CountTeams(t *testing.T) {
	test.Short(t)
	ctx, dbx := test.DbSetup()
	_ = dbx.RunInTx(func(dbxx database.Dbx) error {
		adapter := stores.NewStorageAdapter(dbxx)
		teamStore := adapter.TeamGroup()
		userStore := adapter.User()

		// Create users
		user1, err := userStore.CreateUser(ctx, &models.User{Email: "countuser1@example.com"})
		if err != nil {
			t.Fatalf("CreateUser() error = %v", err)
		}
		user2, err := userStore.CreateUser(ctx, &models.User{Email: "countuser2@example.com"})
		if err != nil {
			t.Fatalf("CreateUser() error = %v", err)
		}

		// Create teams
		teamA, err := teamStore.CreateTeam(ctx, "CountAlpha", "count-alpha-slug")
		if err != nil {
			t.Fatalf("CreateTeam() error = %v", err)
		}
		teamB, err := teamStore.CreateTeam(ctx, "CountBeta", "count-beta-slug")
		if err != nil {
			t.Fatalf("CreateTeam() error = %v", err)
		}
		teamC, err := teamStore.CreateTeam(ctx, "CountGamma", "count-gamma-slug")
		if err != nil {
			t.Fatalf("CreateTeam() error = %v", err)
		}

		// Add members
		_, err = adapter.TeamMember().CreateTeamMember(ctx, teamA.ID, user1.ID, models.TeamMemberRoleMember, true)
		if err != nil {
			t.Fatalf("CreateTeamMember() error = %v", err)
		}
		_, err = adapter.TeamMember().CreateTeamMember(ctx, teamB.ID, user1.ID, models.TeamMemberRoleMember, true)
		if err != nil {
			t.Fatalf("CreateTeamMember() error = %v", err)
		}
		_, err = adapter.TeamMember().CreateTeamMember(ctx, teamC.ID, user2.ID, models.TeamMemberRoleMember, true)
		if err != nil {
			t.Fatalf("CreateTeamMember() error = %v", err)
		}

		// Count all teams
		count, err := teamStore.CountTeams(ctx, nil)
		if err != nil {
			t.Fatalf("CountTeams(nil) error = %v", err)
		}
		if count < 3 {
			t.Errorf("CountTeams(nil) = %d, want at least 3", count)
		}

		// Count teams for user1
		params := &stores.TeamFilter{
			UserIds: []uuid.UUID{user1.ID},
		}
		user1Count, err := teamStore.CountTeams(ctx, params)
		if err != nil {
			t.Fatalf("CountTeams(user1) error = %v", err)
		}
		if user1Count != 2 {
			t.Errorf("CountTeams(user1) = %d, want 2", user1Count)
		}

		// Count teams with search query
		paramsQ := &stores.TeamFilter{
			Q: "CountAlpha",
		}
		alphaCount, err := teamStore.CountTeams(ctx, paramsQ)
		if err != nil {
			t.Fatalf("CountTeams(Q) error = %v", err)
		}
		if alphaCount == 0 {
			t.Errorf("CountTeams(Q) = %d, want at least 1", alphaCount)
		}

		// Count teams with no matches
		paramsNone := &stores.TeamFilter{
			Q: "NoSuchTeam",
		}
		noneCount, err := teamStore.CountTeams(ctx, paramsNone)
		if err != nil {
			t.Fatalf("CountTeams(no match) error = %v", err)
		}
		if noneCount != 0 {
			t.Errorf("CountTeams(no match) = %d, want 0", noneCount)
		}

		return errors.New("rollback")
	})
}
func TestTeamStore_FindTeamBySlug(t *testing.T) {
	test.Short(t)
	ctx, dbx := test.DbSetup()
	_ = dbx.RunInTx(func(dbxx database.Dbx) error {
		adapter := stores.NewStorageAdapter(dbxx)
		teamStore := adapter.TeamGroup()

		// Create a team with a unique slug
		team, err := teamStore.CreateTeam(ctx, "SlugTeam", "unique-slug-123")
		if err != nil {
			t.Fatalf("CreateTeam() error = %v", err)
		}

		// Should find the team by slug
		found, err := teamStore.FindTeamBySlug(ctx, "unique-slug-123")
		if err != nil {
			t.Fatalf("FindTeamBySlug() error = %v", err)
		}
		if found == nil || found.ID != team.ID {
			t.Errorf("FindTeamBySlug() = %v, want ID %v", found, team.ID)
		}

		// Should return nil for non-existent slug
		notFound, err := teamStore.FindTeamBySlug(ctx, "non-existent-slug")
		if err != nil {
			t.Fatalf("FindTeamBySlug(non-existent) error = %v", err)
		}
		if notFound != nil {
			t.Errorf("FindTeamBySlug(non-existent) = %v, want nil", notFound)
		}

		return errors.New("rollback")
	})
}
