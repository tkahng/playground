package cmd

import (
	"errors"
	"slices"

	"github.com/alexedwards/argon2id"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/spf13/cobra"
	"github.com/tkahng/authgo/internal/conf"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/queries"
	"github.com/tkahng/authgo/internal/stores"
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
		confdb := conf.GetConfig[conf.DBConfig]()

		dbx := database.CreateQueries(ctx, confdb.DatabaseUrl)
		userStore := stores.NewDbUserStore(dbx)
		authStore := stores.NewDbAuthStore(dbx)

		rbacStore := stores.NewDbRBACStore(dbx)
		err := rbacStore.EnsureRoleAndPermissions(ctx, "superuser", "superuser")
		if err != nil {
			return err
		}

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
		if user == nil {
			hash, err := security.CreateHash(args[1], argon2id.DefaultParams)
			if err != nil {
				return err
			}
			user, err = authStore.CreateUser(ctx, &models.User{
				Email: args[0],
			})
			if err != nil {
				return err
			}
			account := &models.UserAccount{
				Provider:          models.ProvidersCredentials,
				ProviderAccountID: args[0],
				UserID:            user.ID,
				Type:              models.ProviderTypeCredentials,
				Password:          types.Pointer(hash),
			}
			_, err = authStore.CreateUserAccount(ctx, account)
			if err != nil {
				return err
			}
		}
		if user != nil {
			claims, err := queries.FindUserWithRolesAndPermissionsByEmail(ctx, dbx, args[0])
			if err != nil {
				return err
			}
			if !slices.Contains(claims.Roles, "superuser") {
				err = queries.CreateUserRoles(ctx, dbx, user.ID, role.ID)
				if err != nil {
					return err
				}
			}
		}
		return nil
	},
}
