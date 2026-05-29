package services_test

// Tests for DbPlayerNotifier:
//
//   Notify persists a notification row for the player's user and sends SSE.
//   When the player does not exist, the notification is skipped but SSE still fires.
//   Notification type is stored verbatim in the type column.

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/notification"
	"github.com/tkahng/playground/internal/services"
	"github.com/tkahng/playground/internal/stores"
	"github.com/tkahng/playground/internal/tools/sse"
)

func setupPlayerNotifyFixture(t *testing.T, ctx context.Context, db database.Dbx) (
	adapter stores.StorageAdapterInterface,
	player *models.Player,
) {
	t.Helper()
	adapter = stores.NewStorageAdapter(db)

	user, err := adapter.User().CreateUser(ctx, &models.User{Email: "player-notif@example.com"})
	require.NoError(t, err)

	player, err = adapter.Gaming().CreatePlayer(ctx, &models.Player{
		Email:  user.Email,
		UserID: &user.ID,
	})
	require.NoError(t, err)
	return
}

// recordingSSE captures channels and payloads sent via the SSE manager.
type recordingSSE struct {
	noopSSE
	sent []struct {
		channel string
		payload any
	}
}

func (r *recordingSSE) Send(channel string, payload any) error {
	r.sent = append(r.sent, struct {
		channel string
		payload any
	}{channel, payload})
	return nil
}

// TestPlayerNotifier_PersistsAndSendsSSE verifies that Notify creates a
// notification row for the player's user_id and sends to the player channel.
func TestPlayerNotifier_PersistsAndSendsSSE(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter, player := setupPlayerNotifyFixture(t, ctx, db)

		recorder := &recordingSSE{}
		notifier := services.NewDbPlayerNotifier(recorder, adapter)

		payload := notification.NewNotificationPayload(
			"New challenge",
			"Someone challenged you to RPS",
			notification.RpsGameChallengedData{
				GameID:             uuid.New(),
				RequestingPlayerID: uuid.New(),
				RequestingEmail:    "challenger@example.com",
			},
		)
		require.NoError(t, notifier.Notify(ctx, player.ID, "rps_game_challenged", payload))

		notifs, err := adapter.Notification().FindNotifications(ctx, &stores.NotificationFilter{
			UserIds: []uuid.UUID{*player.UserID},
			Types:   []string{"rps_game_challenged"},
		})
		require.NoError(t, err)
		require.Len(t, notifs, 1, "exactly one notification row should be created")

		assert.Equal(t, sse.PlayerChannel(player.ID.String()), notifs[0].Channel)
		assert.Equal(t, "rps_game_challenged", notifs[0].Type)

		var stored notification.NotificationPayload[notification.RpsGameChallengedData]
		require.NoError(t, json.Unmarshal(notifs[0].Payload, &stored))
		assert.Equal(t, "New challenge", stored.Notification.Title)

		require.Len(t, recorder.sent, 1)
		assert.Equal(t, sse.PlayerChannel(player.ID.String()), recorder.sent[0].channel)
	})
}

// TestPlayerNotifier_UnknownPlayerSkipsPersist verifies that when the player
// does not exist, no notification row is created but SSE is still attempted.
func TestPlayerNotifier_UnknownPlayerSkipsPersist(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter, _ := setupPlayerNotifyFixture(t, ctx, db)

		recorder := &recordingSSE{}
		notifier := services.NewDbPlayerNotifier(recorder, adapter)

		unknownID := uuid.New()
		payload := notification.NewNotificationPayload(
			"Test", "test",
			notification.RpsGameChallengedData{GameID: uuid.New(), RequestingPlayerID: uuid.New(), RequestingEmail: "x@y.com"},
		)
		require.NoError(t, notifier.Notify(ctx, unknownID, "rps_game_challenged", payload))

		notifs, err := adapter.Notification().FindNotifications(ctx, &stores.NotificationFilter{
			Channels: []string{sse.PlayerChannel(unknownID.String())},
		})
		require.NoError(t, err)
		assert.Empty(t, notifs, "no notification row when player is not found")

		assert.Len(t, recorder.sent, 1, "SSE should still fire even without a DB row")
	})
}

// TestPlayerNotifier_NotificationTypeStoredVerbatim verifies that the notifType
// argument is stored as-is in the type column.
func TestPlayerNotifier_NotificationTypeStoredVerbatim(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter, player := setupPlayerNotifyFixture(t, ctx, db)

		notifier := services.NewDbPlayerNotifier(noopSSE{}, adapter)

		friendPayload := notification.NewNotificationPayload(
			"Friend request", "someone wants to be friends",
			notification.FriendRequestNotificationData{
				RequestingPlayerID: uuid.New(),
				RequestingEmail:    "friend@example.com",
				FriendshipID:       uuid.New(),
			},
		)
		require.NoError(t, notifier.Notify(ctx, player.ID, "friend_request", friendPayload))

		notif, err := adapter.Notification().FindNotification(ctx, &stores.NotificationFilter{
			UserIds: []uuid.UUID{*player.UserID},
			Types:   []string{"friend_request"},
		})
		require.NoError(t, err)
		require.NotNil(t, notif)
		assert.Equal(t, "friend_request", notif.Type)
	})
}
