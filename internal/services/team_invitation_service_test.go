package services

import (
	"context"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/tkahng/authgo/internal/conf"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/stores"
	"github.com/tkahng/authgo/internal/tools/mailer"
)

func TestNewInvitationService(t *testing.T) {
	mockStore := stores.NewAdapterDecorators()
	opts := conf.NewSettings()
	wg := new(sync.WaitGroup)
	mockRoutineService := new(MockRoutineService)
	mockRoutineService.Wg = wg
	service := NewInvitationService(mockStore, NewMailService(&mailer.LogMailer{}), *opts, mockRoutineService)

	assert.NotNil(t, service, "NewInvitationService should not return nil")
}

func TestInvitationService_CreateInvitation(t *testing.T) {
	ctx := context.Background()
	store := stores.NewAdapterDecorators()
	opts := conf.NewSettings()
	mailService := NewMockMailService()
	wg := new(sync.WaitGroup)
	mockRoutineService := new(MockRoutineService)
	mockRoutineService.Wg = wg
	service := NewInvitationService(store, mailService, *opts, mockRoutineService)
	member := &models.TeamMember{ID: uuid.New()}
	inviteeEmail := "invitee@example.com"
	invitingUser := &models.User{ID: uuid.New(), Email: "inviting@example.com"}
	team := &models.Team{ID: uuid.New(), Name: "Test Team"}
	store.TeamMemberFunc.FindTeamMemberByTeamAndUserIdFunc = func(ctx context.Context, teamId, userId uuid.UUID) (*models.TeamMember, error) {
		return member, nil
	}
	// store.On("FindTeamMemberByTeamAndUserId", ctx, team.ID, invitingUser.ID).Return(member, nil)
	store.UserFunc.FindUserByIDFunc = func(ctx context.Context, userId uuid.UUID) (*models.User, error) {
		return invitingUser, nil
	}
	// store.On("FindUserByID", ctx, invitingUser.ID).Return(invitingUser, nil)
	store.TeamGroupFunc.FindTeamByIDFunc = func(ctx context.Context, teamId uuid.UUID) (*models.Team, error) {
		return team, nil
	}
	// store.On("FindTeamByID", ctx, team.ID).Return(team, nil)
	store.TeamInvitationFunc.FindPendingInvitationFunc = func(ctx context.Context, teamId uuid.UUID, email string) (*models.TeamInvitation, error) {
		return nil, nil
	}
	// store.On("FindPendingInvitation", ctx, team.ID, inviteeEmail).Return(nil, nil)
	store.TeamInvitationFunc.CreateInvitationFunc = func(ctx context.Context, invitation *models.TeamInvitation) error {
		return nil
	}
	// store.On("CreateInvitation", ctx, mock.AnythingOfType("*models.TeamInvitation")).Return(nil)

	err := service.CreateInvitation(ctx, team.ID, invitingUser.ID, inviteeEmail, models.TeamMemberRoleMember, true)
	wg.Wait()
	params := mailService.param
	assert.NotNil(t, params)
	assert.NoError(t, err)
	// store.AssertExpectations(t)
}

