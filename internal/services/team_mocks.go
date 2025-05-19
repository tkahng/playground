package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/shared"
)

type mockTeamService struct {
	mock.Mock
}

// AddMember implements TeamService.
func (m *mockTeamService) AddMember(ctx context.Context, teamId uuid.UUID, userId uuid.UUID, role models.TeamMemberRole) (*models.TeamMember, error) {
	args := m.Called(ctx, teamId, userId, role)
	var member *models.TeamMember
	if args.Get(0) != nil {
		member = args.Get(0).(*models.TeamMember)
	}
	return member, args.Error(1)
}

// FindLatestTeamInfo implements TeamService.
func (m *mockTeamService) FindLatestTeamInfo(ctx context.Context, userId uuid.UUID) (*shared.TeamInfo, error) {
	args := m.Called(ctx, userId)
	var info *shared.TeamInfo
	if args.Get(0) != nil {
		info = args.Get(0).(*shared.TeamInfo)
	}
	return info, args.Error(1)
}

// FindTeamInfo implements TeamService.
func (m *mockTeamService) FindTeamInfo(ctx context.Context, teamId uuid.UUID, userId uuid.UUID) (*shared.TeamInfo, error) {
	args := m.Called(ctx, teamId, userId)
	var info *shared.TeamInfo
	if args.Get(0) != nil {
		info = args.Get(0).(*shared.TeamInfo)
	}
	return info, args.Error(1)
}

// GetActiveTeam implements TeamService.
func (m *mockTeamService) GetActiveTeam(ctx context.Context, userId uuid.UUID) (*models.TeamMember, error) {
	args := m.Called(ctx, userId)
	var member *models.TeamMember
	if args.Get(0) != nil {
		member = args.Get(0).(*models.TeamMember)
	}
	return member, args.Error(1)
}

// RemoveMember implements TeamService.
func (m *mockTeamService) RemoveMember(ctx context.Context, teamId uuid.UUID, userId uuid.UUID) error {
	args := m.Called(ctx, teamId, userId)
	return args.Error(0)
}

// SetActiveTeam implements TeamService.
func (m *mockTeamService) SetActiveTeam(ctx context.Context, userId uuid.UUID, teamId uuid.UUID) (*models.TeamMember, error) {
	args := m.Called(ctx, userId, teamId)
	var member *models.TeamMember
	if args.Get(0) != nil {
		member = args.Get(0).(*models.TeamMember)
	}
	return member, args.Error(1)
}

// Store implements TeamService.
func (m *mockTeamService) Store() TeamStore {
	args := m.Called()
	return args.Get(0).(TeamStore)
}

var _ TeamService = (*mockTeamService)(nil)

type mockTeamStore struct {
	mock.Mock
}

// CheckTeamSlug implements TeamStore.
func (m *mockTeamStore) CheckTeamSlug(ctx context.Context, slug string) (bool, error) {
	args := m.Called(ctx, slug)
	return args.Bool(0), args.Error(1)
}

// CountTeamMembers implements TeamStore.
func (m *mockTeamStore) CountTeamMembers(ctx context.Context, teamId uuid.UUID) (int64, error) {
	args := m.Called(ctx, teamId)
	var count int64
	if args.Get(0) != nil {
		count = args.Get(0).(int64)
	}
	return count, args.Error(1)
}

// CreateTeam implements TeamStore.
func (m *mockTeamStore) CreateTeam(ctx context.Context, name string, slug string, stripeCustomerId *string) (*models.Team, error) {
	args := m.Called(ctx, name, slug, stripeCustomerId)
	var team *models.Team
	if args.Get(0) != nil {
		team = args.Get(0).(*models.Team)
	}
	return team, args.Error(1)
}

// CreateTeamMember implements TeamStore.
func (m *mockTeamStore) CreateTeamMember(ctx context.Context, teamId uuid.UUID, userId uuid.UUID, role models.TeamMemberRole) (*models.TeamMember, error) {
	args := m.Called(ctx, teamId, userId, role)
	var member *models.TeamMember
	if args.Get(0) != nil {
		member = args.Get(0).(*models.TeamMember)
	}
	return member, args.Error(1)
}

// DeleteTeam implements TeamStore.
func (m *mockTeamStore) DeleteTeam(ctx context.Context, teamId uuid.UUID) error {
	args := m.Called(ctx, teamId)
	return args.Error(0)
}

