package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/conf"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/stores"
	"github.com/tkahng/authgo/internal/tools/security"
	"github.com/tkahng/authgo/internal/workers"
)

type TeamInvitationMailParams struct {
	Email           string
	InvitedByEmail  string
	TeamName        string
	TokenHash       string
	ConfirmationURL string
}

type TeamInvitationService interface {
	CreateInvitation(
		ctx context.Context,
		teamId uuid.UUID,
		invitingUserId uuid.UUID,
		inviteeEmail string,
		role models.TeamMemberRole,
		resend bool,
	) error
	CheckValidInvitation(
		ctx context.Context,
		userId uuid.UUID,
		invitationToken string,
	) (bool, error)
	GetInvitation(
		ctx context.Context,
		invitationToken string,
	) (*models.TeamInvitation, error)
	AcceptInvitation(
		ctx context.Context,
		userId uuid.UUID,
		invitationToken string,
	) error
	RejectInvitation(
		ctx context.Context,
		userId uuid.UUID,
		invitationToken string,
	) error

	CancelInvitation(
		ctx context.Context,
		teamId uuid.UUID,
		userId uuid.UUID,
		invitationId uuid.UUID,
	) error

	FindInvitations(
		ctx context.Context,
		teamId uuid.UUID,
	) ([]*models.TeamInvitation, error)
}

var _ TeamInvitationService = (*InvitationService)(nil)

type InvitationService struct {
	adapter    stores.StorageAdapterInterface
	settings   conf.AppOptions
	jobService JobService
}

// GetInvitation implements TeamInvitationService.
func (i *InvitationService) GetInvitation(ctx context.Context, invitationToken string) (*models.TeamInvitation, error) {
	inv, err := i.adapter.TeamInvitation().FindInvitationByToken(ctx, invitationToken)
	if err != nil {
		return nil, err
	}
	if inv == nil {
		return nil, fmt.Errorf("invitation not found")
	}
	if inv.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("invitation is expired")
	}
	return inv, nil

}

func NewInvitationService(
	adapter stores.StorageAdapterInterface,
	settings conf.AppOptions,
	jobService JobService,
) TeamInvitationService {
	return &InvitationService{
		settings:   settings,
		adapter:    adapter,
		jobService: jobService,
	}
}
func (i *InvitationService) CancelInvitation(
	ctx context.Context,
	teamId uuid.UUID,
	userId uuid.UUID,
	invitationId uuid.UUID,
) error {
	member, err := i.adapter.TeamMember().FindTeamMember(ctx, &stores.TeamMemberFilter{
		TeamIds: []uuid.UUID{teamId},
		UserIds: []uuid.UUID{userId},
	})
	if err != nil {
		return err
	}
	if member == nil {
		return fmt.Errorf("user is not a member of the team")
	}
	if member.Role != models.TeamMemberRoleOwner {
		return fmt.Errorf("user is not an owner of the team")
	}
	invitation, err := i.adapter.TeamInvitation().FindInvitationByID(ctx, invitationId)
	if err != nil {
		return err
	}
	if invitation == nil {
		return fmt.Errorf("invitation not found")
	}
	if invitation.TeamID != teamId {
		return fmt.Errorf("invitation does not match team")
	}
	invitation.Status = models.TeamInvitationStatusCanceled

	return i.adapter.TeamInvitation().UpdateInvitation(ctx, invitation)
}

// CheckValidInvitation implements TeamInvitationService.
func (i *InvitationService) CheckValidInvitation(
	ctx context.Context,
	userId uuid.UUID,
	invitationToken string,
) (bool, error) {
	invite, err := i.adapter.TeamInvitation().FindInvitationByToken(ctx, invitationToken)
	if err != nil {
		return false, err
	}
	if invite == nil {
		return false, fmt.Errorf("invitation not found")
	}
	user, err := i.adapter.User().FindUserByID(ctx, userId)
	if err != nil {
		return false, err
	}
	if user == nil {
		return false, fmt.Errorf("user not found")
	}
	if invite.Email != user.Email {
		return false, fmt.Errorf("user does not match invitation")
	}
	if invite.ExpiresAt.After(time.Now()) {
		return false, fmt.Errorf("invitation is expired")
	}
	if invite.Status != models.TeamInvitationStatusPending {
		return false, fmt.Errorf("invitation is not pending")
	}

	return true, nil
}

