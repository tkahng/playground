package apis

import (
	"context"

	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"

	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/contextstore"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/workers"

	"github.com/tkahng/playground/internal/shared"
	"github.com/tkahng/playground/internal/stores"
	"github.com/tkahng/playground/internal/tools/mapper"
	"github.com/tkahng/playground/internal/tools/types"
)

type TeamMember struct {
	_                struct{}       `db:"team_members" json:"-"`
	ID               uuid.UUID      `db:"id" json:"id" format:"uuid"`
	TeamID           uuid.UUID      `db:"team_id" json:"team_id" format:"uuid"`
	UserID           *uuid.UUID     `db:"user_id" json:"user_id" nullable:"true" format:"uuid"`
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

type FindTeamTeamMemberByIDInput struct {
	TeamID       string `path:"team-id" required:"true" format:"uuid"`
	TeamMemberID string `path:"team-member-id" required:"true" format:"uuid"`
}

func (api *Api) bindFindTeamMemberByID(aapi huma.API) {
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
				api.Middlewares().GetRequireTeamInfo(),
			},
		},
		func(ctx context.Context, input *FindTeamTeamMemberByIDInput) (*ApiOutput[*TeamMember], error) {
			teamInfo := contextstore.GetContextTeamInfo(ctx)
			if teamInfo == nil {
				return nil, huma.Error404NotFound("team not found")
			}

			member := contextstore.GetContextTeamMember(ctx)
			if member == nil {
				return nil, huma.Error404NotFound("team member not found")
			}
			if member.TeamID != teamInfo.Team.ID {
				return nil, huma.Error422UnprocessableEntity("team member's team_id does not match team_id in path")
			}
			member.Team = &teamInfo.Team
			if member.UserID != nil {
				user, err := api.App().Adapter().User().FindUserByID(ctx, *member.UserID)
				if err != nil {
					return nil, err
				}
				if user != nil {
					member.User = user
				}
			}

			teamMember := fromTeamMemberModel(member)
			return &ApiOutput[*TeamMember]{
				Body: teamMember,
			}, nil
		},
	)
}

type FindTeamTeamMembersInput struct {
	PaginatedInput
	SortParams
	Roles  []TeamMemberRole          `query:"roles,omitempty" required:"false" minimum:"1" maximum:"100" enum:"owner,member,guest"`
	Q      string                    `query:"q,omitempty" required:"false"`
	TeamID string                    `path:"team-id" required:"true" format:"uuid"`
	Active types.OptionalParam[bool] `query:"active,omitempty" required:"false"`
}

func (api *Api) bindFindTeamTeamMembers(humaApi huma.API) {
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
			Middlewares: huma.Middlewares{
				api.Middlewares().GetRequireTeamInfo(),
			},
		},
		func(ctx context.Context, input *FindTeamTeamMembersInput) (*ApiPaginatedOutput[*TeamMember], error) {
			team := contextstore.GetContextTeam(ctx)
			if team == nil {
				return nil, huma.Error404NotFound("team not found")
			}
			teamID := team.ID
			filter := &stores.TeamMemberFilter{}
			filter.Page = input.Page
			filter.PerPage = input.PerPage
			filter.Active = input.Active
			filter.SortBy = input.SortBy
			filter.SortOrder = input.SortOrder
			filter.TeamIds = []uuid.UUID{teamID}
			filter.Q = input.Q
			filter.Roles = mapper.Map(input.Roles, func(role TeamMemberRole) models.TeamMemberRole {
				return models.TeamMemberRole(role)
			})
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
					member.Team = team
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
		},
	)
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

type UpdateTeamMemberDto struct {
	Role TeamMemberRole `json:"role" enum:"owner,member,guest"`
}
type UpdateTeamsTeamMemberInput struct {
	TeamMemberID string `path:"team-member-id" required:"true" format:"uuid"`
	Body         UpdateTeamMemberDto
}

