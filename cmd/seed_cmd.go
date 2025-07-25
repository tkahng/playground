package cmd

import (
	"errors"
	"log/slog"
	"time"

	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/spf13/cobra"
	"github.com/tkahng/playground/internal/conf"
	"github.com/tkahng/playground/internal/core"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/services"
	"github.com/tkahng/playground/internal/stores"
	"github.com/tkahng/playground/internal/tools/slug"
)

func NewSeedCmd() *cobra.Command {
	seedCmd.AddCommand(seedRolesCmd)
	seedCmd.AddCommand(seedUserCmd)
	seedCmd.AddCommand(seedTeam)
	return seedCmd
}

var seedCmd = &cobra.Command{
	Use:   "seed",
	Short: "seed",
}

var seedRolesCmd = &cobra.Command{
	Use:   "roles",
	Short: "seed roles",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		conf := conf.GetConfig[conf.DBConfig]()

		dbx := database.CreateQueriesContext(ctx, conf.DatabaseUrl)
		rbacStore := stores.NewDbRBACStore(dbx)
		err := rbacStore.EnsureRoleAndPermissions(ctx, "superuser", "superuser", "advanced", "pro", "basic")
		if err != nil {
			slog.Error(
				"error at createing roles",
				"error", err,
				"role", "superuser",
			)
		}
		err = rbacStore.EnsureRoleAndPermissions(ctx, "advanced", "advanced", "pro", "basic")
		if err != nil {
			slog.Error(
				"error at createing roles",
				"error", err,
				"role", "basic",
			)
		}
		err = rbacStore.EnsureRoleAndPermissions(ctx, "pro", "pro", "basic")
		if err != nil {
			slog.Error(
				"error at createing roles",
				"error", err,
				"role", "basic",
			)
		}
		err = rbacStore.EnsureRoleAndPermissions(ctx, "basic", "basic")
		if err != nil {
			slog.Error(
				"error at createing roles",
				"error", err,
				"role", "basic",
			)
		}
		return err
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
		var verifiedAt *time.Time
		if verirfied == "true" {
			t := time.Now()
			verifiedAt = &t
		}
		ctx := cmd.Context()
		cfg := conf.AppConfigGetter()
		app := core.BootstrappedApp(cfg)
		params := &services.AuthenticationInput{
			Email:           email,
			Password:        &password,
			EmailVerifiedAt: verifiedAt,
			Provider:        models.ProvidersCredentials,
			Type:            models.ProviderTypeCredentials,
		}
		_, err := app.Auth().Authenticate(
			ctx,
			params,
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
		app := core.BootstrappedApp(cfg)
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
