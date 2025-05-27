package apis

import (
	"context"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/sse"
	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/shared"
)

func BindSSE(api huma.API, appApi *Api) {
	// Register the SSE endpoint
	sse.Register(api, huma.Operation{
		OperationID: "sse",
		Method:      http.MethodGet,
		Path:        "/sse",
		Summary:     "Server sent events example",
	}, map[string]any{
		// Mapping of event type name to Go struct for that event.
		"notification": shared.Notification{},
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
					shared.Notification{
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
