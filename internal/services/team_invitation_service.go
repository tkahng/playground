package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/tools/security"
)

type TeamInvitationService interface {
	CreateInvitation(
		ctx context.Context,
		teamId uuid.UUID,
		userId uuid.UUID,
		email string,
		role models.TeamMemberRole,
	) error
	CheckValidInvitation(
		ctx context.Context,
		invitationToken string,
		userId uuid.UUID,
	) (bool, error)
	AcceptInvitation(
		ctx context.Context,
		invitationToken string,
		userId uuid.UUID,
	) error
	RejectInvitation(
		ctx context.Context,
		invitationToken string,
		userId uuid.UUID,
	) error
	FindInvitations(
		ctx context.Context,
		teamId uuid.UUID,
	) ([]*models.TeamInvitation, error)
}

type TeamInvitationStore interface {
	CreateTeamMember(ctx context.Context, teamId, userId uuid.UUID, role models.TeamMemberRole) (*models.TeamMember, error)
	DeleteTeamMember(ctx context.Context, teamId, userId uuid.UUID) error
	FindTeamMemberByTeamAndUserId(
		ctx context.Context,
		teamId uuid.UUID,
		userId uuid.UUID,
	) (*models.TeamMember, error)
	FindUserByID(ctx context.Context, userId uuid.UUID) (*models.User, error)
	CreateInvitation(
		ctx context.Context,
		invitation *models.TeamInvitation,
	) error
	UpdateInvitation(
		ctx context.Context,
		invitation *models.TeamInvitation,
	) error
	FindInvitationByToken(
		ctx context.Context,
		token string,
	) (*models.TeamInvitation, error)
	FindInvitationByID(
		ctx context.Context,
		invitationId uuid.UUID,
	) (*models.TeamInvitation, error)
	FindTeamInvitations(
		ctx context.Context,
		teamId uuid.UUID,
	) ([]*models.TeamInvitation, error)
}

type InnvitationService struct {
	store TeamInvitationStore
}

// CheckValidInvitation implements TeamInvitationService.
func (i *InnvitationService) CheckValidInvitation(ctx context.Context, invitationToken string, userId uuid.UUID) (bool, error) {
	invite, err := i.store.FindInvitationByToken(ctx, invitationToken)
	if err != nil {
		return false, err
	}
	if invite == nil {
		return false, fmt.Errorf("invitation not found")
	}
	user, err := i.store.FindUserByID(ctx, userId)
	if err != nil {
		return false, err
	}
	if user == nil {
		return false, fmt.Errorf("user not found")
	}
	if invite.Email != user.Email {
		return false, fmt.Errorf("user does not match invitation")
	}
	if invite.Status != models.TeamInvitationStatusPending {
		return false, fmt.Errorf("invitation is not pending")
	}
	return true, nil
}

var _ TeamInvitationService = (*InnvitationService)(nil)

func NewInvitationService(store TeamInvitationStore) TeamInvitationService {
	return &InnvitationService{
		store: store,
	}
}

// AcceptInvitation implements TeamInvitationService.
func (i *InnvitationService) AcceptInvitation(ctx context.Context, invitationToken string, userId uuid.UUID) error {
	invite, err := i.store.FindInvitationByToken(ctx, invitationToken)
	if err != nil {
		return err
	}
	if invite == nil {
		return fmt.Errorf("invitation not found")
	}
	user, err := i.store.FindUserByID(ctx, userId)
	if err != nil {
		return err
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}
	if invite.Email != user.Email {
		return fmt.Errorf("user does not match invitation")
	}
	if invite.Status != models.TeamInvitationStatusPending {
		return fmt.Errorf("invitation is not pending")
	}
	invite.Status = models.TeamInvitationStatusAccepted
	_, err = i.store.CreateTeamMember(ctx, invite.TeamID, user.ID, invite.Role)
	if err != nil {
		return err
	}
	err = i.store.UpdateInvitation(ctx, invite)
	if err != nil {
		return err
	}
	return nil
}

// CreateInvitation implements TeamInvitationService.
func (i *InnvitationService) CreateInvitation(
	ctx context.Context,
	teamId uuid.UUID,
	userId uuid.UUID,
	email string,
	role models.TeamMemberRole,
) error {
	token := security.GenerateTokenKey()

	var invitation = &models.TeamInvitation{
		TeamID: teamId,
		Email:  email,
		Role:   role,
		Token:  token,
		Status: models.TeamInvitationStatusPending,
	}
	member, err := i.store.FindTeamMemberByTeamAndUserId(ctx, teamId, userId)
	if err != nil {
		return err
	}
	if member == nil {
		return fmt.Errorf("user is not a member of the team")
	}
	invitation.InviterMemberID = member.ID
	invitation.ExpiresAt = time.Now().Add(24 * time.Hour)
	invitation.CreatedAt = time.Now()
	invitation.UpdatedAt = time.Now()

	return i.store.CreateInvitation(ctx, invitation)
}

// FindInvitations implements TeamInvitationService.
func (i *InnvitationService) FindInvitations(ctx context.Context, teamId uuid.UUID) ([]*models.TeamInvitation, error) {
	invitations, err := i.store.FindTeamInvitations(ctx, teamId)
	if err != nil {
		return nil, err
	}
	return invitations, nil
}

// RejectInvitation implements TeamInvitationService.
func (i *InnvitationService) RejectInvitation(ctx context.Context, invitationToken string, userId uuid.UUID) error {
	invite, err := i.store.FindInvitationByToken(ctx, invitationToken)
	if err != nil {
		return err
	}
	if invite == nil {
		return fmt.Errorf("invitation not found")
	}
	user, err := i.store.FindUserByID(ctx, userId)
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
	return nil
}
