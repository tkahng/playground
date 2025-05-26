package stores

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/crudrepo"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/models"
)

type PostgresNotificationStore struct {
	db database.Dbx
}

func NewPostgresNotificationStore(db database.Dbx) *PostgresNotificationStore {
	return &PostgresNotificationStore{
		db: db,
	}
}

func (s *PostgresNotificationStore) CreateNotification(ctx context.Context, notification *models.Notification) (*models.Notification, error) {
	return crudrepo.Notification.PostOne(
		ctx,
		s.db,
		notification,
	)
}

func (s *PostgresNotificationStore) CreateManyNotifications(ctx context.Context, notifications []models.Notification) ([]*models.Notification, error) {
	return crudrepo.Notification.Post(
		ctx,
		s.db,
		notifications,
	)
}

func (s *PostgresNotificationStore) FindNotification(ctx context.Context, args *models.Notification) (*models.Notification, error) {
	if args == nil {
		return nil, nil
	}
	where := map[string]any{}
	if args.ID != uuid.Nil {
		where["id"] = map[string]any{
			"_eq": args.ID.String(),
		}
	}
	if args.UserID != nil {
		where["user_id"] = map[string]any{
			"_eq": args.UserID.String(),
		}
	}
	if args.TeamID != nil {
		where["team_id"] = map[string]any{
			"_eq": args.TeamID.String(),
		}
	}
	if args.TeamMemberID != nil {
		where["team_member_id"] = map[string]any{
			"_eq": args.TeamMemberID.String(),
		}
	}
	if args.Type != "" {
		where["type"] = map[string]any{
			"_eq": args.Type,
		}
	}
	if args.Channel != "" {
		where["channel"] = map[string]any{
			"_eq": args.Channel,
		}
	}
	if args.ReadAt != nil {
		where["read_at"] = map[string]any{
			"_gte": args.ReadAt.Format(time.RFC3339Nano),
		}
	}

	return crudrepo.Notification.GetOne(
		ctx,
		s.db,
		&where,
	)
}
