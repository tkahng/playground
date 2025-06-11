package resource

import (
	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/models"
)

type (
	UserResource        = Resource[models.User, uuid.UUID, UserFilter]
	PermissionResource  = Resource[models.Permission, uuid.UUID, PermissionsFilter]
	UserAccountResource = Resource[models.UserAccount, uuid.UUID, UserAccountFilter]
	TokenResource       = Resource[models.Token, uuid.UUID, TokenFilter]
)

type ResourceAdapter struct {
	User        UserResource
	Permission  PermissionResource
	UserAccount UserAccountResource
	Token       TokenResource
}
