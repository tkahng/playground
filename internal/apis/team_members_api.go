package apis

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/danielgtaylor/huma/v2/sse"
	"github.com/tkahng/authgo/internal/contextstore"
	"github.com/tkahng/authgo/internal/models"
)

func (api *Api) TeamMembersSseEvents(ctx context.Context, input *struct{}, send sse.Sender) {
	ch := make(chan *models.Notification)
	teamInfo := contextstore.GetContextTeamInfo(ctx)
	if teamInfo == nil {
		return
	}

	subscription := api.App().Notifier().Subscribe("team_member_id:" + teamInfo.Member.ID.String())
	defer close(ch)
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
				var notification models.Notification
				json.Unmarshal(payload, &notification)
				ch <- &notification
			}
		}
	}()

}
