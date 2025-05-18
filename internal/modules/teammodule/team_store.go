package teammodule

import (
	"context"

	"github.com/google/uuid"

	"github.com/tkahng/authgo/internal/models"
)

type TeamStore interface {
	CreateTeamFromUser(ctx context.Context, user *models.User) (*models.Team, error)
	FindTeamByID(ctx context.Context, teamId uuid.UUID) (*models.Team, error)
	CreateTeam(ctx context.Context, name string, stripeCustomerId *string) (*models.Team, error)
	UpdateTeam(ctx context.Context, teamId uuid.UUID, name string, stripeCustomerId *string) (*models.Team, error)
	DeleteTeam(ctx context.Context, teamId uuid.UUID) error
	FindTeamMembersByUserID(ctx context.Context, userId uuid.UUID) ([]*models.TeamMember, error)
	FindTeamMemberByUserAndTeamID(ctx context.Context, teamId uuid.UUID, userId uuid.UUID) (*models.TeamMember, error)
	FindLatestTeamMemberByUserID(ctx context.Context, userId uuid.UUID) (*models.TeamMember, error)
	CreateTeamMember(ctx context.Context, teamId, userId uuid.UUID, role models.TeamMemberRole) (*models.TeamMember, error)
	UpdateTeamMemberUpdatedAt(ctx context.Context, teamMemberId uuid.UUID) error
}
