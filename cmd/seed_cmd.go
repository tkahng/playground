package cmd

import (
	"log/slog"

	"github.com/spf13/cobra"
	"github.com/tkahng/authgo/internal/conf"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/stores"
)

func NewSeedCmd() *cobra.Command {
	seedCmd.AddCommand(seedRolesCmd)
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

		dbx := database.CreateQueries(ctx, conf.DatabaseUrl)
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

// var seedUserCmd = &cobra.Command{
// 	Use:   "users",
// 	Short: "seed users",
// 	RunE: func(cmd *cobra.Command, args []string) error {
// 		ctx := cmd.Context()
// 		conf := conf.GetConfig[conf.DBConfig]()

// 		pool := db.CreatePool(ctx, conf)
// 		dbx := db.NewQueries(pool)
// 		// role, err := queries.FindRoleByName(ctx, dbx, "basic")
// 		// if err != nil {
// 		// 	return fmt.Errorf("error at createing users: %w", err)
// 		// }

// 		// // _, err = seeders.UserCredentialsRolesFactory(ctx, dbx, 20, role)
// 		// // err := repository.PopulateRoles(ctx, dbx)
// 		// if err != nil {

// 		// 	return fmt.Errorf("error at createing users: %w", err)
// 		// }
// 		// err = seeders.UserOauthFactory(ctx, dbx, 10, models.ProvidersGoogle)
// 		// if err != nil {
// 		// 	return fmt.Errorf("error at createing users: %w", err)
// 		// }
// 		// err = seeders.UserOauthFactory(ctx, dbx, 10, models.ProvidersGithub)
// 		// if err != nil {
// 		// 	return fmt.Errorf("error at createing users: %w", err)
// 		// }
// 		return nil
// 	},
// }
