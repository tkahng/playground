package stores

import (
	"context"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/crudrepo"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/models"
)

type DbNotificationStore struct {
	db database.Dbx
}

func NewDbNotificationStore(db database.Dbx) *DbNotificationStore {
	return &DbNotificationStore{
		db: db,
	}
}

func (s *DbNotificationStore) CreateNotification(ctx context.Context, notification *models.Notification) (*models.Notification, error) {
	return crudrepo.Notification.PostOne(
		ctx,
		s.db,
		notification,
	)
}

func (s *DbNotificationStore) CreateManyNotifications(ctx context.Context, notifications []models.Notification) ([]*models.Notification, error) {
	return crudrepo.Notification.Post(
		ctx,
		s.db,
		notifications,
	)
}

func (s *DbNotificationStore) FindNotification(ctx context.Context, args *models.Notification) (*models.Notification, error) {
	if args == nil {
		return nil, nil
	}
	where := map[string]any{}
	if args.ID != uuid.Nil {
		where["id"] = map[string]any{
			"_eq": args.ID,
		}
	}
	if args.UserID != nil {
		where["user_id"] = map[string]any{
			"_eq": args.UserID,
		}
	}
	if args.TeamID != nil {
		where["team_id"] = map[string]any{
			"_eq": args.TeamID,
		}
	}
	if args.TeamMemberID != nil {
		where["team_member_id"] = map[string]any{
			"_eq": args.TeamMemberID,
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
			"_gte": args.ReadAt,
		}
	}

	return crudrepo.Notification.GetOne(
		ctx,
		s.db,
		&where,
	)
}
