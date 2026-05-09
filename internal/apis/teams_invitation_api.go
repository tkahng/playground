package apis

import (
	"context"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/contextstore"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/stores"
	"github.com/tkahng/playground/internal/tools/mapper"
)

type InviteTeamMemberDto struct {
	Email string `json:"email" required:"true"`
	Role  string `json:"role" required:"true" enum:"owner,member,guest"`
}
type InviteTeamMemberInput struct {
	TeamID string              `path:"team-id" required:"true" format:"uuid"`
	Body   InviteTeamMemberDto `json:"body" required:"true"`
}

func (api *Api) CreateInvitation(ctx context.Context, input *InviteTeamMemberInput) (*struct{}, error) {
	userInfo := contextstore.GetContextUserInfo(ctx)
	if userInfo == nil {
		return nil, huma.Error401Unauthorized("Unauthorized")
	}
	parsedTeamId, err := uuid.Parse(input.TeamID)
	if err != nil {
		return nil, err
	}

	err = api.App().TeamInvitation().CreateInvitation(
		ctx,
		parsedTeamId,
		userInfo.User.ID,
		input.Body.Email,
		models.TeamMemberRole(input.Body.Role),
		true,
	)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

type CheckValidInvitationDto struct {
	Token string `json:"token" required:"true"`
}
type CheckValidInvitationInput struct {
	Body CheckValidInvitationDto
}

func (api *Api) CheckValidInvitation(ctx context.Context, input *CheckValidInvitationInput) (*struct{}, error) {
	// userInfo := contextstore.GetContextUserInfo(ctx)
	// if userInfo == nil {
	// 	return nil, huma.Error401Unauthorized("Unauthorized. No user info")
	// }
	res, err := api.App().TeamInvitation().GetInvitation(
		ctx,
		// userInfo.User.ID,
		input.Body.Token,
	)
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, huma.Error400BadRequest("Invalid invitation")
	}
	return nil, nil
}

