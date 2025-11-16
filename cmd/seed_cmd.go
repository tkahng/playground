package cmd

import (
	"errors"

	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/spf13/cobra"
	"github.com/tkahng/playground/internal/auth"
	"github.com/tkahng/playground/internal/conf"
	"github.com/tkahng/playground/internal/core"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/shared"
	"github.com/tkahng/playground/internal/stores"
	"github.com/tkahng/playground/internal/tools/slug"
)

func NewSeedCmd() *cobra.Command {

	seedCmd.AddCommand(seedRolesCmd)
	seedCmd.AddCommand(seedUserCmd)
	seedCmd.AddCommand(seedAllCmd)
	seedCmd.AddCommand(seedTeam)
	return seedCmd
}

var seedCmd = &cobra.Command{
	Use:   "seed",
	Short: "seed",
}

var seedAllCmd = &cobra.Command{
	Use:   "all",
	Short: "seed all",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := seedRolesCmd.RunE(cmd, args); err != nil {
			return err
		}
		if err := stripeSyncCmd.RunE(cmd, args); err != nil {
			return err
		}
		if err := stripeRolesCmd.RunE(cmd, args); err != nil {
			return err
		}
		return nil
	},
}

var seedRolesCmd = &cobra.Command{
	Use:   "roles",
	Short: "seed roles",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		conf := conf.GetConfig[conf.DBConfig]()

		dbx := database.CreateNewQueriesContext(ctx, conf.GetDatabaseUrl())
		defer dbx.Close()

		rbacStore := stores.NewDbRBACStore(dbx)
		return rbacStore.CreateRolesAndPermissions(ctx, shared.KnownRoleNamesPermissionsMap)
	},
}

var seedUserCmd = &cobra.Command{
	Use:     "user",
	Short:   "seed user",
	Example: "seed user admin@k2dv.io Password123! true",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 3 {
			return errors.New("missing email and password arguments")
		}

		if args[0] == "" || is.EmailFormat.Validate(args[0]) != nil {
			return errors.New("mrror missing or invalid email address")
		}
		email := args[0]
		password := args[1]
		verirfied := args[2]
		ctx := cmd.Context()
		cfg := conf.AppConfigGetter()
		app := core.NewApp(cfg)
		defer app.Close()
		_, err := app.Auth().Signup(
			ctx,
			&auth.SignupInput{
				Email:    email,
				Password: password,
				Verified: verirfied == "true",
			},
		)
		return err
	},
}

var seedTeam = &cobra.Command{
	Use:     "team",
	Short:   "seed team",
	Example: "seed team admin@k2dv.io teamSlug",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			return errors.New("missing email and password arguments")
		}

		if args[0] == "" || is.EmailFormat.Validate(args[0]) != nil {
			return errors.New("mrror missing or invalid email address")
		}
		email := args[0]
		slug := slug.NewSlug(args[1])

		ctx := cmd.Context()
		cfg := conf.AppConfigGetter()
		app := core.NewApp(cfg)
		defer app.Close()
		user, err := app.Adapter().User().FindUser(ctx, &stores.UserFilter{
			Emails: []string{email},
		})
		if err != nil {
			return err
		}
		if user == nil {
			return errors.New("user not found")
		}

		team, err := app.Team().CreateTeamWithOwner(
			ctx,
			slug,
			slug,
			user.ID,
		)
		if err != nil {
			return err
		}
		if team == nil {
			return errors.New("team not found")
		}
		return err
	},
}
