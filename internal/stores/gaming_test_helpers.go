package stores

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/tools/security"
)

type PlayerOptionFunc func(opt *models.Player)

func WithPlayerEmail(email string) PlayerOptionFunc {
	return func(opt *models.Player) {
		opt.Email = email
	}
}

func WithPlayerDisplayName(displayName string) PlayerOptionFunc {
	return func(opt *models.Player) {
		opt.DisplayName = &displayName
	}
}

func WithPlayerMetadata(metadata string) PlayerOptionFunc {
	return func(opt *models.Player) {
		opt.Metadata = []byte(metadata)
	}
}
func WithUserID(userID uuid.UUID) PlayerOptionFunc {
	return func(opt *models.Player) {
		opt.UserID = &userID
	}
}

func MustCreatePlayer(t testing.TB, ctx context.Context, gamingStore *DBGamingStore, fns ...PlayerOptionFunc) *models.Player {
	key := security.RandomString(16)
	player := &models.Player{
		Email:    fmt.Sprintf("%s@example", key),
		Metadata: []byte(fmt.Sprintf(`{"key": "%s"}`, key)),
	}
	for _, fn := range fns {
		fn(player)
	}
	newPlayer, err := gamingStore.CreatePlayer(ctx, player)
	if err != nil {
		t.Fatalf("CreatePlayer() error = %v", err)
	}
	if newPlayer.ID == uuid.Nil {
		t.Errorf("CreatePlayer() = %v, want id not nil", newPlayer)
	}
	if newPlayer.Email != player.Email {
		t.Errorf("CreatePlayer() = %v, want email %v", newPlayer.Email, player.Email)
	}
	if newPlayer.DisplayName != nil && player.DisplayName != nil {
		if *newPlayer.DisplayName != *player.DisplayName {
			t.Errorf("CreatePlayer() = %v, want display name %v", *newPlayer.DisplayName, *player.DisplayName)
		}
	} else if newPlayer.DisplayName == nil && player.DisplayName == nil {

	} else {
		t.Errorf("CreatePlayer() = %v, want display name %v", newPlayer.DisplayName, player.DisplayName)
	}
	return newPlayer
}

type FriendshipOptionFunc func(opt *models.Frindship)

func WithStatus(status models.FriendshipStatus) FriendshipOptionFunc {
	return func(opt *models.Frindship) {
		opt.Status = status
	}
}
func MustCreateFriendship(t testing.TB, ctx context.Context, gamingStore *DBGamingStore, player1, player2 *models.Player, fns ...FriendshipOptionFunc) *models.Frindship {
	friendship := &models.Frindship{
		InvitedPlayerID:    player1.ID,
		RequestingPlayerID: player2.ID,
		Status:             models.FriendshipStatusPending,
	}
	for _, fn := range fns {
		fn(friendship)
	}
	newFriendship, err := gamingStore.CreateFriendship(ctx, friendship)
	if err != nil {
		t.Fatalf("CreateFriendship() error = %v", err)
	}
	if newFriendship.ID == uuid.Nil {
		t.Errorf("CreateFriendship() = %v, want id not nil", newFriendship)
	}
	if newFriendship.InvitedPlayerID != player1.ID {
		t.Errorf("CreateFriendship() = %v, want invited player id %v", newFriendship, player1.ID)
	}
	if newFriendship.RequestingPlayerID != player2.ID {
		t.Errorf("CreateFriendship() = %v, want requesting player id %v", newFriendship, player2.ID)
	}
	return newFriendship
}
