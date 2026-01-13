package stores

import (
	"context"
	"fmt"
	"testing"
	"time"

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

func MustCreatePlayer(t testing.TB, ctx context.Context, gamingStore GamingStore, fns ...PlayerOptionFunc) *models.Player {
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
func MustCreateFriendship(t testing.TB, ctx context.Context, gamingStore GamingStore, player1, player2 *models.Player, fns ...FriendshipOptionFunc) *models.Frindship {
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

type RpsGameOptionFunc func(opt *models.RpsGame)

func WithRpsGameStatus(status models.RpsGameStatus) RpsGameOptionFunc {
	return func(opt *models.RpsGame) {
		opt.Status = status
	}
}

func WithRpsGameExpiresAt(expiresAt time.Time) RpsGameOptionFunc {
	return func(opt *models.RpsGame) {
		opt.ExpiresAt = expiresAt
	}
}

func WithRpsGameMetadata(metadata string) RpsGameOptionFunc {
	return func(opt *models.RpsGame) {
		opt.Metadata = []byte(metadata)
	}
}

func WithRpsGameParticipants(participants []*models.RpsParticipant) RpsGameOptionFunc {
	return func(opt *models.RpsGame) {
		opt.Participants = participants
	}
}

func MustCreateRpsGame(t testing.TB, gamingStore GamingStore, fns ...RpsGameOptionFunc) *models.RpsGame {
	key := security.RandomString(16)
	game := &models.RpsGame{
		Status:    models.RpsGameStatusPending,
		ExpiresAt: time.Now().UTC().Add(time.Hour * 36),
		Metadata:  []byte(fmt.Sprintf(`{"key": "%s"}`, key)),
	}
	for _, fn := range fns {
		fn(game)
	}
	newGame, err := gamingStore.CreateRpsGame(t.Context(), game)
	if err != nil {
		t.Fatalf("CreateRpsGame() error = %v", err)
	}
	return newGame
}

type RpsParticipantOptionFunc func(opt *models.RpsParticipant)

func WithRpsParticipantStatus(status models.RpsParticipantStatus) RpsParticipantOptionFunc {
	return func(opt *models.RpsParticipant) {
		opt.Status = status
	}
}

func WithRpsParticipantMove(move models.RpsParticipantMove) RpsParticipantOptionFunc {
	return func(opt *models.RpsParticipant) {
		opt.Move = move
	}
}

func WithRpsParticipantResult(result models.RpsParticipantResult) RpsParticipantOptionFunc {
	return func(opt *models.RpsParticipant) {
		opt.Result = result
	}
}

func WithRpsParticipantMetadata(metadata string) RpsParticipantOptionFunc {
	return func(opt *models.RpsParticipant) {
		opt.Metadata = []byte(metadata)
	}
}

func WithRpsParticipantType(typ models.RpsParticipantType) RpsParticipantOptionFunc {
	return func(opt *models.RpsParticipant) {
		opt.Type = typ
	}
}

func WithRpsParticipantRespondedAt(repondedAt time.Time) RpsParticipantOptionFunc {
	return func(opt *models.RpsParticipant) {
		opt.RespondedAt = &repondedAt
	}
}

func MustCreateRpsParticipant(t testing.TB, gamingStore GamingStore, fns ...RpsParticipantOptionFunc) *models.RpsParticipant {
	participant := &models.RpsParticipant{
		Type:   models.RpsParticipantTypeHost,
		Status: models.RpsParticipantStatusPending,
		Move:   models.RpsParticipantMovePaper,
		Result: models.RpsParticipantResultTie,
	}
	for _, fn := range fns {
		fn(participant)
	}
	newParticipant, err := gamingStore.CreateRpsParticipant(t.Context(), participant)
	if err != nil {
		t.Fatalf("CreateRpsParticipant() error = %v", err)
	}
	return newParticipant
}
