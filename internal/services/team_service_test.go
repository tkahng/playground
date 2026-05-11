package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/services"
	"github.com/tkahng/playground/internal/stores"
)

func TestTeamService_ProcessSlug_FirstAvailable(t *testing.T) {
	t.Parallel()
	adapterDecorator := stores.NewAdapterDecorators()
	service := services.NewTeamService(adapterDecorator)

	ctx := context.Background()
	adapterDecorator.TeamGroupFunc.FindTeamBySlugFunc = func(ctx context.Context, slug string) (*models.Team, error) {
		return nil, nil
	}

	result, err := service.ProcessSlug(ctx, "My Team")
	assert.NoError(t, err)
	assert.Equal(t, "my-team", result)
}

func TestTeamService_ProcessSlug_NumericSuffixOnConflict(t *testing.T) {
	t.Parallel()
	adapterDecorator := stores.NewAdapterDecorators()
	service := services.NewTeamService(adapterDecorator)

	ctx := context.Background()
	adapterDecorator.TeamGroupFunc.FindTeamBySlugFunc = func(ctx context.Context, slug string) (*models.Team, error) {
		// base slug is taken; my-team-1 is also taken; my-team-2 is free
		if slug == "my-team" || slug == "my-team-1" {
			return &models.Team{Slug: slug}, nil
		}
		return nil, nil
	}

	result, err := service.ProcessSlug(ctx, "My Team")
	assert.NoError(t, err)
	assert.Equal(t, "my-team-2", result)
}

func TestTeamService_ProcessSlug_DBError(t *testing.T) {
	t.Parallel()
	adapterDecorator := stores.NewAdapterDecorators()
	service := services.NewTeamService(adapterDecorator)

	ctx := context.Background()
	expectedErr := errors.New("db error")
	adapterDecorator.TeamGroupFunc.FindTeamBySlugFunc = func(ctx context.Context, slug string) (*models.Team, error) {
		return nil, expectedErr
	}

	result, err := service.ProcessSlug(ctx, "My Team")
	assert.Empty(t, result)
	assert.Equal(t, expectedErr, err)
}

func TestTeamService_ProcessSlug_SpecialCharsOnlyFallsBackToUUID(t *testing.T) {
	t.Parallel()
	adapterDecorator := stores.NewAdapterDecorators()
	service := services.NewTeamService(adapterDecorator)

	ctx := context.Background()
	adapterDecorator.TeamGroupFunc.FindTeamBySlugFunc = func(ctx context.Context, slug string) (*models.Team, error) {
		return nil, nil
	}

	// "!!!" produces an empty slug from NewSlug; service should fall back to a UUID
	result, err := service.ProcessSlug(ctx, "!!!")
	assert.NoError(t, err)
	assert.NotEmpty(t, result)
	_, parseErr := uuid.Parse(result)
	assert.NoError(t, parseErr, "fallback slug should be a valid UUID")
}

func TestTeamService_ProcessSlug_ExhaustedSuffixes(t *testing.T) {
	t.Parallel()
	adapterDecorator := stores.NewAdapterDecorators()
	service := services.NewTeamService(adapterDecorator)

	ctx := context.Background()
	// every candidate is taken
	adapterDecorator.TeamGroupFunc.FindTeamBySlugFunc = func(ctx context.Context, slug string) (*models.Team, error) {
		return &models.Team{Slug: slug}, nil
	}

	result, err := service.ProcessSlug(ctx, "My Team")
	assert.Empty(t, result)
	assert.EqualError(t, err, "cannot generate unique team slug")
}

func TestTeamService_CreateTeam_DBError(t *testing.T) {
	t.Parallel()
	adapterDecorator := stores.NewAdapterDecorators()
	service := services.NewTeamService(adapterDecorator)

	ctx := context.Background()
	name := "Test Team"
	userID := uuid.New()

	expectedErr := errors.New("db error")
	adapterDecorator.UserFunc.FindUserByIDFunc = func(ctx context.Context, userID uuid.UUID) (*models.User, error) {
		return &models.User{ID: userID}, nil
	}
	adapterDecorator.TeamGroupFunc.FindTeamBySlugFunc = func(ctx context.Context, slug string) (*models.Team, error) {
		return nil, expectedErr
	}

	teamInfo, err := service.CreateTeamWithOwner(ctx, name, userID)
	assert.Nil(t, teamInfo)
	assert.Equal(t, expectedErr, err)
}

func TestTeamService_CreateTeam_CreateTeamWithOwnerMemberError(t *testing.T) {
	t.Parallel()
	adapterDecorator := stores.NewAdapterDecorators()
	service := services.NewTeamService(adapterDecorator)

	ctx := context.Background()
	name := "Test Team"
	userID := uuid.New()

	expectedErr := errors.New("create team error")

	adapterDecorator.UserFunc.FindUserByIDFunc = func(ctx context.Context, userID uuid.UUID) (*models.User, error) {
		return &models.User{ID: userID}, nil
	}
	adapterDecorator.TeamGroupFunc.FindTeamBySlugFunc = func(ctx context.Context, slug string) (*models.Team, error) {
		return nil, nil
	}
	adapterDecorator.TeamGroupFunc.CreateTeamFunc = func(ctx context.Context, name, slug string) (*models.Team, error) {
		return nil, expectedErr
	}

	teamInfo, err := service.CreateTeamWithOwner(ctx, name, userID)
	assert.Nil(t, teamInfo)
	assert.Equal(t, expectedErr, err)
}

func TestTeamService_UpdateTeam_Success(t *testing.T) {
	t.Parallel()
	adapterDecorator := stores.NewAdapterDecorators()
	service := services.NewTeamService(adapterDecorator)

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
	t.Parallel()
	adapterDecorator := stores.NewAdapterDecorators()
	service := services.NewTeamService(adapterDecorator)

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
	t.Parallel()
	adapterDecorator := stores.NewAdapterDecorators()
	service := services.NewTeamService(adapterDecorator)

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
