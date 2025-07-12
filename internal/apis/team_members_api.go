package apis

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	humasse "github.com/danielgtaylor/huma/v2/sse"
	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/contextstore"
	"github.com/tkahng/authgo/internal/middleware"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/notification"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/stores"
	"github.com/tkahng/authgo/internal/tools/mapper"
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
		1*time.Second,
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
			"new_team_member":  &notification.NotificationPayload[notification.NewTeamMemberNotificationData]{},
			"assigned_to_task": &notification.NotificationPayload[notification.AssignedToTaskNotificationData]{},
			"ping":             &PingMessage{},
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

type TeamMembersNotificationsInput struct {
	PaginatedInput
	SortParams
	TeamMemberID string `path:"team-member-id" required:"true" format:"uuid"`
}

type Notification struct {
	_            struct{}       `db:"notifications" json:"-"`
	ID           uuid.UUID      `db:"id,pk" json:"id"`
	CreatedAt    time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time      `db:"updated_at" json:"updated_at"`
	ReadAt       *time.Time     `db:"read_at" json:"read_at,omitempty"`
	Channel      string         `db:"channel" json:"channel"`
	Payload      string         `db:"payload" json:"payload"`
	UserID       *uuid.UUID     `db:"user_id" json:"user_id,omitempty"`
	TeamMemberID *uuid.UUID     `db:"team_member_id" json:"team_member_id,omitempty"`
	TeamID       *uuid.UUID     `db:"team_id" json:"team_id,omitempty"`
	Metadata     map[string]any `db:"metadata" json:"metadata"`
	Type         string         `db:"type" json:"type"`
	User         *ApiUser       `db:"user" src:"user_id" dest:"id" table:"users" json:"user,omitempty"`
	TeamMember   *TeamMember    `db:"team_member" src:"team_member_id" dest:"id" table:"team_members" json:"team_member,omitempty"`
	Team         *Team          `db:"team" src:"team_id" dest:"id" table:"teams" json:"team,omitempty"`
}

func FromModelNotification(notification *models.Notification) *Notification {
	return &Notification{
		ID:           notification.ID,
		CreatedAt:    notification.CreatedAt,
		UpdatedAt:    notification.UpdatedAt,
		ReadAt:       notification.ReadAt,
		Channel:      notification.Channel,
		UserID:       notification.UserID,
		Payload:      string(notification.Payload),
		TeamMemberID: notification.TeamMemberID,
		TeamID:       notification.TeamID,
		Metadata:     notification.Metadata,
		Type:         notification.Type,
		User:         FromUserModel(notification.User),
		TeamMember:   FromTeamMemberModel(notification.TeamMember),
		Team:         FromTeamModel(notification.Team),
	}
}
func (api *Api) BindFindTeamMembersNotifications(aapi huma.API) {
	teamInfoFromMember := middleware.TeamInfoFromTeamMemberID(aapi, api.app)
	huma.Register(
		aapi,
		huma.Operation{
			OperationID: "find-team-members-notifications",
			Method:      http.MethodGet,
			Path:        "/team-members/{team-member-id}/notifications",
			Summary:     "find-team-members-notifications",
			Description: "find team members notifications",
			Tags:        []string{"Team Members"},
			Errors:      []int{http.StatusInternalServerError, http.StatusBadRequest},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
			Middlewares: huma.Middlewares{
				teamInfoFromMember,
			},
		},
		func(ctx context.Context, input *TeamMembersNotificationsInput) (*ApiPaginatedOutput[*Notification], error) {
			teamInfo := contextstore.GetContextTeamInfo(ctx)
			if teamInfo == nil {
				return nil, huma.Error401Unauthorized("no team info")
			}
			filter := &stores.NotificationFilter{
				TeamMemberIds: []uuid.UUID{
					teamInfo.Member.ID,
				},
			}
			filter.Page = input.Page
			filter.PerPage = input.PerPage
			filter.SortBy = input.SortBy
			filter.SortOrder = input.SortOrder
			notifications, err := api.App().Adapter().Notification().FindNotifications(ctx, filter)
			if err != nil {
				return nil, err
			}
			count, err := api.App().Adapter().Notification().CountNotification(ctx, filter)
			if err != nil {
				return nil, err
			}
			return &ApiPaginatedOutput[*Notification]{
				Body: ApiPaginatedResponse[*Notification]{
					Meta: ApiGenerateMeta(&input.PaginatedInput, count),
					Data: mapper.Map(notifications, FromModelNotification),
				},
			}, nil
		},
	)
}

