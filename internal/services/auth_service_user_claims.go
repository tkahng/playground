package services

import (
	"github.com/tkahng/authgo/internal/models"
)

type UserClaims struct {
	Roles       []string           `db:"roles" json:"roles"`
	Permissions []string           `db:"permissions" json:"permissions"`
	Providers   []models.Providers `db:"providers" json:"providers" enum:"google,apple,facebook,github,credentials"`
}

// func (s *BaseAuthService) GetUserClaims(ctx context.Context, userId uuid.UUID) (*shared.UserClaims, error) {
// }
