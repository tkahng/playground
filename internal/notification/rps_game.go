package notification

import "github.com/google/uuid"

type RpsGameChallengedData struct {
	GameID             uuid.UUID `json:"game_id" required:"true"`
	RequestingPlayerID uuid.UUID `json:"requesting_player_id" required:"true"`
	RequestingEmail    string    `json:"requesting_email" required:"true"`
}

func (RpsGameChallengedData) Kind() string { return "rps_game_challenged" }

type RpsGameCompletedData struct {
	GameID  uuid.UUID `json:"game_id" required:"true"`
	Result  string    `json:"result" required:"true"`
	YourMove string   `json:"your_move" required:"true"`
	OpponentMove string `json:"opponent_move" required:"true"`
}

func (RpsGameCompletedData) Kind() string { return "rps_game_completed" }

type RpsRematchRequestedData struct {
	RematchRequestID   uuid.UUID `json:"rematch_request_id" required:"true"`
	OriginalGameID     uuid.UUID `json:"original_game_id" required:"true"`
	RequestingPlayerID uuid.UUID `json:"requesting_player_id" required:"true"`
	RequestingEmail    string    `json:"requesting_email" required:"true"`
	ExpiresAt          string    `json:"expires_at" required:"true"`
}

func (RpsRematchRequestedData) Kind() string { return "rps_rematch_requested" }

type RpsRematchAcceptedData struct {
	RematchRequestID uuid.UUID `json:"rematch_request_id" required:"true"`
	NewGameID        uuid.UUID `json:"new_game_id" required:"true"`
}

func (RpsRematchAcceptedData) Kind() string { return "rps_rematch_accepted" }

type RpsRematchDeclinedData struct {
	RematchRequestID uuid.UUID `json:"rematch_request_id" required:"true"`
	OriginalGameID   uuid.UUID `json:"original_game_id" required:"true"`
}

func (RpsRematchDeclinedData) Kind() string { return "rps_rematch_declined" }

type RpsRematchExpiredData struct {
	RematchRequestID uuid.UUID `json:"rematch_request_id" required:"true"`
	OriginalGameID   uuid.UUID `json:"original_game_id" required:"true"`
}

func (RpsRematchExpiredData) Kind() string { return "rps_rematch_expired" }

type RpsGameCancelledData struct {
	GameID             uuid.UUID `json:"game_id" required:"true"`
	CancellingPlayerID uuid.UUID `json:"cancelling_player_id" required:"true"`
}

func (RpsGameCancelledData) Kind() string { return "rps_game_cancelled" }
