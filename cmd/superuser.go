package cmd

import (
	"errors"

	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/spf13/cobra"
	"github.com/tkahng/authgo/internal/conf"
	"github.com/tkahng/authgo/internal/core"
	"github.com/tkahng/authgo/internal/db/models"
	"github.com/tkahng/authgo/internal/repository"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/types"
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
	Example: "superuser create test@example.com Password123!",
	Short:   "create superuser",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			return errors.New("missing email and password arguments")
		}

		if args[0] == "" || is.EmailFormat.Validate(args[0]) != nil {
			return errors.New("mrror missing or invalid email address")
		}

		ctx := cmd.Context()
		conf := conf.AppConfigGetter()
		dbx := core.NewBobFromConf(ctx, conf.Db)
		err := repository.EnsureRoleAndPermissions(ctx, dbx, "superuser", "superuser", "superuser")
		if err != nil {
			return err
		}

		user, err := repository.GetUserByEmail(ctx, dbx, args[0])
		if err != nil {
			return err
		}
		role, err := repository.FindRoleByName(ctx, dbx, "superuser")
		if err != nil {
			return err
		}
		if user == nil {
			data := &shared.AuthenticateUserParams{
				Email:             args[0],
				Provider:          models.ProvidersCredentials,
				ProviderAccountID: args[0],
				HashPassword:      types.Pointer(args[1]),
				Type:              models.ProviderTypesCredentials,
			}
			user, err = repository.CreateUser(ctx, dbx, data)
			if err != nil {
				return err
			}
			_, err = repository.CreateAccount(ctx, dbx, user, data)
			if err != nil {
				return err
			}
		}
		if user != nil {
			err = user.AttachRoles(ctx, dbx, role)
			if err != nil {
				return err
			}
		}
		return nil
	},
}

// var seedRolesCmd = &cobra.Command{
// 	Use:   "roles",
// 	Short: "seed roles",
// 	RunE: func(cmd *cobra.Command, args []string) error {
// 		ctx := cmd.Context()
// 		conf := conf.AppConfigGetter()

// 		dbx := core.NewBobFromConf(ctx, conf.Db)
// 		err := repository.PopulateRoles(ctx, dbx)
// 		if err != nil {
// 			return err
// 		}
// 		return nil
// 	},
// }
// var seedUserCmd = &cobra.Command{
// 	Use:   "users",
// 	Short: "seed users",
// 	RunE: func(cmd *cobra.Command, args []string) error {
// 		ctx := cmd.Context()
// 		conf := conf.AppConfigGetter()

// 		dbx := core.NewBobFromConf(ctx, conf.Db)

// 		err := seeders.UserCredentialsFactory(ctx, dbx, 10)
// 		// err := repository.PopulateRoles(ctx, dbx)
// 		if err != nil {

// 			return fmt.Errorf("error at createing users: %w", err)
// 		}
// 		err = seeders.UserOauthFactory(ctx, dbx, 10, models.ProvidersGoogle)
// 		if err != nil {
// 			return fmt.Errorf("error at createing users: %w", err)
// 		}
// 		err = seeders.UserOauthFactory(ctx, dbx, 10, models.ProvidersGithub)
// 		if err != nil {
// 			return fmt.Errorf("error at createing users: %w", err)
// 		}
// 		return nil
// 	},
// }
