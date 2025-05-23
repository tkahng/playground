package services

import (
	"context"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/tkahng/authgo/internal/conf"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/tools/mailer"
)

func TestNewInvitationService(t *testing.T) {
	mockStore := &mockTeamInvitationStore{}
	opts := conf.NewSettings()
	wg := new(sync.WaitGroup)
	mockRoutineService := new(mockRoutineService)
	mockRoutineService.wg = wg
	service := NewInvitationService(mockStore, NewMailService(&mailer.LogMailer{}), *opts, mockRoutineService)

	assert.NotNil(t, service, "NewInvitationService should not return nil")
}

func TestInvitationService_CreateInvitation(t *testing.T) {
	ctx := context.Background()
	store := new(mockTeamInvitationStore)
	opts := conf.NewSettings()
	mailService := NewMockMailService()
	wg := new(sync.WaitGroup)
	mockRoutineService := new(mockRoutineService)
	mockRoutineService.wg = wg
	service := NewInvitationService(store, mailService, *opts, mockRoutineService)
	member := &models.TeamMember{ID: uuid.New()}
	inviteeEmail := "invitee@example.com"
	invitingUser := &models.User{ID: uuid.New(), Email: "inviting@example.com"}
	team := &models.Team{ID: uuid.New(), Name: "Test Team"}
	store.On("FindTeamMemberByTeamAndUserId", ctx, team.ID, invitingUser.ID).Return(member, nil)
	store.On("FindUserByID", ctx, invitingUser.ID).Return(invitingUser, nil)
	store.On("FindTeamByID", ctx, team.ID).Return(team, nil)
	store.On("FindPendingInvitation", ctx, team.ID, inviteeEmail).Return(nil, nil)
	store.On("CreateInvitation", ctx, mock.AnythingOfType("*models.TeamInvitation")).Return(nil)

	err := service.CreateInvitation(ctx, team.ID, invitingUser.ID, inviteeEmail, models.TeamMemberRoleMember, true)
	wg.Wait()
	params := mailService.param
	assert.NotNil(t, params)
	assert.NoError(t, err)
	store.AssertExpectations(t)
}

func TestInvitationService_CreateInvitation_NotMember(t *testing.T) {
	ctx := context.Background()
	store := new(mockTeamInvitationStore)
	opts := conf.NewSettings()
	mailService := NewMockMailService()
	wg := new(sync.WaitGroup)
	mockRoutineService := new(mockRoutineService)
	mockRoutineService.wg = wg
	service := NewInvitationService(store, mailService, *opts, mockRoutineService)
	teamId := uuid.New()
	userId := uuid.New()
	store.On("FindTeamMemberByTeamAndUserId", ctx, teamId, userId).Return((*models.TeamMember)(nil), nil)
	err := service.CreateInvitation(ctx, teamId, userId, "test@example.com", models.TeamMemberRoleMember, true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not a member")
}

func TestInvitationService_AcceptInvitation(t *testing.T) {
	ctx := context.Background()
	store := new(mockTeamInvitationStore)
	opts := conf.NewSettings()
	mailService := NewMockMailService()
	wg := new(sync.WaitGroup)
	mockRoutineService := new(mockRoutineService)
	mockRoutineService.wg = wg
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
	store.On("FindInvitationByToken", ctx, "token").Return(invitation, nil)
	store.On("FindUserByID", ctx, userId).Return(user, nil)
	store.On("CreateTeamMember", ctx, teamId, userId, models.TeamMemberRoleMember, false).Return(&models.TeamMember{}, nil)
	store.On("UpdateInvitation", ctx, invitation).Return(nil)
	err := service.AcceptInvitation(ctx, userId, "token")
	assert.NoError(t, err)
	assert.Equal(t, models.TeamInvitationStatusAccepted, invitation.Status)
}

func TestInvitationService_AcceptInvitation_UserMismatch(t *testing.T) {
	ctx := context.Background()
	store := new(mockTeamInvitationStore)
	opts := conf.NewSettings()
	mailService := NewMockMailService()
	wg := new(sync.WaitGroup)
	mockRoutineService := new(mockRoutineService)
	mockRoutineService.wg = wg
	service := NewInvitationService(store, mailService, *opts, mockRoutineService)
	teamId := uuid.New()
	userId := uuid.New()
	invitation := &models.TeamInvitation{
		TeamID: teamId,
		Email:  "test@example.com",
		Status: models.TeamInvitationStatusPending,
	}
	user := &models.User{ID: userId, Email: "other@example.com"}
	store.On("FindInvitationByToken", ctx, "token").Return(invitation, nil)
	store.On("FindUserByID", ctx, userId).Return(user, nil)
	err := service.AcceptInvitation(ctx, userId, "token")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not match invitation")
}

func TestInvitationService_RejectInvitation(t *testing.T) {
	ctx := context.Background()
	store := new(mockTeamInvitationStore)
	opts := conf.NewSettings()
	mailService := NewMockMailService()
	wg := new(sync.WaitGroup)
	mockRoutineService := new(mockRoutineService)
	mockRoutineService.wg = wg
	service := NewInvitationService(store, mailService, *opts, mockRoutineService)
	teamId := uuid.New()
	userId := uuid.New()
	invitation := &models.TeamInvitation{
		TeamID: teamId,
		Email:  "test@example.com",
		Status: models.TeamInvitationStatusPending,
	}
	user := &models.User{ID: userId, Email: "test@example.com"}
	store.On("FindInvitationByToken", ctx, "token").Return(invitation, nil)
	store.On("FindUserByID", ctx, userId).Return(user, nil)
	err := service.RejectInvitation(ctx, userId, "token")
	assert.NoError(t, err)
	assert.Equal(t, models.TeamInvitationStatusDeclined, invitation.Status)
}

func TestInvitationService_FindInvitations(t *testing.T) {
	ctx := context.Background()
	store := new(mockTeamInvitationStore)
	opts := conf.NewSettings()
	mailService := NewMockMailService()
	wg := new(sync.WaitGroup)
	mockRoutineService := new(mockRoutineService)
	mockRoutineService.wg = wg
	service := NewInvitationService(store, mailService, *opts, mockRoutineService)
	teamId := uuid.New()
	invitations := []*models.TeamInvitation{{TeamID: teamId, Email: "test@example.com"}}
	store.On("FindTeamInvitations", ctx, teamId).Return(invitations, nil)
	result, err := service.FindInvitations(ctx, teamId)
	assert.NoError(t, err)
	assert.Equal(t, invitations, result)
}
