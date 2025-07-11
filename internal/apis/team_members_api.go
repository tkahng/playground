package apis

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	humasse "github.com/danielgtaylor/huma/v2/sse"
	"github.com/tkahng/authgo/internal/contextstore"
	"github.com/tkahng/authgo/internal/middleware"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/notification"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/sse"
)

func TeamChannel(teamMemberId string) string {
	return "team_member_id:" + teamMemberId
}

type TeamMemberSseInput struct {
	TeamMemberID string `path:"team-member-id"`
	AccessToken  string `query:"access_token"`
}

type MiddlewareFunc func(ctx huma.Context, next func(huma.Context))

func (api *Api) BindTeamMembersSseEvents(humapi huma.API) {
	membermiddleware := middleware.TeamInfoFromTeamMemberID(humapi, api.App())
	hanlder := sse.ServeSSE[TeamMemberSseInput](
		func(ctx context.Context, f func(any) error) sse.Client {
			teamInfo := contextstore.GetContextTeamInfo(ctx)
			return sse.NewClient(TeamChannel(teamInfo.Member.ID.String()), f, slog.Default(), func() any {
				return &PingMessage{
					Message: "ping",
				}
			})
		},
		func(ctx context.Context, cf context.CancelFunc, c sse.Client) {
			api.app.SseManager().RegisterClient(ctx, cf, c)
		},
		func(c sse.Client) {
			fmt.Println("unregistering client")
			api.app.SseManager().UnregisterClient(c)
		},
		60*time.Second,
	)
	humasse.Register(
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
			"new_team_member": &notification.NotificationPayload[notification.NewTeamMemberNotificationData]{},
			"ping":            &PingMessage{},
		},
		// api.TeamMembersSseEvents2,
		hanlder,
	)

}

type PingMessage struct {
	Message string `json:"message"`
}

func (PingMessage) Kind() string {
	return "ping"
}

func (api *Api) TeamMembersSseEvents2(ctx context.Context, input *struct {
	TeamMemberID string `path:"team-member-id"`
	AccessToken  string `query:"access_token"`
}, send humasse.Sender) {
	ctx, cancelRequest := context.WithCancel(ctx)
	defer cancelRequest()
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			// subscription.Unlisten(ctx)
			slog.Debug("Subscription closed")
			return

		case <-ticker.C:
			var pl PingMessage
			pl.Message = "ping"
			if err := send.Data(&pl); err != nil {
				return
			}
		}
	}

}
func (api *Api) TeamMembersSseEvents(ctx context.Context, input *struct {
	TeamMemberID string `path:"team-member-id"`
	AccessToken  string `query:"access_token"`
}, send humasse.Sender) {
	teamInfo := contextstore.GetContextTeamInfo(ctx)
	if teamInfo == nil {
		return
	}
	ctx, cancelRequest := context.WithCancel(ctx)
	defer cancelRequest()
	subscription := api.App().Notifier().Subscribe("team_member_id:" + teamInfo.Member.ID.String())
	defer subscription.Unlisten(ctx)

	fmt.Println("EstablishedC")
	<-subscription.EstablishedC()
	fmt.Println("EstablishedC done")

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	go func() {

		for {
			select {

			case <-ctx.Done():
				subscription.Unlisten(ctx)
				slog.Debug("Subscription closed")
				return

			case <-ticker.C:
				var pl notification.NotificationPayload[PingMessage]
				pl.Data.Message = "ping"
				pl.Notification.Body = "ping"
				pl.Notification.Title = "ping"
				if err := send.Data(pl); err != nil {
					return
				}
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
