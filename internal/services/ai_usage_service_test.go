package services_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/services"
	"github.com/tkahng/playground/internal/stores"
	"github.com/tkahng/playground/internal/tools/gemini"
	"github.com/tkahng/playground/internal/tools/types"
)

func TestAiUsageService_RecordAndCheckQuota(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		svc := services.NewAiUsageService(adapter)

		user, err := adapter.User().CreateUser(ctx, &models.User{Email: "quota@example.com"})
		if err != nil {
			t.Fatalf("CreateUser() error = %v", err)
		}
		team, err := adapter.TeamGroup().CreateTeam(ctx, "quota-team", "quota-team")
		if err != nil {
			t.Fatalf("CreateTeam() error = %v", err)
		}
		member, err := adapter.TeamMember().CreateTeamMember(ctx, &models.TeamMember{
			TeamID:           team.ID,
			UserID:           &user.ID,
			Role:             models.TeamMemberRoleOwner,
			Active:           true,
			HasBillingAccess: true,
		})
		if err != nil {
			t.Fatalf("CreateTeamMember() error = %v", err)
		}

		// Should be within quota before any usage
		if err := svc.CheckQuota(ctx, team.ID); err != nil {
			t.Errorf("CheckQuota() before usage returned error: %v", err)
		}

		// Record usage below free tier limit
		err = svc.RecordUsage(ctx, user.ID, member.ID, team.ID, gemini.Usage{
			PromptTokens:     5_000,
			CompletionTokens: 2_000,
			TotalTokens:      7_000,
		})
		if err != nil {
			t.Fatalf("RecordUsage() error = %v", err)
		}

		// Still within quota
		if err := svc.CheckQuota(ctx, team.ID); err != nil {
			t.Errorf("CheckQuota() after partial usage returned error: %v", err)
		}
	})
}

func TestAiUsageService_GetDailyLimit_FreeTier(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		svc := services.NewAiUsageService(adapter)

		_, err := adapter.TeamGroup().CreateTeam(ctx, "free-team", "free-team")
		if err != nil {
			t.Fatalf("CreateTeam() error = %v", err)
		}

		// A team with no subscription gets the free tier limit
		// We need a team ID — use a real one we just created
		team, err := adapter.TeamGroup().CreateTeam(ctx, "free-team-2", "free-team-2")
		if err != nil {
			t.Fatalf("CreateTeam() error = %v", err)
		}

		limit, err := svc.GetDailyLimit(ctx, team.ID)
		if err != nil {
			t.Fatalf("GetDailyLimit() error = %v", err)
		}
		if limit != services.FreeTierDailyTokenLimit {
			t.Errorf("GetDailyLimit() = %d, want %d (free tier)", limit, services.FreeTierDailyTokenLimit)
		}
	})
}

func TestAiUsageService_GetDailyLimit_PlanFeaturesRow(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		svc := services.NewAiUsageService(adapter)

		user, err := adapter.User().CreateUser(ctx, &models.User{Email: "paid@example.com"})
		if err != nil {
			t.Fatalf("CreateUser() error = %v", err)
		}
		team, err := adapter.TeamGroup().CreateTeam(ctx, "paid-team", "paid-team")
		if err != nil {
			t.Fatalf("CreateTeam() error = %v", err)
		}

		// Create stripe product + price + customer + subscription
		product := &models.StripeProduct{ID: "prod_ai_test", Active: true, Name: "Pro", Metadata: map[string]string{}}
		if err := adapter.Product().UpsertProduct(ctx, product); err != nil {
			t.Fatalf("UpsertProduct() error = %v", err)
		}
		price := &models.StripePrice{
			ID:        "price_ai_test",
			ProductID: product.ID,
			Active:    true,
			UnitAmount: types.Pointer(int64(5000)),
			Currency:  "usd",
			Type:      models.StripePricingTypeRecurring,
			Metadata:  map[string]string{},
		}
		if err := adapter.Price().UpsertPrice(ctx, price); err != nil {
			t.Fatalf("UpsertPrice() error = %v", err)
		}
		customer := &models.StripeCustomer{
			ID:           "cus_ai_test",
			Email:        "paid@example.com",
			CustomerType: models.StripeCustomerTypeTeam,
			TeamID:       types.Pointer(team.ID),
		}
		if _, err := adapter.Customer().CreateCustomer(ctx, customer); err != nil {
			t.Fatalf("CreateCustomer() error = %v", err)
		}
		sub := &models.StripeSubscription{
			ID:                 "sub_ai_test",
			StripeCustomerID:   customer.ID,
			Status:             models.StripeSubscriptionStatusActive,
			Metadata:           map[string]string{},
			ItemID:             "item_ai_test",
			PriceID:            price.ID,
			Quantity:           1,
			CancelAtPeriodEnd:  false,
			Created:            time.Now(),
			CurrentPeriodStart: time.Now(),
			CurrentPeriodEnd:   time.Now().Add(30 * 24 * time.Hour),
		}
		if err := adapter.Subscription().UpsertSubscription(ctx, sub); err != nil {
			t.Fatalf("UpsertSubscription() error = %v", err)
		}
		_ = user

		// No plan_features row yet → falls back to paid tier default
		limit, err := svc.GetDailyLimit(ctx, team.ID)
		if err != nil {
			t.Fatalf("GetDailyLimit() without plan_features error = %v", err)
		}
		if limit != services.PaidTierDailyTokenLimit {
			t.Errorf("GetDailyLimit() without plan_features = %d, want %d", limit, services.PaidTierDailyTokenLimit)
		}

		// Insert a plan_features row with a custom limit
		customLimit := int64(250_000)
		if _, err := adapter.PlanFeatures().Upsert(ctx, &models.PlanFeatures{
			StripeProductID: product.ID,
			DailyAiTokens:   customLimit,
		}); err != nil {
			t.Fatalf("PlanFeatures().Upsert() error = %v", err)
		}

		limit, err = svc.GetDailyLimit(ctx, team.ID)
		if err != nil {
			t.Fatalf("GetDailyLimit() with plan_features error = %v", err)
		}
		if limit != customLimit {
			t.Errorf("GetDailyLimit() with plan_features = %d, want %d", limit, customLimit)
		}
	})
}

