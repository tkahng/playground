//go:build integration

package stores_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/stores"
	"github.com/tkahng/playground/internal/test"
)

func TestAiUsageStore_CreateAndQuery(t *testing.T) {
	t.Parallel()
	test.SkipIfShort(t)
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewStorageAdapter(db)

		user := stores.CreateUser(adapter, ctx, "aiusage@example.com")
		team := stores.CreateTeam(adapter, ctx, "aiusage-team")
		member := stores.CreateTeamMember(adapter, ctx, team, user, models.TeamMemberRoleOwner, true)

		usage := &models.AiUsage{
			UserID:           user.ID,
			TeamMemberID:     &member.ID,
			TeamID:           &team.ID,
			PromptTokens:     100,
			CompletionTokens: 50,
			TotalTokens:      150,
		}
		created, err := adapter.AiUsage().CreateAiUsage(ctx, usage)
		if err != nil {
			t.Fatalf("CreateAiUsage() error = %v", err)
		}
		if created == nil {
			t.Fatal("CreateAiUsage() returned nil")
		}
		if created.TotalTokens != 150 {
			t.Errorf("TotalTokens = %d, want 150", created.TotalTokens)
		}

		tokens, err := adapter.AiUsage().GetDailyTokensByTeamMember(ctx, member.ID, time.Now().UTC())
		if err != nil {
			t.Fatalf("GetDailyTokensByTeamMember() error = %v", err)
		}
		if tokens != 150 {
			t.Errorf("GetDailyTokensByTeamMember() = %d, want 150", tokens)
		}

		teamTokens, err := adapter.AiUsage().GetDailyTokensByTeam(ctx, team.ID, time.Now().UTC())
		if err != nil {
			t.Fatalf("GetDailyTokensByTeam() error = %v", err)
		}
		if teamTokens != 150 {
			t.Errorf("GetDailyTokensByTeam() = %d, want 150", teamTokens)
		}
	})
}

func TestAiUsageStore_MultipleUsageRows_Aggregates(t *testing.T) {
	t.Parallel()
	test.SkipIfShort(t)
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewStorageAdapter(db)

		user := stores.CreateUser(adapter, ctx, "aiusage2@example.com")
		team := stores.CreateTeam(adapter, ctx, "aiusage2-team")
		member := stores.CreateTeamMember(adapter, ctx, team, user, models.TeamMemberRoleOwner, true)

		for i := range 3 {
			_ = i
			_, err := adapter.AiUsage().CreateAiUsage(ctx, &models.AiUsage{
				UserID:           user.ID,
				TeamMemberID:     &member.ID,
				TeamID:           &team.ID,
				PromptTokens:     200,
				CompletionTokens: 100,
				TotalTokens:      300,
			})
			if err != nil {
				t.Fatalf("CreateAiUsage() error = %v", err)
			}
		}

		tokens, err := adapter.AiUsage().GetDailyTokensByTeam(ctx, team.ID, time.Now().UTC())
		if err != nil {
			t.Fatalf("GetDailyTokensByTeam() error = %v", err)
		}
		if tokens != 900 {
			t.Errorf("GetDailyTokensByTeam() = %d, want 900", tokens)
		}
	})
}

func TestAiUsageStore_ZeroWhenNoRows(t *testing.T) {
	t.Parallel()
	test.SkipIfShort(t)
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewStorageAdapter(db)

		randomTeamID := uuid.New()
		randomMemberID := uuid.New()

		tokens, err := adapter.AiUsage().GetDailyTokensByTeam(ctx, randomTeamID, time.Now().UTC())
		if err != nil {
			t.Fatalf("GetDailyTokensByTeam() error = %v", err)
		}
		if tokens != 0 {
			t.Errorf("GetDailyTokensByTeam() = %d, want 0", tokens)
		}

		memberTokens, err := adapter.AiUsage().GetDailyTokensByTeamMember(ctx, randomMemberID, time.Now().UTC())
		if err != nil {
			t.Fatalf("GetDailyTokensByTeamMember() error = %v", err)
		}
		if memberTokens != 0 {
			t.Errorf("GetDailyTokensByTeamMember() = %d, want 0", memberTokens)
		}
	})
}

func TestAiUsageStore_DayBoundary(t *testing.T) {
	t.Parallel()
	test.SkipIfShort(t)
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewStorageAdapter(db)

		user := stores.CreateUser(adapter, ctx, "aiusage3@example.com")
		team := stores.CreateTeam(adapter, ctx, "aiusage3-team")
		member := stores.CreateTeamMember(adapter, ctx, team, user, models.TeamMemberRoleOwner, true)

		_, err := adapter.AiUsage().CreateAiUsage(ctx, &models.AiUsage{
			UserID:           user.ID,
			TeamMemberID:     &member.ID,
			TeamID:           &team.ID,
			PromptTokens:     50,
			CompletionTokens: 50,
			TotalTokens:      100,
		})
		if err != nil {
			t.Fatalf("CreateAiUsage() error = %v", err)
		}

		// Query for yesterday — should return 0
		yesterday := time.Now().UTC().Add(-24 * time.Hour)
		tokens, err := adapter.AiUsage().GetDailyTokensByTeam(ctx, team.ID, yesterday)
		if err != nil {
			t.Fatalf("GetDailyTokensByTeam() error = %v", err)
		}
		if tokens != 0 {
			t.Errorf("GetDailyTokensByTeam(yesterday) = %d, want 0", tokens)
		}
	})
}

