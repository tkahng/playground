package cmd

import (
	"github.com/spf13/cobra"
)

func NewVerifactionCmd() *cobra.Command {
	verificationCmd.AddCommand(seedRolesCmd, seedUserCmd)
	return verificationCmd
}

var verificationCmd = &cobra.Command{
	Use:   "verification",
	Short: "verification",
	RunE: func(cmd *cobra.Command, args []string) error {
		// ctx := cmd.Context()
		// conf := conf.AppConfigGetter()

		// dbx := core.NewBobFromConf(ctx, conf.Db)
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