// AcceptInvitation implements TeamInvitationService.
func (i *InvitationService) AcceptInvitation(
	ctx context.Context,
	userId uuid.UUID,
	invitationToken string,
) error {
	teamMember := &models.TeamMember{}
	err := i.adapter.TeamInvitation().AcceptInvitation(ctx, i.adapter, userId, invitationToken, teamMember)
	if err != nil {
		return err
	}
	err = i.jobService.EnqueueRefreshSubscriptionQuantityJob(ctx, &workers.RefreshSubscriptionQuantityJobArgs{
		TeamID: teamMember.TeamID,
	})
	if err != nil {
		return err
	}
	err = i.jobService.EnqueueTeamMemberAddedJob(ctx, &workers.NewMemberNotificationJobArgs{
		TeamMemberID: teamMember.ID,
	})
	if err != nil {
		return err
	}
	return nil
}

// CreateInvitation implements TeamInvitationService.
func (i *InvitationService) CreateInvitation(
	ctx context.Context,
	teamId uuid.UUID,
	invitingUserId uuid.UUID,
	inviteeEmail string,
	role models.TeamMemberRole,
	resend bool,
) error {

	member, err := i.adapter.TeamMember().FindTeamMember(ctx, &stores.TeamMemberFilter{
		TeamIds: []uuid.UUID{teamId},
		UserIds: []uuid.UUID{invitingUserId},
	})
	if err != nil {
		return err
	}
	if member == nil {
		return fmt.Errorf("user is not a member of the team")
	}
	if user, err := i.adapter.User().FindUserByID(ctx, invitingUserId); err != nil {
		return err
	} else if user == nil {
		return fmt.Errorf("user not found")
	} else {
		member.User = user
	}
	if team, err := i.adapter.TeamGroup().FindTeamByID(ctx, teamId); err != nil {
		return err
	} else if team == nil {
		return fmt.Errorf("team not found")
	} else {
		member.Team = team
	}
	invitation := new(models.TeamInvitation)
	existingInvite, err := i.adapter.TeamInvitation().FindPendingInvitation(ctx, teamId, inviteeEmail)
	if err != nil {
		return err
	}
	if existingInvite == nil {
		token := security.GenerateTokenKey()
		invitation.Status = models.TeamInvitationStatusPending
		invitation.Token = token
		invitation.Email = inviteeEmail
		invitation.Role = role
		invitation.TeamID = teamId
		invitation.InviterMemberID = member.ID
		invitation.ExpiresAt = i.settings.Auth.InviteToken.Expires()
		err = i.adapter.TeamInvitation().CreateInvitation(ctx, invitation)
		if err != nil {
			return err
		}
	} else {
		if !resend {
			return fmt.Errorf("invitation already exists")
		}
		existingInvite.Status = models.TeamInvitationStatusPending
		existingInvite.Role = role
		existingInvite.ExpiresAt = i.settings.Auth.InviteToken.Expires()
		err = i.adapter.TeamInvitation().UpdateInvitation(ctx, existingInvite)
		if err != nil {
			return err
		}
		invitation = existingInvite
	}
	err = i.jobService.EnqueueTeamInvitationJob(ctx, &workers.TeamInvitationJobArgs{
		Email:          invitation.Email,
		InvitedByEmail: member.User.Email,
		TeamName:       member.Team.Name,
		TokenHash:      invitation.Token,
	})
	if err != nil {
		return err
	}

	return nil
}

// FindInvitations implements TeamInvitationService.
func (i *InvitationService) FindInvitations(ctx context.Context, teamId uuid.UUID) ([]*models.TeamInvitation, error) {
	invitations, err := i.adapter.TeamInvitation().FindTeamInvitations(ctx, &stores.TeamInvitationFilter{
		TeamIds: []uuid.UUID{teamId},
	})
	if err != nil {
		return nil, err
	}
	return invitations, nil
}

// RejectInvitation implements TeamInvitationService.
func (i *InvitationService) RejectInvitation(ctx context.Context, userId uuid.UUID, invitationToken string) error {
	invite, err := i.adapter.TeamInvitation().FindInvitationByToken(ctx, invitationToken)
	if err != nil {
		return err
	}
	if invite == nil {
		return fmt.Errorf("invitation not found")
	}
	user, err := i.adapter.User().FindUserByID(ctx, userId)
	if err != nil {
		return err
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}
	if invite.Email != user.Email {
		return fmt.Errorf("user does not match invitation")
	}
	invite.Status = models.TeamInvitationStatusDeclined
	err = i.adapter.TeamInvitation().UpdateInvitation(ctx, invite)
	if err != nil {
		return err
	}
	return nil
}
