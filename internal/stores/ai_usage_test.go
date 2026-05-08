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
