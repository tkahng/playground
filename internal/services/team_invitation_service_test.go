package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// mockTeamInvitationStore is a mock implementation of TeamInvitationStore for testing.

func TestNewInvitationService(t *testing.T) {
	mockStore := &mockTeamInvitationStore{}
	service := NewInvitationService(mockStore)

	assert.NotNil(t, service, "NewInvitationService should not return nil")
}
