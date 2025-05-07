package cmd

import (
	"errors"

	"github.com/alexedwards/argon2id"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/spf13/cobra"
	"github.com/tkahng/authgo/internal/conf"
	"github.com/tkahng/authgo/internal/db"
	"github.com/tkahng/authgo/internal/queries"
	"github.com/tkahng/authgo/internal/repository"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/security"
	"github.com/tkahng/authgo/internal/tools/types"
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
		if len(args) != 2 {
			return errors.New("missing email and password arguments")
		}

		if args[0] == "" || is.EmailFormat.Validate(args[0]) != nil {
			return errors.New("mrror missing or invalid email address")
		}

		ctx := cmd.Context()
		conf := conf.GetConfig[conf.DBConfig]()

		dbx := db.CreateQueries(ctx, conf.DatabaseUrl)
		err := queries.EnsureRoleAndPermissions(ctx, dbx, "superuser", "superuser")
		if err != nil {
			return err
		}

		user, err := queries.FindUserByEmail(ctx, dbx, args[0])
		if err != nil {
			return err
		}
		role, err := repository.Role.GetOne(
			ctx,
			dbx,
			&map[string]any{
				"name": map[string]any{
					"_eq": "superuser",
				},
			},
		)
		if err != nil {
			return err
		}
		if user == nil {
			hash, err := security.CreateHash(args[1], argon2id.DefaultParams)
			if err != nil {
				return err
			}
			data := &shared.AuthenticationInput{
				Email:             args[0],
				Provider:          shared.ProvidersCredentials,
				ProviderAccountID: args[0],
				HashPassword:      types.Pointer(hash),
				Type:              shared.ProviderTypeCredentials,
			}
			user, err = queries.CreateUser(ctx, dbx, data)
			if err != nil {
				return err
			}
			_, err = queries.CreateAccount(ctx, dbx, user.ID, data)
			if err != nil {
				return err
			}
		}
		if user != nil {
			err = queries.CreateUserRoles(ctx, dbx, user.ID, role.ID)
			if err != nil {
				return err
			}
		}
		return nil
	},
}
