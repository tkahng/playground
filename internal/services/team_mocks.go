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

// FindTeamMembersByUserID implements TeamService.
func (m *mockTeamService) FindTeamMembersByUserID(ctx context.Context, userId uuid.UUID, paginate *shared.TeamMemberListInput) ([]*models.TeamMember, error) {
	args := m.Called(ctx, userId, paginate)
	var members []*models.TeamMember
	if args.Get(0) != nil {
		members = args.Get(0).([]*models.TeamMember)
	}
	return members, args.Error(1)
}

// LeaveTeam implements TeamService.
func (m *mockTeamService) LeaveTeam(ctx context.Context, teamId uuid.UUID, userId uuid.UUID) error {
	args := m.Called(ctx, teamId, userId)
	return args.Error(0)
}

// DeleteTeam implements TeamService.
func (m *mockTeamService) DeleteTeam(ctx context.Context, teamId uuid.UUID, userId uuid.UUID) error {
	args := m.Called(ctx, teamId, userId)
	return args.Error(0)
}

// UpdateTeam implements TeamService.
func (m *mockTeamService) UpdateTeam(ctx context.Context, teamId uuid.UUID, name string) (*models.Team, error) {
	args := m.Called(ctx, teamId, name)
	var team *models.Team
	if args.Get(0) != nil {
		team = args.Get(0).(*models.Team)
	}
	return team, args.Error(1)
}

// CreateTeam implements TeamService.
func (m *mockTeamService) CreateTeam(ctx context.Context, name string, slug string, userId uuid.UUID) (*shared.TeamInfo, error) {
	args := m.Called(ctx, name, slug, userId)
	var teamInfo *shared.TeamInfo
	if args.Get(0) != nil {
		teamInfo = args.Get(0).(*shared.TeamInfo)
	}
	return teamInfo, args.Error(1)
}

