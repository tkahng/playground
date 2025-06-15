package apis

import (
	"context"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/sse"
	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/models"
)

type Notification struct {
	_            struct{}       `db:"notifications" json:"-"`
	ID           uuid.UUID      `db:"id,pk" json:"id"`
	CreatedAt    time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time      `db:"updated_at" json:"updated_at"`
	ReadAt       *time.Time     `db:"read_at" json:"read_at,omitempty"`
	Channel      string         `db:"channel" json:"channel"`
	Payload      []byte         `db:"payload" json:"payload"`
	UserID       *uuid.UUID     `db:"user_id" json:"user_id,omitempty"`
	TeamMemberID *uuid.UUID     `db:"team_member_id" json:"team_member_id,omitempty"`
	TeamID       *uuid.UUID     `db:"team_id" json:"team_id,omitempty"`
	Metadata     map[string]any `db:"metadata" json:"metadata"`
	Type         string         `db:"type" json:"type"`
	User         *ApiUser       `db:"user" src:"user_id" dest:"id" table:"users" json:"user,omitempty"`
	TeamMember   *TeamMember    `db:"team_member" src:"team_member_id" dest:"id" table:"team_members" json:"team_member,omitempty"`
	Team         *Team          `db:"team" src:"team_id" dest:"id" table:"teams" json:"team,omitempty"`
}

func FromModelNotification(notification *models.Notification) *Notification {
	return &Notification{
		ID:           notification.ID,
		CreatedAt:    notification.CreatedAt,
		UpdatedAt:    notification.UpdatedAt,
		ReadAt:       notification.ReadAt,
		Channel:      notification.Channel,
		UserID:       notification.UserID,
		TeamMemberID: notification.TeamMemberID,
		TeamID:       notification.TeamID,
		Metadata:     notification.Metadata,
		Type:         notification.Type,
		User:         FromUserModel(notification.User),
		TeamMember:   FromTeamMemberModel(notification.TeamMember),
		Team:         FromTeamModel(notification.Team),
	}
}

func BindSSE(api huma.API, appApi *Api) {
	// Register the SSE endpoint
	sse.Register(api, huma.Operation{
		OperationID: "sse",
		Method:      http.MethodGet,
		Path:        "/sse",
		Summary:     "Server sent events example",
	}, map[string]any{
		// Mapping of event type name to Go struct for that event.
		"notification": Notification{},
	}, func(ctx context.Context, input *struct{}, send sse.Sender) {
		// Send an event every second for 10 seconds.
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				ctx.Err()
			case <-ticker.C:
				send.Data(
					Notification{
						ID:        uuid.New(),
						Metadata:  map[string]any{},
						Type:      "notification",
						Payload:   []byte("Hello, this is a server sent event!"),
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
				)
			}
		}
	})
}
