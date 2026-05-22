//go:build integration

package services

import (
	"context"
	"testing"

	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/stores"
	"github.com/tkahng/playground/internal/tools/types"
)

func TestIsHouseEnabled_DefaultsToTrue(t *testing.T) {
	cases := []struct {
		name     string
		metadata []byte
		want     bool
	}{
		{"empty", []byte("{}"), true},
		{"nil", nil, true},
		{"explicit true", []byte(`{"house_enabled":true}`), true},
		{"explicit false", []byte(`{"house_enabled":false}`), false},
		{"other keys", []byte(`{"foo":"bar"}`), true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := IsHouseEnabled(tc.metadata); got != tc.want {
				t.Errorf("IsHouseEnabled(%s) = %v, want %v", tc.metadata, got, tc.want)
			}
		})
	}
}

func TestSetHouseEnabled_RoundTrip(t *testing.T) {
	meta := []byte("{}")

	disabled, err := SetHouseEnabled(meta, false)
	if err != nil {
		t.Fatalf("SetHouseEnabled(false) error = %v", err)
	}
	if IsHouseEnabled(disabled) {
		t.Error("expected disabled after SetHouseEnabled(false)")
	}

	enabled, err := SetHouseEnabled(disabled, true)
	if err != nil {
		t.Fatalf("SetHouseEnabled(true) error = %v", err)
	}
	if !IsHouseEnabled(enabled) {
		t.Error("expected enabled after SetHouseEnabled(true)")
	}
}

func TestChallengeHouse_Forbidden_WhenDisabled(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		svc := newHouseTestService(t, ctx, adapter)

		// Disable the house player.
		house, _ := adapter.Gaming().FindHousePlayer(ctx)
		disabledMeta, _ := SetHouseEnabled(house.Metadata, false)
		house.Metadata = disabledMeta
		if _, err := adapter.Gaming().UpdatePlayer(ctx, house); err != nil {
			t.Fatalf("UpdatePlayer: %v", err)
		}

		player := stores.MustCreatePlayer(t, ctx, adapter.Gaming(), stores.WithPlayerEmail("disabled_h@example.com"))
		_, err := svc.ChallengeHouse(ctx, &ChallengeHouseInput{
			RequestingPlayerID:   player.ID,
			RequestingPlayerMove: models.RpsParticipantMoveRock,
		})
		if err == nil {
			t.Fatal("expected error when house is disabled, got nil")
		}
	})
}

func TestGetHousePlayerStats_ErrorWhenNotSeeded(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)

		// Remove any pre-seeded house player within this transaction (rolled back after test).
		_, _ = adapter.Gaming().DeletePlayers(ctx, &stores.PlayersFilter{
			IsHouse: types.OptionalParam[bool]{IsSet: true, Value: true},
		})

		_, err := GetHousePlayerStats(ctx, adapter)
		if err == nil {
			t.Fatal("expected error when house player is not seeded, got nil")
		}
	})
}

func TestGetHousePlayerStats_Empty(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		if err := SeedHousePlayer(ctx, adapter); err != nil {
			t.Fatalf("SeedHousePlayer: %v", err)
		}

		stats, err := GetHousePlayerStats(ctx, adapter)
		if err != nil {
			t.Fatalf("GetHousePlayerStats() error = %v", err)
		}
		if stats.TotalGames != 0 {
			t.Errorf("TotalGames = %d, want 0", stats.TotalGames)
		}
		if !stats.Enabled {
			t.Error("Enabled = false, want true for freshly seeded house")
		}
	})
}

func TestGetHousePlayerStats_AfterGames(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		svc := newHouseTestService(t, ctx, adapter)

		// Play 3 games: wins, losses, and ties are random, but totals should add up.
		for i := range 3 {
			email := ""
			switch i {
			case 0:
				email = "stats_p0@example.com"
			case 1:
				email = "stats_p1@example.com"
			case 2:
				email = "stats_p2@example.com"
			}
			p := stores.MustCreatePlayer(t, ctx, adapter.Gaming(), stores.WithPlayerEmail(email))
			if _, err := svc.ChallengeHouse(ctx, &ChallengeHouseInput{
				RequestingPlayerID:   p.ID,
				RequestingPlayerMove: models.RpsParticipantMoveRock,
			}); err != nil {
				t.Fatalf("ChallengeHouse() error = %v", err)
			}
		}

		stats, err := GetHousePlayerStats(ctx, adapter)
		if err != nil {
			t.Fatalf("GetHousePlayerStats() error = %v", err)
		}
		if stats.TotalGames != 3 {
			t.Errorf("TotalGames = %d, want 3", stats.TotalGames)
		}
		if stats.HouseWins+stats.UserWins+stats.Ties != 3 {
			t.Errorf("win/lose/tie sum = %d, want 3", stats.HouseWins+stats.UserWins+stats.Ties)
		}
		if stats.BettedGames != 0 {
			t.Errorf("BettedGames = %d, want 0 (no bets)", stats.BettedGames)
		}
	})
}

func TestGetHousePlayerStats_WithBets(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)
		svc := newHouseTestService(t, ctx, adapter)

		user := mustCreateUser(t, ctx, adapter, "stats_bet@example.com")
		player := stores.MustCreatePlayer(t, ctx, adapter.Gaming(),
			stores.WithPlayerEmail("stats_bet@example.com"),
			stores.WithUserID(user.ID),
		)
		mustFundPlayerWallet(t, ctx, adapter, ledger, &user.ID, 500)

		betAmt := int64(50)
		if _, err := svc.ChallengeHouse(ctx, &ChallengeHouseInput{
			RequestingPlayerID:   player.ID,
			RequestingPlayerMove: models.RpsParticipantMoveRock,
			BetAmount:            &betAmt,
			HostUserID:           &user.ID,
		}); err != nil {
			t.Fatalf("ChallengeHouse() error = %v", err)
		}

		stats, err := GetHousePlayerStats(ctx, adapter)
		if err != nil {
			t.Fatalf("GetHousePlayerStats() error = %v", err)
		}
		if stats.BettedGames != 1 {
			t.Errorf("BettedGames = %d, want 1", stats.BettedGames)
		}
		if stats.TotalBetAmount != 50 {
			t.Errorf("TotalBetAmount = %d, want 50", stats.TotalBetAmount)
		}
	})
}