func TestInvitationService_CreateInvitation_NotMember(t *testing.T) {
	ctx := context.Background()
	store := stores.NewAdapterDecorators()
	opts := conf.NewSettings()
	mailService := NewMockMailService()
	wg := new(sync.WaitGroup)
	mockRoutineService := new(MockRoutineService)
	mockRoutineService.Wg = wg
	service := NewInvitationService(store, mailService, *opts, mockRoutineService)
	teamId := uuid.New()
	userId := uuid.New()
	store.TeamMemberFunc.FindTeamMemberByTeamAndUserIdFunc = func(ctx context.Context, teamId, userId uuid.UUID) (*models.TeamMember, error) {
		return nil, nil
	}
	// store.On("FindTeamMemberByTeamAndUserId", ctx, teamId, userId).Return((*models.TeamMember)(nil), nil)
	err := service.CreateInvitation(ctx, teamId, userId, "test@example.com", models.TeamMemberRoleMember, true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not a member")
}

func TestInvitationService_AcceptInvitation(t *testing.T) {
	ctx := context.Background()
	store := stores.NewAdapterDecorators()
	opts := conf.NewSettings()
	mailService := NewMockMailService()
	wg := new(sync.WaitGroup)
	mockRoutineService := new(MockRoutineService)
	mockRoutineService.Wg = wg
	// Mock the mail service
	service := NewInvitationService(store, mailService, *opts, mockRoutineService)
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
	// store.On("FindInvitationByToken", ctx, "token").Return(invitation, nil)
	store.UserFunc.FindUserByIDFunc = func(ctx context.Context, userId uuid.UUID) (*models.User, error) {
		return user, nil
	}
	store.TeamMemberFunc.CreateTeamMemberFunc = func(ctx context.Context, teamId, userId uuid.UUID, role models.TeamMemberRole, hasBillingAccess bool) (*models.TeamMember, error) {
		return &models.TeamMember{}, nil
	}
	// store.On("CreateTeamMember", ctx, teamId, userId, invitation.Role, false).Return(&models.TeamMember{}, nil)
	store.TeamInvitationFunc.UpdateInvitationFunc = func(ctx context.Context, invitation *models.TeamInvitation) error {
		return nil
	}

	err := service.AcceptInvitation(ctx, userId, "token")
	assert.NoError(t, err)
	assert.Equal(t, models.TeamInvitationStatusAccepted, invitation.Status)
}

// func TestInvitationService_AcceptInvitation_UserMismatch(t *testing.T) {
// 	ctx := context.Background()
// 	store := stores.NewAdapterDecorators()
// 	opts := conf.NewSettings()
// 	mailService := NewMockMailService()
// 	wg := new(sync.WaitGroup)
// 	mockRoutineService := new(MockRoutineService)
// 	mockRoutineService.Wg = wg
// 	service := NewInvitationService(store, mailService, *opts, mockRoutineService)
// 	teamId := uuid.New()
// 	userId := uuid.New()
// 	invitation := &models.TeamInvitation{
// 		TeamID: teamId,
// 		Email:  "test@example.com",
// 		Status: models.TeamInvitationStatusPending,
// 	}
// 	user := &models.User{ID: userId, Email: "other@example.com"}
// 	store.TeamInvitationFunc.FindInvitationByTokenFunc = func(ctx context.Context, token string) (*models.TeamInvitation, error) {
// 		return invitation, nil
// 	}
// 	// store.On("FindInvitationByToken", ctx, "token").Return(invitation, nil)
// 	store.UserFunc.FindUserByIDFunc = func(ctx context.Context, userId uuid.UUID) (*models.User, error) {
// 		return user, nil
// 	}
// 	err := service.AcceptInvitation(ctx, userId, "token")
// 	assert.Error(t, err)
// 	assert.Contains(t, err.Error(), "does not match invitation")
// }

// func TestInvitationService_RejectInvitation(t *testing.T) {
// 	ctx := context.Background()
// 	store := stores.NewAdapterDecorators()
// 	opts := conf.NewSettings()
// 	mailService := NewMockMailService()
// 	wg := new(sync.WaitGroup)
// 	mockRoutineService := new(MockRoutineService)
// 	mockRoutineService.Wg = wg
// 	service := NewInvitationService(store, mailService, *opts, mockRoutineService)
// 	teamId := uuid.New()
// 	userId := uuid.New()
// 	invitation := &models.TeamInvitation{
// 		TeamID: teamId,
// 		Email:  "test@example.com",
// 		Status: models.TeamInvitationStatusPending,
// 	}
// 	user := &models.User{ID: userId, Email: "test@example.com"}
// 	store.TeamInvitationFunc.FindInvitationByTokenFunc = func(ctx context.Context, token string) (*models.TeamInvitation, error) {
// 		return invitation, nil
// 	}
// 	// store.On("FindInvitationByToken", ctx, "token").Return(invitation, nil)
// 	store.UserFunc.FindUserByIDFunc = func(ctx context.Context, userId uuid.UUID) (*models.User, error) {
// 		return user, nil
// 	}
// 	err := service.RejectInvitation(ctx, userId, "token")
// 	assert.NoError(t, err)
// 	assert.Equal(t, models.TeamInvitationStatusDeclined, invitation.Status)
// }

// func TestInvitationService_FindInvitations(t *testing.T) {
// 	ctx := context.Background()
// 	store := stores.NewAdapterDecorators()
// 	opts := conf.NewSettings()
// 	mailService := NewMockMailService()
// 	wg := new(sync.WaitGroup)
// 	mockRoutineService := new(MockRoutineService)
// 	mockRoutineService.Wg = wg
// 	service := NewInvitationService(store, mailService, *opts, mockRoutineService)
// 	teamId := uuid.New()
// 	invitations := []*models.TeamInvitation{{TeamID: teamId, Email: "test@example.com"}}
// 	store.TeamInvitationFunc.FindTeamInvitationsFunc = func(ctx context.Context, teamId uuid.UUID) ([]*models.TeamInvitation, error) {
// 		return invitations, nil
// 	}
// 	// store.On("FindTeamInvitations", ctx, teamId).Return(invitations, nil)
// 	result, err := service.FindInvitations(ctx, teamId)
// 	assert.NoError(t, err)
// 	assert.Equal(t, invitations, result)
// }
// func TestInvitationService_CancelInvitation_Success(t *testing.T) {
// 	ctx := context.Background()
// 	store := stores.NewAdapterDecorators()
// 	opts := conf.NewSettings()
// 	mailService := NewMockMailService()
// 	wg := new(sync.WaitGroup)
// 	mockRoutineService := new(MockRoutineService)
// 	mockRoutineService.Wg = wg
// 	service := NewInvitationService(store, mailService, *opts, mockRoutineService)

// 	teamId := uuid.New()
// 	userId := uuid.New()
// 	invitationId := uuid.New()
// 	member := &models.TeamMember{
// 		ID:   uuid.New(),
// 		Role: models.TeamMemberRoleOwner,
// 	}
// 	invitation := &models.TeamInvitation{
// 		ID:     invitationId,
// 		TeamID: teamId,
// 		Status: models.TeamInvitationStatusPending,
// 	}

// 	store.On("FindTeamMemberByTeamAndUserId", ctx, teamId, userId).Return(member, nil)
// 	store.On("FindInvitationByID", ctx, invitationId).Return(invitation, nil)
// 	store.TeamInvitationFunc.UpdateInvitationFunc = func(ctx context.Context, invitation *models.TeamInvitation) error {
// 		return nil
// 	}

// 	err := service.CancelInvitation(ctx, teamId, userId, invitationId)
// 	assert.NoError(t, err)
// 	assert.Equal(t, models.TeamInvitationStatusCanceled, invitation.Status)
// 	store.AssertExpectations(t)
// }

// func TestInvitationService_CancelInvitation_NotMember(t *testing.T) {
// 	ctx := context.Background()
// 	store := stores.NewAdapterDecorators()
// 	opts := conf.NewSettings()
// 	mailService := NewMockMailService()
// 	wg := new(sync.WaitGroup)
// 	mockRoutineService := new(MockRoutineService)
// 	mockRoutineService.Wg = wg
// 	service := NewInvitationService(store, mailService, *opts, mockRoutineService)

// 	teamId := uuid.New()
// 	userId := uuid.New()
// 	invitationId := uuid.New()

// 	store.On("FindTeamMemberByTeamAndUserId", ctx, teamId, userId).Return((*models.TeamMember)(nil), nil)

// 	err := service.CancelInvitation(ctx, teamId, userId, invitationId)
// 	assert.Error(t, err)
// 	assert.Contains(t, err.Error(), "not a member")
// }

// func TestInvitationService_CancelInvitation_NotOwner(t *testing.T) {
// 	ctx := context.Background()
// 	store := stores.NewAdapterDecorators()
// 	opts := conf.NewSettings()
// 	mailService := NewMockMailService()
// 	wg := new(sync.WaitGroup)
// 	mockRoutineService := new(MockRoutineService)
// 	mockRoutineService.Wg = wg
// 	service := NewInvitationService(store, mailService, *opts, mockRoutineService)

// 	teamId := uuid.New()
// 	userId := uuid.New()
// 	invitationId := uuid.New()
// 	member := &models.TeamMember{
// 		ID:   uuid.New(),
// 		Role: models.TeamMemberRoleMember,
// 	}

// 	store.On("FindTeamMemberByTeamAndUserId", ctx, teamId, userId).Return(member, nil)

// 	err := service.CancelInvitation(ctx, teamId, userId, invitationId)
// 	assert.Error(t, err)
// 	assert.Contains(t, err.Error(), "not an owner")
// }

// func TestInvitationService_CancelInvitation_InvitationNotFound(t *testing.T) {
// 	ctx := context.Background()
// 	store := stores.NewAdapterDecorators()
// 	opts := conf.NewSettings()
// 	mailService := NewMockMailService()
// 	wg := new(sync.WaitGroup)
// 	mockRoutineService := new(MockRoutineService)
// 	mockRoutineService.Wg = wg
// 	service := NewInvitationService(store, mailService, *opts, mockRoutineService)

// 	teamId := uuid.New()
// 	userId := uuid.New()
// 	invitationId := uuid.New()
// 	member := &models.TeamMember{
// 		ID:   uuid.New(),
// 		Role: models.TeamMemberRoleOwner,
// 	}

// 	store.On("FindTeamMemberByTeamAndUserId", ctx, teamId, userId).Return(member, nil)
// 	store.On("FindInvitationByID", ctx, invitationId).Return((*models.TeamInvitation)(nil), nil)

// 	err := service.CancelInvitation(ctx, teamId, userId, invitationId)
// 	assert.Error(t, err)
// 	assert.Contains(t, err.Error(), "invitation not found")
// }

// func TestInvitationService_CancelInvitation_InvitationTeamMismatch(t *testing.T) {
// 	ctx := context.Background()
// 	store := stores.NewAdapterDecorators()
// 	opts := conf.NewSettings()
// 	mailService := NewMockMailService()
// 	wg := new(sync.WaitGroup)
// 	mockRoutineService := new(MockRoutineService)
// 	mockRoutineService.Wg = wg
// 	service := NewInvitationService(store, mailService, *opts, mockRoutineService)

// 	teamId := uuid.New()
// 	userId := uuid.New()
// 	invitationId := uuid.New()
// 	member := &models.TeamMember{
// 		ID:   uuid.New(),
// 		Role: models.TeamMemberRoleOwner,
// 	}
// 	invitation := &models.TeamInvitation{
// 		ID:     invitationId,
// 		TeamID: uuid.New(), // Different team ID
// 		Status: models.TeamInvitationStatusPending,
// 	}

// 	store.On("FindTeamMemberByTeamAndUserId", ctx, teamId, userId).Return(member, nil)
// 	store.On("FindInvitationByID", ctx, invitationId).Return(invitation, nil)

// 	err := service.CancelInvitation(ctx, teamId, userId, invitationId)
// 	assert.Error(t, err)
// 	assert.Contains(t, err.Error(), "does not match team")
// }

// func TestInvitationService_CancelInvitation_FindTeamMemberError(t *testing.T) {
// 	ctx := context.Background()
// 	store := stores.NewAdapterDecorators()
// 	opts := conf.NewSettings()
// 	mailService := NewMockMailService()
// 	wg := new(sync.WaitGroup)
// 	mockRoutineService := new(MockRoutineService)
// 	mockRoutineService.Wg = wg
// 	service := NewInvitationService(store, mailService, *opts, mockRoutineService)

// 	teamId := uuid.New()
// 	userId := uuid.New()
// 	invitationId := uuid.New()

// 	store.On("FindTeamMemberByTeamAndUserId", ctx, teamId, userId).Return((*models.TeamMember)(nil), assert.AnError)

// 	err := service.CancelInvitation(ctx, teamId, userId, invitationId)
// 	assert.Error(t, err)
// 	assert.Equal(t, assert.AnError, err)
// }

// func TestInvitationService_CancelInvitation_FindInvitationError(t *testing.T) {
// 	ctx := context.Background()
// 	store := stores.NewAdapterDecorators()
// 	opts := conf.NewSettings()
// 	mailService := NewMockMailService()
// 	wg := new(sync.WaitGroup)
// 	mockRoutineService := new(MockRoutineService)
// 	mockRoutineService.Wg = wg
// 	service := NewInvitationService(store, mailService, *opts, mockRoutineService)

// 	teamId := uuid.New()
// 	userId := uuid.New()
// 	invitationId := uuid.New()
// 	member := &models.TeamMember{
// 		ID:   uuid.New(),
// 		Role: models.TeamMemberRoleOwner,
// 	}

// 	store.On("FindTeamMemberByTeamAndUserId", ctx, teamId, userId).Return(member, nil)
// 	store.On("FindInvitationByID", ctx, invitationId).Return((*models.TeamInvitation)(nil), assert.AnError)

// 	err := service.CancelInvitation(ctx, teamId, userId, invitationId)
// 	assert.Error(t, err)
// 	assert.Equal(t, assert.AnError, err)
// }

// func TestInvitationService_CancelInvitation_UpdateInvitationError(t *testing.T) {
// 	ctx := context.Background()
// 	store := stores.NewAdapterDecorators()
// 	opts := conf.NewSettings()
// 	mailService := NewMockMailService()
// 	wg := new(sync.WaitGroup)
// 	mockRoutineService := new(MockRoutineService)
// 	mockRoutineService.Wg = wg
// 	service := NewInvitationService(store, mailService, *opts, mockRoutineService)

// 	teamId := uuid.New()
// 	userId := uuid.New()
// 	invitationId := uuid.New()
// 	member := &models.TeamMember{
// 		ID:   uuid.New(),
// 		Role: models.TeamMemberRoleOwner,
// 	}
// 	invitation := &models.TeamInvitation{
// 		ID:     invitationId,
// 		TeamID: teamId,
// 		Status: models.TeamInvitationStatusPending,
// 	}

// 	store.On("FindTeamMemberByTeamAndUserId", ctx, teamId, userId).Return(member, nil)
// 	store.On("FindInvitationByID", ctx, invitationId).Return(invitation, nil)
// 	store.On("UpdateInvitation", ctx, invitation).Return(assert.AnError)

// 	err := service.CancelInvitation(ctx, teamId, userId, invitationId)
// 	assert.Error(t, err)
// 	assert.Equal(t, assert.AnError, err)
// }
// func TestInvitationService_CheckValidInvitation_Success(t *testing.T) {
// 	ctx := context.Background()
// 	store := stores.NewAdapterDecorators()
// 	opts := conf.NewSettings()
// 	mailService := NewMockMailService()
// 	wg := new(sync.WaitGroup)
// 	mockRoutineService := new(MockRoutineService)
// 	mockRoutineService.Wg = wg
// 	service := NewInvitationService(store, mailService, *opts, mockRoutineService)

// 	userId := uuid.New()
// 	invitationToken := "token"
// 	invitation := &models.TeamInvitation{
// 		Email:  "test@example.com",
// 		Status: models.TeamInvitationStatusPending,
// 	}
// 	user := &models.User{
// 		ID:    userId,
// 		Email: "test@example.com",
// 	}

// 	store.On("FindInvitationByToken", ctx, invitationToken).Return(invitation, nil)
// 	store.UserFunc.FindUserByIDFunc = func(ctx context.Context, userId uuid.UUID) (*models.User, error) {
// 		return user, nil
// 	}

// 	ok, err := service.CheckValidInvitation(ctx, userId, invitationToken)
// 	assert.True(t, ok)
// 	assert.NoError(t, err)
// }

// func TestInvitationService_CheckValidInvitation_InvitationNotFound(t *testing.T) {
// 	ctx := context.Background()
// 	store := stores.NewAdapterDecorators()
// 	opts := conf.NewSettings()
// 	mailService := NewMockMailService()
// 	wg := new(sync.WaitGroup)
// 	mockRoutineService := new(MockRoutineService)
// 	mockRoutineService.Wg = wg
// 	service := NewInvitationService(store, mailService, *opts, mockRoutineService)

// 	userId := uuid.New()
// 	invitationToken := "token"

// 	store.On("FindInvitationByToken", ctx, invitationToken).Return((*models.TeamInvitation)(nil), nil)

// 	ok, err := service.CheckValidInvitation(ctx, userId, invitationToken)
// 	assert.False(t, ok)
// 	assert.Error(t, err)
// 	assert.Contains(t, err.Error(), "invitation not found")
// }

// func TestInvitationService_CheckValidInvitation_FindInvitationError(t *testing.T) {
// 	ctx := context.Background()
// 	store := stores.NewAdapterDecorators()
// 	opts := conf.NewSettings()
// 	mailService := NewMockMailService()
// 	wg := new(sync.WaitGroup)
// 	mockRoutineService := new(MockRoutineService)
// 	mockRoutineService.Wg = wg
// 	service := NewInvitationService(store, mailService, *opts, mockRoutineService)

// 	userId := uuid.New()
// 	invitationToken := "token"

// 	store.On("FindInvitationByToken", ctx, invitationToken).Return((*models.TeamInvitation)(nil), assert.AnError)

// 	ok, err := service.CheckValidInvitation(ctx, userId, invitationToken)
// 	assert.False(t, ok)
// 	assert.Equal(t, assert.AnError, err)
// }

// func TestInvitationService_CheckValidInvitation_UserNotFound(t *testing.T) {
// 	ctx := context.Background()
// 	store := stores.NewAdapterDecorators()
// 	opts := conf.NewSettings()
// 	mailService := NewMockMailService()
// 	wg := new(sync.WaitGroup)
// 	mockRoutineService := new(MockRoutineService)
// 	mockRoutineService.Wg = wg
// 	service := NewInvitationService(store, mailService, *opts, mockRoutineService)

// 	userId := uuid.New()
// 	invitationToken := "token"
// 	invitation := &models.TeamInvitation{
// 		Email:  "test@example.com",
// 		Status: models.TeamInvitationStatusPending,
// 	}

// 	store.On("FindInvitationByToken", ctx, invitationToken).Return(invitation, nil)
// 	store.On("FindUserByID", ctx, userId).Return((*models.User)(nil), nil)

// 	ok, err := service.CheckValidInvitation(ctx, userId, invitationToken)
// 	assert.False(t, ok)
// 	assert.Error(t, err)
// 	assert.Contains(t, err.Error(), "user not found")
// }

// func TestInvitationService_CheckValidInvitation_FindUserError(t *testing.T) {
// 	ctx := context.Background()
// 	store := stores.NewAdapterDecorators()
// 	opts := conf.NewSettings()
// 	mailService := NewMockMailService()
// 	wg := new(sync.WaitGroup)
// 	mockRoutineService := new(MockRoutineService)
// 	mockRoutineService.Wg = wg
// 	service := NewInvitationService(store, mailService, *opts, mockRoutineService)

// 	userId := uuid.New()
// 	invitationToken := "token"
// 	invitation := &models.TeamInvitation{
// 		Email:  "test@example.com",
// 		Status: models.TeamInvitationStatusPending,
// 	}

// 	store.On("FindInvitationByToken", ctx, invitationToken).Return(invitation, nil)
// 	store.On("FindUserByID", ctx, userId).Return((*models.User)(nil), assert.AnError)

// 	ok, err := service.CheckValidInvitation(ctx, userId, invitationToken)
// 	assert.False(t, ok)
// 	assert.Equal(t, assert.AnError, err)
// }

// func TestInvitationService_CheckValidInvitation_UserEmailMismatch(t *testing.T) {
// 	ctx := context.Background()
// 	store := stores.NewAdapterDecorators()
// 	opts := conf.NewSettings()
// 	mailService := NewMockMailService()
// 	wg := new(sync.WaitGroup)
// 	mockRoutineService := new(MockRoutineService)
// 	mockRoutineService.Wg = wg
// 	service := NewInvitationService(store, mailService, *opts, mockRoutineService)

// 	userId := uuid.New()
// 	invitationToken := "token"
// 	invitation := &models.TeamInvitation{
// 		Email:  "invitee@example.com",
// 		Status: models.TeamInvitationStatusPending,
// 	}
// 	user := &models.User{
// 		ID:    userId,
// 		Email: "other@example.com",
// 	}

// 	store.On("FindInvitationByToken", ctx, invitationToken).Return(invitation, nil)
// 	store.UserFunc.FindUserByIDFunc = func(ctx context.Context, userId uuid.UUID) (*models.User, error) {
// 		return user, nil
// 	}

// 	ok, err := service.CheckValidInvitation(ctx, userId, invitationToken)
// 	assert.False(t, ok)
// 	assert.Error(t, err)
// 	assert.Contains(t, err.Error(), "user does not match invitation")
// }

// func TestInvitationService_CheckValidInvitation_InvitationNotPending(t *testing.T) {
// 	ctx := context.Background()
// 	store := stores.NewAdapterDecorators()
// 	opts := conf.NewSettings()
// 	mailService := NewMockMailService()
// 	wg := new(sync.WaitGroup)
// 	mockRoutineService := new(MockRoutineService)
// 	mockRoutineService.Wg = wg
// 	service := NewInvitationService(store, mailService, *opts, mockRoutineService)

// 	userId := uuid.New()
// 	invitationToken := "token"
// 	invitation := &models.TeamInvitation{
// 		Email:  "test@example.com",
// 		Status: models.TeamInvitationStatusAccepted,
// 	}
// 	user := &models.User{
// 		ID:    userId,
// 		Email: "test@example.com",
// 	}

// 	store.On("FindInvitationByToken", ctx, invitationToken).Return(invitation, nil)
// 	store.UserFunc.FindUserByIDFunc = func(ctx context.Context, userId uuid.UUID) (*models.User, error) {
// 		return user, nil
// 	}

// 	ok, err := service.CheckValidInvitation(ctx, userId, invitationToken)
// 	assert.False(t, ok)
// 	assert.Error(t, err)
// 	assert.Contains(t, err.Error(), "invitation is not pending")
// }
// func TestInvitationService_AcceptInvitation_Success(t *testing.T) {
// 	ctx := context.Background()
// 	store := stores.NewAdapterDecorators()
// 	opts := conf.NewSettings()
// 	mailService := NewMockMailService()
// 	wg := new(sync.WaitGroup)
// 	mockRoutineService := new(MockRoutineService)
// 	mockRoutineService.Wg = wg
// 	service := NewInvitationService(store, mailService, *opts, mockRoutineService)

// 	teamId := uuid.New()
// 	userId := uuid.New()
// 	invitation := &models.TeamInvitation{
// 		TeamID: teamId,
// 		Email:  "test@example.com",
// 		Status: models.TeamInvitationStatusPending,
// 		Role:   models.TeamMemberRoleMember,
// 	}
// 	user := &models.User{ID: userId, Email: "test@example.com"}

// 	store.TeamInvitationFunc.FindInvitationByTokenFunc = func(ctx context.Context, token string) (*models.TeamInvitation, error) {
// 		return invitation, nil
// 	}
// 	// store.On("FindInvitationByToken", ctx, "token").Return(invitation, nil)
// 	store.UserFunc.FindUserByIDFunc = func(ctx context.Context, userId uuid.UUID) (*models.User, error) {
// 		return user, nil
// 	}
// 	store.TeamMemberFunc.CreateTeamMemberFunc = func(ctx context.Context, teamId, userId uuid.UUID, role models.TeamMemberRole, hasBillingAccess bool) (*models.TeamMember, error) {
// 		return &models.TeamMember{}, nil
// 	}
// 	store.TeamInvitationFunc.UpdateInvitationFunc = func(ctx context.Context, invitation *models.TeamInvitation) error {
// 		return nil
// 	}

// 	err := service.AcceptInvitation(ctx, userId, "token")
// 	assert.NoError(t, err)
// 	assert.Equal(t, models.TeamInvitationStatusAccepted, invitation.Status)
// }

// func TestInvitationService_AcceptInvitation_InvitationNotFound(t *testing.T) {
// 	ctx := context.Background()
// 	store := stores.NewAdapterDecorators()
// 	opts := conf.NewSettings()
// 	mailService := NewMockMailService()
// 	wg := new(sync.WaitGroup)
// 	mockRoutineService := new(MockRoutineService)
// 	mockRoutineService.Wg = wg
// 	service := NewInvitationService(store, mailService, *opts, mockRoutineService)

// 	userId := uuid.New()
// 	store.On("FindInvitationByToken", ctx, "token").Return((*models.TeamInvitation)(nil), nil)

// 	err := service.AcceptInvitation(ctx, userId, "token")
// 	assert.Error(t, err)
// 	assert.Contains(t, err.Error(), "invitation not found")
// }

// func TestInvitationService_AcceptInvitation_FindInvitationError(t *testing.T) {
// 	ctx := context.Background()
// 	store := stores.NewAdapterDecorators()
// 	opts := conf.NewSettings()
// 	mailService := NewMockMailService()
// 	wg := new(sync.WaitGroup)
// 	mockRoutineService := new(MockRoutineService)
// 	mockRoutineService.Wg = wg
// 	service := NewInvitationService(store, mailService, *opts, mockRoutineService)

// 	userId := uuid.New()
// 	store.On("FindInvitationByToken", ctx, "token").Return((*models.TeamInvitation)(nil), assert.AnError)

// 	err := service.AcceptInvitation(ctx, userId, "token")
// 	assert.Error(t, err)
// 	assert.Equal(t, assert.AnError, err)
// }

// func TestInvitationService_AcceptInvitation_UserNotFound(t *testing.T) {
// 	ctx := context.Background()
// 	store := stores.NewAdapterDecorators()
// 	opts := conf.NewSettings()
// 	mailService := NewMockMailService()
// 	wg := new(sync.WaitGroup)
// 	mockRoutineService := new(MockRoutineService)
// 	mockRoutineService.Wg = wg
// 	service := NewInvitationService(store, mailService, *opts, mockRoutineService)

// 	teamId := uuid.New()
// 	userId := uuid.New()
// 	invitation := &models.TeamInvitation{
// 		TeamID: teamId,
// 		Email:  "test@example.com",
// 		Status: models.TeamInvitationStatusPending,
// 	}

// 	store.TeamInvitationFunc.FindInvitationByTokenFunc = func(ctx context.Context, token string) (*models.TeamInvitation, error) {
// 		return invitation, nil
// 	}
// 	// store.On("FindInvitationByToken", ctx, "token").Return(invitation, nil)
// 	store.On("FindUserByID", ctx, userId).Return((*models.User)(nil), nil)

// 	err := service.AcceptInvitation(ctx, userId, "token")
// 	assert.Error(t, err)
// 	assert.Contains(t, err.Error(), "user not found")
// }

// func TestInvitationService_AcceptInvitation_FindUserError(t *testing.T) {
// 	ctx := context.Background()
// 	store := stores.NewAdapterDecorators()
// 	opts := conf.NewSettings()
// 	mailService := NewMockMailService()
// 	wg := new(sync.WaitGroup)
// 	mockRoutineService := new(MockRoutineService)
// 	mockRoutineService.Wg = wg
// 	service := NewInvitationService(store, mailService, *opts, mockRoutineService)

// 	teamId := uuid.New()
// 	userId := uuid.New()
// 	invitation := &models.TeamInvitation{
// 		TeamID: teamId,
// 		Email:  "test@example.com",
// 		Status: models.TeamInvitationStatusPending,
// 	}

// 	store.TeamInvitationFunc.FindInvitationByTokenFunc = func(ctx context.Context, token string) (*models.TeamInvitation, error) {
// 		return invitation, nil
// 	}
// 	// store.On("FindInvitationByToken", ctx, "token").Return(invitation, nil)
// 	store.On("FindUserByID", ctx, userId).Return((*models.User)(nil), assert.AnError)

// 	err := service.AcceptInvitation(ctx, userId, "token")
// 	assert.Error(t, err)
// 	assert.Equal(t, assert.AnError, err)
// }

// func TestInvitationService_AcceptInvitation_UserEmailMismatch(t *testing.T) {
// 	ctx := context.Background()
// 	store := stores.NewAdapterDecorators()
// 	opts := conf.NewSettings()
// 	mailService := NewMockMailService()
// 	wg := new(sync.WaitGroup)
// 	mockRoutineService := new(MockRoutineService)
// 	mockRoutineService.Wg = wg
// 	service := NewInvitationService(store, mailService, *opts, mockRoutineService)

// 	teamId := uuid.New()
// 	userId := uuid.New()
// 	invitation := &models.TeamInvitation{
// 		TeamID: teamId,
// 		Email:  "invitee@example.com",
// 		Status: models.TeamInvitationStatusPending,
// 	}
// 	user := &models.User{ID: userId, Email: "other@example.com"}

// 	store.TeamInvitationFunc.FindInvitationByTokenFunc = func(ctx context.Context, token string) (*models.TeamInvitation, error) {
// 		return invitation, nil
// 	}
// 	// store.On("FindInvitationByToken", ctx, "token").Return(invitation, nil)
// 	store.UserFunc.FindUserByIDFunc = func(ctx context.Context, userId uuid.UUID) (*models.User, error) {
// 		return user, nil
// 	}

// 	err := service.AcceptInvitation(ctx, userId, "token")
// 	assert.Error(t, err)
// 	assert.Contains(t, err.Error(), "user does not match invitation")
// }

// func TestInvitationService_AcceptInvitation_InvitationNotPending(t *testing.T) {
// 	ctx := context.Background()
// 	store := stores.NewAdapterDecorators()
// 	opts := conf.NewSettings()
// 	mailService := NewMockMailService()
// 	wg := new(sync.WaitGroup)
// 	mockRoutineService := new(MockRoutineService)
// 	mockRoutineService.Wg = wg
// 	service := NewInvitationService(store, mailService, *opts, mockRoutineService)

// 	teamId := uuid.New()
// 	userId := uuid.New()
// 	invitation := &models.TeamInvitation{
// 		TeamID: teamId,
// 		Email:  "test@example.com",
// 		Status: models.TeamInvitationStatusAccepted,
// 	}
// 	user := &models.User{ID: userId, Email: "test@example.com"}

// 	store.TeamInvitationFunc.FindInvitationByTokenFunc = func(ctx context.Context, token string) (*models.TeamInvitation, error) {
// 		return invitation, nil
// 	}
// 	// store.On("FindInvitationByToken", ctx, "token").Return(invitation, nil)
// 	store.UserFunc.FindUserByIDFunc = func(ctx context.Context, userId uuid.UUID) (*models.User, error) {
// 		return user, nil
// 	}

// 	err := service.AcceptInvitation(ctx, userId, "token")
// 	assert.Error(t, err)
// 	assert.Contains(t, err.Error(), "invitation is not pending")
// }

// func TestInvitationService_AcceptInvitation_CreateTeamMemberError(t *testing.T) {
// 	ctx := context.Background()
// 	store := stores.NewAdapterDecorators()
// 	opts := conf.NewSettings()
// 	mailService := NewMockMailService()
// 	wg := new(sync.WaitGroup)
// 	mockRoutineService := new(MockRoutineService)
// 	mockRoutineService.Wg = wg
// 	service := NewInvitationService(store, mailService, *opts, mockRoutineService)

// 	teamId := uuid.New()
// 	userId := uuid.New()
// 	invitation := &models.TeamInvitation{
// 		TeamID: teamId,
// 		Email:  "test@example.com",
// 		Status: models.TeamInvitationStatusPending,
// 		Role:   models.TeamMemberRoleMember,
// 	}
// 	user := &models.User{ID: userId, Email: "test@example.com"}

// 	store.TeamInvitationFunc.FindInvitationByTokenFunc = func(ctx context.Context, token string) (*models.TeamInvitation, error) {
// 		return invitation, nil
// 	}
// 	// store.On("FindInvitationByToken", ctx, "token").Return(invitation, nil)
// 	store.UserFunc.FindUserByIDFunc = func(ctx context.Context, userId uuid.UUID) (*models.User, error) {
// 		return user, nil
// 	}
// 	store.On("CreateTeamMember", ctx, teamId, userId, models.TeamMemberRoleMember, false).Return((*models.TeamMember)(nil), assert.AnError)

// 	err := service.AcceptInvitation(ctx, userId, "token")
// 	assert.Error(t, err)
// 	assert.Equal(t, assert.AnError, err)
// }

// func TestInvitationService_AcceptInvitation_UpdateInvitationError(t *testing.T) {
// 	ctx := context.Background()
// 	store := stores.NewAdapterDecorators()
// 	opts := conf.NewSettings()
// 	mailService := NewMockMailService()
// 	wg := new(sync.WaitGroup)
// 	mockRoutineService := new(MockRoutineService)
// 	mockRoutineService.Wg = wg
// 	service := NewInvitationService(store, mailService, *opts, mockRoutineService)

// 	teamId := uuid.New()
// 	userId := uuid.New()
// 	invitation := &models.TeamInvitation{
// 		TeamID: teamId,
// 		Email:  "test@example.com",
// 		Status: models.TeamInvitationStatusPending,
// 		Role:   models.TeamMemberRoleMember,
// 	}
// 	user := &models.User{ID: userId, Email: "test@example.com"}

// 	store.TeamInvitationFunc.FindInvitationByTokenFunc = func(ctx context.Context, token string) (*models.TeamInvitation, error) {
// 		return invitation, nil
// 	}
// 	// store.On("FindInvitationByToken", ctx, "token").Return(invitation, nil)
// 	store.UserFunc.FindUserByIDFunc = func(ctx context.Context, userId uuid.UUID) (*models.User, error) {
// 		return user, nil
// 	}
// 	store.TeamMemberFunc.CreateTeamMemberFunc = func(ctx context.Context, teamId, userId uuid.UUID, role models.TeamMemberRole, hasBillingAccess bool) (*models.TeamMember, error) {
// 		return &models.TeamMember{}, nil
// 	}
// 	store.On("UpdateInvitation", ctx, invitation).Return(assert.AnError)

// 	err := service.AcceptInvitation(ctx, userId, "token")
// 	assert.Error(t, err)
// 	assert.Equal(t, assert.AnError, err)
// }
// func TestInvitationService_CreateInvitation_NewInvite_Success(t *testing.T) {
// 	ctx := context.Background()
// 	store := stores.NewAdapterDecorators()
// 	opts := conf.NewSettings()
// 	mailService := NewMockMailService()
// 	wg := new(sync.WaitGroup)
// 	mockRoutineService := new(MockRoutineService)
// 	mockRoutineService.Wg = wg
// 	service := NewInvitationService(store, mailService, *opts, mockRoutineService)

// 	teamId := uuid.New()
// 	invitingUserId := uuid.New()
// 	inviteeEmail := "invitee@example.com"
// 	role := models.TeamMemberRoleMember

// 	member := &models.TeamMember{ID: uuid.New()}
// 	user := &models.User{ID: invitingUserId, Email: "inviter@example.com"}
// 	team := &models.Team{ID: teamId, Name: "TeamName"}

// 	store.On("FindTeamMemberByTeamAndUserId", ctx, teamId, invitingUserId).Return(member, nil)
// 	store.On("FindUserByID", ctx, invitingUserId).Return(user, nil)
// 	store.On("FindTeamByID", ctx, teamId).Return(team, nil)
// 	store.On("FindPendingInvitation", ctx, teamId, inviteeEmail).Return(nil, nil)
// 	store.On("CreateInvitation", ctx, mock.AnythingOfType("*models.TeamInvitation")).Return(nil)

// 	err := service.CreateInvitation(ctx, teamId, invitingUserId, inviteeEmail, role, false)
// 	wg.Wait()
// 	assert.NoError(t, err)
// 	store.AssertExpectations(t)
// }

// func TestInvitationService_CreateInvitation_ExistingInvite_Resend_Success(t *testing.T) {
// 	ctx := context.Background()
// 	store := stores.NewAdapterDecorators()
// 	opts := conf.NewSettings()
// 	mailService := NewMockMailService()
// 	wg := new(sync.WaitGroup)
// 	mockRoutineService := new(MockRoutineService)
// 	mockRoutineService.Wg = wg
// 	service := NewInvitationService(store, mailService, *opts, mockRoutineService)

// 	teamId := uuid.New()
// 	invitingUserId := uuid.New()
// 	inviteeEmail := "invitee@example.com"
// 	role := models.TeamMemberRoleMember

// 	member := &models.TeamMember{ID: uuid.New()}
// 	user := &models.User{ID: invitingUserId, Email: "inviter@example.com"}
// 	team := &models.Team{ID: teamId, Name: "TeamName"}
// 	existingInvite := &models.TeamInvitation{ID: uuid.New(), Email: inviteeEmail, TeamID: teamId}

// 	store.On("FindTeamMemberByTeamAndUserId", ctx, teamId, invitingUserId).Return(member, nil)
// 	store.On("FindUserByID", ctx, invitingUserId).Return(user, nil)
// 	store.On("FindTeamByID", ctx, teamId).Return(team, nil)
// 	store.On("FindPendingInvitation", ctx, teamId, inviteeEmail).Return(existingInvite, nil)
// 	store.On("UpdateInvitation", ctx, existingInvite).Return(nil)

// 	err := service.CreateInvitation(ctx, teamId, invitingUserId, inviteeEmail, role, true)
// 	wg.Wait()
// 	assert.NoError(t, err)
// 	store.AssertExpectations(t)
// }

// func TestInvitationService_CreateInvitation_ExistingInvite_NoResend_Error(t *testing.T) {
// 	ctx := context.Background()
// 	store := stores.NewAdapterDecorators()
// 	opts := conf.NewSettings()
// 	mailService := NewMockMailService()
// 	wg := new(sync.WaitGroup)
// 	mockRoutineService := new(MockRoutineService)
// 	mockRoutineService.Wg = wg
// 	service := NewInvitationService(store, mailService, *opts, mockRoutineService)

// 	teamId := uuid.New()
// 	invitingUserId := uuid.New()
// 	inviteeEmail := "invitee@example.com"
// 	role := models.TeamMemberRoleMember

// 	member := &models.TeamMember{ID: uuid.New()}
// 	user := &models.User{ID: invitingUserId, Email: "inviter@example.com"}
// 	team := &models.Team{ID: teamId, Name: "TeamName"}
// 	existingInvite := &models.TeamInvitation{ID: uuid.New(), Email: inviteeEmail, TeamID: teamId}

// 	store.On("FindTeamMemberByTeamAndUserId", ctx, teamId, invitingUserId).Return(member, nil)
// 	store.On("FindUserByID", ctx, invitingUserId).Return(user, nil)
// 	store.On("FindTeamByID", ctx, teamId).Return(team, nil)
// 	store.On("FindPendingInvitation", ctx, teamId, inviteeEmail).Return(existingInvite, nil)

// 	err := service.CreateInvitation(ctx, teamId, invitingUserId, inviteeEmail, role, false)
// 	assert.Error(t, err)
// 	assert.Contains(t, err.Error(), "invitation already exists")
// 	store.AssertExpectations(t)
// }

// func TestInvitationService_CreateInvitation_FindTeamMemberError(t *testing.T) {
// 	ctx := context.Background()
// 	store := stores.NewAdapterDecorators()
// 	opts := conf.NewSettings()
// 	mailService := NewMockMailService()
// 	wg := new(sync.WaitGroup)
// 	mockRoutineService := new(MockRoutineService)
// 	mockRoutineService.Wg = wg
// 	service := NewInvitationService(store, mailService, *opts, mockRoutineService)

// 	teamId := uuid.New()
// 	invitingUserId := uuid.New()
// 	inviteeEmail := "invitee@example.com"
// 	role := models.TeamMemberRoleMember

// 	store.On("FindTeamMemberByTeamAndUserId", ctx, teamId, invitingUserId).Return((*models.TeamMember)(nil), assert.AnError)

// 	err := service.CreateInvitation(ctx, teamId, invitingUserId, inviteeEmail, role, false)
// 	assert.Error(t, err)
// 	assert.Equal(t, assert.AnError, err)
// }

// func TestInvitationService_CreateInvitation_UserNotMember_Error(t *testing.T) {
// 	ctx := context.Background()
// 	store := stores.NewAdapterDecorators()
// 	opts := conf.NewSettings()
// 	mailService := NewMockMailService()
// 	wg := new(sync.WaitGroup)
// 	mockRoutineService := new(MockRoutineService)
// 	mockRoutineService.Wg = wg
// 	service := NewInvitationService(store, mailService, *opts, mockRoutineService)

// 	teamId := uuid.New()
// 	invitingUserId := uuid.New()
// 	inviteeEmail := "invitee@example.com"
// 	role := models.TeamMemberRoleMember

// 	store.On("FindTeamMemberByTeamAndUserId", ctx, teamId, invitingUserId).Return((*models.TeamMember)(nil), nil)

// 	err := service.CreateInvitation(ctx, teamId, invitingUserId, inviteeEmail, role, false)
// 	assert.Error(t, err)
// 	assert.Contains(t, err.Error(), "not a member")
// }

// func TestInvitationService_CreateInvitation_FindUserByID_Error(t *testing.T) {
// 	ctx := context.Background()
// 	store := stores.NewAdapterDecorators()
// 	opts := conf.NewSettings()
// 	mailService := NewMockMailService()
// 	wg := new(sync.WaitGroup)
// 	mockRoutineService := new(MockRoutineService)
// 	mockRoutineService.Wg = wg
// 	service := NewInvitationService(store, mailService, *opts, mockRoutineService)

// 	teamId := uuid.New()
// 	invitingUserId := uuid.New()
// 	inviteeEmail := "invitee@example.com"
// 	role := models.TeamMemberRoleMember

// 	member := &models.TeamMember{ID: uuid.New()}
// 	store.On("FindTeamMemberByTeamAndUserId", ctx, teamId, invitingUserId).Return(member, nil)
// 	store.On("FindUserByID", ctx, invitingUserId).Return((*models.User)(nil), assert.AnError)

// 	err := service.CreateInvitation(ctx, teamId, invitingUserId, inviteeEmail, role, false)
// 	assert.Error(t, err)
// 	assert.Equal(t, assert.AnError, err)
// }

// func TestInvitationService_CreateInvitation_UserNotFound_Error(t *testing.T) {
// 	ctx := context.Background()
// 	store := stores.NewAdapterDecorators()
// 	opts := conf.NewSettings()
// 	mailService := NewMockMailService()
// 	wg := new(sync.WaitGroup)
// 	mockRoutineService := new(MockRoutineService)
// 	mockRoutineService.Wg = wg
// 	service := NewInvitationService(store, mailService, *opts, mockRoutineService)

// 	teamId := uuid.New()
// 	invitingUserId := uuid.New()
// 	inviteeEmail := "invitee@example.com"
// 	role := models.TeamMemberRoleMember

// 	member := &models.TeamMember{ID: uuid.New()}
// 	store.On("FindTeamMemberByTeamAndUserId", ctx, teamId, invitingUserId).Return(member, nil)
// 	store.On("FindUserByID", ctx, invitingUserId).Return((*models.User)(nil), nil)

// 	err := service.CreateInvitation(ctx, teamId, invitingUserId, inviteeEmail, role, false)
// 	assert.Error(t, err)
// 	assert.Contains(t, err.Error(), "user not found")
// }

// func TestInvitationService_CreateInvitation_FindTeamByID_Error(t *testing.T) {
// 	ctx := context.Background()
// 	store := stores.NewAdapterDecorators()
// 	opts := conf.NewSettings()
// 	mailService := NewMockMailService()
// 	wg := new(sync.WaitGroup)
// 	mockRoutineService := new(MockRoutineService)
// 	mockRoutineService.Wg = wg
// 	service := NewInvitationService(store, mailService, *opts, mockRoutineService)

// 	teamId := uuid.New()
// 	invitingUserId := uuid.New()
// 	inviteeEmail := "invitee@example.com"
// 	role := models.TeamMemberRoleMember

// 	member := &models.TeamMember{ID: uuid.New()}
// 	user := &models.User{ID: invitingUserId, Email: "inviter@example.com"}
// 	store.On("FindTeamMemberByTeamAndUserId", ctx, teamId, invitingUserId).Return(member, nil)
// 	store.On("FindUserByID", ctx, invitingUserId).Return(user, nil)
// 	store.On("FindTeamByID", ctx, teamId).Return((*models.Team)(nil), assert.AnError)

// 	err := service.CreateInvitation(ctx, teamId, invitingUserId, inviteeEmail, role, false)
// 	assert.Error(t, err)
// 	assert.Equal(t, assert.AnError, err)
// }

// func TestInvitationService_CreateInvitation_TeamNotFound_Error(t *testing.T) {
// 	ctx := context.Background()
// 	store := stores.NewAdapterDecorators()
// 	opts := conf.NewSettings()
// 	mailService := NewMockMailService()
// 	wg := new(sync.WaitGroup)
// 	mockRoutineService := new(MockRoutineService)
// 	mockRoutineService.Wg = wg
// 	service := NewInvitationService(store, mailService, *opts, mockRoutineService)

// 	teamId := uuid.New()
// 	invitingUserId := uuid.New()
// 	inviteeEmail := "invitee@example.com"
// 	role := models.TeamMemberRoleMember

// 	member := &models.TeamMember{ID: uuid.New()}
// 	user := &models.User{ID: invitingUserId, Email: "inviter@example.com"}
// 	store.On("FindTeamMemberByTeamAndUserId", ctx, teamId, invitingUserId).Return(member, nil)
// 	store.On("FindUserByID", ctx, invitingUserId).Return(user, nil)
// 	store.On("FindTeamByID", ctx, teamId).Return((*models.Team)(nil), nil)

// 	err := service.CreateInvitation(ctx, teamId, invitingUserId, inviteeEmail, role, false)
// 	assert.Error(t, err)
// 	assert.Contains(t, err.Error(), "team not found")
// }

// func TestInvitationService_CreateInvitation_FindPendingInvitation_Error(t *testing.T) {
// 	ctx := context.Background()
// 	store := stores.NewAdapterDecorators()
// 	opts := conf.NewSettings()
// 	mailService := NewMockMailService()
// 	wg := new(sync.WaitGroup)
// 	mockRoutineService := new(MockRoutineService)
// 	mockRoutineService.Wg = wg
// 	service := NewInvitationService(store, mailService, *opts, mockRoutineService)

// 	teamId := uuid.New()
// 	invitingUserId := uuid.New()
// 	inviteeEmail := "invitee@example.com"
// 	role := models.TeamMemberRoleMember

// 	member := &models.TeamMember{ID: uuid.New()}
// 	user := &models.User{ID: invitingUserId, Email: "inviter@example.com"}
// 	team := &models.Team{ID: teamId, Name: "TeamName"}

// 	store.On("FindTeamMemberByTeamAndUserId", ctx, teamId, invitingUserId).Return(member, nil)
// 	store.On("FindUserByID", ctx, invitingUserId).Return(user, nil)
// 	store.On("FindTeamByID", ctx, teamId).Return(team, nil)
// 	store.On("FindPendingInvitation", ctx, teamId, inviteeEmail).Return((*models.TeamInvitation)(nil), assert.AnError)

// 	err := service.CreateInvitation(ctx, teamId, invitingUserId, inviteeEmail, role, false)
// 	assert.Error(t, err)
// 	assert.Equal(t, assert.AnError, err)
// }

// func TestInvitationService_CreateInvitation_CreateInvitation_Error(t *testing.T) {
// 	ctx := context.Background()
// 	store := stores.NewAdapterDecorators()
// 	opts := conf.NewSettings()
// 	mailService := NewMockMailService()
// 	wg := new(sync.WaitGroup)
// 	mockRoutineService := new(MockRoutineService)
// 	mockRoutineService.Wg = wg
// 	service := NewInvitationService(store, mailService, *opts, mockRoutineService)

// 	teamId := uuid.New()
// 	invitingUserId := uuid.New()
// 	inviteeEmail := "invitee@example.com"
// 	role := models.TeamMemberRoleMember

// 	member := &models.TeamMember{ID: uuid.New()}
// 	user := &models.User{ID: invitingUserId, Email: "inviter@example.com"}
// 	team := &models.Team{ID: teamId, Name: "TeamName"}

// 	store.On("FindTeamMemberByTeamAndUserId", ctx, teamId, invitingUserId).Return(member, nil)
// 	store.On("FindUserByID", ctx, invitingUserId).Return(user, nil)
// 	store.On("FindTeamByID", ctx, teamId).Return(team, nil)
// 	store.On("FindPendingInvitation", ctx, teamId, inviteeEmail).Return(nil, nil)
// 	store.On("CreateInvitation", ctx, mock.AnythingOfType("*models.TeamInvitation")).Return(assert.AnError)

// 	err := service.CreateInvitation(ctx, teamId, invitingUserId, inviteeEmail, role, false)
// 	assert.Error(t, err)
// 	assert.Equal(t, assert.AnError, err)
// }

// func TestInvitationService_CreateInvitation_UpdateInvitation_Error(t *testing.T) {
// 	ctx := context.Background()
// 	store := stores.NewAdapterDecorators()
// 	opts := conf.NewSettings()
// 	mailService := NewMockMailService()
// 	wg := new(sync.WaitGroup)
// 	mockRoutineService := new(MockRoutineService)
// 	mockRoutineService.Wg = wg
// 	service := NewInvitationService(store, mailService, *opts, mockRoutineService)

// 	teamId := uuid.New()
// 	invitingUserId := uuid.New()
// 	inviteeEmail := "invitee@example.com"
// 	role := models.TeamMemberRoleMember

// 	member := &models.TeamMember{ID: uuid.New()}
// 	user := &models.User{ID: invitingUserId, Email: "inviter@example.com"}
// 	team := &models.Team{ID: teamId, Name: "TeamName"}
// 	existingInvite := &models.TeamInvitation{ID: uuid.New(), Email: inviteeEmail, TeamID: teamId}

// 	store.On("FindTeamMemberByTeamAndUserId", ctx, teamId, invitingUserId).Return(member, nil)
// 	store.On("FindUserByID", ctx, invitingUserId).Return(user, nil)
// 	store.On("FindTeamByID", ctx, teamId).Return(team, nil)
// 	store.On("FindPendingInvitation", ctx, teamId, inviteeEmail).Return(existingInvite, nil)
// 	store.On("UpdateInvitation", ctx, existingInvite).Return(assert.AnError)

// 	err := service.CreateInvitation(ctx, teamId, invitingUserId, inviteeEmail, role, true)
// 	assert.Error(t, err)
// 	assert.Equal(t, assert.AnError, err)
// }
// func TestInvitationService_FindInvitations_Success(t *testing.T) {
// 	ctx := context.Background()
// 	store := stores.NewAdapterDecorators()
// 	opts := conf.NewSettings()
// 	mailService := NewMockMailService()
// 	wg := new(sync.WaitGroup)
// 	mockRoutineService := new(MockRoutineService)
// 	mockRoutineService.Wg = wg
// 	service := NewInvitationService(store, mailService, *opts, mockRoutineService)

// 	teamId := uuid.New()
// 	expectedInvitations := []*models.TeamInvitation{
// 		{TeamID: teamId, Email: "user1@example.com"},
// 		{TeamID: teamId, Email: "user2@example.com"},
// 	}
// 	store.On("FindTeamInvitations", ctx, teamId).Return(expectedInvitations, nil)

// 	invitations, err := service.FindInvitations(ctx, teamId)
// 	assert.NoError(t, err)
// 	assert.Equal(t, expectedInvitations, invitations)
// 	store.AssertExpectations(t)
// }

// func TestInvitationService_FindInvitations_Error(t *testing.T) {
// 	ctx := context.Background()
// 	store := stores.NewAdapterDecorators()
// 	opts := conf.NewSettings()
// 	mailService := NewMockMailService()
// 	wg := new(sync.WaitGroup)
// 	mockRoutineService := new(MockRoutineService)
// 	mockRoutineService.Wg = wg
// 	service := NewInvitationService(store, mailService, *opts, mockRoutineService)

// 	teamId := uuid.New()
// 	store.On("FindTeamInvitations", ctx, teamId).Return(nil, assert.AnError)

// 	invitations, err := service.FindInvitations(ctx, teamId)
// 	assert.Error(t, err)
// 	assert.Nil(t, invitations)
// 	assert.Equal(t, assert.AnError, err)
// 	store.AssertExpectations(t)
// }
// func TestInvitationService_RejectInvitation_Success(t *testing.T) {
// 	ctx := context.Background()
// 	store := stores.NewAdapterDecorators()
// 	opts := conf.NewSettings()
// 	mailService := NewMockMailService()
// 	wg := new(sync.WaitGroup)
// 	mockRoutineService := new(MockRoutineService)
// 	mockRoutineService.Wg = wg
// 	service := NewInvitationService(store, mailService, *opts, mockRoutineService)

// 	userId := uuid.New()
// 	invitationToken := "token"
// 	invitation := &models.TeamInvitation{
// 		Email:  "test@example.com",
// 		Status: models.TeamInvitationStatusPending,
// 	}
// 	user := &models.User{
// 		ID:    userId,
// 		Email: "test@example.com",
// 	}

// 	store.On("FindInvitationByToken", ctx, invitationToken).Return(invitation, nil)
// 	store.UserFunc.FindUserByIDFunc = func(ctx context.Context, userId uuid.UUID) (*models.User, error) {
// 		return user, nil
// 	}

// 	err := service.RejectInvitation(ctx, userId, invitationToken)
// 	assert.NoError(t, err)
// 	assert.Equal(t, models.TeamInvitationStatusDeclined, invitation.Status)
// 	store.AssertExpectations(t)
// }

// func TestInvitationService_RejectInvitation_InvitationNotFound(t *testing.T) {
// 	ctx := context.Background()
// 	store := stores.NewAdapterDecorators()
// 	opts := conf.NewSettings()
// 	mailService := NewMockMailService()
// 	wg := new(sync.WaitGroup)
// 	mockRoutineService := new(MockRoutineService)
// 	mockRoutineService.Wg = wg
// 	service := NewInvitationService(store, mailService, *opts, mockRoutineService)

// 	userId := uuid.New()
// 	invitationToken := "token"

// 	store.On("FindInvitationByToken", ctx, invitationToken).Return((*models.TeamInvitation)(nil), nil)

// 	err := service.RejectInvitation(ctx, userId, invitationToken)
// 	assert.Error(t, err)
// 	assert.Contains(t, err.Error(), "invitation not found")
// 	store.AssertExpectations(t)
// }

// func TestInvitationService_RejectInvitation_FindInvitationError(t *testing.T) {
// 	ctx := context.Background()
// 	store := stores.NewAdapterDecorators()
// 	opts := conf.NewSettings()
// 	mailService := NewMockMailService()
// 	wg := new(sync.WaitGroup)
// 	mockRoutineService := new(MockRoutineService)
// 	mockRoutineService.Wg = wg
// 	service := NewInvitationService(store, mailService, *opts, mockRoutineService)

// 	userId := uuid.New()
// 	invitationToken := "token"

// 	store.On("FindInvitationByToken", ctx, invitationToken).Return((*models.TeamInvitation)(nil), assert.AnError)

// 	err := service.RejectInvitation(ctx, userId, invitationToken)
// 	assert.Error(t, err)
// 	assert.Equal(t, assert.AnError, err)
// 	store.AssertExpectations(t)
// }

// func TestInvitationService_RejectInvitation_UserNotFound(t *testing.T) {
// 	ctx := context.Background()
// 	store := stores.NewAdapterDecorators()
// 	opts := conf.NewSettings()
// 	mailService := NewMockMailService()
// 	wg := new(sync.WaitGroup)
// 	mockRoutineService := new(MockRoutineService)
// 	mockRoutineService.Wg = wg
// 	service := NewInvitationService(store, mailService, *opts, mockRoutineService)

// 	userId := uuid.New()
// 	invitationToken := "token"
// 	invitation := &models.TeamInvitation{
// 		Email:  "test@example.com",
// 		Status: models.TeamInvitationStatusPending,
// 	}

// 	store.On("FindInvitationByToken", ctx, invitationToken).Return(invitation, nil)
// 	store.On("FindUserByID", ctx, userId).Return((*models.User)(nil), nil)

// 	err := service.RejectInvitation(ctx, userId, invitationToken)
// 	assert.Error(t, err)
// 	assert.Contains(t, err.Error(), "user not found")
// 	store.AssertExpectations(t)
// }

// func TestInvitationService_RejectInvitation_FindUserError(t *testing.T) {
// 	ctx := context.Background()
// 	store := stores.NewAdapterDecorators()
// 	opts := conf.NewSettings()
// 	mailService := NewMockMailService()
// 	wg := new(sync.WaitGroup)
// 	mockRoutineService := new(MockRoutineService)
// 	mockRoutineService.Wg = wg
// 	service := NewInvitationService(store, mailService, *opts, mockRoutineService)

// 	userId := uuid.New()
// 	invitationToken := "token"
// 	invitation := &models.TeamInvitation{
// 		Email:  "test@example.com",
// 		Status: models.TeamInvitationStatusPending,
// 	}

// 	store.On("FindInvitationByToken", ctx, invitationToken).Return(invitation, nil)
// 	store.On("FindUserByID", ctx, userId).Return((*models.User)(nil), assert.AnError)

// 	err := service.RejectInvitation(ctx, userId, invitationToken)
// 	assert.Error(t, err)
// 	assert.Equal(t, assert.AnError, err)
// 	store.AssertExpectations(t)
// }

// func TestInvitationService_RejectInvitation_UserEmailMismatch(t *testing.T) {
// 	ctx := context.Background()
// 	store := stores.NewAdapterDecorators()
// 	opts := conf.NewSettings()
// 	mailService := NewMockMailService()
// 	wg := new(sync.WaitGroup)
// 	mockRoutineService := new(MockRoutineService)
// 	mockRoutineService.Wg = wg
// 	service := NewInvitationService(store, mailService, *opts, mockRoutineService)

// 	userId := uuid.New()
// 	invitationToken := "token"
// 	invitation := &models.TeamInvitation{
// 		Email:  "invitee@example.com",
// 		Status: models.TeamInvitationStatusPending,
// 	}
// 	user := &models.User{
// 		ID:    userId,
// 		Email: "other@example.com",
// 	}

// 	store.On("FindInvitationByToken", ctx, invitationToken).Return(invitation, nil)
// 	store.UserFunc.FindUserByIDFunc = func(ctx context.Context, userId uuid.UUID) (*models.User, error) {
// 		return user, nil
// 	}

// 	err := service.RejectInvitation(ctx, userId, invitationToken)
// 	assert.Error(t, err)
// 	assert.Contains(t, err.Error(), "user does not match invitation")
// 	store.AssertExpectations(t)
// }
