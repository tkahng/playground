package stores_test

import (
	"context"
	"reflect"
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

		got, err := store.CreateManyNotifications(ctx, notifications)
		assert.NoError(t, err)
		assert.Equal(t, len(notifications), len(got))
		for i, notification := range notifications {
			assert.Equal(t, notification.Channel, got[i].Channel)
			assert.Equal(t, notification.Type, got[i].Type)
			assert.Equal(t, notification.Metadata, got[i].Metadata)
		}
	})
}

func TestNotificationStore_FindNotification(t *testing.T) {
	test.DbSetup()
	test.WithTx(t, func(ctx context.Context, db database.Dbx) {
		store := stores.NewDbNotificationStore(db)

		notification := &models.Notification{
			ID:        uuid.New(),
			Channel:   "find-channel",
			Type:      "find-type",
			CreatedAt: time.Now(),
			Metadata:  map[string]any{"test": "test"},
			Payload:   []byte("{\"key\": \"value\"}"),
		}
		_, err := store.CreateNotification(ctx, notification)
		assert.NoError(t, err)

		args := &stores.NotificationFilter{
			Channels: []string{notification.Channel},
		}

		got, err := store.FindNotification(ctx, args)
		assert.NoError(t, err)
		assert.NotNil(t, got)
		assert.Equal(t, notification.Channel, got.Channel)
		assert.Equal(t, notification.Type, got.Type)
		assert.Equal(t, notification.Metadata, got.Metadata)
	})
}

func TestNotificationStore_FindNotification2(t *testing.T) {
	test.DbSetup()
	test.WithTx(t, func(ctx context.Context, db database.Dbx) {
		type fields struct {
			db database.Dbx
		}
		type args struct {
			ctx  context.Context
			args *stores.NotificationFilter
		}
		tests := []struct {
			name    string
			fields  fields
			args    args
			want    *models.Notification
			wantErr bool
		}{
			// TODO: Add test cases.
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				s := stores.NewDbNotificationStore(tt.fields.db)
				got, err := s.FindNotification(tt.args.ctx, tt.args.args)
				if (err != nil) != tt.wantErr {
					t.Errorf("PostgresNotificationStore.FindNotification() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("PostgresNotificationStore.FindNotification() = %v, want %v", got, tt.want)
				}
			})
		}
	})
}
