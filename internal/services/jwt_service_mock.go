package services

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/mock"
	"github.com/tkahng/authgo/internal/conf"
)

type MockJwtService struct {
	mock.Mock
}

var _ JwtService = (*MockJwtService)(nil)

func NewMockJwtService() *MockJwtService {
	return new(MockJwtService)
}

// CreateJwtToken implements TokenManager.
func (m *MockJwtService) CreateJwtToken(payload jwt.Claims, signingKey string) (string, error) {
	args := m.Called(payload, signingKey)
	return args.String(0), args.Error(1)
}

// ParseToken implements TokenManager.
func (m *MockJwtService) ParseToken(token string, config conf.TokenOption, data any) error {
	args := m.Called(token, config, data)
	return args.Error(0)
}
