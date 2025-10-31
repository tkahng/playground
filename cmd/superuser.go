package cmd

import (
	"context"
	"errors"
	"slices"

	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/spf13/cobra"
	"github.com/tkahng/playground/internal/auth"
	"github.com/tkahng/playground/internal/conf"
	"github.com/tkahng/playground/internal/core"
	"github.com/tkahng/playground/internal/stores"
)

func NewSuperuserCmd() *cobra.Command {
	superuserCmd.AddCommand(superuserCreate)
	return superuserCmd
}

var superuserCmd = &cobra.Command{
	Use:   "superuser",
	Short: "superuser",
}

var superuserCreate = &cobra.Command{
	Use:     "create",
	Example: "superuser create admin@k2dv.io Password123!",
	Short:   "create superuser",
	RunE: func(cmd *cobra.Command, args []string) error {
		return CreateSuperuser(cmd.Context(), core.NewApp(conf.AppConfigGetter()), args)
	},
}

func CreateSuperuser(ctx context.Context, app core.App, args []string) error {
	if len(args) != 2 {
		return errors.New("missing email and password arguments")
	}

	if args[0] == "" || is.EmailFormat.Validate(args[0]) != nil {
		return errors.New("mrror missing or invalid email address")
	}
	defer app.Close()
	adapter := app.Adapter()
	userStore := adapter.User()
	rbacStore := adapter.Rbac()
	user, err := userStore.FindUser(ctx, &stores.UserFilter{
		Emails: []string{args[0]},
	})
	if err != nil {
		return err
	}
	role, err := rbacStore.FindRoleByName(ctx, "superuser")
	if err != nil {
		return err
	}
	if role == nil {
		return errors.New("superuser role not found")
	}
	if user == nil {
		txErr := app.Adapter().RunInTxCtx(ctx, func(txCtx context.Context) error {
			userInfo, err := app.Auth().Signup(txCtx, &auth.SignupInput{
				Email:    args[0],
				Password: args[1],
				Verified: true,
			})
			if err != nil {
				return err
			}
			err = adapter.Rbac().CreateUserRoles(txCtx, userInfo.User.ID, role.ID)
			if err != nil {
				return err
			}
			_, err = app.Payment().CreateUserCustomer(txCtx, &userInfo.User)
			if err != nil {
				return err
			}
			return nil
		})
		if txErr != nil {
			return err
		}
	}
	if user != nil {
		claims, err := adapter.User().GetUserInfo(ctx, args[0])
		if err != nil {
			return err
		}
		if !slices.Contains(claims.Roles, "superuser") {
			err = adapter.Rbac().CreateUserRoles(ctx, user.ID, role.ID)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
