//go:build integration

package services

import (
	"context"
	"testing"

	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/stores"
	"github.com/tkahng/playground/internal/tools/types"
)

func TestSeedHousePlayer_CreatesOnFirstRun(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)

		if err := SeedHousePlayer(ctx, adapter); err != nil {
			t.Fatalf("SeedHousePlayer() error = %v", err)
		}

		house, err := adapter.Gaming().FindHousePlayer(ctx)
		if err != nil {
			t.Fatalf("FindHousePlayer() error = %v", err)
		}
		if house == nil {
			t.Fatal("FindHousePlayer() returned nil after seed")
		}
		if !house.IsHouse {
			t.Error("house player IsHouse = false, want true")
		}
		if house.Email != HousePlayerEmail {
			t.Errorf("house player email = %q, want %q", house.Email, HousePlayerEmail)
		}
		if house.DisplayName == nil || *house.DisplayName != HousePlayerDisplayName {
			t.Errorf("house player display_name = %v, want %q", house.DisplayName, HousePlayerDisplayName)
		}
		if house.UserID != nil {
			t.Errorf("house player user_id = %v, want nil", house.UserID)
		}
	})
}

func TestSeedHousePlayer_IdempotentOnRepeatRuns(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)

		for range 3 {
			if err := SeedHousePlayer(ctx, adapter); err != nil {
				t.Fatalf("SeedHousePlayer() error = %v", err)
			}
		}

		// FindPlayers(nil) excludes house by default — query explicitly.
		housePlayers, err := adapter.Gaming().FindPlayers(ctx, &stores.PlayersFilter{
			IsHouse: types.OptionalParam[bool]{IsSet: true, Value: true},
		})
		if err != nil {
			t.Fatalf("FindPlayers(house) error = %v", err)
		}
		if len(housePlayers) != 1 {
			t.Errorf("house player count = %d after 3 seed calls, want 1", len(housePlayers))
		}
	})
}

func TestGetHousePlayer_ReturnsHouseAfterSeed(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)

		if err := SeedHousePlayer(ctx, adapter); err != nil {
			t.Fatalf("SeedHousePlayer() error = %v", err)
		}

		house, err := GetHousePlayer(ctx, adapter)
		if err != nil {
			t.Fatalf("GetHousePlayer() error = %v", err)
		}
		if house == nil {
			t.Fatal("GetHousePlayer() returned nil")
		}
		if !house.IsHouse {
			t.Error("IsHouse = false, want true")
		}
	})
}

func TestGetHousePlayer_ErrorWhenNotSeeded(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)

		// Remove any pre-seeded house player within this transaction (rolled back after test).
		_, _ = adapter.Gaming().DeletePlayers(ctx, &stores.PlayersFilter{
			IsHouse: types.OptionalParam[bool]{IsSet: true, Value: true},
		})

		_, err := GetHousePlayer(ctx, adapter)
		if err == nil {
			t.Fatal("GetHousePlayer() expected error when house player absent, got nil")
		}
	})
}

func TestIsHousePlayer(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)

		if err := SeedHousePlayer(ctx, adapter); err != nil {
			t.Fatalf("SeedHousePlayer() error = %v", err)
		}
		house, err := adapter.Gaming().FindHousePlayer(ctx)
		if err != nil || house == nil {
			t.Fatalf("FindHousePlayer() error = %v, player = %v", err, house)
		}

		regular := stores.MustCreatePlayer(t, ctx, adapter.Gaming(), stores.WithPlayerEmail("regular@example.com"))

		if !IsHousePlayer(house) {
			t.Error("IsHousePlayer(house) = false, want true")
		}
		if IsHousePlayer(regular) {
			t.Error("IsHousePlayer(regular) = true, want false")
		}
		if IsHousePlayer(nil) {
			t.Error("IsHousePlayer(nil) = true, want false")
		}
	})
}

func TestExcludeHousePlayers_FilterHidesHouse(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)

		if err := SeedHousePlayer(ctx, adapter); err != nil {
			t.Fatalf("SeedHousePlayer() error = %v", err)
		}
		stores.MustCreatePlayer(t, ctx, adapter.Gaming(), stores.WithPlayerEmail("p1@example.com"))
		stores.MustCreatePlayer(t, ctx, adapter.Gaming(), stores.WithPlayerEmail("p2@example.com"))

		filter := ExcludeHousePlayers()
		players, err := adapter.Gaming().FindPlayers(ctx, &filter)
		if err != nil {
			t.Fatalf("FindPlayers() error = %v", err)
		}
		for _, p := range players {
			if p.IsHouse {
				t.Errorf("ExcludeHousePlayers filter returned house player %s", p.Email)
			}
		}
		if len(players) != 2 {
			t.Errorf("player count = %d, want 2 (house excluded)", len(players))
		}
	})
}
