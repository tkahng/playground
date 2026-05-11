package cmd

import (
	"context"
	"errors"

	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/spf13/cobra"
	"github.com/tkahng/playground/internal/auth"
	"github.com/tkahng/playground/internal/conf"
	"github.com/tkahng/playground/internal/core"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/shared"
	"github.com/tkahng/playground/internal/stores"
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
		cfg := conf.GetConfig[conf.DBConfig]()
		dbx, err := database.CreateNewQueriesContext(ctx, cfg.GetDatabaseURL())
		if err != nil {
			return err
		}
		defer dbx.Close()
		return SeedRoles(ctx, dbx)
	},
}

func SeedRoles(ctx context.Context, db database.Dbx) error {
	rbacStore := stores.NewDbRBACStore(db)
	if err := rbacStore.CreateRolesAndPermissions(ctx, shared.KnownRoleNamesPermissionsMap); err != nil {
		return err
	}
	return rbacStore.CreateTeamRolePermissions(ctx, shared.KnownTeamRolePermissionsMap)
}

var seedUserCmd = &cobra.Command{
	Use:     "user",
	Short:   "seed user",
	Example: "seed user admin@k2dv.io Password123! true",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		cfg := conf.AppConfigGetter()
		app := core.NewApp(cfg)
		defer app.Close()
		return SeedUser(ctx, app, args)
	},
}

func SeedUser(ctx context.Context, app core.App, args []string) error {
	if len(args) != 3 {
		return errors.New("missing email, password, and verified arguments")
	}
	if args[0] == "" || is.EmailFormat.Validate(args[0]) != nil {
		return errors.New("missing or invalid email address")
	}
	email := args[0]
	password := args[1]
	verified := args[2]
	_, err := app.Auth().Signup(ctx, &auth.SignupInput{
		Email:    email,
		Password: password,
		Verified: verified == "true",
	})
	return err
}

var seedTeam = &cobra.Command{
	Use:     "team",
	Short:   "seed team",
	Example: "seed team admin@k2dv.io teamSlug",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		cfg := conf.AppConfigGetter()
		app := core.NewApp(cfg)
		defer app.Close()
		return SeedTeam(ctx, app, args)
	},
}

func SeedTeam(ctx context.Context, app core.App, args []string) error {
	if len(args) != 2 {
		return errors.New("missing email and team name arguments")
	}
	if args[0] == "" || is.EmailFormat.Validate(args[0]) != nil {
		return errors.New("missing or invalid email address")
	}
	email := args[0]
	teamName := args[1]

	user, err := app.Adapter().User().FindUser(ctx, &stores.UserFilter{
		Emails: []string{email},
	})
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}
	team, err := app.Team().CreateTeamWithOwner(ctx, teamName, user.ID)
	if err != nil {
		return err
	}
	if team == nil {
		return errors.New("team not found")
	}
	return nil
}
