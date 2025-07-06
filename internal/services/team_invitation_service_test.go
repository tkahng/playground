package services

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/tkahng/authgo/internal/conf"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/stores"
)

func TestNewInvitationService(t *testing.T) {
	mockStore := stores.NewAdapterDecorators()
	opts := conf.NewSettings()

	JobService := &JobServiceDecorator{}
	service := NewInvitationService(mockStore, *opts, JobService)

	assert.NotNil(t, service, "NewInvitationService should not return nil")
}

func TestInvitationService_CreateInvitation(t *testing.T) {
	ctx := context.Background()
	store := stores.NewAdapterDecorators()
	opts := conf.NewSettings()

	jobService := &JobServiceDecorator{}
	service := NewInvitationService(store, *opts, jobService)

	member := &models.TeamMember{ID: uuid.New()}
	inviteeEmail := "invitee@example.com"
	invitingUser := &models.User{ID: uuid.New(), Email: "inviting@example.com"}
	team := &models.Team{ID: uuid.New(), Name: "Test Team"}
	store.TeamMemberFunc.FindTeamMemberFunc = func(ctx context.Context, filter *stores.TeamMemberFilter) (*models.TeamMember, error) {
		return member, nil
	}
	store.UserFunc.FindUserByIDFunc = func(ctx context.Context, userId uuid.UUID) (*models.User, error) {
		return invitingUser, nil
	}
	store.TeamGroupFunc.FindTeamByIDFunc = func(ctx context.Context, teamId uuid.UUID) (*models.Team, error) {
		return team, nil
	}
	store.TeamInvitationFunc.FindPendingInvitationFunc = func(ctx context.Context, teamId uuid.UUID, email string) (*models.TeamInvitation, error) {
		return nil, nil
	}
	store.TeamInvitationFunc.CreateInvitationFunc = func(ctx context.Context, invitation *models.TeamInvitation) error {
		return nil
	}

	err := service.CreateInvitation(ctx, team.ID, invitingUser.ID, inviteeEmail, models.TeamMemberRoleMember, true)
	assert.NoError(t, err)
}

