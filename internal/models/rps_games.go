package models

import (
	"time"

	"github.com/google/uuid"
)

// enum:"pending,accepted,declined,expired"
type RpsRematchStatus string

const (
	RpsRematchStatusPending  RpsRematchStatus = "pending"
	RpsRematchStatusAccepted RpsRematchStatus = "accepted"
	RpsRematchStatusDeclined RpsRematchStatus = "declined"
	RpsRematchStatusExpired  RpsRematchStatus = "expired"
)

type RpsRematchRequest struct {
	_                   struct{}         `db:"rps_rematch_requests" schema:"gaming" json:"-"`
	ID                  uuid.UUID        `db:"id,pk" json:"id"`
	OriginalGameID      uuid.UUID        `db:"original_game_id" json:"original_game_id"`
	RequestingPlayerID  uuid.UUID        `db:"requesting_player_id" json:"requesting_player_id"`
	InvitedPlayerID     uuid.UUID        `db:"invited_player_id" json:"invited_player_id"`
	Status              RpsRematchStatus `db:"status" json:"status" default:"pending" enum:"pending,accepted,declined,expired"`
	NewGameID           *uuid.UUID       `db:"new_game_id" json:"new_game_id,omitempty"`
	ExpiresAt           time.Time        `db:"expires_at" json:"expires_at"`
	Metadata            []byte           `db:"metadata" json:"metadata"`
	CreatedAt           time.Time        `db:"created_at" json:"created_at"`
	UpdatedAt           time.Time        `db:"updated_at" json:"updated_at"`
	RequestingPlayer    *Player          `db:"requesting_player" src:"requesting_player_id" dest:"id" table:"gaming.players" json:"requesting_player,omitempty"`
	InvitedPlayer       *Player          `db:"invited_player" src:"invited_player_id" dest:"id" table:"gaming.players" json:"invited_player,omitempty"`
	OriginalGame        *RpsGame         `db:"original_game" src:"original_game_id" dest:"id" table:"gaming.rps_games" json:"original_game,omitempty"`
}

// enum:"pending,cancelled,completed"
type RpsGameStatus string

const (
	RpsGameStatusPending   RpsGameStatus = "pending"
	RpsGameStatusCancelled RpsGameStatus = "cancelled"
	RpsGameStatusCompleted RpsGameStatus = "completed"
)

type RpsGame struct {
	_           struct{}      `db:"rps_games" schema:"gaming" json:"-"`
	ID          uuid.UUID     `db:"id,pk" json:"id"`
	CompletedAt *time.Time    `db:"completed_at" json:"completed_at,omitempty"`
	ExpiresAt   time.Time     `db:"expires_at" json:"expires_at"`
	Status      RpsGameStatus `db:"status" json:"status" default:"pending" enum:"pending,cancelled,completed"`
	Metadata    []byte        `db:"metadata" json:"metadata"`
	// BetAmount is the number of points wagered by each player.
	// nil means no bet is associated with this game.
	BetAmount *int64 `db:"bet_amount" json:"bet_amount,omitempty"`
	// HostBetTransferID is the pending ledger transfer ID for the host's escrow bet.
	HostBetTransferID *uuid.UUID `db:"host_bet_transfer_id" json:"host_bet_transfer_id,omitempty"`
	// GuestBetTransferID is the pending ledger transfer ID for the guest's escrow bet.
	GuestBetTransferID *uuid.UUID        `db:"guest_bet_transfer_id" json:"guest_bet_transfer_id,omitempty"`
	CreatedAt          time.Time         `db:"created_at" json:"created_at"`
	UpdatedAt          time.Time         `db:"updated_at" json:"updated_at"`
	Participants       []*RpsParticipant `db:"rps_participants" src:"id" dest:"game_id" table:"gaming.rps_participants" json:"participants,omitempty"`
}

// enum:"pending,declined,completed"
type RpsParticipantStatus string

const (
	RpsParticipantStatusPending   RpsParticipantStatus = "pending"
	RpsParticipantStatusDeclined  RpsParticipantStatus = "declined"
	RpsParticipantStatusCompleted RpsParticipantStatus = "completed"
)

