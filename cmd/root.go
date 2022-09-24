package cmd

import (
	"github.com/Riven-Spell/hydaelyn/bot/rolereact"
	"github.com/Riven-Spell/hydaelyn/database"
	"log"
	"os"
	"os/signal"

	"github.com/spf13/cobra"

	"github.com/Riven-Spell/hydaelyn/bot"
	"github.com/Riven-Spell/hydaelyn/common"
)

var rootCmd = &cobra.Command{
	Use:   "hydaelyn",
	Short: "hydaelyn is a discord bot for the Know Not Knights discord.",
	Long:  "hydaelyn is a discord bot for the Know Not Knights that allows for role-reacts.",

	Run: func(cmd *cobra.Command, args []string) {
		services := []common.LCMService{
			bot.LCMServiceBot(),
			rolereact.LCMServiceRoleReact,
		}
		services = append(services, DefaultServices...)

		// first, get LCM and it's friends open.
		lcm := common.CreateLifeCycleManager(services)
		logger := lcm.Services[common.LCMServiceNameLog].GetSvc().(*log.Logger)
		db := lcm.Services[common.LCMServiceNameSQL].GetSvc().(*database.Database)

		defer lcm.Shutdown()

		err := db.DBSetupCorrectly()
		if err != nil {
			logger.Println("Shutting down. The database was not set up correctly; encountered error:", err)
			logger.Println("Run hydaelyn setup --destructive or manually repair the database to continue.")
			return
		}

		logger.Println("Hydaelyn is running...")
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		for _ = range c {
			// sig is a ^C, handle it
			logger.Println("Shutting down...")
			return
		}
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&common.ConfigTarget, "cfg", "~/.hydaelyn/hydaelyn.yaml", "Specify a config rather than the assumed location (~/.hydaelyn/hydaelyn.yaml)")
}

func Execute() {
	rootCmd.Execute()
}
