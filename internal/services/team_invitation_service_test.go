package services_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/tkahng/playground/internal/conf"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/services"
	"github.com/tkahng/playground/internal/stores"
)

func TestInvitationService_CheckValidInvitation_ValidFutureExpiry(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	userID := uuid.New()
	adapter := stores.NewAdapterDecorators()
	service := services.NewInvitationService(adapter, conf.EnvConfig{}, nil, nil)

	adapter.TeamInvitationFunc.FindInvitationByTokenFunc = func(ctx context.Context, token string) (*models.TeamInvitation, error) {
		return &models.TeamInvitation{
			Token:     token,
			Email:     "invitee@example.com",
			Status:    models.TeamInvitationStatusPending,
			ExpiresAt: time.Now().Add(time.Hour),
		}, nil
	}
	adapter.UserFunc.FindUserByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.User, error) {
		assert.Equal(t, userID, id)
		return &models.User{
			ID:    id,
			Email: "invitee@example.com",
		}, nil
	}

	valid, err := service.CheckValidInvitation(ctx, userID, "valid-token")

	assert.NoError(t, err)
	assert.True(t, valid)
}

func TestInvitationService_CheckValidInvitation_ExpiredInvitation(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	userID := uuid.New()
	adapter := stores.NewAdapterDecorators()
	service := services.NewInvitationService(adapter, conf.EnvConfig{}, nil, nil)

	adapter.TeamInvitationFunc.FindInvitationByTokenFunc = func(ctx context.Context, token string) (*models.TeamInvitation, error) {
		return &models.TeamInvitation{
			Token:     token,
			Email:     "invitee@example.com",
			Status:    models.TeamInvitationStatusPending,
			ExpiresAt: time.Now().Add(-time.Hour),
		}, nil
	}
	adapter.UserFunc.FindUserByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.User, error) {
		return &models.User{
			ID:    id,
			Email: "invitee@example.com",
		}, nil
	}

	valid, err := service.CheckValidInvitation(ctx, userID, "expired-token")

	assert.False(t, valid)
	assert.EqualError(t, err, "invitation is expired")
}
