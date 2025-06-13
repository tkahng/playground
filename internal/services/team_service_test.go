package services

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/stores"
)

// MockTeamStore implements TeamStore for testing

func TestTeamService_AddMember_Success(t *testing.T) {
	adapterDecorator := stores.NewAdapterDecorators()
	service := &teamService{
		adapter: adapterDecorator,
	}

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
	adapterDecorator.TeamMemberFunc.CreateTeamMemberFunc = func(ctx context.Context, teamId uuid.UUID, userId uuid.UUID, role models.TeamMemberRole, hasBillingAccess bool) (*models.TeamMember, error) {
		return expectedMember, nil
	}
	// mockStore.On("CreateTeamMember", ctx, teamID, userID, role, true).Return(expectedMember, nil)

	member, err := service.AddMember(ctx, teamID, userID, role, true)
	assert.NoError(t, err)
	assert.Equal(t, expectedMember, member)
}

func TestTeamService_AddMember_Error(t *testing.T) {
	adapterDecorator := stores.NewAdapterDecorators()
	service := &teamService{
		adapter: adapterDecorator,
	}

	ctx := context.Background()
	teamID := uuid.New()
	userID := uuid.New()
	role := models.TeamMemberRoleMember

	expectedErr := errors.New("failed to create member")
	adapterDecorator.TeamMemberFunc.CreateTeamMemberFunc = func(ctx context.Context, teamId uuid.UUID, userId uuid.UUID, role models.TeamMemberRole, hasBillingAccess bool) (*models.TeamMember, error) {
		return nil, expectedErr
	}
	// mockStore.On("CreateTeamMember", ctx, teamID, userID, role, true).Return(nil, expectedErr)

	member, err := service.AddMember(ctx, teamID, userID, role, true)
	assert.Nil(t, member)
	assert.Equal(t, expectedErr, err)
}
func TestTeamService_CreateTeam_Success(t *testing.T) {
	adapterDecorator := stores.NewAdapterDecorators()
	service := &teamService{
		adapter: adapterDecorator,
	}

	ctx := context.Background()
	name := "Test Team"
	slug := "test-team"
	userID := uuid.New()
	teamID := uuid.New()
	teamMemberID := uuid.New()
	// Team slug is available

	// mockStore.On("CheckTeamSlug", ctx, slug).Return(true, nil)

	expectedTeamInfo := &shared.TeamInfoModel{
		User: models.User{ID: userID},
		Team: models.Team{ID: teamID, Name: name, Slug: slug},
		Member: models.TeamMember{
			ID:               teamMemberID,
			TeamID:           teamID,
			UserID:           &userID,
			Role:             models.TeamMemberRoleOwner,
			HasBillingAccess: true,
		},
	}
	adapterDecorator.UserFunc.FindUserByIDFunc = func(ctx context.Context, userID uuid.UUID) (*models.User, error) {
		return &expectedTeamInfo.User, nil
	}
	adapterDecorator.TeamGroupFunc.CheckTeamSlugFunc = func(ctx context.Context, slug string) (bool, error) {
		return true, nil
	}
	adapterDecorator.TeamGroupFunc.CreateTeamFunc = func(ctx context.Context, name, slug string) (*models.Team, error) {
		return &expectedTeamInfo.Team, nil
	}
	adapterDecorator.TeamMemberFunc.CreateTeamMemberFunc = func(ctx context.Context, teamId uuid.UUID, userId uuid.UUID, role models.TeamMemberRole, hasBillingAccess bool) (*models.TeamMember, error) {
		return &expectedTeamInfo.Member, nil
	}

	teamInfo, err := service.CreateTeamWithOwner(ctx, name, slug, userID)
	assert.NoError(t, err)
	assert.Equal(t, expectedTeamInfo, teamInfo)
}

func TestTeamService_CreateTeam_SlugExists(t *testing.T) {
	adapterDecorator := stores.NewAdapterDecorators()
	service := &teamService{
		adapter: adapterDecorator,
	}

	ctx := context.Background()
	name := "Test Team"
	slug := "existing-slug"
	userID := uuid.New()
	adapterDecorator.UserFunc.FindUserByIDFunc = func(ctx context.Context, userID uuid.UUID) (*models.User, error) {
		return &models.User{ID: userID}, nil
	}
	// Team slug already exists
	adapterDecorator.TeamGroupFunc.CheckTeamSlugFunc = func(ctx context.Context, slug string) (bool, error) {
		return false, nil
	}
	// mockStore.On("CheckTeamSlug", ctx, slug).Return(false, nil)

	teamInfo, err := service.CreateTeamWithOwner(ctx, name, slug, userID)
	assert.Nil(t, teamInfo)
	assert.EqualError(t, err, "team slug already exists")
}

