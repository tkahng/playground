package core

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/tkahng/authgo/internal/shared"
)

func TestHandleRefreshToken(t *testing.T) {
	ctx := context.Background()
	mockStorage := new(mockAuthStorage)
	mockToken := new(mockTokenManager)
	app := &AuthActionsBase{
		token:   mockToken,
		storage: mockStorage,
		options: &AppOptions{
			Auth: AuthOptions{
				RefreshToken: TokenOption{
					Type:     shared.TokenTypesRefreshToken,
					Secret:   string(shared.TokenTypesRefreshToken),
					Duration: 604800, // 7days
				},
			},
		},
	}

	testCases := []struct {
		name          string
		token         string
		setupMocks    func()
		expectedError bool
	}{
		{
			name:  "valid refresh token",
			token: "valid.token.here",
			setupMocks: func() {
				mockToken.On("ParseToken", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				mockToken.On("CreateJwtToken", mock.Anything, mock.Anything).Return("new.valid.token.here", nil)
				mockStorage.On("VerifyTokenStorage", ctx, mock.Anything).Return(nil)
				mockStorage.On("GetUserInfo", ctx, mock.Anything).Return(&shared.UserInfo{
					User: shared.User{
						ID:    uuid.New(),
						Email: "test@example.com",
					},
				}, nil)
				mockStorage.On("SaveToken", ctx, mock.Anything).Return(nil)
			},
			expectedError: false,
		},
		{
			name:  "invalid token format",
			token: "invalid-token",
			setupMocks: func() {
				mockToken.On("ParseToken", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("invalid token"))
			},
			expectedError: true,
		},
		{
			name:  "token verification fails",
			token: "valid.token.here",
			setupMocks: func() {
				mockToken.On("ParseToken", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("invalid token"))
			},
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockToken.ExpectedCalls = nil
			mockToken.Calls = nil
			mockStorage.ExpectedCalls = nil
			mockStorage.Calls = nil

			if tc.setupMocks != nil {
				tc.setupMocks()
			}

			result, err := app.HandleRefreshToken(ctx, tc.token)

			if tc.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}

			mockToken.AssertExpectations(t)
			mockStorage.AssertExpectations(t)
		})
	}
}
