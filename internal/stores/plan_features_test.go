//go:build integration

package stores_test

import (
	"context"
	"testing"

	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/stores"
	"github.com/tkahng/playground/internal/test"
)

func seedStripeProduct(t testing.TB, ctx context.Context, adapter stores.StorageAdapterInterface, id string) {
	t.Helper()
	err := adapter.Product().UpsertProduct(ctx, &models.StripeProduct{
		ID:       id,
		Active:   true,
		Name:     id,
		Metadata: map[string]string{},
	})
	if err != nil {
		t.Fatalf("UpsertProduct() error = %v", err)
	}
}

func TestPlanFeaturesStore_UpsertAndFind(t *testing.T) {
	t.Parallel()
	test.SkipIfShort(t)
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewStorageAdapter(db)
		seedStripeProduct(t, ctx, adapter, "prod_pf_1")

		pf, err := adapter.PlanFeatures().Upsert(ctx, &models.PlanFeatures{
			StripeProductID: "prod_pf_1",
			DailyAiTokens:   50_000,
		})
		if err != nil {
			t.Fatalf("Upsert() error = %v", err)
		}
		if pf == nil {
			t.Fatal("Upsert() returned nil")
		}
		if pf.DailyAiTokens != 50_000 {
			t.Errorf("DailyAiTokens = %d, want 50000", pf.DailyAiTokens)
		}

		found, err := adapter.PlanFeatures().FindByProductID(ctx, "prod_pf_1")
		if err != nil {
			t.Fatalf("FindByProductID() error = %v", err)
		}
		if found == nil {
			t.Fatal("FindByProductID() returned nil")
		}
		if found.DailyAiTokens != 50_000 {
			t.Errorf("DailyAiTokens = %d, want 50000", found.DailyAiTokens)
		}
	})
}

func TestPlanFeaturesStore_Upsert_UpdatesExisting(t *testing.T) {
	t.Parallel()
	test.SkipIfShort(t)
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewStorageAdapter(db)
		seedStripeProduct(t, ctx, adapter, "prod_pf_2")

		_, err := adapter.PlanFeatures().Upsert(ctx, &models.PlanFeatures{
			StripeProductID: "prod_pf_2",
			DailyAiTokens:   10_000,
		})
		if err != nil {
			t.Fatalf("first Upsert() error = %v", err)
		}

		updated, err := adapter.PlanFeatures().Upsert(ctx, &models.PlanFeatures{
			StripeProductID: "prod_pf_2",
			DailyAiTokens:   200_000,
		})
		if err != nil {
			t.Fatalf("second Upsert() error = %v", err)
		}
		if updated.DailyAiTokens != 200_000 {
			t.Errorf("DailyAiTokens after update = %d, want 200000", updated.DailyAiTokens)
		}
	})
}

func TestPlanFeaturesStore_InsertIfMissing_Idempotent(t *testing.T) {
	t.Parallel()
	test.SkipIfShort(t)
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewStorageAdapter(db)
		seedStripeProduct(t, ctx, adapter, "prod_pf_3")

		// Insert the initial row via Upsert with a custom value
		_, err := adapter.PlanFeatures().Upsert(ctx, &models.PlanFeatures{
			StripeProductID: "prod_pf_3",
			DailyAiTokens:   75_000,
		})
		if err != nil {
			t.Fatalf("Upsert() error = %v", err)
		}

		// InsertIfMissing should not overwrite the existing value
		err = adapter.PlanFeatures().InsertIfMissing(ctx, &models.PlanFeatures{
			StripeProductID: "prod_pf_3",
			DailyAiTokens:   10_000,
		})
		if err != nil {
			t.Fatalf("InsertIfMissing() error = %v", err)
		}

		found, err := adapter.PlanFeatures().FindByProductID(ctx, "prod_pf_3")
		if err != nil {
			t.Fatalf("FindByProductID() error = %v", err)
		}
		if found.DailyAiTokens != 75_000 {
			t.Errorf("DailyAiTokens after InsertIfMissing = %d, want 75000 (existing should not be overwritten)", found.DailyAiTokens)
		}
	})
}

func TestPlanFeaturesStore_InsertIfMissing_InsertsWhenAbsent(t *testing.T) {
	t.Parallel()
	test.SkipIfShort(t)
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewStorageAdapter(db)
		seedStripeProduct(t, ctx, adapter, "prod_pf_4")

		err := adapter.PlanFeatures().InsertIfMissing(ctx, &models.PlanFeatures{
			StripeProductID: "prod_pf_4",
			DailyAiTokens:   10_000,
		})
		if err != nil {
			t.Fatalf("InsertIfMissing() error = %v", err)
		}

		found, err := adapter.PlanFeatures().FindByProductID(ctx, "prod_pf_4")
		if err != nil {
			t.Fatalf("FindByProductID() error = %v", err)
		}
		if found == nil {
			t.Fatal("FindByProductID() returned nil after InsertIfMissing")
		}
		if found.DailyAiTokens != 10_000 {
			t.Errorf("DailyAiTokens = %d, want 10000", found.DailyAiTokens)
		}
	})
}