func TestAiUsageStore_ListAndCount(t *testing.T) {
	t.Parallel()
	test.SkipIfShort(t)
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewStorageAdapter(db)

		user := stores.CreateUser(adapter, ctx, "list@example.com")
		team := stores.CreateTeam(adapter, ctx, "list-team")
		member := stores.CreateTeamMember(adapter, ctx, team, user, models.TeamMemberRoleOwner, true)

		for range 3 {
			if _, err := adapter.AiUsage().CreateAiUsage(ctx, &models.AiUsage{
				UserID: user.ID, TeamMemberID: &member.ID, TeamID: &team.ID,
				PromptTokens: 100, CompletionTokens: 50, TotalTokens: 150,
			}); err != nil {
				t.Fatalf("CreateAiUsage() error = %v", err)
			}
		}

		rows, err := adapter.AiUsage().ListAiUsages(ctx, &stores.AiUsageFilter{
			PaginatedInput: stores.PaginatedInput{Page: 0, PerPage: 10},
		})
		if err != nil {
			t.Fatalf("ListAiUsages() error = %v", err)
		}
		if len(rows) < 3 {
			t.Errorf("ListAiUsages() = %d rows, want at least 3", len(rows))
		}

		total, err := adapter.AiUsage().CountAiUsages(ctx, &stores.AiUsageFilter{})
		if err != nil {
			t.Fatalf("CountAiUsages() error = %v", err)
		}
		if total < 3 {
			t.Errorf("CountAiUsages() = %d, want at least 3", total)
		}
	})
}

func TestAiUsageStore_ListFilterByTeam(t *testing.T) {
	t.Parallel()
	test.SkipIfShort(t)
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewStorageAdapter(db)

		user := stores.CreateUser(adapter, ctx, "filter@example.com")
		teamA := stores.CreateTeam(adapter, ctx, "filter-team-a")
		teamB := stores.CreateTeam(adapter, ctx, "filter-team-b")
		memberA := stores.CreateTeamMember(adapter, ctx, teamA, user, models.TeamMemberRoleOwner, true)
		memberB := stores.CreateTeamMember(adapter, ctx, teamB, user, models.TeamMemberRoleOwner, true)

		if _, err := adapter.AiUsage().CreateAiUsage(ctx, &models.AiUsage{
			UserID: user.ID, TeamMemberID: &memberA.ID, TeamID: &teamA.ID,
			PromptTokens: 10, CompletionTokens: 5, TotalTokens: 15,
		}); err != nil {
			t.Fatalf("CreateAiUsage(teamA) error = %v", err)
		}
		if _, err := adapter.AiUsage().CreateAiUsage(ctx, &models.AiUsage{
			UserID: user.ID, TeamMemberID: &memberB.ID, TeamID: &teamB.ID,
			PromptTokens: 20, CompletionTokens: 10, TotalTokens: 30,
		}); err != nil {
			t.Fatalf("CreateAiUsage(teamB) error = %v", err)
		}

		rows, err := adapter.AiUsage().ListAiUsages(ctx, &stores.AiUsageFilter{
			PaginatedInput: stores.PaginatedInput{Page: 0, PerPage: 10},
			TeamID:         &teamA.ID,
		})
		if err != nil {
			t.Fatalf("ListAiUsages(teamA) error = %v", err)
		}
		if len(rows) != 1 {
			t.Errorf("ListAiUsages(teamA) = %d rows, want 1", len(rows))
		}
		if rows[0].TotalTokens != 15 {
			t.Errorf("TotalTokens = %d, want 15", rows[0].TotalTokens)
		}

		count, err := adapter.AiUsage().CountAiUsages(ctx, &stores.AiUsageFilter{TeamID: &teamA.ID})
		if err != nil {
			t.Fatalf("CountAiUsages(teamA) error = %v", err)
		}
		if count != 1 {
			t.Errorf("CountAiUsages(teamA) = %d, want 1", count)
		}
	})
}

func TestAiUsageStore_ListPagination(t *testing.T) {
	t.Parallel()
	test.SkipIfShort(t)
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewStorageAdapter(db)

		user := stores.CreateUser(adapter, ctx, "page@example.com")
		team := stores.CreateTeam(adapter, ctx, "page-team")
		member := stores.CreateTeamMember(adapter, ctx, team, user, models.TeamMemberRoleOwner, true)

		for range 5 {
			if _, err := adapter.AiUsage().CreateAiUsage(ctx, &models.AiUsage{
				UserID: user.ID, TeamMemberID: &member.ID, TeamID: &team.ID,
				PromptTokens: 1, CompletionTokens: 1, TotalTokens: 2,
			}); err != nil {
				t.Fatalf("CreateAiUsage() error = %v", err)
			}
		}

		page0, err := adapter.AiUsage().ListAiUsages(ctx, &stores.AiUsageFilter{
			PaginatedInput: stores.PaginatedInput{Page: 0, PerPage: 3},
			TeamID:         &team.ID,
		})
		if err != nil {
			t.Fatalf("ListAiUsages(page 0) error = %v", err)
		}
		if len(page0) != 3 {
			t.Errorf("page 0 = %d rows, want 3", len(page0))
		}

		page1, err := adapter.AiUsage().ListAiUsages(ctx, &stores.AiUsageFilter{
			PaginatedInput: stores.PaginatedInput{Page: 1, PerPage: 3},
			TeamID:         &team.ID,
		})
		if err != nil {
			t.Fatalf("ListAiUsages(page 1) error = %v", err)
		}
		if len(page1) != 2 {
			t.Errorf("page 1 = %d rows, want 2", len(page1))
		}
	})
}