func (api *Api) AcceptInvitation(ctx context.Context, input *CheckValidInvitationInput) (*struct{}, error) {
	userInfo := contextstore.GetContextUserInfo(ctx)
	if userInfo == nil {
		return nil, huma.Error401Unauthorized("Unauthorized. No user info")
	}
	err := api.App().TeamInvitation().AcceptInvitation(
		ctx,
		userInfo.User.ID,
		input.Body.Token,
	)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (api *Api) DeclineInvitation(ctx context.Context, input *CheckValidInvitationInput) (*struct{}, error) {
	userInfo := contextstore.GetContextUserInfo(ctx)
	if userInfo == nil {
		return nil, huma.Error401Unauthorized("Unauthorized. No user info")
	}
	err := api.App().TeamInvitation().RejectInvitation(
		ctx,
		userInfo.User.ID,
		input.Body.Token,
	)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

type CancelInvitationInput struct {
	TeamID       string `path:"team-id" required:"true" format:"uuid"`
	InvitationID string `path:"invitation-id" required:"true" format:"uuid"`
}

func (api *Api) CencelInvitation(ctx context.Context, input *CancelInvitationInput) (*struct{}, error) {
	userInfo := contextstore.GetContextUserInfo(ctx)
	if userInfo == nil {
		return nil, huma.Error401Unauthorized("Unauthorized. No user info")
	}
	teamInfoTx := contextstore.GetContextTeamInfo(ctx)
	if teamInfoTx == nil {
		return nil, huma.Error401Unauthorized("Unauthorized. No team info")
	}
	parsedTeamId, err := uuid.Parse(input.TeamID)
	if err != nil {
		return nil, err
	}
	parsedInvitationId, err := uuid.Parse(input.InvitationID)
	if err != nil {
		return nil, err
	}
	err = api.App().TeamInvitation().CancelInvitation(
		ctx,
		parsedTeamId,
		teamInfoTx.User.ID,
		parsedInvitationId,
	)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

type FindInvitationsInput struct {
	PaginatedInput
	SortParams
	TeamID   string                 `path:"team-id" required:"true" format:"uuid"`
	Statuses []TeamInvitationStatus `query:"statuses,omitempty" required:"false" minimum:"1" maximum:"10" enum:"pending,accepted,declined,cancelled"`
}
type TeamInvitationStatus string

const (
	TeamInvitationStatusPending  TeamInvitationStatus = "pending"
	TeamInvitationStatusAccepted TeamInvitationStatus = "accepted"
	TeamInvitationStatusDeclined TeamInvitationStatus = "declined"
	TeamInvitationStatusCanceled TeamInvitationStatus = "cancelled"
)

type TeamInvitation struct {
	_               struct{}             `db:"team_invitations" json:"-"`
	ID              uuid.UUID            `db:"id" json:"id"`
	TeamID          uuid.UUID            `db:"team_id" json:"team_id"`
	InviterMemberID uuid.UUID            `db:"inviter_member_id" json:"inviter_member_id"`
	Email           string               `db:"email" json:"email"`
	Role            TeamMemberRole       `db:"role" json:"role"`
	Token           string               `db:"token" json:"token"`
	Status          TeamInvitationStatus `db:"status" json:"status" enum:"pending,accepted,declined,cancelled"`
	ExpiresAt       time.Time            `db:"expires_at" json:"expires_at"`
	CreatedAt       time.Time            `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time            `db:"updated_at" json:"updated_at"`
	Team            *Team                `db:"team" src:"team_id" dest:"id" table:"teams" json:"team,omitempty"`
	InviterMember   *TeamMember          `db:"inviter_member" src:"inviter_member_id" dest:"id" table:"member" json:"inviter_member,omitempty"`
}

func fromTeamInvitationModel(team *models.TeamInvitation) *TeamInvitation {
	if team == nil {
		return nil
	}
	return &TeamInvitation{
		ID:              team.ID,
		TeamID:          team.TeamID,
		InviterMemberID: team.InviterMemberID,
		Email:           team.Email,
		Role:            TeamMemberRole(team.Role),
		Token:           team.Token,
		Status:          TeamInvitationStatus(team.Status),
		ExpiresAt:       team.ExpiresAt,
		CreatedAt:       team.CreatedAt,
		UpdatedAt:       team.UpdatedAt,
		Team:            fromTeamModel(team.Team),
		InviterMember:   fromTeamMemberModel(team.InviterMember),
	}
}
func (api *Api) FindInvitations(ctx context.Context, input *FindInvitationsInput) (*ApiPaginatedOutput[*TeamInvitation], error) {
	userInfo := contextstore.GetContextUserInfo(ctx)
	if userInfo == nil {
		return nil, huma.Error401Unauthorized("Unauthorized. No user info")
	}
	parsedTeamId, err := uuid.Parse(input.TeamID)
	if err != nil {
		return nil, err
	}
	filter := &stores.TeamInvitationFilter{}
	filter.Page = input.Page
	filter.PerPage = input.PerPage
	filter.SortBy = input.SortBy
	filter.SortOrder = input.SortOrder
	filter.TeamIds = []uuid.UUID{parsedTeamId}
	filter.Statuses = mapper.Map(input.Statuses, func(s TeamInvitationStatus) models.TeamInvitationStatus {
		return models.TeamInvitationStatus(s)
	})
	invitations, err := api.App().Adapter().TeamInvitation().FindTeamInvitations(
		ctx,
		filter,
	)
	if err != nil {
		return nil, err
	}
	if len(invitations) > 0 {
		err := api.LoadTeamInvitationRelations(ctx, invitations)
		if err != nil {
			return nil, err
		}
	}
	count, err := api.App().Adapter().TeamInvitation().CountTeamInvitations(
		ctx,
		filter,
	)
	if err != nil {
		return nil, err
	}

	return &ApiPaginatedOutput[*TeamInvitation]{
		Body: ApiPaginatedResponse[*TeamInvitation]{
			Data: mapper.Map(invitations, fromTeamInvitationModel),
			Meta: ApiGenerateMeta(&input.PaginatedInput, count),
		},
	}, nil
}

func (api *Api) LoadTeamInvitationRelations(ctx context.Context, invitations []*models.TeamInvitation) error {
	teamIds := mapper.Map(
		invitations,
		func(t *models.TeamInvitation) uuid.UUID {
			return t.TeamID
		},
	)
	teams, err := api.App().Adapter().TeamGroup().LoadTeamsByIds(ctx, teamIds...)
	if err != nil {
		return err
	}
	for idx, invitation := range invitations {
		team := teams[idx]
		invitation.Team = team
	}

	memberIds := mapper.Map(
		invitations,
		func(t *models.TeamInvitation) uuid.UUID {
			return t.InviterMemberID
		},
	)
	members, err := api.App().Adapter().TeamMember().LoadTeamMembersByIds(ctx, memberIds...)
	if err != nil {
		return err
	}
	for idx, invitation := range invitations {
		member := members[idx]
		invitation.InviterMember = member
	}
	return nil
}

type FindUserTeamInvitationsInput struct {
	PaginatedInput
	SortParams
	Statuses []TeamInvitationStatus `query:"statuses,omitempty" required:"false" minimum:"1" maximum:"10" enum:"pending,accepted,declined,cancelled"`
}

func (api *Api) GetUserTeamInvitations(ctx context.Context, input *FindUserTeamInvitationsInput) (*ApiPaginatedOutput[*TeamInvitation], error) {
	userInfo := contextstore.GetContextUserInfo(ctx)
	if userInfo == nil {
		return nil, huma.Error401Unauthorized("Unauthorized. No user info")
	}

	filter := &stores.TeamInvitationFilter{}
	filter.Page = input.Page
	filter.PerPage = input.PerPage
	filter.SortBy = input.SortBy
	filter.SortOrder = input.SortOrder
	filter.Emails = []string{userInfo.User.Email}
	filter.Statuses = mapper.Map(input.Statuses, func(s TeamInvitationStatus) models.TeamInvitationStatus {
		return models.TeamInvitationStatus(s)
	})
	invitations, err := api.App().Adapter().TeamInvitation().FindTeamInvitations(
		ctx,
		filter,
	)
	if err != nil {
		return nil, err
	}
	if len(invitations) > 0 {
		err := api.LoadTeamInvitationRelations(ctx, invitations)
		if err != nil {
			return nil, err
		}
	}
	count, err := api.App().Adapter().TeamInvitation().CountTeamInvitations(
		ctx,
		filter,
	)
	if err != nil {
		return nil, err
	}
	return &ApiPaginatedOutput[*TeamInvitation]{
		Body: ApiPaginatedResponse[*TeamInvitation]{
			Data: mapper.Map(invitations, fromTeamInvitationModel),
			Meta: ApiGenerateMeta(&input.PaginatedInput, count),
		},
	}, nil
}

type GetInvitationByTokenInput struct {
	Token string `path:"token" required:"true"`
}

func (api *Api) GetInvitationByTokenBind(aapi huma.API) {
	huma.Register(
		aapi,
		huma.Operation{
			OperationID: "find-user-team-invitation-by-token",
			Method:      http.MethodGet,
			Path:        "/team-invitations/token/{token}",
			Summary:     "find-user-team-invitation-by-token",
			Description: "find user team invitation by token",
			Tags:        []string{"Team Invitations"},
			Errors:      []int{http.StatusInternalServerError, http.StatusBadRequest},
			Security:    []map[string][]string{{
				// shared.BearerAuthSecurityKey: {},
			}},
			Middlewares: huma.Middlewares{},
		},
		func(ctx context.Context, input *GetInvitationByTokenInput) (*ApiOutput[*TeamInvitation], error) {
			invitation, err := api.App().TeamInvitation().GetInvitation(ctx, input.Token)
			if err != nil {
				return nil, err
			}
			if invitation == nil {
				return nil, huma.Error404NotFound("invitation not found")
			}
			team, err := api.App().Adapter().TeamGroup().FindTeamByID(ctx, invitation.TeamID)
			if err != nil {
				return nil, err
			}
			invitation.Team = team
			member, err := api.App().Adapter().TeamMember().FindTeamMember(ctx, &stores.TeamMemberFilter{
				Ids: []uuid.UUID{invitation.InviterMemberID},
			})
			if err != nil {
				return nil, err
			}
			invitation.InviterMember = member
			if member != nil {
				if member.UserID != nil {
					user, err := api.App().Adapter().User().FindUserByID(ctx, *member.UserID)
					if err != nil {
						return nil, err
					}
					invitation.InviterMember.User = user
				}
			}
			return &ApiOutput[*TeamInvitation]{
				Body: fromTeamInvitationModel(invitation),
			}, nil
		},
	)
}
