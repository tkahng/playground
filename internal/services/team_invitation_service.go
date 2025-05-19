package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/models"
)

type TeamInvitationService interface {
	SendInvitationMail(
		ctx context.Context,
	)
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
	GetInvitationByID(
		ctx context.Context,
		invitationId uuid.UUID,
	) (*models.TeamInvitation, error)
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
	FindTeamInvitations(
		ctx context.Context,
		teamId uuid.UUID,
	) ([]*models.TeamInvitation, error)
}

type invitationService struct {
	store TeamInvitationStore
}
