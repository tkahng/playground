package apis

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	humasse "github.com/danielgtaylor/huma/v2/sse"
	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/contextstore"
	"github.com/tkahng/playground/internal/middleware/humamiddleware"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/notification"
	"github.com/tkahng/playground/internal/shared"
	"github.com/tkahng/playground/internal/stores"
	"github.com/tkahng/playground/internal/tools/mapper"
	"github.com/tkahng/playground/internal/tools/sse"
)

type TeamMember struct {
	_                struct{}       `db:"team_members" json:"-"`
	ID               uuid.UUID      `db:"id" json:"id"`
	TeamID           uuid.UUID      `db:"team_id" json:"team_id"`
	UserID           *uuid.UUID     `db:"user_id" json:"user_id"`
	Active           bool           `db:"active" json:"active"`
	Role             TeamMemberRole `db:"role" json:"role" enum:"owner,member,guest"`
	HasBillingAccess bool           `db:"has_billing_access" json:"has_billing_access"`
	LastSelectedAt   time.Time      `db:"last_selected_at" json:"last_selected_at"`
	CreatedAt        time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time      `db:"updated_at" json:"updated_at"`
	Team             *Team          `db:"team" src:"team_id" dest:"id" table:"team" json:"team,omitempty"`
	User             *ApiUser       `db:"user" src:"user_id" dest:"id" table:"users" json:"user,omitempty"`
}
type TeamMemberOutput struct {
	Body *TeamMember `json:"body"`
}

func (api *Api) GetActiveTeamMember(
	ctx context.Context,
	input *struct{},
) (
	*TeamMemberOutput,
	error,
) {
	info := contextstore.GetContextUserInfo(ctx)
	if info == nil {
		return nil, huma.Error401Unauthorized("unauthorized")
	}
	team, err := api.App().Team().GetActiveTeamMember(ctx, info.User.ID)
	if err != nil {
		return nil, err
	}
	if team == nil {
		return nil, huma.Error404NotFound("team not found")
	}
	return &TeamMemberOutput{
		Body: fromTeamMemberModel(team),
	}, nil
}
func TeamChannel(teamMemberId string) string {
	return "team_member_id:" + teamMemberId
}

type TeamMemberSseInput struct {
	TeamMemberID string `path:"team-member-id"`
	AccessToken  string `query:"access_token"`
}

type MiddlewareFunc func(ctx huma.Context, next func(huma.Context))

func (api *Api) bindTeamMembersSseEvents(humapi huma.API) {
	membermiddleware := humamiddleware.TeamInfoFromTeamMemberID(humapi, api.App())
	hanlder := sse.ServeSSE(
		func(ctx context.Context, f func(any) error, input *TeamMemberSseInput) sse.Client {
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
			api.app.SseManager().UnregisterClient(c)
		},
		30*time.Second,
	)
	humasse.Register(
		humapi,
		huma.Operation{
			OperationID: "team-members-sse-team-member-notifications",
			Method:      http.MethodGet,
			Path:        "/team-members/{team-member-id}/sse",
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
			"task_completed":   &notification.NotificationPayload[notification.TaskCompletedNotificationData]{},
			"task_due_today":   &notification.NotificationPayload[notification.TaskDueTodayNotificationData]{},
			"new_team_member":  &notification.NotificationPayload[notification.NewTeamMemberNotificationData]{},
			"assigned_to_task": &notification.NotificationPayload[notification.AssignedToTaskNotificationData]{},
			"ping":             &PingMessage{},
		},
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

func fromModelNotification(notification *models.Notification) *Notification {
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
		User:         fromUserModel(notification.User),
		TeamMember:   fromTeamMemberModel(notification.TeamMember),
		Team:         fromTeamModel(notification.Team),
	}
}
func (api *Api) bindFindTeamMembersNotifications(aapi huma.API) {
	teamInfoFromMember := humamiddleware.TeamInfoFromTeamMemberID(aapi, api.app)
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
					Data: mapper.Map(notifications, fromModelNotification),
				},
			}, nil
		},
	)
}

type ReadTeamMembersNotificationsInput struct {
	NotificationID string `path:"notification-id" required:"true" format:"uuid"`
	TeamMemberID   string `path:"team-member-id" required:"true" format:"uuid"`
}

