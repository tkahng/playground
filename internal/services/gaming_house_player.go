package services

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/database/repository"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/stores"
	"github.com/tkahng/playground/internal/tools/types"
)

type houseMeta struct {
	HouseEnabled *bool `json:"house_enabled,omitempty"`
}

// IsHouseEnabled reports whether the house player's metadata has house_enabled = true (default).
func IsHouseEnabled(metadata []byte) bool {
	if len(metadata) == 0 {
		return true
	}
	var m houseMeta
	if err := json.Unmarshal(metadata, &m); err != nil {
		return true
	}
	return m.HouseEnabled == nil || *m.HouseEnabled
}

// SetHouseEnabled sets the house_enabled flag in the metadata JSON blob and returns the updated bytes.
func SetHouseEnabled(metadata []byte, enabled bool) ([]byte, error) {
	raw := map[string]any{}
	if len(metadata) > 0 {
		if err := json.Unmarshal(metadata, &raw); err != nil {
			return nil, fmt.Errorf("parse house metadata: %w", err)
		}
	}
	raw["house_enabled"] = enabled
	return json.Marshal(raw)
}

// HousePlayerStats summarises all-time activity for the house player.
type HousePlayerStats struct {
	TotalGames     int64 `json:"total_games"`
	BettedGames    int64 `json:"betted_games"`
	HouseWins      int64 `json:"house_wins"`
	UserWins       int64 `json:"user_wins"`
	Ties           int64 `json:"ties"`
	TotalBetAmount int64 `json:"total_bet_amount"`
	Enabled        bool  `json:"enabled"`
}

// GetHousePlayerStats computes stats for the house player across all completed games.
func GetHousePlayerStats(ctx context.Context, adapter stores.StorageAdapterInterface) (*HousePlayerStats, error) {
	house, err := adapter.Gaming().FindHousePlayer(ctx)
	if err != nil {
		return nil, fmt.Errorf("find house player: %w", err)
	}
	if house == nil {
		return nil, fmt.Errorf("house player not found")
	}

	totalGames, err := adapter.Gaming().CountRpsGames(ctx, &stores.RpsGameFilter{
		ParticipantIds: []uuid.UUID{house.ID},
	})
	if err != nil {
		return nil, fmt.Errorf("count house games: %w", err)
	}

	// House participant records give win/lose/tie from the house's perspective.
	houseParticipants, err := adapter.Gaming().FindRpsParticipants(ctx, &stores.RpsParticipantFilter{
		PlayerIds: []uuid.UUID{house.ID},
		Statuses:  []models.RpsParticipantStatus{models.RpsParticipantStatusCompleted},
		PaginatedInput: repository.PaginatedInput{
			Page:    0,
			PerPage: 1_000_000,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("find house participants: %w", err)
	}

	var houseWins, userWins, ties int64
	for _, p := range houseParticipants {
		switch p.Result {
		case models.RpsParticipantResultWin:
			houseWins++
		case models.RpsParticipantResultLose:
			userWins++
		case models.RpsParticipantResultTie:
			ties++
		}
	}

	// Sum bet amounts from all house games.
	houseGames, err := adapter.Gaming().FindRpsGames(ctx, &stores.RpsGameFilter{
		ParticipantIds: []uuid.UUID{house.ID},
		PaginatedInput: repository.PaginatedInput{
			Page:    0,
			PerPage: 1_000_000,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("find house games: %w", err)
	}

	var bettedGames, totalBetAmount int64
	for _, g := range houseGames {
		if g.BetAmount != nil && *g.BetAmount > 0 {
			bettedGames++
			totalBetAmount += *g.BetAmount
		}
	}

	return &HousePlayerStats{
		TotalGames:     totalGames,
		BettedGames:    bettedGames,
		HouseWins:      houseWins,
		UserWins:       userWins,
		Ties:           ties,
		TotalBetAmount: totalBetAmount,
		Enabled:        IsHouseEnabled(house.Metadata),
	}, nil
}

const (
	HousePlayerEmail       = "house@system.internal"
	HousePlayerDisplayName = "The House"
)

// SeedHousePlayer ensures exactly one house player record exists. Safe to call
// on every startup — it is a no-op when the record already exists.
func SeedHousePlayer(ctx context.Context, adapter stores.StorageAdapterInterface) error {
	existing, err := adapter.Gaming().FindHousePlayer(ctx)
	if err != nil {
		return fmt.Errorf("find house player: %w", err)
	}
	if existing != nil {
		return nil
	}
	displayName := HousePlayerDisplayName
	_, err = adapter.Gaming().CreatePlayer(ctx, &models.Player{
		Email:       HousePlayerEmail,
		DisplayName: &displayName,
		IsHouse:     true,
		Metadata:    []byte("{}"),
		UserID:      nil,
	})
	if err != nil {
		return fmt.Errorf("create house player: %w", err)
	}
	return nil
}

// GetHousePlayer returns the house player, returning an error if it doesn't exist.
func GetHousePlayer(ctx context.Context, adapter stores.StorageAdapterInterface) (*models.Player, error) {
	house, err := adapter.Gaming().FindHousePlayer(ctx)
	if err != nil {
		return nil, fmt.Errorf("find house player: %w", err)
	}
	if house == nil {
		return nil, fmt.Errorf("house player not found")
	}
	return house, nil
}

// IsHousePlayer reports whether a player ID belongs to the house.
func IsHousePlayer(player *models.Player) bool {
	return player != nil && player.IsHouse
}

// ExcludeHousePlayers returns a filter option that hides the house player from
// results — use this for leaderboards, search, and friend suggestions.
func ExcludeHousePlayers() stores.PlayersFilter {
	return stores.PlayersFilter{
		IsHouse: types.OptionalParam[bool]{IsSet: true, Value: false},
	}
}
