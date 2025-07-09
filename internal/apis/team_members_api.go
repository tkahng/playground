package apis

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/sse"
	"github.com/tkahng/authgo/internal/contextstore"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/notification"
	"github.com/tkahng/authgo/internal/shared"
)

type MiddlewareFunc func(ctx huma.Context, next func(huma.Context))

func (api *Api) BindTeamMembersSseEvents(humapi huma.API, middlewares ...func(ctx huma.Context, next func(huma.Context))) {
	sse.Register(
		humapi,
		huma.Operation{
			OperationID: "team-members-sse-events",
			Method:      http.MethodGet,
			Path:        "/team-members/sse-events",
			Summary:     "team-members-sse-events",
			Description: "team-members-sse-events",
			Tags:        []string{"Team Members"},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
			Errors: []int{http.StatusInternalServerError, http.StatusBadRequest},
		},
		map[string]any{
			"new_team_member": notification.NotificationPayload[notification.NewTeamMemberNotificationData]{},
		},
		api.TeamMembersSseEvents,
	)

}

func (api *Api) TeamMembersSseEvents(ctx context.Context, input *struct{}, send sse.Sender) {
	teamInfo := contextstore.GetContextTeamInfo(ctx)
	if teamInfo == nil {
		return
	}

	subscription := api.App().Notifier().Subscribe("team_member_id:" + teamInfo.Member.ID.String())
	defer subscription.Unlisten(ctx)

	<-subscription.EstablishedC()

	go func() {

		for {
			select {
			case <-ctx.Done():
				subscription.Unlisten(ctx)
				slog.Debug("Subscription closed")
				return
			case payload := <-subscription.NotificationC():
				var noti models.Notification
				err := json.Unmarshal(payload, &noti)
				if err != nil {
					slog.Error("Failed to unmarshal notification", slog.Any("error", err))
					continue
				}
				var pl notification.NotificationPayload[notification.NewTeamMemberNotificationData]
				err = json.Unmarshal(noti.Payload, &pl)
				if err != nil {
					slog.Error("Failed to unmarshal notification payload", slog.Any("error", err))
					continue
				}
				if err := send.Data(pl); err != nil {
					return
				}
			}
		}
	}()

}