func TestTeamService_CreateTeam_CheckTeamSlugError(t *testing.T) {
	adapterDecorator := stores.NewAdapterDecorators()
	service := &teamService{
		adapter: adapterDecorator,
	}

	ctx := context.Background()
	name := "Test Team"
	slug := "test-team"
	userID := uuid.New()

	expectedErr := errors.New("db error")
	adapterDecorator.UserFunc.FindUserByIDFunc = func(ctx context.Context, userID uuid.UUID) (*models.User, error) {
		return &models.User{ID: userID}, nil
	}
	adapterDecorator.TeamGroupFunc.CheckTeamSlugFunc = func(ctx context.Context, slug string) (bool, error) {
		return false, expectedErr
	}
	// mockStore.On("CheckTeamSlug", ctx, slug).Return(false, expectedErr)

	teamInfo, err := service.CreateTeamWithOwner(ctx, name, slug, userID)
	assert.Nil(t, teamInfo)
	assert.Equal(t, expectedErr, err)
}

func TestTeamService_CreateTeam_CreateTeamWithOwnerMemberError(t *testing.T) {
	adapterDecorator := stores.NewAdapterDecorators()
	service := &teamService{
		adapter: adapterDecorator,
	}

	ctx := context.Background()
	name := "Test Team"
	slug := "test-team"
	userID := uuid.New()

	// mockStore.On("CheckTeamSlug", ctx, slug).Return(true, nil)
	expectedErr := errors.New("create team error")

	adapterDecorator.UserFunc.FindUserByIDFunc = func(ctx context.Context, userID uuid.UUID) (*models.User, error) {
		return &models.User{ID: userID}, nil
	}
	adapterDecorator.TeamGroupFunc.CheckTeamSlugFunc = func(ctx context.Context, slug string) (bool, error) {
		return true, nil
	}
	adapterDecorator.TeamGroupFunc.CreateTeamFunc = func(ctx context.Context, name, slug string) (*models.Team, error) {
		return nil, expectedErr
	}

	// mockStore.On("CreateTeamWithOwnerMember", ctx, name, slug, userID).Return(nil, expectedErr)

	teamInfo, err := service.CreateTeamWithOwner(ctx, name, slug, userID)
	assert.Nil(t, teamInfo)
	assert.Equal(t, expectedErr, err)
	// mockStore.AssertExpectations(t)
}

func TestTeamService_UpdateTeam_Success(t *testing.T) {
	adapterDecorator := stores.NewAdapterDecorators()
	service := &teamService{
		adapter: adapterDecorator,
	}

	ctx := context.Background()
	teamID := uuid.New()
	name := "Updated Team"
	expectedTeam := &models.Team{ID: teamID, Name: name}

	adapterDecorator.TeamGroupFunc.UpdateTeamFunc = func(ctx context.Context, teamId uuid.UUID, name string) (*models.Team, error) {
		return expectedTeam, nil
	}
	// mockStore.On("UpdateTeam", ctx, teamID, name).Return(expectedTeam, nil)

	team, err := service.UpdateTeam(ctx, teamID, name)
	assert.NoError(t, err)
	assert.Equal(t, expectedTeam, team)
	// mockStore.AssertExpectations(t)
}

func TestTeamService_UpdateTeam_Error(t *testing.T) {
	adapterDecorator := stores.NewAdapterDecorators()
	service := &teamService{
		adapter: adapterDecorator,
	}

	ctx := context.Background()
	teamID := uuid.New()
	name := "Updated Team"
	expectedErr := errors.New("update error")

	adapterDecorator.TeamGroupFunc.UpdateTeamFunc = func(ctx context.Context, teamId uuid.UUID, name string) (*models.Team, error) {
		return nil, expectedErr
	}
	// mockStore.On("UpdateTeam", ctx, teamID, name).Return(nil, expectedErr)

	team, err := service.UpdateTeam(ctx, teamID, name)
	assert.Nil(t, team)
	assert.Equal(t, expectedErr, err)
	// mockStore.AssertExpectations(t)
}

func TestTeamService_UpdateTeam_TeamNotFound(t *testing.T) {
	adapterDecorator := stores.NewAdapterDecorators()
	service := &teamService{
		adapter: adapterDecorator,
	}

	ctx := context.Background()
	teamID := uuid.New()
	name := "Updated Team"

	adapterDecorator.TeamGroupFunc.UpdateTeamFunc = func(ctx context.Context, teamId uuid.UUID, name string) (*models.Team, error) {
		return nil, nil
	}
	// mockStore.On("UpdateTeam", ctx, teamID, name).Return(nil, nil)

	team, err := service.UpdateTeam(ctx, teamID, name)
	assert.Nil(t, team)
	assert.EqualError(t, err, "team not found")
	// mockStore.AssertExpectations(t)
}
