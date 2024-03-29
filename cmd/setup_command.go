package cmd

import (
	"github.com/Riven-Spell/hydaelyn/database"
	"github.com/Riven-Spell/hydaelyn/database/setup"
	"log"

	"github.com/spf13/cobra"

	"github.com/Riven-Spell/hydaelyn/common"
)

var setupDestructive bool
var setupUpgrade bool

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Builds config file; if config is ready, sets up database tables. WARNING: This command is destructive.",

	Run: func(cmd *cobra.Command, args []string) {
		lcm := common.CreateLifeCycleManager(DefaultServices)
		logger := lcm.Services[common.LCMServiceNameLog].GetSvc().(*log.Logger)
		db := lcm.Services[common.LCMServiceNameSQL].GetSvc().(*database.Database)

		err := setup.DBSetupCorrectly(db)

		if err == nil && !setupDestructive {
			logger.Println("The database already has the necessary tables and potentially information. Use --destructive to reset the database. Existing role-react messages, prefixes, etc. will be broken.")
			return
		}

		setupMode := common.Ternary(setupDestructive, setup.SetupModeDestructive, common.Ternary(setupUpgrade, setup.SetupModeUpgrade, setup.SetupModeInitial))
		err = setup.SetupDatabase(db, setupMode)

		if err != nil {
			logger.Println("Failed to initialize database:", err)
			return
		}

		logger.Println("Successfully set up database.")
	},
}

func init() {
	// todo: break one more time and change to setup-mode and default to upgrade
	setupCmd.PersistentFlags().BoolVar(&setupDestructive, "destructive", false, "DESTROY any and all hydaelyn-related data in the database & remake it. Existing role-react messages, etc. will be broken.")
	setupCmd.PersistentFlags().BoolVar(&setupUpgrade, "upgrade", true, "Upgrades the existing database to the latest setup.")

	rootCmd.AddCommand(setupCmd)
}