// enum:"host,guest"
type RpsParticipantType string

const (
	RpsParticipantTypeHost  RpsParticipantType = "host"
	RpsParticipantTypeGuest RpsParticipantType = "guest"
)

// enum:"rock,paper,scissors"
type RpsParticipantMove string

const (
	RpsParticipantMoveRock     RpsParticipantMove = "rock"
	RpsParticipantMovePaper    RpsParticipantMove = "paper"
	RpsParticipantMoveScissors RpsParticipantMove = "scissors"
)

// enum:"tie,win,lose"
type RpsParticipantResult string

const (
	RpsParticipantResultTie  RpsParticipantResult = "tie"
	RpsParticipantResultWin  RpsParticipantResult = "win"
	RpsParticipantResultLose RpsParticipantResult = "lose"
)

type RpsParticipant struct {
	_           struct{}             `db:"rps_participants" schema:"gaming" json:"-"`
	ID          uuid.UUID            `db:"id,pk" json:"id"`
	GameID      uuid.UUID            `db:"game_id" json:"game_id"`
	PlayerID    uuid.UUID            `db:"player_id" json:"player_id"`
	Type        RpsParticipantType   `db:"type" json:"type" default:"host" enum:"host,guest"`
	Status      RpsParticipantStatus `db:"status" json:"status" default:"pending" enum:"pending,declined,completed"`
	Move        RpsParticipantMove   `db:"move" json:"move" default:"rock" enum:"rock,paper,scissors"`
	Result      RpsParticipantResult `db:"result" json:"result" default:"tie" enum:"tie,win,lose"`
	RespondedAt *time.Time           `db:"responded_at" json:"responded_at,omitempty"`
	Metadata    []byte               `db:"metadata" json:"metadata"`
	CreatedAt   time.Time            `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time            `db:"updated_at" json:"updated_at"`
	Game        *RpsGame             `db:"game" src:"game_id" dest:"id" table:"gaming.rps_games" json:"game,omitempty"`
	Player      *Player              `db:"player" src:"player_id" dest:"id" table:"gaming.players" json:"player,omitempty"`
}

// create table gaming.rps_game_invites (
//
//		id uuid primary key default uuidv7(),
//		game_id uuid not null references gaming.rps_games(id),
//		requesting_player_id uuid not null references gaming.players(id),
//		invited_player_id uuid not null references gaming.players(id),
//		token text not null unique,
//	    expires_at timestamptz NOT NULL,
//		metadata jsonb not null default '{}'::jsonb,
//		created_at timestamptz not null default clock_timestamp(),
//		updated_at timestamptz not null default clock_timestamp(),
//		constraint rps_game_invites_token check (utility.not_empty(token))
//
// );
type RpsGameInvite struct {
	_                  struct{}  `db:"rps_game_invites" schema:"gaming" json:"-"`
	ID                 uuid.UUID `db:"id,pk" json:"id"`
	GameID             uuid.UUID `db:"game_id" json:"game_id"`
	RequestingPlayerID uuid.UUID `db:"requesting_player_id" json:"requesting_player_id"`
	InvitedPlayerID    uuid.UUID `db:"invited_player_id" json:"invited_player_id"`
	Token              string    `db:"token" json:"token"`
	ExpiresAt          time.Time `db:"expires_at" json:"expires_at"`
	Metadata           []byte    `db:"metadata" json:"metadata"`
	CreatedAt          time.Time `db:"created_at" json:"created_at"`
	UpdatedAt          time.Time `db:"updated_at" json:"updated_at"`
	Game               *RpsGame  `db:"game" src:"game_id" dest:"id" table:"gaming.rps_games" json:"game,omitempty"`
	RequestingPlayer   *Player   `db:"requesting_player" src:"requesting_player_id" dest:"id" table:"gaming.players" json:"requesting_player,omitempty"`
	InvitedPlayer      *Player   `db:"invited_player" src:"invited_player_id" dest:"id" table:"gaming.players" json:"invited_player,omitempty"`
}
