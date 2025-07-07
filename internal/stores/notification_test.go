package stores_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/stores"
	"github.com/tkahng/authgo/internal/test"
)

func TestNotificationStore_CreateNotification(t *testing.T) {
	test.DbSetup()
	test.WithTx(t, func(ctx context.Context, db database.Dbx) {
		store := stores.NewDbNotificationStore(db)

		notification := &models.Notification{
			Channel:   "test-channel",
			Type:      "test-type",
			CreatedAt: time.Now(),
			Metadata:  map[string]any{"test": "test"},
			Payload:   []byte("{\"key\": \"value\"}"),
		}

		got, err := store.CreateNotification(ctx, notification)
		assert.NoError(t, err)
		assert.NotNil(t, got)
	})
}

func TestNotificationStore_CreateManyNotifications(t *testing.T) {
	test.DbSetup()
	test.WithTx(t, func(ctx context.Context, db database.Dbx) {
		store := stores.NewDbNotificationStore(db)

		notifications := []models.Notification{
			{
				ID:        uuid.New(),
				Channel:   "channel-1",
				Type:      "type-1",
				CreatedAt: time.Now(),
				Metadata:  map[string]any{"test": "test"},
				Payload:   []byte("{\"key\": \"value\"}"),
			},
			{
				ID:        uuid.New(),
				Channel:   "channel-2",
				Type:      "type-2",
				CreatedAt: time.Now(),
				Metadata:  map[string]any{"test": "test"},
				Payload:   []byte("{\"key\": \"value\"}"),
			},
		}

		got, err := store.InsertManyNotifications(ctx, notifications)
		assert.NoError(t, err)
		assert.Equal(t, len(notifications), int(got))

	})
}
