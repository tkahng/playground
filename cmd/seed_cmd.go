package cmd

import (
	"fmt"

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
		err := repository.PopulateRolesFromTree(ctx, dbx)
		if err != nil {
			return err
		}
		return nil
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
