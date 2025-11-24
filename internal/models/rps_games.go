package models

import (
	"time"

	"github.com/google/uuid"
)

// enum:"pending,cancelled,completed"
type RpsGameStatus string

const (
	RpsGameStatusPending   RpsGameStatus = "pending"
	RpsGameStatusCancelled RpsGameStatus = "cancelled"
	RpsGameStatusCompleted RpsGameStatus = "completed"
)

type RpsGame struct {
	_            struct{}          `db:"rps_games" schema:"gaming" json:"-"`
	ID           uuid.UUID         `db:"id,pk" json:"id"`
	CompletedAt  *time.Time        `db:"completed_at" json:"completed_at,omitempty"`
	ExpiresAt    time.Time         `db:"expires_at" json:"expires_at"`
	Status       RpsGameStatus     `db:"status" json:"status" default:"pending" enum:"pending,cancelled,completed"`
	Metadata     []byte            `db:"metadata" json:"metadata"`
	CreatedAt    time.Time         `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time         `db:"updated_at" json:"updated_at"`
	Participants []*RpsParticipant `db:"rps_participants" src:"id" dest:"game_id" table:"gaming.rps_participants" json:"participants,omitempty"`
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
