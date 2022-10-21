package bot

import (
	"github.com/Riven-Spell/hydaelyn/bot/commands"
	"github.com/Riven-Spell/hydaelyn/database"
	"github.com/bwmarrin/discordgo"
	log2 "log"

	"github.com/Riven-Spell/hydaelyn/common"
)

func LCMServiceBot() common.LCMService {
	var session *discordgo.Session

	return common.LCMService{
		Name:         common.LCMServiceNameBot,
		Dependencies: []string{"SQL", "log", "config"},
		GetSvc: func() interface{} {
			return session
		},
		Startup: func(deps []interface{}) error {
			lcm := common.GetLifeCycleManager()
			db := deps[0].(*database.Database)
			log := deps[1].(*log2.Logger)
			config := deps[2].(*common.Config)

			var err error
			session, err = discordgo.New("Bot " + config.Discord.BotToken)

			if err != nil {
				return err
			}

			commands.GlobalCommands["help"] = commands.HelpCommand // HelpCommand creates a reference loop, so it has to be resolved by forcefully making use of HelpCommand later.

			err = commands.RegisterGlobalCommands(session, config.Discord)
			if err != nil {
				log.Fatal(err)
			}

			session.Identify.Intents = discordgo.IntentGuilds | discordgo.IntentGuildMessages | discordgo.IntentGuildMessageReactions | discordgo.IntentGuildScheduledEvents

			session.AddHandler(commands.GetCommandHandler(lcm, db, log))

			err = session.Open()

			return err
		},
		Shutdown: func() error {
			return session.Close()
		},
	}
}
