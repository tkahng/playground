package services

import (
	"context"
	"fmt"

	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/stores"
	"github.com/tkahng/playground/internal/tools/types"
)

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