// DeleteTeamMember implements TeamStore.
func (m *mockTeamStore) DeleteTeamMember(ctx context.Context, teamId uuid.UUID, userId uuid.UUID) error {
	args := m.Called(ctx, teamId, userId)
	return args.Error(0)
}

// FindLatestTeamMemberByUserID implements TeamStore.
func (m *mockTeamStore) FindLatestTeamMemberByUserID(ctx context.Context, userId uuid.UUID) (*models.TeamMember, error) {
	args := m.Called(ctx, userId)
	var member *models.TeamMember
	if args.Get(0) != nil {
		member = args.Get(0).(*models.TeamMember)
	}
	return member, args.Error(1)
}

// FindTeamByID implements TeamStore.
func (m *mockTeamStore) FindTeamByID(ctx context.Context, teamId uuid.UUID) (*models.Team, error) {
	args := m.Called(ctx, teamId)
	var team *models.Team
	if args.Get(0) != nil {
		team = args.Get(0).(*models.Team)
	}
	return team, args.Error(1)
}

// FindTeamByStripeCustomerId implements TeamStore.
func (m *mockTeamStore) FindTeamByStripeCustomerId(ctx context.Context, stripeCustomerId string) (*models.Team, error) {
	args := m.Called(ctx, stripeCustomerId)
	var team *models.Team
	if args.Get(0) != nil {
		team = args.Get(0).(*models.Team)
	}
	return team, args.Error(1)
}

// FindTeamMemberByTeamAndUserId implements TeamStore.
func (m *mockTeamStore) FindTeamMemberByTeamAndUserId(ctx context.Context, teamId uuid.UUID, userId uuid.UUID) (*models.TeamMember, error) {
	args := m.Called(ctx, teamId, userId)
	var member *models.TeamMember
	if args.Get(0) != nil {
		member = args.Get(0).(*models.TeamMember)
	}
	return member, args.Error(1)
}

// FindTeamMembersByUserID implements TeamStore.
func (m *mockTeamStore) FindTeamMembersByUserID(ctx context.Context, userId uuid.UUID) ([]*models.TeamMember, error) {
	args := m.Called(ctx, userId)
	var members []*models.TeamMember
	if args.Get(0) != nil {
		members = args.Get(0).([]*models.TeamMember)
	}
	return members, args.Error(1)
}

// UpdateTeam implements TeamStore.
func (m *mockTeamStore) UpdateTeam(ctx context.Context, teamId uuid.UUID, name string, stripeCustomerId *string) (*models.Team, error) {
	args := m.Called(ctx, teamId, name, stripeCustomerId)
	var team *models.Team
	if args.Get(0) != nil {
		team = args.Get(0).(*models.Team)
	}
	return team, args.Error(1)
}

// UpdateTeamMember implements TeamStore.
func (m *mockTeamStore) UpdateTeamMember(ctx context.Context, member *models.TeamMember) (*models.TeamMember, error) {
	args := m.Called(ctx, member)
	var updated *models.TeamMember
	if args.Get(0) != nil {
		updated = args.Get(0).(*models.TeamMember)
	}
	return updated, args.Error(1)
}

// UpdateTeamMemberSelectedAt implements TeamStore.
func (m *mockTeamStore) UpdateTeamMemberSelectedAt(ctx context.Context, teamId uuid.UUID, userId uuid.UUID) error {
	args := m.Called(ctx, teamId, userId)
	return args.Error(0)
}

var _ TeamStore = (*mockTeamStore)(nil)

type mockTeamInvitationService struct {
	mock.Mock
}

// AcceptInvitation implements TeamInvitationService.
func (m *mockTeamInvitationService) AcceptInvitation(ctx context.Context, invitationToken string, userId uuid.UUID) error {
	args := m.Called(ctx, invitationToken, userId)
	return args.Error(0)
}

// CreateInvitation implements TeamInvitationService.
func (m *mockTeamInvitationService) CreateInvitation(ctx context.Context, teamId uuid.UUID, userId uuid.UUID, email string, role models.TeamMemberRole) error {
	args := m.Called(ctx, teamId, userId, email, role)
	return args.Error(0)
}

