package cmd

import (
	"database/sql"
	"log"
	"os"
	"os/signal"
	
	"github.com/spf13/cobra"
	
	"github.com/virepri/hydaelyn/bot"
	"github.com/virepri/hydaelyn/common"
)

var rootCmd = &cobra.Command{
	Use:   "hydaelyn",
	Short: "hydaelyn is a discord bot for the Know Not Knights discord.",
	Long:  "hydaelyn is a discord bot for the Know Not Knights that allows for role-reacts.",
	
	Run: func(cmd *cobra.Command, args []string) {
		services := []common.LCMService{
			bot.LCMServiceBot(),
		}
		services = append(services, common.DefaultServices...)
		
		// first, get LCM and it's friends open.
		lcm := common.CreateLifeCycleManager(services)
		logger := lcm.Services["log"].GetSvc().(*log.Logger)
		db := lcm.Services["SQL"].GetSvc().(*sql.DB)
		
		defer lcm.Shutdown()
		
		err := common.DBSetupCorrectly(db)
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
