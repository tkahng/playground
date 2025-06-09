package resource

import (
	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/models"
)

type ResourceProvider interface {
	User() Resource[models.User, uuid.UUID, UserFilter]
	UserAccount() Resource[models.UserAccount, uuid.UUID, UserAccountFilter]
	Permission() Resource[models.Permission, uuid.UUID, PermissionsFilter]
}
