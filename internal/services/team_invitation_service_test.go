package services

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/tkahng/authgo/internal/models"
)

// mockTeamInvitationStore is a mock implementation of TeamInvitationStore for testing.

func TestNewInvitationService(t *testing.T) {
	mockStore := &mockTeamInvitationStore{}
	service := NewInvitationService(mockStore)

	assert.NotNil(t, service, "NewInvitationService should not return nil")
}

func TestInnvitationService_CreateInvitation(t *testing.T) {
	mockStore := &mockTeamInvitationStore{}
	service := NewInvitationService(mockStore)
	ctx := context.Background()
	teamId := uuid.New()
	userId := uuid.New()
	email := "test@example.com"
	role := models.TeamMemberRoleMember
	member := &models.TeamMember{ID: uuid.New()}

	mockStore.On("FindTeamMemberByTeamAndUserId", ctx, teamId, userId).Return(member, nil)
	mockStore.On("CreateInvitation", ctx, mock.AnythingOfType("*models.TeamInvitation")).Return(nil)

	err := service.CreateInvitation(ctx, teamId, userId, email, role)
	assert.NoError(t, err)
	mockStore.AssertExpectations(t)
}

func TestInnvitationService_AcceptInvitation(t *testing.T) {
	mockStore := &mockTeamInvitationStore{}
	service := NewInvitationService(mockStore)
	ctx := context.Background()
	userId := uuid.New()
	teamId := uuid.New()
	invitation := &models.TeamInvitation{
		TeamID: teamId,
		Email:  "user@example.com",
		Role:   models.TeamMemberRoleMember,
		Status: models.TeamInvitationStatusPending,
	}
	user := &models.User{ID: userId, Email: "user@example.com"}
	member := &models.TeamMember{ID: uuid.New()}

	mockStore.On("FindInvitationByToken", ctx, "token").Return(invitation, nil)
	mockStore.On("FindUserByID", ctx, userId).Return(user, nil)
	mockStore.On("CreateTeamMember", ctx, teamId, userId, invitation.Role).Return(member, nil)
	mockStore.On("UpdateInvitation", ctx, invitation).Return(nil)

	err := service.AcceptInvitation(ctx, "token", userId)
	assert.NoError(t, err)
	mockStore.AssertExpectations(t)
}

func TestInnvitationService_FindInvitations(t *testing.T) {
	mockStore := &mockTeamInvitationStore{}
	service := NewInvitationService(mockStore)
	ctx := context.Background()
	teamId := uuid.New()
	invitations := []*models.TeamInvitation{{TeamID: teamId}}

	mockStore.On("FindTeamInvitations", ctx, teamId).Return(invitations, nil)

	result, err := service.FindInvitations(ctx, teamId)
	assert.NoError(t, err)
	assert.Equal(t, invitations, result)
	mockStore.AssertExpectations(t)
}

func TestInnvitationService_RejectInvitation(t *testing.T) {
	mockStore := &mockTeamInvitationStore{}
	service := NewInvitationService(mockStore)
	ctx := context.Background()
	userId := uuid.New()
	invitation := &models.TeamInvitation{
		Email: "user@example.com",
	}
	user := &models.User{ID: userId, Email: "user@example.com"}

	mockStore.On("FindInvitationByToken", ctx, "token").Return(invitation, nil)
	mockStore.On("FindUserByID", ctx, userId).Return(user, nil)

	err := service.RejectInvitation(ctx, "token", userId)
	assert.NoError(t, err)
	assert.Equal(t, models.TeamInvitationStatusDeclined, invitation.Status)
	mockStore.AssertExpectations(t)
}

func TestInnvitationService_CreateInvitation_Errors(t *testing.T) {
	mockStore := &mockTeamInvitationStore{}
	service := NewInvitationService(mockStore)
	ctx := context.Background()
	teamId := uuid.New()
	userId := uuid.New()
	email := "test@example.com"
	role := models.TeamMemberRoleMember

	mockStore.On("FindTeamMemberByTeamAndUserId", ctx, teamId, userId).Return(&models.TeamMember{}, errors.New("not found"))
	err := service.CreateInvitation(ctx, teamId, userId, email, role)
	assert.Error(t, err)
	mockStore.AssertExpectations(t)
}

func TestInnvitationService_AcceptInvitation_Errors(t *testing.T) {
	mockStore := &mockTeamInvitationStore{}
	service := NewInvitationService(mockStore)
	ctx := context.Background()
	userId := uuid.New()

	mockStore.On("FindInvitationByToken", ctx, "token").Return(&models.TeamInvitation{}, errors.New("invitation not found"))
	err := service.AcceptInvitation(ctx, "token", userId)
	assert.Error(t, err)
	mockStore.AssertExpectations(t)
}
