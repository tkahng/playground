package cmd

import (
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
	"github.com/tkahng/authgo/internal/conf"
	"github.com/tkahng/authgo/internal/core"
	"github.com/tkahng/authgo/internal/db/models"
	"github.com/tkahng/authgo/internal/db/seeders"
	"github.com/tkahng/authgo/internal/repository"
)

func NewSeedCmd() *cobra.Command {
	seedCmd.AddCommand(seedRolesCmd, seedUserCmd)
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
		conf := conf.AppConfigGetter()

		dbx := core.NewBobFromConf(ctx, conf.Db)
		err := repository.EnsureRoleAndPermissions(ctx, dbx, "superuser", "superuser", "advanced", "pro", "basic")
		if err != nil {
			slog.Error(
				"error at createing roles",
				"error", err,
				"role", "superuser",
			)
		}
		err = repository.EnsureRoleAndPermissions(ctx, dbx, "advanced", "advanced", "pro", "basic")
		if err != nil {
			slog.Error(
				"error at createing roles",
				"error", err,
				"role", "basic",
			)
		}
		err = repository.EnsureRoleAndPermissions(ctx, dbx, "pro", "pro", "basic")
		if err != nil {
			slog.Error(
				"error at createing roles",
				"error", err,
				"role", "basic",
			)
		}
		err = repository.EnsureRoleAndPermissions(ctx, dbx, "basic", "basic")
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
	Use:   "users",
	Short: "seed users",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		conf := conf.AppConfigGetter()

		dbx := core.NewBobFromConf(ctx, conf.Db)

		err := seeders.UserCredentialsFactory(ctx, dbx, 10)
		// err := repository.PopulateRoles(ctx, dbx)
		if err != nil {

			return fmt.Errorf("error at createing users: %w", err)
		}
		err = seeders.UserOauthFactory(ctx, dbx, 10, models.ProvidersGoogle)
		if err != nil {
			return fmt.Errorf("error at createing users: %w", err)
		}
		err = seeders.UserOauthFactory(ctx, dbx, 10, models.ProvidersGithub)
		if err != nil {
			return fmt.Errorf("error at createing users: %w", err)
		}
		return nil
	},
}
