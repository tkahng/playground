package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/models"
)

type TeamInvitationService interface {
	// SendInvitationMail(
	// 	ctx context.Context,
	// )
	CreateInvitation(
		ctx context.Context,
		teamId uuid.UUID,
		userId uuid.UUID,
		email string,
		role models.TeamMemberRole,
	) error
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

type invitationService struct {
	store TeamInvitationStore
}

// AcceptInvitation implements TeamInvitationService.
func (i *invitationService) AcceptInvitation(ctx context.Context, invitationToken string, userId uuid.UUID) error {
	panic("unimplemented")
}

// CreateInvitation implements TeamInvitationService.
func (i *invitationService) CreateInvitation(ctx context.Context, teamId uuid.UUID, userId uuid.UUID, email string, role models.TeamMemberRole) error {
	panic("unimplemented")
}

// FindInvitations implements TeamInvitationService.
func (i *invitationService) FindInvitations(ctx context.Context, teamId uuid.UUID) ([]*models.TeamInvitation, error) {
	panic("unimplemented")
}

// RejectInvitation implements TeamInvitationService.
func (i *invitationService) RejectInvitation(ctx context.Context, invitationToken string, userId uuid.UUID) error {
	panic("unimplemented")
}

func NewInvitationService(store TeamInvitationStore) TeamInvitationService {
	return &invitationService{
		store: store,
	}
}
