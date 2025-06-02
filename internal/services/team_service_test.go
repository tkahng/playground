package services

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/shared"
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
		ID:               uuid.New(),
		TeamID:           teamID,
		UserID:           &userID,
		Role:             role,
		HasBillingAccess: true,
	}

	mockStore.On("CreateTeamMember", ctx, teamID, userID, role, true).Return(expectedMember, nil)

	member, err := service.AddMember(ctx, teamID, userID, role, true)
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
	mockStore.On("CreateTeamMember", ctx, teamID, userID, role, true).Return(nil, expectedErr)

	member, err := service.AddMember(ctx, teamID, userID, role, true)
	assert.Nil(t, member)
	assert.Equal(t, expectedErr, err)
	mockStore.AssertExpectations(t)
}
func TestTeamService_CreateTeam_Success(t *testing.T) {
	mockStore := new(mockTeamStore)
	service := &teamService{teamStore: mockStore}

	ctx := context.Background()
	name := "Test Team"
	slug := "test-team"
	userID := uuid.New()

	// Team slug is available
	mockStore.On("CheckTeamSlug", ctx, slug).Return(true, nil)

	expectedTeamInfo := &shared.TeamInfoModel{}
	mockStore.On("CreateTeamWithOwnerMember", ctx, name, slug, userID).Return(expectedTeamInfo, nil)

	teamInfo, err := service.CreateTeam(ctx, name, slug, userID)
	assert.NoError(t, err)
	assert.Equal(t, expectedTeamInfo, teamInfo)
	mockStore.AssertExpectations(t)
}

func TestTeamService_CreateTeam_SlugExists(t *testing.T) {
	mockStore := new(mockTeamStore)
	service := &teamService{teamStore: mockStore}

	ctx := context.Background()
	name := "Test Team"
	slug := "existing-slug"
	userID := uuid.New()

	// Team slug already exists
	mockStore.On("CheckTeamSlug", ctx, slug).Return(false, nil)

	teamInfo, err := service.CreateTeam(ctx, name, slug, userID)
	assert.Nil(t, teamInfo)
	assert.EqualError(t, err, "team slug already exists")
	mockStore.AssertExpectations(t)
}

func TestTeamService_CreateTeam_CheckTeamSlugError(t *testing.T) {
	mockStore := new(mockTeamStore)
	service := &teamService{teamStore: mockStore}

	ctx := context.Background()
	name := "Test Team"
	slug := "test-team"
	userID := uuid.New()

	expectedErr := errors.New("db error")
	mockStore.On("CheckTeamSlug", ctx, slug).Return(false, expectedErr)

	teamInfo, err := service.CreateTeam(ctx, name, slug, userID)
	assert.Nil(t, teamInfo)
	assert.Equal(t, expectedErr, err)
	mockStore.AssertExpectations(t)
}

func TestTeamService_CreateTeam_CreateTeamWithOwnerMemberError(t *testing.T) {
	mockStore := new(mockTeamStore)
	service := &teamService{teamStore: mockStore}

	ctx := context.Background()
	name := "Test Team"
	slug := "test-team"
	userID := uuid.New()

	mockStore.On("CheckTeamSlug", ctx, slug).Return(true, nil)
	expectedErr := errors.New("create team error")
	mockStore.On("CreateTeamWithOwnerMember", ctx, name, slug, userID).Return(nil, expectedErr)

	teamInfo, err := service.CreateTeam(ctx, name, slug, userID)
	assert.Nil(t, teamInfo)
	assert.Equal(t, expectedErr, err)
	mockStore.AssertExpectations(t)
}

func TestTeamService_CreateTeam_CreateTeamWithOwnerMemberNil(t *testing.T) {
	mockStore := new(mockTeamStore)
	service := &teamService{teamStore: mockStore}

	ctx := context.Background()
	name := "Test Team"
	slug := "test-team"
	userID := uuid.New()

	mockStore.On("CheckTeamSlug", ctx, slug).Return(true, nil)
	mockStore.On("CreateTeamWithOwnerMember", ctx, name, slug, userID).Return(nil, nil)

	teamInfo, err := service.CreateTeam(ctx, name, slug, userID)
	assert.Nil(t, teamInfo)
	assert.EqualError(t, err, "team not found")
	mockStore.AssertExpectations(t)
}
func TestTeamService_UpdateTeam_Success(t *testing.T) {
	mockStore := new(mockTeamStore)
	service := &teamService{teamStore: mockStore}

	ctx := context.Background()
	teamID := uuid.New()
	name := "Updated Team"
	expectedTeam := &models.Team{ID: teamID, Name: name}

	mockStore.On("UpdateTeam", ctx, teamID, name).Return(expectedTeam, nil)

	team, err := service.UpdateTeam(ctx, teamID, name)
	assert.NoError(t, err)
	assert.Equal(t, expectedTeam, team)
	mockStore.AssertExpectations(t)
}

func TestTeamService_UpdateTeam_Error(t *testing.T) {
	mockStore := new(mockTeamStore)
	service := &teamService{teamStore: mockStore}

	ctx := context.Background()
	teamID := uuid.New()
	name := "Updated Team"
	expectedErr := errors.New("update error")

	mockStore.On("UpdateTeam", ctx, teamID, name).Return(nil, expectedErr)

	team, err := service.UpdateTeam(ctx, teamID, name)
	assert.Nil(t, team)
	assert.Equal(t, expectedErr, err)
	mockStore.AssertExpectations(t)
}

func TestTeamService_UpdateTeam_TeamNotFound(t *testing.T) {
	mockStore := new(mockTeamStore)
	service := &teamService{teamStore: mockStore}

	ctx := context.Background()
	teamID := uuid.New()
	name := "Updated Team"

	mockStore.On("UpdateTeam", ctx, teamID, name).Return(nil, nil)

	team, err := service.UpdateTeam(ctx, teamID, name)
	assert.Nil(t, team)
	assert.EqualError(t, err, "team not found")
	mockStore.AssertExpectations(t)
}
