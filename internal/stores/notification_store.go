package stores

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/tools/types"
)

type NotificationFilter struct {
	Ids           []uuid.UUID                    `query:"ids" json:"ids,omitempty" required:"false" minimum:"1" maximum:"100" uniqueItems:"true"`
	UserIds       []uuid.UUID                    `query:"user_ids" json:"user_ids,omitempty" required:"false" minimum:"1" maximum:"100" uniqueItems:"true"`
	TeamIds       []uuid.UUID                    `query:"team_ids" json:"team_ids,omitempty" required:"false" minimum:"1" maximum:"100" uniqueItems:"true"`
	TeamMemberIds []uuid.UUID                    `query:"team_member_ids" json:"team_member_ids,omitempty" required:"false" minimum:"1" maximum:"100" uniqueItems:"true"`
	Channels      []string                       `query:"channels" json:"channels,omitempty" required:"false" minimum:"1" maximum:"100" uniqueItems:"true"`
	Types         []string                       `query:"types" json:"types,omitempty" required:"false" minimum:"1" maximum:"100" uniqueItems:"true"`
	ReadAt        types.OptionalParam[time.Time] `query:"read_at" json:"read_at" required:"false"`
}

type NotificationStore interface {
	CreateNotification(ctx context.Context, notification *models.Notification) (*models.Notification, error)
	CreateManyNotifications(ctx context.Context, notifications []*models.Notification) ([]*models.Notification, error)
	FindNotification(ctx context.Context, args *models.Notification) (*models.Notification, error)
}
