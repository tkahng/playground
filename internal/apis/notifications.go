package apis

import (
	"context"

	"github.com/danielgtaylor/huma/v2/sse"
)

func (api *Api) NotificationsSseEvents() map[string]any {
	// sse.
	return map[string]any{
		"notification": "notifications",
	}
}

func (api *Api) NotificationsSsefunc(ctx context.Context, input *struct{},
	send sse.Sender,
) {

}