func TestPlanFeaturesStore_FindByProductID_ReturnsNilWhenMissing(t *testing.T) {
	t.Parallel()
	test.SkipIfShort(t)
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewStorageAdapter(db)

		found, err := adapter.PlanFeatures().FindByProductID(ctx, "prod_does_not_exist")
		if err != nil {
			t.Fatalf("FindByProductID() error = %v", err)
		}
		if found != nil {
			t.Errorf("FindByProductID() = %v, want nil", found)
		}
	})
}

func TestPlanFeaturesStore_List(t *testing.T) {
	t.Parallel()
	test.SkipIfShort(t)
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewStorageAdapter(db)
		seedStripeProduct(t, ctx, adapter, "prod_list_1")
		seedStripeProduct(t, ctx, adapter, "prod_list_2")

		for _, id := range []string{"prod_list_1", "prod_list_2"} {
			_, err := adapter.PlanFeatures().Upsert(ctx, &models.PlanFeatures{
				StripeProductID: id,
				DailyAiTokens:   10_000,
			})
			if err != nil {
				t.Fatalf("Upsert(%s) error = %v", id, err)
			}
		}

		rows, err := adapter.PlanFeatures().List(ctx, &stores.PlanFeaturesFilter{
			PaginatedInput: stores.PaginatedInput{Page: 0, PerPage: 100},
		})
		if err != nil {
			t.Fatalf("List() error = %v", err)
		}
		if len(rows) < 2 {
			t.Errorf("List() = %d rows, want at least 2", len(rows))
		}
	})
}

func TestPlanFeaturesStore_CountPlanFeatures(t *testing.T) {
	t.Parallel()
	test.SkipIfShort(t)
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewStorageAdapter(db)
		seedStripeProduct(t, ctx, adapter, "prod_count_1")
		seedStripeProduct(t, ctx, adapter, "prod_count_2")

		for _, id := range []string{"prod_count_1", "prod_count_2"} {
			_, err := adapter.PlanFeatures().Upsert(ctx, &models.PlanFeatures{
				StripeProductID: id,
				DailyAiTokens:   10_000,
			})
			if err != nil {
				t.Fatalf("Upsert(%s) error = %v", id, err)
			}
		}

		count, err := adapter.PlanFeatures().CountPlanFeatures(ctx, &stores.PlanFeaturesFilter{})
		if err != nil {
			t.Fatalf("CountPlanFeatures() error = %v", err)
		}
		if count < 2 {
			t.Errorf("CountPlanFeatures() = %d, want at least 2", count)
		}
	})
}

func TestPlanFeaturesStore_ListPagination(t *testing.T) {
	t.Parallel()
	test.SkipIfShort(t)
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewStorageAdapter(db)
		ids := []string{"prod_page_1", "prod_page_2", "prod_page_3"}
		for _, id := range ids {
			seedStripeProduct(t, ctx, adapter, id)
			_, err := adapter.PlanFeatures().Upsert(ctx, &models.PlanFeatures{
				StripeProductID: id,
				DailyAiTokens:   10_000,
			})
			if err != nil {
				t.Fatalf("Upsert(%s) error = %v", id, err)
			}
		}

		page0, err := adapter.PlanFeatures().List(ctx, &stores.PlanFeaturesFilter{
			PaginatedInput: stores.PaginatedInput{Page: 0, PerPage: 2},
		})
		if err != nil {
			t.Fatalf("List(page=0, per_page=2) error = %v", err)
		}
		if len(page0) != 2 {
			t.Errorf("List(page=0, per_page=2) = %d rows, want 2", len(page0))
		}

		page1, err := adapter.PlanFeatures().List(ctx, &stores.PlanFeaturesFilter{
			PaginatedInput: stores.PaginatedInput{Page: 1, PerPage: 2},
		})
		if err != nil {
			t.Fatalf("List(page=1, per_page=2) error = %v", err)
		}
		if len(page1) < 1 {
			t.Errorf("List(page=1, per_page=2) = %d rows, want at least 1", len(page1))
		}

		if page0[0].StripeProductID == page1[0].StripeProductID {
			t.Error("page 0 and page 1 returned the same first row")
		}
	})
}