func (api *Api) bindReadTeamMembersNotifications(aapi huma.API) {
	teamMemberMiddleware := humamiddleware.TeamInfoFromTeamMemberID(aapi, api.app)
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

func (api *Api) bindDeleteTeamMembersNotifications(aapi huma.API) {
	teamMemberMiddleware := humamiddleware.TeamInfoFromTeamMemberID(aapi, api.app)
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

type FindTeamTeamMemberByIDInput struct {
	TeamID       string `path:"team-id" required:"true" format:"uuid"`
	TeamMemberID string `path:"team-member-id" required:"true" format:"uuid"`
}

func (api *Api) bindFindTeamMemberByID(aapi huma.API) {
	middleware := humamiddleware.TeamInfoFromParam(aapi, api.app)
	huma.Register(
		aapi,
		huma.Operation{
			OperationID: "find-team-team-member-by-id",
			Method:      http.MethodGet,
			Path:        "/teams/{team-id}/team-members/{team-member-id}",
			Summary:     "find-team-team-member-by-id",
			Description: "find team team member by id",
			Tags:        []string{"Team Members"},
			Errors:      []int{http.StatusInternalServerError, http.StatusBadRequest},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
			Middlewares: huma.Middlewares{
				middleware,
			},
		},
		func(ctx context.Context, input *FindTeamTeamMemberByIDInput) (*ApiOutput[*TeamMember], error) {
			teamInfo := contextstore.GetContextTeamInfo(ctx)
			if teamInfo == nil {
				return nil, huma.Error401Unauthorized("no team info")
			}
			memberId, err := uuid.Parse(input.TeamMemberID)
			if err != nil {
				return nil, err
			}
			otherTeamInfo, err := api.App().Team().FindTeamInfoByMemberID(
				ctx,
				memberId,
			)
			if err != nil {
				return nil, err
			}
			teamMember := fromTeamMemberModel(&otherTeamInfo.Member)
			teamMember.Team = fromTeamModel(&otherTeamInfo.Team)
			teamMember.User = fromUserModel(&otherTeamInfo.User)
			return &ApiOutput[*TeamMember]{
				Body: teamMember,
			}, nil
		},
	)
}

func (api *Api) bindFindTeamTeamMembers(
	humaApi huma.API,
) {
	huma.Register(
		humaApi,
		huma.Operation{
			OperationID: "get-team-team-members",
			Method:      http.MethodGet,
			Path:        "/teams/{team-id}/members",
			Summary:     "get-team-team-members",
			Description: "get members of a team by team team ID",
			Tags:        []string{"Teams", "Team Members"},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
			Errors: []int{http.StatusInternalServerError, http.StatusBadRequest},
		},
		api.FindTeamTeamMembers,
	)
}

type FindTeamTeamMembersInput struct {
	PaginatedInput
	SortParams
	Q      string `query:"q,omitempty" required:"false"`
	TeamID string `path:"team-id" required:"true" format:"uuid"`
}

func (api *Api) FindTeamTeamMembers(
	ctx context.Context,
	input *FindTeamTeamMembersInput,
) (
	*ApiPaginatedOutput[*TeamMember],
	error,
) {
	teamID, err := uuid.Parse(input.TeamID)
	if err != nil {
		return nil, huma.Error400BadRequest("invalid team ID")
	}
	info := contextstore.GetContextUserInfo(ctx)
	if info == nil {
		return nil, huma.Error401Unauthorized("unauthorized")
	}
	filter := &stores.TeamMemberFilter{}
	filter.Page = input.Page
	filter.PerPage = input.PerPage
	filter.SortBy = input.SortBy
	filter.SortOrder = input.SortOrder
	filter.TeamIds = []uuid.UUID{teamID}
	filter.Q = input.Q
	members, err := api.App().Adapter().TeamMember().FindTeamMembers(ctx, filter)
	if err != nil {
		return nil, err
	}
	if len(members) > 0 {
		userIds := make([]uuid.UUID, len(members))
		for idx, member := range members {
			if member == nil {
				continue
			}
			if member.UserID == nil {
				continue
			}
			userIds[idx] = *member.UserID
		}
		users, err := api.App().Adapter().User().LoadUsersByUserIds(ctx, userIds...)
		if err != nil {
			return nil, err
		}
		for idx := range userIds {
			member := members[idx]
			if member == nil {
				continue
			}
			user := users[idx]
			if user == nil {
				continue
			}
			member.User = user
		}

	}
	count, err := api.App().Adapter().TeamMember().CountTeamMembers(ctx, filter)
	if err != nil {
		return nil, err
	}
	return &ApiPaginatedOutput[*TeamMember]{
		Body: ApiPaginatedResponse[*TeamMember]{
			Data: mapper.Map(members, fromTeamMemberModel),
			Meta: ApiGenerateMeta(&input.PaginatedInput, count),
		},
	}, nil
}

func fromTeamMemberModel(member *models.TeamMember) *TeamMember {
	if member == nil {
		return nil
	}
	return &TeamMember{
		ID:               member.ID,
		TeamID:           member.TeamID,
		UserID:           member.UserID,
		Active:           member.Active,
		Role:             TeamMemberRole(member.Role),
		HasBillingAccess: member.HasBillingAccess,
		LastSelectedAt:   member.LastSelectedAt,
		CreatedAt:        member.CreatedAt,
		UpdatedAt:        member.UpdatedAt,
		Team:             fromTeamModel(member.Team),
		User:             fromUserModel(member.User),
	}
}