func (api *Api) UpdateTeamMemberBind(humaApi huma.API) {
	huma.Register(
		humaApi,
		huma.Operation{
			OperationID: "update-team-member",
			Method:      http.MethodPut,
			Path:        "/team-members/{team-member-id}",
			Summary:     "update-team-member",
			Description: "update a team member",
			Tags:        []string{"Team Members"},
			Errors:      []int{http.StatusInternalServerError, http.StatusBadRequest},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
			Middlewares: huma.Middlewares{
				api.Middlewares().GetRequireTeamInfo(),
				api.Middlewares().GetTeamRequiredOwnerMember(),
			},
		},
		func(ctx context.Context, input *UpdateTeamsTeamMemberInput) (*struct{}, error) {
			teamInfo := contextstore.GetContextTeamInfo(ctx)
			if teamInfo == nil {
				return nil, huma.Error401Unauthorized("no team info")
			}
			memberId, err := uuid.Parse(input.TeamMemberID)
			if err != nil {
				return nil, err
			}
			// find the member to be updated
			member, err := api.App().Adapter().TeamMember().FindTeamMember(ctx, &stores.TeamMemberFilter{
				Ids: []uuid.UUID{memberId},
			})
			if err != nil {
				return nil, err
			}
			if member == nil {
				return nil, huma.Error404NotFound("team member not found")
			}
			// check if the member can be deleted
			//
			if !member.Active { // already deleted
				return nil, huma.Error400BadRequest("cannot update inactive user.")
			}
			// update member
			member.Role = models.TeamMemberRole(input.Body.Role)
			_, err = api.App().Adapter().TeamMember().UpdateTeamMember(ctx, member)
			if err != nil {
				return nil, err
			}

			return nil, nil
		},
	)
}

type DeactivateTeamMemberInput struct {
	TeamMemberID string `path:"team-member-id" required:"true" format:"uuid"`
}

func (api *Api) DeactivateTeamMemberBind(humaApi huma.API) {
	huma.Register(
		humaApi,
		huma.Operation{
			OperationID: "delete-team-member",
			Method:      http.MethodDelete,
			Path:        "/team-members/{team-member-id}",
			Summary:     "delete-team-member",
			Description: "delete a team member",
			Tags:        []string{"Team Members"},
			Errors:      []int{http.StatusInternalServerError, http.StatusBadRequest},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
			Middlewares: huma.Middlewares{
				api.Middlewares().GetRequireTeamInfo(),
				api.Middlewares().GetTeamRequiredOwnerMember(),
			},
		},
		func(ctx context.Context, input *DeactivateTeamMemberInput) (*struct{}, error) {
			teamInfo := contextstore.GetContextTeamInfo(ctx)
			if teamInfo == nil {
				return nil, huma.Error401Unauthorized("no team info")
			}
			memberId, err := uuid.Parse(input.TeamMemberID)
			if err != nil {
				return nil, err
			}
			// find the member to be updated
			member, err := api.App().Adapter().TeamMember().FindTeamMember(ctx, &stores.TeamMemberFilter{
				Ids: []uuid.UUID{memberId},
				Active: types.OptionalParam[bool]{
					Value: true, IsSet: true,
				},
			})
			if err != nil {
				return nil, err
			}
			if member == nil {
				return nil, huma.Error404NotFound("team member not found")
			}
			// update member
			txErr := api.App().Adapter().RunInTxCtx(ctx, func(txCtx context.Context) error {
				member.Active = false
				member.UserID = nil
				_, err = api.App().Adapter().TeamMember().UpdateTeamMember(txCtx, member)
				if err != nil {
					return err
				}
				err = api.App().JobService().EnqueueRefreshSubscriptionQuantityJob(txCtx, &workers.RefreshSubscriptionQuantityJobArgs{
					TeamID: member.TeamID,
				})
				if err != nil {
					return err
				}
				return nil
			})
			if txErr != nil {
				return nil, txErr
			}
			return nil, nil
		},
	)
}