type ReadTeamMembersNotificationsInput struct {
	NotificationID string `path:"notification-id" required:"true" format:"uuid"`
	TeamMemberID   string `path:"team-member-id" required:"true" format:"uuid"`
}

func (api *Api) BindReadTeamMembersNotifications(aapi huma.API) {
	teamMemberMiddleware := middleware.TeamInfoFromTeamMemberID(aapi, api.app)
	huma.Register(
		aapi,
		huma.Operation{
			OperationID: "read-team-members-notifications",
			Method:      http.MethodPost,
			Path:        "/team-members/{team-member-id}/notifications/{notification-id}/read",
			Summary:     "read-team-members-notifications",
			Description: "read team members notifications",
			Tags:        []string{"Team Members"},
			Errors:      []int{http.StatusInternalServerError, http.StatusBadRequest},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
			Middlewares: huma.Middlewares{
				teamMemberMiddleware,
			},
		},
		func(ctx context.Context, input *ReadTeamMembersNotificationsInput) (*struct{}, error) {
			teamInfo := contextstore.GetContextTeamInfo(ctx)
			if teamInfo == nil {
				return nil, huma.Error401Unauthorized("unauthorized")
			}
			notificationID, err := uuid.Parse(input.NotificationID)
			if err != nil {
				return nil, err
			}
			notification, err := api.App().Adapter().Notification().FindNotification(ctx, &stores.NotificationFilter{
				Ids: []uuid.UUID{
					notificationID,
				},
				TeamMemberIds: []uuid.UUID{
					teamInfo.Member.ID,
				},
			})
			if err != nil {
				return nil, err
			}
			now := time.Now()
			notification.ReadAt = &now
			err = api.App().Adapter().Notification().UpdateNotification(ctx, notification)
			if err != nil {
				return nil, err
			}

			return nil, nil
		},
	)
}

func (api *Api) BindDeleteTeamMembersNotifications(aapi huma.API) {
	teamMemberMiddleware := middleware.TeamInfoFromTeamMemberID(aapi, api.app)
	huma.Register(
		aapi,
		huma.Operation{
			OperationID: "delete-team-members-notifications",
			Method:      http.MethodDelete,
			Path:        "/team-members/{team-member-id}/notifications/{notification-id}",
			Summary:     "delete-team-members-notifications",
			Description: "delete team members notifications",
			Tags:        []string{"Team Members"},
			Errors:      []int{http.StatusInternalServerError, http.StatusBadRequest},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
			Middlewares: huma.Middlewares{
				teamMemberMiddleware,
			},
		},
		func(ctx context.Context, input *ReadTeamMembersNotificationsInput) (*struct{}, error) {
			teamInfo := contextstore.GetContextTeamInfo(ctx)
			if teamInfo == nil {
				return nil, huma.Error401Unauthorized("unauthorized")
			}
			notificationID, err := uuid.Parse(input.NotificationID)
			if err != nil {
				return nil, err
			}
			_, err = api.App().Adapter().Notification().DeleteNotifications(ctx, &stores.NotificationFilter{
				Ids: []uuid.UUID{
					notificationID,
				},
				TeamMemberIds: []uuid.UUID{
					teamInfo.Member.ID,
				},
			})
			if err != nil {
				return nil, err
			}
			return nil, nil
		},
	)
}
