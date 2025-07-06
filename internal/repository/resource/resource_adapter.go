package resource

import (
	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/models"
)

type (
	UserResource        = Resource[models.User, uuid.UUID, UserFilter]
	PermissionResource  = Resource[models.Permission, uuid.UUID, PermissionsFilter]
	UserAccountResource = Resource[models.UserAccount, uuid.UUID, UserAccountFilter]
	TokenResource       = Resource[models.Token, uuid.UUID, TokenFilter]
)

type ResourceAdapterInterface interface {
	User() UserResource
	Permission() PermissionResource
	UserAccount() UserAccountResource
	Token() TokenResource
}

type ResourceAdapter struct {
	db          database.Dbx
	user        UserResource
	permission  PermissionResource
	userAccount UserAccountResource
	token       TokenResource
}

func NewResourceAdapter(
	db database.Dbx,
) *ResourceAdapter {
	return &ResourceAdapter{
		db:          db,
		user:        NewUserRepositoryResource(db),
		permission:  NewPermissionQueryResource(db),
		userAccount: NewUserAccountRepositoryResource(db),
		token:       NewTokenRepositoryResource(db),
	}
}
