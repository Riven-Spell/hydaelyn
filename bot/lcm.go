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
		Startup: func() error {
			lcm := common.GetLifeCycleManager()
			config := lcm.Services[common.LCMServiceNameConfig].GetSvc().(*common.Config)
			log := lcm.Services[common.LCMServiceNameLog].GetSvc().(*log2.Logger)
			db := lcm.Services[common.LCMServiceNameSQL].GetSvc().(*database.Database)

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

			session.Identify.Intents = discordgo.IntentsGuilds | discordgo.IntentsGuildMessages | discordgo.IntentsGuildMessageReactions

			session.AddHandler(commands.GetCommandHandler(lcm, db, log))

			err = session.Open()

			return err
		},
		Shutdown: func() error {
			return session.Close()
		},
	}
}