// AddMember implements TeamService.
func (m *mockTeamService) AddMember(ctx context.Context, teamId uuid.UUID, userId uuid.UUID, role models.TeamMemberRole, hasBillingAccess bool) (*models.TeamMember, error) {
	args := m.Called(ctx, teamId, userId, role, hasBillingAccess)
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

// GetActiveTeamMember implements TeamService.
func (m *mockTeamService) GetActiveTeamMember(ctx context.Context, userId uuid.UUID) (*models.TeamMember, error) {
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

// SetActiveTeamMember implements TeamService.
func (m *mockTeamService) SetActiveTeamMember(ctx context.Context, userId uuid.UUID, teamId uuid.UUID) (*models.TeamMember, error) {
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

// CountTeams implements TeamServiceStore.
func (m *mockTeamStore) CountTeams(ctx context.Context, params *shared.ListTeamsParams) (int64, error) {
	args := m.Called(ctx, params)
	var count int64
	if args.Get(0) != nil {
		count = args.Get(0).(int64)
	}
	return count, args.Error(1)
}

// ListTeams implements TeamServiceStore.
func (m *mockTeamStore) ListTeams(ctx context.Context, params *shared.ListTeamsParams) ([]*models.Team, error) {
	args := m.Called(ctx, params)
	var teams []*models.Team
	if args.Get(0) != nil {
		teams = args.Get(0).([]*models.Team)
	}
	return teams, args.Error(1)
}

// FindUserByID implements TeamServiceStore.
func (m *mockTeamStore) FindUserByID(ctx context.Context, userId uuid.UUID) (*models.User, error) {
	args := m.Called(ctx, userId)
	var user *models.User
	if args.Get(0) != nil {
		user = args.Get(0).(*models.User)
	}
	return user, args.Error(1)
}

// FindTeam implements TeamServiceStore.
func (m *mockTeamStore) FindTeam(ctx context.Context, team *models.Team) (*models.Team, error) {
	args := m.Called(ctx, team)
	var teamInfo *models.Team
	if args.Get(0) != nil {
		teamInfo = args.Get(0).(*models.Team)
	}
	return teamInfo, args.Error(1)
}

// FindTeamMember implements TeamServiceStore.
func (m *mockTeamStore) FindTeamMember(ctx context.Context, member *models.TeamMember) (*models.TeamMember, error) {
	args := m.Called(ctx, member)
	var teamMember *models.TeamMember
	if args.Get(0) != nil {
		teamMember = args.Get(0).(*models.TeamMember)
	}
	return teamMember, args.Error(1)
}

// LoadTeamsByIds implements TeamServiceStore.
func (m *mockTeamStore) LoadTeamsByIds(ctx context.Context, teamIds ...uuid.UUID) ([]*models.Team, error) {
	args := m.Called(ctx, teamIds)
	var teams []*models.Team
	if args.Get(0) != nil {
		teams = args.Get(0).([]*models.Team)
	}
	return teams, args.Error(1)
}

// CountTeamMembersByUserID implements TeamServiceStore.
func (m *mockTeamStore) CountTeamMembersByUserID(ctx context.Context, userId uuid.UUID) (int64, error) {
	args := m.Called(ctx, userId)
	var count int64
	if args.Get(0) != nil {
		count = args.Get(0).(int64)
	}
	return count, args.Error(1)
}

// CountOwnerTeamMembers implements TeamServiceStore.
func (m *mockTeamStore) CountOwnerTeamMembers(ctx context.Context, teamId uuid.UUID) (int64, error) {
	args := m.Called(ctx, teamId)
	var count int64
	if args.Get(0) != nil {
		count = args.Get(0).(int64)
	}
	return count, args.Error(1)
}

// FindLatestActiveSubscriptionWithPriceByCustomerId implements TeamServiceStore.
func (m *mockTeamStore) FindLatestActiveSubscriptionWithPriceByCustomerId(ctx context.Context, customerId string) (*models.SubscriptionWithPrice, error) {
	args := m.Called(ctx, customerId)
	var subscription *models.SubscriptionWithPrice
	if args.Get(0) != nil {
		subscription = args.Get(0).(*models.SubscriptionWithPrice)
	}
	return subscription, args.Error(1)
}

// CreateTeamWithOwnerMember implements TeamStore.
func (m *mockTeamStore) CreateTeamWithOwnerMember(ctx context.Context, name string, slug string, userId uuid.UUID) (*shared.TeamInfo, error) {
	args := m.Called(ctx, name, slug, userId)
	var teamInfo *shared.TeamInfo
	if args.Get(0) != nil {
		teamInfo = args.Get(0).(*shared.TeamInfo)
	}
	return teamInfo, args.Error(1)
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
func (m *mockTeamStore) CreateTeam(ctx context.Context, name string, slug string) (*models.Team, error) {
	args := m.Called(ctx, name, slug)
	var team *models.Team
	if args.Get(0) != nil {
		team = args.Get(0).(*models.Team)
	}
	return team, args.Error(1)
}

// CreateTeamMember implements TeamStore.
func (m *mockTeamStore) CreateTeamMember(ctx context.Context, teamId uuid.UUID, userId uuid.UUID, role models.TeamMemberRole, hasBillingAccess bool) (*models.TeamMember, error) {
	args := m.Called(ctx, teamId, userId, role, hasBillingAccess)
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
func (m *mockTeamStore) FindTeamMembersByUserID(ctx context.Context, userId uuid.UUID, paginate *shared.TeamMemberListInput) ([]*models.TeamMember, error) {
	args := m.Called(ctx, userId, paginate)
	var members []*models.TeamMember
	if args.Get(0) != nil {
		members = args.Get(0).([]*models.TeamMember)
	}
	return members, args.Error(1)
}

// UpdateTeam implements TeamStore.
func (m *mockTeamStore) UpdateTeam(ctx context.Context, teamId uuid.UUID, name string) (*models.Team, error) {
	args := m.Called(ctx, teamId, name)
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

var _ TeamServiceStore = (*mockTeamStore)(nil)

type mockTeamInvitationService struct {
	mock.Mock
}

// CancelInvitation implements TeamInvitationService.
func (m *mockTeamInvitationService) CancelInvitation(ctx context.Context, teamId uuid.UUID, userId uuid.UUID, invitationId uuid.UUID) error {
	args := m.Called(ctx, teamId, userId, invitationId)
	return args.Error(0)
}

// FireAndForget implements TeamInvitationService.
func (m *mockTeamInvitationService) FireAndForget(f func()) {
	args := m.Called(f)
	if args.Get(0) != nil {
		f()
	}
}

// SendInvitationEmail implements TeamInvitationService.
func (m *mockTeamInvitationService) SendInvitationEmail(ctx context.Context, params *TeamInvitationMailParams) error {
	args := m.Called(ctx, params)
	return args.Error(0)
}

// CheckValidInvitation implements TeamInvitationService.
func (m *mockTeamInvitationService) CheckValidInvitation(ctx context.Context, userId uuid.UUID, invitationToken string) (bool, error) {
	args := m.Called(ctx, invitationToken, userId)
	var valid bool
	if args.Get(0) != nil {
		valid = args.Get(0).(bool)
	}
	return valid, args.Error(1)
}

// AcceptInvitation implements TeamInvitationService.
func (m *mockTeamInvitationService) AcceptInvitation(ctx context.Context, userId uuid.UUID, invitationToken string) error {
	args := m.Called(ctx, invitationToken, userId)
	return args.Error(0)
}

// CreateInvitation implements TeamInvitationService.
func (m *mockTeamInvitationService) CreateInvitation(ctx context.Context, teamId uuid.UUID, userId uuid.UUID, email string, role models.TeamMemberRole, resend bool) error {
	args := m.Called(ctx, teamId, userId, email, role, resend)
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
func (m *mockTeamInvitationService) RejectInvitation(ctx context.Context, userId uuid.UUID, invitationToken string) error {
	args := m.Called(ctx, userId, invitationToken)
	return args.Error(0)
}

var _ TeamInvitationService = (*mockTeamInvitationService)(nil)

type mockTeamInvitationStore struct {
	mock.Mock
}

// FindTeamByID implements TeamInvitationStore.
func (m *mockTeamInvitationStore) FindTeamByID(ctx context.Context, teamId uuid.UUID) (*models.Team, error) {
	args := m.Called(ctx, teamId)
	var team *models.Team
	if args.Get(0) != nil {
		team = args.Get(0).(*models.Team)
	}
	return team, args.Error(1)
}

// FindPendingInvitation implements TeamInvitationStore.
func (m *mockTeamInvitationStore) FindPendingInvitation(ctx context.Context, teamId uuid.UUID, email string) (*models.TeamInvitation, error) {
	args := m.Called(ctx, teamId, email)
	var invitation *models.TeamInvitation
	if args.Get(0) != nil {
		invitation = args.Get(0).(*models.TeamInvitation)
	}
	return invitation, args.Error(1)
}

// CreateTeamMember implements TeamInvitationStore.
func (m *mockTeamInvitationStore) CreateTeamMember(ctx context.Context, teamId uuid.UUID, userId uuid.UUID, role models.TeamMemberRole, hasBillingAccess bool) (*models.TeamMember, error) {
	args := m.Called(ctx, teamId, userId, role, hasBillingAccess)
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
