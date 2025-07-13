package services

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/shared"
	"github.com/tkahng/playground/internal/stores"
)

type ConstaintCheckerStore interface {
	FindUserById(ctx context.Context, userId uuid.UUID) (*models.User, error)
	FindUserByEmail(ctx context.Context, email string) (*models.User, error)
	FindLatestActiveSubscriptionByUserId(ctx context.Context, userId uuid.UUID) (*models.StripeSubscription, error)
	FindLatestActiveSubscriptionByTeamId(ctx context.Context, teamId uuid.UUID) (*models.StripeSubscription, error)
}

func NewConstraintCheckerService(
	adapter stores.StorageAdapterInterface,
) *ConstraintCheckerService {
	return &ConstraintCheckerService{
		adapter: adapter,
	}
}

type ConstraintChecker interface {
	CannotBeSuperUserID(ctx context.Context, userId uuid.UUID) (bool, error)
	CannotBeSuperUserEmail(ctx context.Context, email string) (bool, error)
	CannotBeAdminOrBasicName(ctx context.Context, permissionName string) (bool, error)
	CannotBeAdminOrBasicRoleAndPermissionName(ctx context.Context, roleName, permissionName string) (bool, error)
	CannotBeSuperUserEmailAndRoleName(ctx context.Context, email, roleName string) (bool, error)
	CannotHaveValidUserSubscription(ctx context.Context, userId uuid.UUID) (bool, error)
	TeamCannotHaveValidSubscription(ctx context.Context, teamId uuid.UUID) (bool, error)
	EmailMustBeVerified(ctx context.Context, email string) (bool, error)
}

type ConstraintCheckerService struct {
	adapter stores.StorageAdapterInterface
}

// EmailMustBeVerified implements ConstraintChecker.
func (c *ConstraintCheckerService) EmailMustBeVerified(ctx context.Context, email string) (bool, error) {
	user, err := c.adapter.User().FindUser(ctx, &stores.UserFilter{
		Emails: []string{email},
	})
	if err != nil {
		return false, err
	}
	if user == nil {
		return false, huma.Error400BadRequest("User not found")
	}
	if user.EmailVerifiedAt == nil {
		return false, huma.Error400BadRequest("Email must be verified")
	}
	return true, nil
}

// TeamCannotHaveValidSubscription implements ConstraintChecker.
func (c *ConstraintCheckerService) TeamCannotHaveValidSubscription(ctx context.Context, teamId uuid.UUID) (bool, error) {
	subscription, err := c.adapter.Subscription().FindActiveSubscriptionsByTeamIds(ctx, teamId)
	if err != nil {
		return false, err
	}

	if len(subscription) > 0 {
		sub := subscription[0]
		if sub != nil {
			return false, huma.Error400BadRequest("Cannot perform this action on a team with a valid subscription")
		}
	}
	return true, nil
}

// CannotHaveValidUserSubscription implements ConstraintChecker.
func (c *ConstraintCheckerService) CannotHaveValidUserSubscription(ctx context.Context, userId uuid.UUID) (bool, error) {
	subscription, err := c.adapter.Subscription().FindActiveSubscriptionsByUserIds(ctx, userId)
	if err != nil {
		return false, err
	}
	if len(subscription) > 0 {
		sub := subscription[0]
		if sub != nil {
			return false, huma.Error400BadRequest("Cannot perform this action on a user with a valid subscription")
		}
	}
	return true, nil
}

// CannotBeAdminOrBasicName implements ConstraintChecker.
func (c *ConstraintCheckerService) CannotBeAdminOrBasicName(ctx context.Context, permissionName string) (bool, error) {
	if permissionName == shared.PermissionNameAdmin || permissionName == shared.PermissionNameBasic {
		return false, huma.Error400BadRequest("Cannot perform this action on the admin or basic permission")
	}
	return true, nil
}

// CannotBeAdminOrBasicRoleAndPermissionName implements ConstraintChecker.
func (c *ConstraintCheckerService) CannotBeAdminOrBasicRoleAndPermissionName(ctx context.Context, roleName string, permissionName string) (bool, error) {
	if (roleName == shared.PermissionNameAdmin && permissionName == shared.PermissionNameAdmin) ||
		(roleName == shared.PermissionNameBasic && permissionName == shared.PermissionNameBasic) {
		return false, huma.Error400BadRequest("Cannot perform this action on the admin role and permission")
	}
	return true, nil
}

// CannotBeSuperUserEmailAndRoleName implements ConstraintChecker.
func (c *ConstraintCheckerService) CannotBeSuperUserEmailAndRoleName(ctx context.Context, email string, roleName string) (bool, error) {
	if email == shared.SuperUserEmail && roleName == shared.PermissionNameAdmin {
		return false, huma.Error400BadRequest("Cannot perform this action on the super user email and admin role")
	}
	return true, nil
}

// CannotBeSuperUserID implements ConstraintChecker.
func (c *ConstraintCheckerService) CannotBeSuperUserID(ctx context.Context, userId uuid.UUID) (bool, error) {
	user, err := c.adapter.User().FindUser(ctx, &stores.UserFilter{
		Ids: []uuid.UUID{userId},
	})
	if err != nil {
		return false, err
	}
	if user == nil {
		return true, nil
	}
	if user.Email == shared.SuperUserEmail {
		return false, huma.Error400BadRequest("Cannot perform this action on the super user")
	}
	return true, nil
}

func (c *ConstraintCheckerService) CannotBeSuperUserEmail(ctx context.Context, email string) (bool, error) {
	if email == shared.SuperUserEmail {
		return false, huma.Error400BadRequest("Cannot perform this action on the super user")
	}
	return true, nil
}

var _ ConstraintChecker = (*ConstraintCheckerService)(nil)
