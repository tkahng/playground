package services

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/tkahng/authgo/internal/models"
)

// MockTeamStore implements TeamStore for testing

func TestTeamService_AddMember_Success(t *testing.T) {
	mockStore := new(mockTeamStore)
	service := &teamService{teamStore: mockStore}

	ctx := context.Background()
	teamID := uuid.New()
	userID := uuid.New()
	role := models.TeamMemberRoleMember

	expectedMember := &models.TeamMember{
		ID:     uuid.New(),
		TeamID: teamID,
		UserID: &userID,
		Role:   role,
	}

	mockStore.On("CreateTeamMember", ctx, teamID, userID, role).Return(expectedMember, nil)

	member, err := service.AddMember(ctx, teamID, userID, role)
	assert.NoError(t, err)
	assert.Equal(t, expectedMember, member)
	mockStore.AssertExpectations(t)
}

func TestTeamService_AddMember_Error(t *testing.T) {
	mockStore := new(mockTeamStore)
	service := &teamService{teamStore: mockStore}

	ctx := context.Background()
	teamID := uuid.New()
	userID := uuid.New()
	role := models.TeamMemberRoleMember

	expectedErr := errors.New("failed to create member")
	mockStore.On("CreateTeamMember", ctx, teamID, userID, role).Return(&models.TeamMember{}, expectedErr)

	member, err := service.AddMember(ctx, teamID, userID, role)
	assert.Nil(t, member)
	assert.Equal(t, expectedErr, err)
	mockStore.AssertExpectations(t)
}