func TestInvitationService_CreateInvitation_NotMember(t *testing.T) {
	ctx := context.Background()
	store := stores.NewAdapterDecorators()
	opts := conf.NewSettings()

	jobService := &JobServiceDecorator{}
	service := NewInvitationService(store, *opts, jobService)

	teamId := uuid.New()
	userId := uuid.New()
	store.TeamMemberFunc.FindTeamMemberFunc = func(ctx context.Context, filter *stores.TeamMemberFilter) (*models.TeamMember, error) {
		return nil, nil
	}

	err := service.CreateInvitation(ctx, teamId, userId, "test@example.com", models.TeamMemberRoleMember, true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not a member")
}

func TestInvitationService_AcceptInvitation(t *testing.T) {
	ctx := context.Background()
	store := stores.NewAdapterDecorators()
	opts := conf.NewSettings()

	// Mock the mail service
	jobService := &JobServiceDecorator{}
	service := NewInvitationService(store, *opts, jobService)

	teamId := uuid.New()
	userId := uuid.New()
	invitation := &models.TeamInvitation{
		TeamID: teamId,
		Email:  "test@example.com",
		Status: models.TeamInvitationStatusPending,
		Role:   models.TeamMemberRoleMember,
	}
	user := &models.User{ID: userId, Email: "test@example.com"}
	store.TeamInvitationFunc.FindInvitationByTokenFunc = func(ctx context.Context, token string) (*models.TeamInvitation, error) {
		return invitation, nil
	}
	store.UserFunc.FindUserByIDFunc = func(ctx context.Context, userId uuid.UUID) (*models.User, error) {
		return user, nil
	}
	store.TeamMemberFunc.CreateTeamMemberFunc = func(ctx context.Context, teamId, userId uuid.UUID, role models.TeamMemberRole, hasBillingAccess bool) (*models.TeamMember, error) {
		return &models.TeamMember{}, nil
	}
	store.TeamInvitationFunc.UpdateInvitationFunc = func(ctx context.Context, invitation *models.TeamInvitation) error {
		return nil
	}

	err := service.AcceptInvitation(ctx, userId, "token")
	assert.NoError(t, err)
	assert.Equal(t, models.TeamInvitationStatusAccepted, invitation.Status)
}

func TestInvitationService_AcceptInvitation_UserMismatch(t *testing.T) {
	ctx := context.Background()
	store := stores.NewAdapterDecorators()
	opts := conf.NewSettings()

	jobService := &JobServiceDecorator{}
	service := NewInvitationService(store, *opts, jobService)

	teamId := uuid.New()
	userId := uuid.New()
	invitation := &models.TeamInvitation{
		TeamID: teamId,
		Email:  "test@example.com",
		Status: models.TeamInvitationStatusPending,
	}
	user := &models.User{ID: userId, Email: "other@example.com"}
	store.TeamInvitationFunc.FindInvitationByTokenFunc = func(ctx context.Context, token string) (*models.TeamInvitation, error) {
		return invitation, nil
	}
	store.UserFunc.FindUserByIDFunc = func(ctx context.Context, userId uuid.UUID) (*models.User, error) {
		return user, nil
	}
	err := service.AcceptInvitation(ctx, userId, "token")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not match invitation")
}

func TestInvitationService_RejectInvitation(t *testing.T) {
	ctx := context.Background()
	store := stores.NewAdapterDecorators()
	opts := conf.NewSettings()

	jobService := &JobServiceDecorator{}
	service := NewInvitationService(store, *opts, jobService)

	teamId := uuid.New()
	userId := uuid.New()
	invitation := &models.TeamInvitation{
		TeamID: teamId,
		Email:  "test@example.com",
		Status: models.TeamInvitationStatusPending,
	}
	user := &models.User{ID: userId, Email: "test@example.com"}
	store.TeamInvitationFunc.FindInvitationByTokenFunc = func(ctx context.Context, token string) (*models.TeamInvitation, error) {
		return invitation, nil
	}
	store.UserFunc.FindUserByIDFunc = func(ctx context.Context, userId uuid.UUID) (*models.User, error) {
		return user, nil
	}
	err := service.RejectInvitation(ctx, userId, "token")
	assert.NoError(t, err)
	assert.Equal(t, models.TeamInvitationStatusDeclined, invitation.Status)
}

func TestInvitationService_FindInvitations(t *testing.T) {
	ctx := context.Background()
	store := stores.NewAdapterDecorators()
	opts := conf.NewSettings()

	jobService := &JobServiceDecorator{}
	service := NewInvitationService(store, *opts, jobService)

	teamId := uuid.New()
	invitations := []*models.TeamInvitation{{TeamID: teamId, Email: "test@example.com"}}
	store.TeamInvitationFunc.FindTeamInvitationsFunc = func(ctx context.Context, params *stores.TeamInvitationFilter) ([]*models.TeamInvitation, error) {
		return invitations, nil
	}

	result, err := service.FindInvitations(ctx, teamId)
	assert.NoError(t, err)
	assert.Equal(t, invitations, result)
}

func TestInvitationService_CancelInvitation_Success(t *testing.T) {
	ctx := context.Background()
	store := stores.NewAdapterDecorators()
	opts := conf.NewSettings()

	jobService := &JobServiceDecorator{}
	service := NewInvitationService(store, *opts, jobService)

	teamId := uuid.New()
	userId := uuid.New()
	invitationId := uuid.New()
	member := &models.TeamMember{
		ID:   uuid.New(),
		Role: models.TeamMemberRoleOwner,
	}
	invitation := &models.TeamInvitation{
		ID:     invitationId,
		TeamID: teamId,
		Status: models.TeamInvitationStatusPending,
	}
	store.TeamMemberFunc.FindTeamMemberFunc = func(ctx context.Context, filter *stores.TeamMemberFilter) (*models.TeamMember, error) {
		return member, nil
	}
	store.TeamInvitationFunc.FindInvitationByIDFunc = func(ctx context.Context, invitationId uuid.UUID) (*models.TeamInvitation, error) {
		return invitation, nil
	}
	store.TeamInvitationFunc.UpdateInvitationFunc = func(ctx context.Context, invitation *models.TeamInvitation) error {
		return nil
	}

	err := service.CancelInvitation(ctx, teamId, userId, invitationId)
	assert.NoError(t, err)
	assert.Equal(t, models.TeamInvitationStatusCanceled, invitation.Status)
}

func TestInvitationService_CancelInvitation_NotMember(t *testing.T) {
	ctx := context.Background()
	store := stores.NewAdapterDecorators()
	opts := conf.NewSettings()

	jobService := &JobServiceDecorator{}
	service := NewInvitationService(store, *opts, jobService)

	teamId := uuid.New()
	userId := uuid.New()
	invitationId := uuid.New()
	store.TeamMemberFunc.FindTeamMemberFunc = func(ctx context.Context, filter *stores.TeamMemberFilter) (*models.TeamMember, error) {
		return nil, nil
	}

	err := service.CancelInvitation(ctx, teamId, userId, invitationId)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not a member")
}

func TestInvitationService_CancelInvitation_NotOwner(t *testing.T) {
	ctx := context.Background()
	store := stores.NewAdapterDecorators()
	opts := conf.NewSettings()

	jobService := &JobServiceDecorator{}
	service := NewInvitationService(store, *opts, jobService)

	teamId := uuid.New()
	userId := uuid.New()
	invitationId := uuid.New()
	member := &models.TeamMember{
		ID:   uuid.New(),
		Role: models.TeamMemberRoleMember,
	}

	store.TeamMemberFunc.FindTeamMemberFunc = func(ctx context.Context, filter *stores.TeamMemberFilter) (*models.TeamMember, error) {
		return member, nil
	}

	err := service.CancelInvitation(ctx, teamId, userId, invitationId)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not an owner")
}

func TestInvitationService_CancelInvitation_InvitationNotFound(t *testing.T) {
	ctx := context.Background()
	store := stores.NewAdapterDecorators()
	opts := conf.NewSettings()

	jobService := &JobServiceDecorator{}
	service := NewInvitationService(store, *opts, jobService)

	teamId := uuid.New()
	userId := uuid.New()
	invitationId := uuid.New()
	member := &models.TeamMember{
		ID:   uuid.New(),
		Role: models.TeamMemberRoleOwner,
	}

	store.TeamMemberFunc.FindTeamMemberFunc = func(ctx context.Context, filter *stores.TeamMemberFilter) (*models.TeamMember, error) {
		return member, nil
	}
	store.TeamInvitationFunc.FindInvitationByIDFunc = func(ctx context.Context, invitationId uuid.UUID) (*models.TeamInvitation, error) {
		return nil, nil
	}

	err := service.CancelInvitation(ctx, teamId, userId, invitationId)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invitation not found")
}

func TestInvitationService_CancelInvitation_InvitationTeamMismatch(t *testing.T) {
	ctx := context.Background()
	store := stores.NewAdapterDecorators()
	opts := conf.NewSettings()

	jobService := &JobServiceDecorator{}
	service := NewInvitationService(store, *opts, jobService)

	teamId := uuid.New()
	userId := uuid.New()
	invitationId := uuid.New()
	member := &models.TeamMember{
		ID:   uuid.New(),
		Role: models.TeamMemberRoleOwner,
	}
	invitation := &models.TeamInvitation{
		ID:     invitationId,
		TeamID: uuid.New(), // Different team ID
		Status: models.TeamInvitationStatusPending,
	}

	store.TeamMemberFunc.FindTeamMemberFunc = func(ctx context.Context, filter *stores.TeamMemberFilter) (*models.TeamMember, error) {
		return member, nil
	}
	store.TeamInvitationFunc.FindInvitationByIDFunc = func(ctx context.Context, invitationId uuid.UUID) (*models.TeamInvitation, error) {
		return invitation, nil
	}

	err := service.CancelInvitation(ctx, teamId, userId, invitationId)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not match team")
}

func TestInvitationService_CancelInvitation_FindTeamMemberError(t *testing.T) {
	ctx := context.Background()
	store := stores.NewAdapterDecorators()
	opts := conf.NewSettings()

	jobService := &JobServiceDecorator{}
	service := NewInvitationService(store, *opts, jobService)

	teamId := uuid.New()
	userId := uuid.New()
	invitationId := uuid.New()

	store.TeamMemberFunc.FindTeamMemberFunc = func(ctx context.Context, filter *stores.TeamMemberFilter) (*models.TeamMember, error) {
		return nil, assert.AnError
	}

	err := service.CancelInvitation(ctx, teamId, userId, invitationId)
	assert.Error(t, err)
	assert.Equal(t, assert.AnError, err)
}

func TestInvitationService_CancelInvitation_FindInvitationError(t *testing.T) {
	ctx := context.Background()
	store := stores.NewAdapterDecorators()
	opts := conf.NewSettings()

	jobService := &JobServiceDecorator{}
	service := NewInvitationService(store, *opts, jobService)

	teamId := uuid.New()
	userId := uuid.New()
	invitationId := uuid.New()
	member := &models.TeamMember{
		ID:   uuid.New(),
		Role: models.TeamMemberRoleOwner,
	}

	store.TeamMemberFunc.FindTeamMemberFunc = func(ctx context.Context, filter *stores.TeamMemberFilter) (*models.TeamMember, error) {
		return member, nil
	}
	store.TeamInvitationFunc.FindInvitationByIDFunc = func(ctx context.Context, invitationId uuid.UUID) (*models.TeamInvitation, error) {
		return nil, assert.AnError
	}

	err := service.CancelInvitation(ctx, teamId, userId, invitationId)
	assert.Error(t, err)
	assert.Equal(t, assert.AnError, err)
}

func TestInvitationService_CancelInvitation_UpdateInvitationError(t *testing.T) {
	ctx := context.Background()
	store := stores.NewAdapterDecorators()
	opts := conf.NewSettings()

	jobService := &JobServiceDecorator{}
	service := NewInvitationService(store, *opts, jobService)

	teamId := uuid.New()
	userId := uuid.New()
	invitationId := uuid.New()
	member := &models.TeamMember{
		ID:   uuid.New(),
		Role: models.TeamMemberRoleOwner,
	}
	invitation := &models.TeamInvitation{
		ID:     invitationId,
		TeamID: teamId,
		Status: models.TeamInvitationStatusPending,
	}

	store.TeamMemberFunc.FindTeamMemberFunc = func(ctx context.Context, filter *stores.TeamMemberFilter) (*models.TeamMember, error) {
		return member, nil
	}
	store.TeamInvitationFunc.FindInvitationByIDFunc = func(ctx context.Context, invitationId uuid.UUID) (*models.TeamInvitation, error) {
		return invitation, nil
	}
	store.TeamInvitationFunc.UpdateInvitationFunc = func(ctx context.Context, invitation *models.TeamInvitation) error {
		return assert.AnError
	}

	err := service.CancelInvitation(ctx, teamId, userId, invitationId)
	assert.Error(t, err)
	assert.Equal(t, assert.AnError, err)
}
func TestInvitationService_CheckValidInvitation_Success(t *testing.T) {
	ctx := context.Background()
	store := stores.NewAdapterDecorators()
	opts := conf.NewSettings()

	jobService := &JobServiceDecorator{}
	service := NewInvitationService(store, *opts, jobService)

	userId := uuid.New()
	invitationToken := "token"
	invitation := &models.TeamInvitation{
		Email:  "test@example.com",
		Status: models.TeamInvitationStatusPending,
	}
	user := &models.User{
		ID:    userId,
		Email: "test@example.com",
	}

	store.TeamInvitationFunc.FindInvitationByTokenFunc = func(ctx context.Context, token string) (*models.TeamInvitation, error) {
		return invitation, nil
	}
	store.UserFunc.FindUserByIDFunc = func(ctx context.Context, userId uuid.UUID) (*models.User, error) {
		return user, nil
	}

	ok, err := service.CheckValidInvitation(ctx, userId, invitationToken)
	assert.True(t, ok)
	assert.NoError(t, err)
}

func TestInvitationService_CheckValidInvitation_InvitationNotFound(t *testing.T) {
	ctx := context.Background()
	store := stores.NewAdapterDecorators()
	opts := conf.NewSettings()

	jobService := &JobServiceDecorator{}
	service := NewInvitationService(store, *opts, jobService)

	userId := uuid.New()
	invitationToken := "token"

	store.TeamInvitationFunc.FindInvitationByTokenFunc = func(ctx context.Context, token string) (*models.TeamInvitation, error) {
		return nil, nil
	}

	ok, err := service.CheckValidInvitation(ctx, userId, invitationToken)
	assert.False(t, ok)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invitation not found")
}

func TestInvitationService_CheckValidInvitation_FindInvitationError(t *testing.T) {
	ctx := context.Background()
	store := stores.NewAdapterDecorators()
	opts := conf.NewSettings()

	jobService := &JobServiceDecorator{}
	service := NewInvitationService(store, *opts, jobService)

	userId := uuid.New()
	invitationToken := "token"

	store.TeamInvitationFunc.FindInvitationByTokenFunc = func(ctx context.Context, token string) (*models.TeamInvitation, error) {
		return nil, assert.AnError
	}

	ok, err := service.CheckValidInvitation(ctx, userId, invitationToken)
	assert.False(t, ok)
	assert.Equal(t, assert.AnError, err)
}

func TestInvitationService_CheckValidInvitation_UserNotFound(t *testing.T) {
	ctx := context.Background()
	store := stores.NewAdapterDecorators()
	opts := conf.NewSettings()

	jobService := &JobServiceDecorator{}
	service := NewInvitationService(store, *opts, jobService)

	userId := uuid.New()
	invitationToken := "token"
	invitation := &models.TeamInvitation{
		Email:  "test@example.com",
		Status: models.TeamInvitationStatusPending,
	}

	store.TeamInvitationFunc.FindInvitationByTokenFunc = func(ctx context.Context, token string) (*models.TeamInvitation, error) {
		return invitation, nil
	}
	store.UserFunc.FindUserByIDFunc = func(ctx context.Context, userId uuid.UUID) (*models.User, error) {
		return nil, nil
	}

	ok, err := service.CheckValidInvitation(ctx, userId, invitationToken)
	assert.False(t, ok)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user not found")
}

func TestInvitationService_CheckValidInvitation_FindUserError(t *testing.T) {
	ctx := context.Background()
	store := stores.NewAdapterDecorators()
	opts := conf.NewSettings()

	jobService := &JobServiceDecorator{}
	service := NewInvitationService(store, *opts, jobService)

	userId := uuid.New()
	invitationToken := "token"
	invitation := &models.TeamInvitation{
		Email:  "test@example.com",
		Status: models.TeamInvitationStatusPending,
	}

	store.TeamInvitationFunc.FindInvitationByTokenFunc = func(ctx context.Context, token string) (*models.TeamInvitation, error) {
		return invitation, nil
	}
	store.UserFunc.FindUserByIDFunc = func(ctx context.Context, userId uuid.UUID) (*models.User, error) {
		return nil, assert.AnError
	}

	ok, err := service.CheckValidInvitation(ctx, userId, invitationToken)
	assert.False(t, ok)
	assert.Equal(t, assert.AnError, err)
}

func TestInvitationService_CheckValidInvitation_UserEmailMismatch(t *testing.T) {
	ctx := context.Background()
	store := stores.NewAdapterDecorators()
	opts := conf.NewSettings()

	jobService := &JobServiceDecorator{}
	service := NewInvitationService(store, *opts, jobService)

	userId := uuid.New()
	invitationToken := "token"
	invitation := &models.TeamInvitation{
		Email:  "invitee@example.com",
		Status: models.TeamInvitationStatusPending,
	}
	user := &models.User{
		ID:    userId,
		Email: "other@example.com",
	}

	store.TeamInvitationFunc.FindInvitationByTokenFunc = func(ctx context.Context, token string) (*models.TeamInvitation, error) {
		return invitation, nil
	}
	store.UserFunc.FindUserByIDFunc = func(ctx context.Context, userId uuid.UUID) (*models.User, error) {
		return user, nil
	}

	ok, err := service.CheckValidInvitation(ctx, userId, invitationToken)
	assert.False(t, ok)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user does not match invitation")
}

func TestInvitationService_CheckValidInvitation_InvitationNotPending(t *testing.T) {
	ctx := context.Background()
	store := stores.NewAdapterDecorators()
	opts := conf.NewSettings()

	jobService := &JobServiceDecorator{}
	service := NewInvitationService(store, *opts, jobService)

	userId := uuid.New()
	invitationToken := "token"
	invitation := &models.TeamInvitation{
		Email:  "test@example.com",
		Status: models.TeamInvitationStatusAccepted,
	}
	user := &models.User{
		ID:    userId,
		Email: "test@example.com",
	}

	store.TeamInvitationFunc.FindInvitationByTokenFunc = func(ctx context.Context, token string) (*models.TeamInvitation, error) {
		return invitation, nil
	}
	store.UserFunc.FindUserByIDFunc = func(ctx context.Context, userId uuid.UUID) (*models.User, error) {
		return user, nil
	}

	ok, err := service.CheckValidInvitation(ctx, userId, invitationToken)
	assert.False(t, ok)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invitation is not pending")
}
func TestInvitationService_AcceptInvitation_Success(t *testing.T) {
	ctx := context.Background()
	store := stores.NewAdapterDecorators()
	opts := conf.NewSettings()

	jobService := &JobServiceDecorator{}
	service := NewInvitationService(store, *opts, jobService)

	teamId := uuid.New()
	userId := uuid.New()
	invitation := &models.TeamInvitation{
		TeamID: teamId,
		Email:  "test@example.com",
		Status: models.TeamInvitationStatusPending,
		Role:   models.TeamMemberRoleMember,
	}
	user := &models.User{ID: userId, Email: "test@example.com"}

	store.TeamInvitationFunc.FindInvitationByTokenFunc = func(ctx context.Context, token string) (*models.TeamInvitation, error) {
		return invitation, nil
	}
	store.UserFunc.FindUserByIDFunc = func(ctx context.Context, userId uuid.UUID) (*models.User, error) {
		return user, nil
	}
	store.TeamMemberFunc.CreateTeamMemberFunc = func(ctx context.Context, teamId, userId uuid.UUID, role models.TeamMemberRole, hasBillingAccess bool) (*models.TeamMember, error) {
		return &models.TeamMember{}, nil
	}
	store.TeamInvitationFunc.UpdateInvitationFunc = func(ctx context.Context, invitation *models.TeamInvitation) error {
		return nil
	}

	err := service.AcceptInvitation(ctx, userId, "token")
	assert.NoError(t, err)
	assert.Equal(t, models.TeamInvitationStatusAccepted, invitation.Status)
}

func TestInvitationService_AcceptInvitation_InvitationNotFound(t *testing.T) {
	ctx := context.Background()
	store := stores.NewAdapterDecorators()
	opts := conf.NewSettings()

	jobService := &JobServiceDecorator{}
	service := NewInvitationService(store, *opts, jobService)

	userId := uuid.New()
	store.TeamInvitationFunc.FindInvitationByTokenFunc = func(ctx context.Context, token string) (*models.TeamInvitation, error) {
		return nil, nil
	}

	err := service.AcceptInvitation(ctx, userId, "token")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invitation not found")
}

func TestInvitationService_AcceptInvitation_FindInvitationError(t *testing.T) {
	ctx := context.Background()
	store := stores.NewAdapterDecorators()
	opts := conf.NewSettings()

	jobService := &JobServiceDecorator{}
	service := NewInvitationService(store, *opts, jobService)

	userId := uuid.New()
	store.TeamInvitationFunc.FindInvitationByTokenFunc = func(ctx context.Context, token string) (*models.TeamInvitation, error) {
		return nil, assert.AnError
	}

	err := service.AcceptInvitation(ctx, userId, "token")
	assert.Error(t, err)
	assert.Equal(t, assert.AnError, err)
}

func TestInvitationService_AcceptInvitation_UserNotFound(t *testing.T) {
	ctx := context.Background()
	store := stores.NewAdapterDecorators()
	opts := conf.NewSettings()

	jobService := &JobServiceDecorator{}
	service := NewInvitationService(store, *opts, jobService)

	teamId := uuid.New()
	userId := uuid.New()
	invitation := &models.TeamInvitation{
		TeamID: teamId,
		Email:  "test@example.com",
		Status: models.TeamInvitationStatusPending,
	}

	store.TeamInvitationFunc.FindInvitationByTokenFunc = func(ctx context.Context, token string) (*models.TeamInvitation, error) {
		return invitation, nil
	}
	store.UserFunc.FindUserByIDFunc = func(ctx context.Context, userId uuid.UUID) (*models.User, error) {
		return nil, nil
	}

	err := service.AcceptInvitation(ctx, userId, "token")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user not found")
}

func TestInvitationService_AcceptInvitation_FindUserError(t *testing.T) {
	ctx := context.Background()
	store := stores.NewAdapterDecorators()
	opts := conf.NewSettings()

	jobService := &JobServiceDecorator{}
	service := NewInvitationService(store, *opts, jobService)

	teamId := uuid.New()
	userId := uuid.New()
	invitation := &models.TeamInvitation{
		TeamID: teamId,
		Email:  "test@example.com",
		Status: models.TeamInvitationStatusPending,
	}

	store.TeamInvitationFunc.FindInvitationByTokenFunc = func(ctx context.Context, token string) (*models.TeamInvitation, error) {
		return invitation, nil
	}
	store.UserFunc.FindUserByIDFunc = func(ctx context.Context, userId uuid.UUID) (*models.User, error) {
		return nil, assert.AnError
	}

	err := service.AcceptInvitation(ctx, userId, "token")
	assert.Error(t, err)
	assert.Equal(t, assert.AnError, err)
}

func TestInvitationService_AcceptInvitation_UserEmailMismatch(t *testing.T) {
	ctx := context.Background()
	store := stores.NewAdapterDecorators()
	opts := conf.NewSettings()

	jobService := &JobServiceDecorator{}
	service := NewInvitationService(store, *opts, jobService)

	teamId := uuid.New()
	userId := uuid.New()
	invitation := &models.TeamInvitation{
		TeamID: teamId,
		Email:  "invitee@example.com",
		Status: models.TeamInvitationStatusPending,
	}
	user := &models.User{ID: userId, Email: "other@example.com"}

	store.TeamInvitationFunc.FindInvitationByTokenFunc = func(ctx context.Context, token string) (*models.TeamInvitation, error) {
		return invitation, nil
	}
	store.UserFunc.FindUserByIDFunc = func(ctx context.Context, userId uuid.UUID) (*models.User, error) {
		return user, nil
	}

	err := service.AcceptInvitation(ctx, userId, "token")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user does not match invitation")
}

func TestInvitationService_AcceptInvitation_InvitationNotPending(t *testing.T) {
	ctx := context.Background()
	store := stores.NewAdapterDecorators()
	opts := conf.NewSettings()

	jobService := &JobServiceDecorator{}
	service := NewInvitationService(store, *opts, jobService)

	teamId := uuid.New()
	userId := uuid.New()
	invitation := &models.TeamInvitation{
		TeamID: teamId,
		Email:  "test@example.com",
		Status: models.TeamInvitationStatusAccepted,
	}
	user := &models.User{ID: userId, Email: "test@example.com"}

	store.TeamInvitationFunc.FindInvitationByTokenFunc = func(ctx context.Context, token string) (*models.TeamInvitation, error) {
		return invitation, nil
	}
	store.UserFunc.FindUserByIDFunc = func(ctx context.Context, userId uuid.UUID) (*models.User, error) {
		return user, nil
	}

	err := service.AcceptInvitation(ctx, userId, "token")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invitation is not pending")
}

// func TestInvitationService_AcceptInvitation_CreateTeamMemberError(t
