package core

import (
	"context"
	"fmt"

	"github.com/stephenafamo/bob"
	"github.com/tkahng/authgo/internal/repository"
	"github.com/tkahng/authgo/internal/shared"
)

func GetUserInfoDTO(ctx context.Context, db bob.Executor, email string) (*shared.UserInfoDto, error) {

	user, err := repository.GetUserByEmail(ctx, db, email)
	if err != nil {
		return nil, fmt.Errorf("error getting user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}
	result := &shared.UserInfoDto{
		User: user,
	}
	roles, err := repository.GetUserWithRolesAndPermissions(ctx, db, email)
	if err != nil {
		return nil, fmt.Errorf("error getting user roles and permissions: %w", err)
	}
	if roles == nil {
		return result, nil
	}
	return &shared.UserInfoDto{
		User:        user,
		Roles:       roles.Roles,
		Permissions: roles.Permissions,
		Providers:   roles.Providers,
	}, nil
}
