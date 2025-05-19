package services

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/queries"
	"github.com/tkahng/authgo/internal/shared"
)

type ConstraintChecker interface {
	CannotBeSuperUserID(ctx context.Context, userId uuid.UUID) error
	CannotBeSuperUserEmail(ctx context.Context, email string) error
	CannotBeAdminOrBasicName(ctx context.Context, permissionName string) error
	CannotBeAdminOrBasicRoleAndPermissionName(ctx context.Context, roleName, permissionName string) error
	CannotBeSuperUserEmailAndRoleName(ctx context.Context, email, roleName string) error
	CannotHaveValidSubscription(ctx context.Context, userId uuid.UUID) error
}

type ConstraintCheckerService struct {
	db database.Dbx
}

// CannotHaveValidSubscription implements ConstraintChecker.
func (c *ConstraintCheckerService) CannotHaveValidSubscription(ctx context.Context, userId uuid.UUID) error {
	subscription, err := queries.FindLatestActiveSubscriptionByTeamId(ctx, c.db, userId)
	if err != nil {
		return err
	}
	if subscription != nil {
		return huma.Error400BadRequest("Cannot perform this action on a user with a valid subscription")
	}
	return nil
}

// CannotBeAdminOrBasicName implements ConstraintChecker.
func (c *ConstraintCheckerService) CannotBeAdminOrBasicName(ctx context.Context, permissionName string) error {
	if permissionName == shared.PermissionNameAdmin || permissionName == shared.PermissionNameBasic {
		return huma.Error400BadRequest("Cannot perform this action on the admin or basic permission")
	}
	return nil
}

// CannotBeAdminOrBasicRoleAndPermissionName implements ConstraintChecker.
func (c *ConstraintCheckerService) CannotBeAdminOrBasicRoleAndPermissionName(ctx context.Context, roleName string, permissionName string) error {
	if (roleName == shared.PermissionNameAdmin && permissionName == shared.PermissionNameAdmin) ||
		(roleName == shared.PermissionNameBasic && permissionName == shared.PermissionNameBasic) {
		return huma.Error400BadRequest("Cannot perform this action on the admin role and permission")
	}
	return nil
}

// CannotBeSuperUserEmailAndRoleName implements ConstraintChecker.
func (c *ConstraintCheckerService) CannotBeSuperUserEmailAndRoleName(ctx context.Context, email string, roleName string) error {
	if email == shared.SuperUserEmail && roleName == shared.PermissionNameAdmin {
		return huma.Error400BadRequest("Cannot perform this action on the super user email and admin role")
	}
	return nil
}

func NewConstraintCheckerService(ctx context.Context, db database.Dbx) *ConstraintCheckerService {
	return &ConstraintCheckerService{
		db: db,
	}
}

// CannotBeSuperUser implements ConstraintChecker.
func (c *ConstraintCheckerService) CannotBeSuperUserID(ctx context.Context, userId uuid.UUID) error {
	user, err := queries.FindUserById(ctx, c.db, userId)
	if err != nil {
		return err
	}
	if user == nil {
		return nil // No user found, so no constraint violation
	}
	if user.Email == shared.SuperUserEmail {
		return huma.Error400BadRequest("Cannot perform this action on the super user")
	}
	return nil
}
func (c *ConstraintCheckerService) CannotBeSuperUserEmail(ctx context.Context, email string) error {
	if email == shared.SuperUserEmail {
		return huma.Error400BadRequest("Cannot perform this action on the super user")
	}
	return nil
}

var _ ConstraintChecker = (*ConstraintCheckerService)(nil)
