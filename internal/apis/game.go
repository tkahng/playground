package apis

import (
	"time"

	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/tools/mapper"
)

type Player struct {
	_           struct{}   `db:"players" schema:"gaming" json:"-"`
	ID          uuid.UUID  `db:"id,pk" json:"id"`
	Email       string     `db:"email" json:"email"`
	DisplayName *string    `db:"display_name" json:"display_name,omitempty" required:"false"`
	UserID      *uuid.UUID `db:"user_id" json:"user_id,omitempty" required:"false"`
	Metadata    []byte     `db:"metadata" json:"metadata"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at" json:"updated_at"`
	User        *ApiUser   `db:"user" src:"user_id" dest:"id" table:"auth.users" json:"user,omitempty"`
}

func ToApiPlayer(player *models.Player) *Player {
	if player == nil {
		return nil
	}
	return &Player{
		ID:          player.ID,
		Email:       player.Email,
		DisplayName: player.DisplayName,
		UserID:      player.UserID,
		Metadata:    player.Metadata,
		CreatedAt:   player.CreatedAt,
		UpdatedAt:   player.UpdatedAt,
		User:        fromUserModel(player.User),
	}
}

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

func toApiRpsGameStatus(status models.RpsGameStatus) RpsGameStatus {
	switch status {
	case models.RpsGameStatusPending:
		return RpsGameStatusPending
	case models.RpsGameStatusCancelled:
		return RpsGameStatusCancelled
	case models.RpsGameStatusCompleted:
		return RpsGameStatusCompleted
	default:
		return RpsGameStatusPending
	}
}

func toApiRpsGame(game *models.RpsGame) *RpsGame {
	if game == nil {
		return nil
	}
	return &RpsGame{
		ID:           game.ID,
		CompletedAt:  game.CompletedAt,
		ExpiresAt:    game.ExpiresAt,
		Status:       toApiRpsGameStatus(game.Status),
		Metadata:     game.Metadata,
		CreatedAt:    game.CreatedAt,
		UpdatedAt:    game.UpdatedAt,
		Participants: mapper.Map(game.Participants, ToApiRpsParticipant),
	}
}

type RpsGameWithParticipants struct {
	RpsGame               *RpsGame
	RequestingParticipant *RpsParticipant
	InvitedParticipant    *RpsParticipant
}

// enum:"pending,declined,completed"
type RpsParticipantStatus string

const (
	RpsParticipantStatusPending   RpsParticipantStatus = "pending"
	RpsParticipantStatusDeclined  RpsParticipantStatus = "declined"
	RpsParticipantStatusCompleted RpsParticipantStatus = "completed"
)

func toApiRpsParticipantStatus(status models.RpsParticipantStatus) RpsParticipantStatus {
	switch status {
	case models.RpsParticipantStatusPending:
		return RpsParticipantStatusPending
	case models.RpsParticipantStatusDeclined:
		return RpsParticipantStatusDeclined
	case models.RpsParticipantStatusCompleted:
		return RpsParticipantStatusCompleted
	default:
		return RpsParticipantStatusPending
	}
}

// enum:"host,guest"
type RpsParticipantType string

const (
	RpsParticipantTypeHost  RpsParticipantType = "host"
	RpsParticipantTypeGuest RpsParticipantType = "guest"
)

func toApiRpsParticipantType(participantType models.RpsParticipantType) RpsParticipantType {
	switch participantType {
	case models.RpsParticipantTypeHost:
		return RpsParticipantTypeHost
	case models.RpsParticipantTypeGuest:
		return RpsParticipantTypeGuest
	default:
		return RpsParticipantTypeHost
	}
}

// enum:"rock,paper,scissors"
type RpsParticipantMove string

const (
	RpsParticipantMoveRock     RpsParticipantMove = "rock"
	RpsParticipantMovePaper    RpsParticipantMove = "paper"
	RpsParticipantMoveScissors RpsParticipantMove = "scissors"
)

func toApiRpsParticipantMove(move models.RpsParticipantMove) RpsParticipantMove {
	switch move {
	case models.RpsParticipantMoveRock:
		return RpsParticipantMoveRock
	case models.RpsParticipantMovePaper:
		return RpsParticipantMovePaper
	case models.RpsParticipantMoveScissors:
		return RpsParticipantMoveScissors
	default:
		return RpsParticipantMoveRock
	}
}

// enum:"tie,win,lose"
type RpsParticipantResult string

const (
	RpsParticipantResultTie  RpsParticipantResult = "tie"
	RpsParticipantResultWin  RpsParticipantResult = "win"
	RpsParticipantResultLose RpsParticipantResult = "lose"
)

func toApiRpsParticipantResult(result models.RpsParticipantResult) RpsParticipantResult {
	switch result {
	case models.RpsParticipantResultTie:
		return RpsParticipantResultTie
	case models.RpsParticipantResultWin:
		return RpsParticipantResultWin
	case models.RpsParticipantResultLose:
		return RpsParticipantResultLose
	default:
		return RpsParticipantResultTie
	}
}

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

func ToApiRpsParticipant(participant *models.RpsParticipant) *RpsParticipant {
	if participant == nil {
		return nil
	}
	return &RpsParticipant{
		ID:          participant.ID,
		GameID:      participant.GameID,
		PlayerID:    participant.PlayerID,
		Type:        toApiRpsParticipantType(participant.Type),
		Status:      toApiRpsParticipantStatus(participant.Status),
		Move:        toApiRpsParticipantMove(participant.Move),
		Result:      toApiRpsParticipantResult(participant.Result),
		RespondedAt: participant.RespondedAt,
		Metadata:    participant.Metadata,
		CreatedAt:   participant.CreatedAt,
		UpdatedAt:   participant.UpdatedAt,
		Game:        toApiRpsGame(participant.Game),
		Player:      ToApiPlayer(participant.Player),
	}
}

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
