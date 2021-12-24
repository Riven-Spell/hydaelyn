package bot

import (
	"github.com/bwmarrin/discordgo"
	
	"github.com/virepri/hydaelyn/common"
)

func LCMServiceBot() common.LCMService {
	var session *discordgo.Session
	
	return common.LCMService{
		Name:         "bot",
		Dependencies: []string{"SQL", "log", "config"},
		GetSvc: func() interface{} {
			return session
		},
		Startup: func() error {
			lcm := common.GetLifeCycleManager()
			config := lcm.Services["config"].GetSvc().(*common.Config)
			
			var err error
			session, err = discordgo.New("Bot " + config.BotToken)
			
			if err != nil {
				return err
			}
			
			session.AddHandler(HandleCommand)
			session.AddHandler(HandleRoleReactAdd)
			session.AddHandler(HandleRoleReactRemove)
			
			session.Identify.Intents = discordgo.IntentsGuilds | discordgo.IntentsGuildMessages | discordgo.IntentsGuildMessageReactions
			
			err = session.Open()
			
			return err
		},
		Shutdown: func() error {
			return session.Close()
		},
	}
}
