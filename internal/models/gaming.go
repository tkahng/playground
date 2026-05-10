package models

import (
	"time"

	"github.com/google/uuid"
)

type Player struct {
	_                     struct{}     `db:"players" schema:"gaming" json:"-"`
	ID                    uuid.UUID    `db:"id,pk" json:"id"`
	Email                 string       `db:"email" json:"email"`
	DisplayName           *string      `db:"display_name" json:"display_name,omitempty" required:"false"`
	UserID                *uuid.UUID   `db:"user_id" json:"user_id,omitempty" required:"false"`
	Metadata              []byte       `db:"metadata" json:"metadata"`
	CreatedAt             time.Time    `db:"created_at" json:"created_at"`
	UpdatedAt             time.Time    `db:"updated_at" json:"updated_at"`
	User                  *User        `db:"user" src:"user_id" dest:"id" table:"auth.users" json:"user,omitempty"`
	RequestingFriendships []*Frindship `db:"requesting_friendships" src:"id" dest:"requesting_player_id" table:"gaming.friendships" json:"requested_friendships,omitempty"`
	InvitedFriendships    []*Frindship `db:"invited_friendships" src:"id" dest:"invited_player_id" table:"gaming.friendships" json:"invited_friendships,omitempty"`
}

// enum:"pending,accepted,declined,blocked"
type FriendshipStatus string

const (
	FriendshipStatusPending  FriendshipStatus = "pending"
	FriendshipStatusAccepted FriendshipStatus = "accepted"
	FriendshipStatusDeclined FriendshipStatus = "declined"
	FriendshipStatusBlocked  FriendshipStatus = "blocked"
)

type Frindship struct {
	_                  struct{}         `db:"friendships" schema:"gaming" json:"-"`
	ID                 uuid.UUID        `db:"id,pk" json:"id"`
	RequestingPlayerID uuid.UUID        `db:"requesting_player_id" json:"requesting_player_id"`
	InvitedPlayerID    uuid.UUID        `db:"invited_player_id" json:"invited_player_id"`
	Status             FriendshipStatus `db:"status" json:"status" default:"pending" enum:"pending,accepted,declined"`
	RespondedAt        *time.Time       `db:"responded_at" json:"responded_at,omitempty"`
	CreatedAt          time.Time        `db:"created_at" json:"created_at"`
	UpdatedAt          time.Time        `db:"updated_at" json:"updated_at"`
	RequestingPlayer   *Player          `db:"requesting_player" src:"requesting_player_id" dest:"id" table:"gaming.players" json:"requesting_player,omitempty"`
	InvitedPlayer      *Player          `db:"invited_player" src:"invited_player_id" dest:"id" table:"gaming.players" json:"invited_player,omitempty"`
}