// FindInvitations implements TeamInvitationService.
func (m *mockTeamInvitationService) FindInvitations(ctx context.Context, teamId uuid.UUID) ([]*models.TeamInvitation, error) {
	args := m.Called(ctx, teamId)
	var invitations []*models.TeamInvitation
	if args.Get(0) != nil {
		invitations = args.Get(0).([]*models.TeamInvitation)
	}
	return invitations, args.Error(1)
}

// RejectInvitation implements TeamInvitationService.
func (m *mockTeamInvitationService) RejectInvitation(ctx context.Context, invitationToken string, userId uuid.UUID) error {
	args := m.Called(ctx, invitationToken, userId)
	return args.Error(0)
}

var _ TeamInvitationService = (*mockTeamInvitationService)(nil)

type mockTeamInvitationStore struct {
	mock.Mock
}

// CreateTeamMember implements TeamInvitationStore.
func (m *mockTeamInvitationStore) CreateTeamMember(ctx context.Context, teamId uuid.UUID, userId uuid.UUID, role models.TeamMemberRole) (*models.TeamMember, error) {
	args := m.Called(ctx, teamId, userId, role)
	var member *models.TeamMember
	if args.Get(0) != nil {
		member = args.Get(0).(*models.TeamMember)
	}
	return member, args.Error(1)
}

// DeleteTeamMember implements TeamInvitationStore.
func (m *mockTeamInvitationStore) DeleteTeamMember(ctx context.Context, teamId uuid.UUID, userId uuid.UUID) error {
	args := m.Called(ctx, teamId, userId)
	return args.Error(0)
}

// FindTeamMemberByTeamAndUserId implements TeamInvitationStore.
func (m *mockTeamInvitationStore) FindTeamMemberByTeamAndUserId(ctx context.Context, teamId uuid.UUID, userId uuid.UUID) (*models.TeamMember, error) {
	args := m.Called(ctx, teamId, userId)
	var member *models.TeamMember
	if args.Get(0) != nil {
		member = args.Get(0).(*models.TeamMember)
	}
	return member, args.Error(1)
}

// FindUserByID implements TeamInvitationStore.
func (m *mockTeamInvitationStore) FindUserByID(ctx context.Context, userId uuid.UUID) (*models.User, error) {
	args := m.Called(ctx, userId)
	var user *models.User
	if args.Get(0) != nil {
		user = args.Get(0).(*models.User)
	}
	return user, args.Error(1)
}

// FindTeamInvitations implements TeamInvitationStore.
func (m *mockTeamInvitationStore) FindTeamInvitations(ctx context.Context, teamId uuid.UUID) ([]*models.TeamInvitation, error) {
	args := m.Called(ctx, teamId)
	var invitations []*models.TeamInvitation
	if args.Get(0) != nil {
		invitations = args.Get(0).([]*models.TeamInvitation)
	}
	return invitations, args.Error(1)
}

// CreateInvitation implements TeamInvitationStore.
func (m *mockTeamInvitationStore) CreateInvitation(ctx context.Context, invitation *models.TeamInvitation) error {
	args := m.Called(ctx, invitation)
	return args.Error(0)
}

// UpdateInvitation implements TeamInvitationStore.
func (m *mockTeamInvitationStore) UpdateInvitation(ctx context.Context, invitation *models.TeamInvitation) error {
	args := m.Called(ctx, invitation)
	return args.Error(0)
}

// FindInvitationByID implements TeamInvitationStore.
func (m *mockTeamInvitationStore) FindInvitationByID(ctx context.Context, invitationId uuid.UUID) (*models.TeamInvitation, error) {
	args := m.Called(ctx, invitationId)
	var invitation *models.TeamInvitation
	if args.Get(0) != nil {
		invitation = args.Get(0).(*models.TeamInvitation)
	}
	return invitation, args.Error(1)
}

// FindInvitationByToken implements TeamInvitationStore.
func (m *mockTeamInvitationStore) FindInvitationByToken(ctx context.Context, token string) (*models.TeamInvitation, error) {
	args := m.Called(ctx, token)
	var invitation *models.TeamInvitation
	if args.Get(0) != nil {
		invitation = args.Get(0).(*models.TeamInvitation)
	}
	return invitation, args.Error(1)
}

var _ TeamInvitationStore = (*mockTeamInvitationStore)(nil)
