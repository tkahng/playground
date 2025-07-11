package apis

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/sse"
	"github.com/tkahng/authgo/internal/contextstore"
	"github.com/tkahng/authgo/internal/middleware"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/notification"
	"github.com/tkahng/authgo/internal/shared"
)

type MiddlewareFunc func(ctx huma.Context, next func(huma.Context))

func (api *Api) BindTeamMembersSseEvents(humapi huma.API) {
	membermiddleware := middleware.TeamInfoFromTeamMemberID(humapi, api.App())

	sse.Register(
		humapi,
		huma.Operation{
			OperationID: "team-members-sse-team-member-notifications",
			Method:      http.MethodGet,
			Path:        "/team-members/{team-member-id}/notifications/sse",
			Summary:     "team-members-sse-team-member-notifications",
			Description: "team-members-sse-team-member-notifications",
			Tags:        []string{"Team Members"},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
			Middlewares: huma.Middlewares{
				membermiddleware,
			},
			Errors: []int{http.StatusInternalServerError, http.StatusBadRequest},
		},
		map[string]any{
			"new_team_member": notification.NotificationPayload[notification.NewTeamMemberNotificationData]{},
		},
		api.TeamMembersSseEvents2,
	)

}
func (api *Api) TeamMembersSseEvents2(ctx context.Context, input *struct{}, send sse.Sender) {
	teamInfo := contextstore.GetContextTeamInfo(ctx)
	if teamInfo == nil {
		return
	}
	ctx, cancelRequest := context.WithCancel(ctx)
	defer cancelRequest()
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
			case payload, ok := <-subscription.NotificationC():
				if !ok {
					slog.Debug("Subscription closed")
					subscription.Unlisten(ctx)
					return
				}
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
func (api *Api) TeamMembersSseEvents(ctx context.Context, input *struct{}, send sse.Sender) {
	teamInfo := contextstore.GetContextTeamInfo(ctx)
	if teamInfo == nil {
		return
	}
	ctx, cancelRequest := context.WithCancel(ctx)
	defer cancelRequest()
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
			case payload, ok := <-subscription.NotificationC():
				if !ok {
					slog.Debug("Subscription closed")
					subscription.Unlisten(ctx)
					return
				}
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
