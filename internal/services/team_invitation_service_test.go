package services

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/tkahng/authgo/internal/models"
)

func TestNewInvitationService(t *testing.T) {
	mockStore := &mockTeamInvitationStore{}
	service := NewInvitationService(mockStore)

	assert.NotNil(t, service, "NewInvitationService should not return nil")
}

func TestInvitationService_CreateInvitation(t *testing.T) {
	ctx := context.Background()
	store := new(mockTeamInvitationStore)
	service := NewInvitationService(store)
	teamId := uuid.New()
	userId := uuid.New()
	member := &models.TeamMember{ID: uuid.New()}
	store.On("FindTeamMemberByTeamAndUserId", ctx, teamId, userId).Return(member, nil)
	store.On("CreateInvitation", ctx, mock.AnythingOfType("*models.TeamInvitation")).Return(nil)
	err := service.CreateInvitation(ctx, teamId, userId, "test@example.com", models.TeamMemberRoleMember)
	assert.NoError(t, err)
	store.AssertExpectations(t)
}

func TestInvitationService_CreateInvitation_NotMember(t *testing.T) {
	ctx := context.Background()
	store := new(mockTeamInvitationStore)
	service := NewInvitationService(store)
	teamId := uuid.New()
	userId := uuid.New()
	store.On("FindTeamMemberByTeamAndUserId", ctx, teamId, userId).Return((*models.TeamMember)(nil), nil)
	err := service.CreateInvitation(ctx, teamId, userId, "test@example.com", models.TeamMemberRoleMember)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not a member")
}

func TestInvitationService_AcceptInvitation(t *testing.T) {
	ctx := context.Background()
	store := new(mockTeamInvitationStore)
	service := NewInvitationService(store)
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
	store.On("CreateTeamMember", ctx, teamId, userId, models.TeamMemberRoleMember).Return(&models.TeamMember{}, nil)
	store.On("UpdateInvitation", ctx, invitation).Return(nil)
	err := service.AcceptInvitation(ctx, "token", userId)
	assert.NoError(t, err)
	assert.Equal(t, models.TeamInvitationStatusAccepted, invitation.Status)
}

func TestInvitationService_AcceptInvitation_UserMismatch(t *testing.T) {
	ctx := context.Background()
	store := new(mockTeamInvitationStore)
	service := NewInvitationService(store)
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
	err := service.AcceptInvitation(ctx, "token", userId)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not match invitation")
}

func TestInvitationService_RejectInvitation(t *testing.T) {
	ctx := context.Background()
	store := new(mockTeamInvitationStore)
	service := NewInvitationService(store)
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
	err := service.RejectInvitation(ctx, "token", userId)
	assert.NoError(t, err)
	assert.Equal(t, models.TeamInvitationStatusDeclined, invitation.Status)
}

func TestInvitationService_FindInvitations(t *testing.T) {
	ctx := context.Background()
	store := new(mockTeamInvitationStore)
	service := NewInvitationService(store)
	teamId := uuid.New()
	invitations := []*models.TeamInvitation{{TeamID: teamId, Email: "test@example.com"}}
	store.On("FindTeamInvitations", ctx, teamId).Return(invitations, nil)
	result, err := service.FindInvitations(ctx, teamId)
	assert.NoError(t, err)
	assert.Equal(t, invitations, result)
}
