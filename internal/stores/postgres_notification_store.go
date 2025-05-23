package stores

import (
	"context"

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

func (s *PostgresNotificationStore) CreateManyNotifications(ctx context.Context, notifications []*models.Notification) ([]*models.Notification, error) {
	var items []models.Notification
	for _, item := range notifications {
		if item == nil {
			continue
		}
		items = append(items, *item)
	}
	return crudrepo.Notification.Post(
		ctx,
		s.db,
		items,
	)
}

func (s *PostgresNotificationStore) GetNotification(ctx context.Context, notificationId uuid.UUID) (*models.Notification, error) {
	return crudrepo.Notification.GetOne(
		ctx,
		s.db,
		&map[string]any{
			"id": map[string]any{
				"_eq": notificationId.String(),
			},
		},
	)
}
