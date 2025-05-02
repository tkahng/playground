package core

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/queries"
	"github.com/tkahng/authgo/internal/shared"
)

type ConstraintChecker interface {
	CannotBeSuperUserID(userId uuid.UUID) error
	CannotBeSuperUserEmail(email string) error
	CannotBeAdminOrBasicName(permissionName string) error
	CannotBeAdminOrBasicRoleAndPermissionName(roleName, permissionName string) error
	CannotBeSuperUserEmailAndRoleName(email, roleName string) error
	CannotHaveValidSubscription(userId uuid.UUID) error
}

type ConstraintCheckerService struct {
	db  queries.Queryer
	ctx context.Context
}

// CannotHaveValidSubscription implements ConstraintChecker.
func (c *ConstraintCheckerService) CannotHaveValidSubscription(userId uuid.UUID) error {
	subscription, err := queries.FindLatestActiveSubscriptionByUserId(c.ctx, c.db, userId)
	if err != nil {
		return err
	}
	if subscription != nil {
		return huma.Error400BadRequest("Cannot perform this action on a user with a valid subscription")
	}
	return nil
}

// CannotBeAdminOrBasicName implements ConstraintChecker.
func (c *ConstraintCheckerService) CannotBeAdminOrBasicName(permissionName string) error {
	if permissionName == shared.PermissionNameAdmin || permissionName == shared.PermissionNameBasic {
		return huma.Error400BadRequest("Cannot perform this action on the admin or basic permission")
	}
	return nil
}

// CannotBeAdminOrBasicRoleAndPermissionName implements ConstraintChecker.
func (c *ConstraintCheckerService) CannotBeAdminOrBasicRoleAndPermissionName(roleName string, permissionName string) error {
	if (roleName == shared.PermissionNameAdmin && permissionName == shared.PermissionNameAdmin) ||
		(roleName == shared.PermissionNameBasic && permissionName == shared.PermissionNameBasic) {
		return huma.Error400BadRequest("Cannot perform this action on the admin role and permission")
	}
	return nil
}

// CannotBeSuperUserEmailAndRoleName implements ConstraintChecker.
func (c *ConstraintCheckerService) CannotBeSuperUserEmailAndRoleName(email string, roleName string) error {
	if email == shared.SuperUserEmail && roleName == shared.PermissionNameAdmin {
		return huma.Error400BadRequest("Cannot perform this action on the super user email and admin role")
	}
	return nil
}

func NewConstraintCheckerService(ctx context.Context, db queries.Queryer) *ConstraintCheckerService {
	return &ConstraintCheckerService{
		db:  db,
		ctx: ctx,
	}
}

// CannotBeSuperUser implements ConstraintChecker.
func (c *ConstraintCheckerService) CannotBeSuperUserID(userId uuid.UUID) error {
	user, err := queries.FindUserById(c.ctx, c.db, userId)
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
func (c *ConstraintCheckerService) CannotBeSuperUserEmail(email string) error {
	if email == shared.SuperUserEmail {
		return huma.Error400BadRequest("Cannot perform this action on the super user")
	}
	return nil
}

var _ ConstraintChecker = (*ConstraintCheckerService)(nil)
