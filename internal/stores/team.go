package stores

import (
	"context"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/shared"
)

type TeamStoreInterface interface {
	// team
	ListTeams(ctx context.Context, params *shared.ListTeamsParams) ([]*models.Team, error)
	CountTeams(ctx context.Context, params *shared.ListTeamsParams) (int64, error)
	CreateTeamWithOwnerMember(ctx context.Context, name string, slug string, userId uuid.UUID) (*shared.TeamInfoModel, error)

	FindTeamByStripeCustomerId(ctx context.Context, stripeCustomerId string) (*models.Team, error)
	FindTeamByID(ctx context.Context, teamId uuid.UUID) (*models.Team, error)
	FindTeamBySlug(ctx context.Context, slug string) (*models.Team, error)

	LoadTeamsByIds(ctx context.Context, teamIds ...uuid.UUID) ([]*models.Team, error)
	CreateTeam(ctx context.Context, name string, slug string) (*models.Team, error)
	CheckTeamSlug(ctx context.Context, slug string) (bool, error)
	UpdateTeam(ctx context.Context, teamId uuid.UUID, name string) (*models.Team, error)
	DeleteTeam(ctx context.Context, teamId uuid.UUID) error

	// find team members
	FindTeamMember(ctx context.Context, member *TeamMemberFilter) (*models.TeamMember, error)
	// FindTeamMembersByUserID(ctx context.Context, userId uuid.UUID, paginate *shared.TeamMemberListInput) ([]*models.TeamMember, error)
	// FindTeamMemberByTeamAndUserId(ctx context.Context, teamId uuid.UUID, userId uuid.UUID) (*models.TeamMember, error)
	FindLatestTeamMemberByUserID(ctx context.Context, userId uuid.UUID) (*models.TeamMember, error)
	CountOwnerTeamMembers(ctx context.Context, teamId uuid.UUID) (int64, error)
	CountTeamMembers(ctx context.Context, filter *TeamMemberFilter) (int64, error)
	CountTeamMembersByUserID(ctx context.Context, userId uuid.UUID) (int64, error)

	CreateTeamMember(ctx context.Context, teamId, userId uuid.UUID, role models.TeamMemberRole, hasBillingAccess bool) (*models.TeamMember, error)
	DeleteTeamMember(ctx context.Context, teamId, userId uuid.UUID) error
	UpdateTeamMember(ctx context.Context, member *models.TeamMember) (*models.TeamMember, error)
	UpdateTeamMemberSelectedAt(ctx context.Context, teamId, userId uuid.UUID) error

	// misc methods
	FindUserByID(ctx context.Context, userId uuid.UUID) (*models.User, error)
}

type DbTeamStore struct {
	db database.Dbx
	*DbTeamGroupStore
	*DbTeamMemberStore
	*DbTeamInvitationStore
}

func (s *DbTeamStore) WithTx(db database.Dbx) *DbTeamStore {
	return &DbTeamStore{

		db:                    db,
		DbTeamGroupStore:      s.DbTeamGroupStore.WithTx(db),
		DbTeamMemberStore:     s.DbTeamMemberStore.WithTx(db),
		DbTeamInvitationStore: s.DbTeamInvitationStore.WithTx(db),
	}
}

func NewDbTeamStore(db database.Dbx) *DbTeamStore {
	return &DbTeamStore{
		db:                    db,
		DbTeamGroupStore:      NewDbTeamGroupStore(db),
		DbTeamMemberStore:     NewDbTeamMemberStore(db),
		DbTeamInvitationStore: NewDbTeamInvitationStore(db),
	}
}