func TestAiUsageService_CheckQuota_ExhaustedLimit(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)

		// Override AiUsage store so GetDailyTokensByTeam returns exactly the free limit
		adapter.AiUsageFunc.GetDailyTokensByTeamFunc = func(_ context.Context, _ uuid.UUID, _ time.Time) (int64, error) {
			return services.FreeTierDailyTokenLimit, nil
		}

		svc := services.NewAiUsageService(adapter)

		team, err := adapter.TeamGroup().CreateTeam(ctx, "exhausted-team", "exhausted-team")
		if err != nil {
			t.Fatalf("CreateTeam() error = %v", err)
		}

		err = svc.CheckQuota(ctx, team.ID)
		if err == nil {
			t.Error("CheckQuota() expected error when limit exhausted, got nil")
		}
	})
}

func TestSyncPlanFeatures_SeedsFreeRow(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)

		// Remove any free row committed by server startup (e.g. seedHousePlayer or
		// syncPlanFeatures called outside a transaction). Deletion is rolled back at
		// the end of the test so it doesn't affect other tests.
		if _, err := db.Exec(ctx, "DELETE FROM billing.plan_features WHERE stripe_product_id = $1", services.FreeTierProductID); err != nil {
			t.Fatalf("cleanup free row: %v", err)
		}

		// No "free" row before sync.
		before, err := adapter.PlanFeatures().FindByProductID(ctx, services.FreeTierProductID)
		if err != nil {
			t.Fatalf("FindByProductID() before sync error = %v", err)
		}
		if before != nil {
			t.Fatal("expected no free row before SyncPlanFeatures")
		}

		if err := services.SyncPlanFeatures(ctx, adapter); err != nil {
			t.Fatalf("SyncPlanFeatures() error = %v", err)
		}

		after, err := adapter.PlanFeatures().FindByProductID(ctx, services.FreeTierProductID)
		if err != nil {
			t.Fatalf("FindByProductID() after sync error = %v", err)
		}
		if after == nil {
			t.Fatal("expected free row after SyncPlanFeatures, got nil")
		}
		if after.DailyAiTokens != services.FreeTierDailyTokenLimit {
			t.Errorf("DailyAiTokens = %d, want %d", after.DailyAiTokens, services.FreeTierDailyTokenLimit)
		}
	})
}

func TestSyncPlanFeatures_DoesNotOverwriteFreeRow(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)

		// Pre-set the free row to a custom limit.
		customLimit := int64(999)
		if _, err := adapter.PlanFeatures().Upsert(ctx, &models.PlanFeatures{
			StripeProductID: services.FreeTierProductID,
			DailyAiTokens:   customLimit,
		}); err != nil {
			t.Fatalf("Upsert() error = %v", err)
		}

		// SyncPlanFeatures should not overwrite it.
		if err := services.SyncPlanFeatures(ctx, adapter); err != nil {
			t.Fatalf("SyncPlanFeatures() error = %v", err)
		}

		row, err := adapter.PlanFeatures().FindByProductID(ctx, services.FreeTierProductID)
		if err != nil {
			t.Fatalf("FindByProductID() error = %v", err)
		}
		if row.DailyAiTokens != customLimit {
			t.Errorf("DailyAiTokens = %d, want %d (existing value should be preserved)", row.DailyAiTokens, customLimit)
		}
	})
}

func TestGetDailyLimit_ReadsFreeRow(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		svc := services.NewAiUsageService(adapter)

		team, err := adapter.TeamGroup().CreateTeam(ctx, "free-row-team", "free-row-team")
		if err != nil {
			t.Fatalf("CreateTeam() error = %v", err)
		}

		// Seed the free row with a non-default value.
		customLimit := int64(7_777)
		if _, err := adapter.PlanFeatures().Upsert(ctx, &models.PlanFeatures{
			StripeProductID: services.FreeTierProductID,
			DailyAiTokens:   customLimit,
		}); err != nil {
			t.Fatalf("Upsert free row error = %v", err)
		}

		limit, err := svc.GetDailyLimit(ctx, team.ID)
		if err != nil {
			t.Fatalf("GetDailyLimit() error = %v", err)
		}
		if limit != customLimit {
			t.Errorf("GetDailyLimit() = %d, want %d (should read from free plan_features row)", limit, customLimit)
		}
	})
}
