package services

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/shared"
)

type ConstaintCheckerStore interface {
	FindUserById(ctx context.Context, userId uuid.UUID) (*models.User, error)
	FindUserByEmail(ctx context.Context, email string) (*models.User, error)
	FindLatestActiveSubscriptionByUserId(ctx context.Context, userId uuid.UUID) (*models.StripeSubscription, error)
	FindLatestActiveSubscriptionByTeamId(ctx context.Context, teamId uuid.UUID) (*models.StripeSubscription, error)
}

func NewConstraintCheckerService(
	store ConstaintCheckerStore,
) *ConstraintCheckerService {
	return &ConstraintCheckerService{
		store: store,
	}
}

type ConstraintChecker interface {
	CannotBeSuperUserID(ctx context.Context, userId uuid.UUID) error
	CannotBeSuperUserEmail(ctx context.Context, email string) error
	CannotBeAdminOrBasicName(ctx context.Context, permissionName string) error
	CannotBeAdminOrBasicRoleAndPermissionName(ctx context.Context, roleName, permissionName string) error
	CannotBeSuperUserEmailAndRoleName(ctx context.Context, email, roleName string) error
	CannotHaveValidUserSubscription(ctx context.Context, userId uuid.UUID) error
	TeamCannotHaveValidSubscription(ctx context.Context, teamId uuid.UUID) error
	EmailMustBeVerified(ctx context.Context, email string) error
}

type ConstraintCheckerService struct {
	store ConstaintCheckerStore
}

// EmailMustBeVerified implements ConstraintChecker.
func (c *ConstraintCheckerService) EmailMustBeVerified(ctx context.Context, email string) error {
	user, err := c.store.FindUserByEmail(ctx, email)
	if err != nil {
		return err
	}
	if user == nil {
		return nil // No user found, so no constraint violation
	}
	if user.EmailVerifiedAt == nil {
		return huma.Error400BadRequest("Email must be verified")
	}
	return nil
}

// TeamCannotHaveValidSubscription implements ConstraintChecker.
func (c *ConstraintCheckerService) TeamCannotHaveValidSubscription(ctx context.Context, teamId uuid.UUID) error {
	subscription, err := c.store.FindLatestActiveSubscriptionByTeamId(ctx, teamId)
	if err != nil {
		return err
	}
	if subscription != nil {
		return huma.Error400BadRequest("Cannot perform this action on a team with a valid subscription")
	}
	return nil
}

// CannotHaveValidUserSubscription implements ConstraintChecker.
func (c *ConstraintCheckerService) CannotHaveValidUserSubscription(ctx context.Context, userId uuid.UUID) error {
	subscription, err := c.store.FindLatestActiveSubscriptionByUserId(ctx, userId)
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

// CannotBeSuperUser implements ConstraintChecker.
func (c *ConstraintCheckerService) CannotBeSuperUserID(ctx context.Context, userId uuid.UUID) error {
	user, err := c.store.FindUserById(ctx, userId)
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
