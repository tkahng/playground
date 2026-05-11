package notification_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/notification"
)

func TestRpsGameCancelledData_Kind(t *testing.T) {
	d := notification.RpsGameCancelledData{
		GameID:             uuid.New(),
		CancellingPlayerID: uuid.New(),
	}
	if d.Kind() != "rps_game_cancelled" {
		t.Errorf("Kind() = %q, want %q", d.Kind(), "rps_game_cancelled")
	}
}
