package commands

import (
	_ "embed"
	"fmt"
	"github.com/Riven-Spell/hydaelyn/common"
	"github.com/Riven-Spell/hydaelyn/database"
	"github.com/bwmarrin/discordgo"
	"log"
)

//go:embed helptext/help_long.txt
var helpHelptext string

var HelpCommand = Command{
	Handler: HelpHandler,
	DgCommand: &discordgo.ApplicationCommand{
		Version:      "1.0",
		Type:         discordgo.ChatApplicationCommand,
		Name:         "help",
		DMPermission: common.Pointer(true),
		Description:  "List commands & information",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "command",
				Description: "Target command to view",
				Required:    false,
			},
		},
	},
	HelpText: helpHelptext,
}

func HelpHandler(s *discordgo.Session, i *discordgo.InteractionCreate, lcm *common.LifeCycleManager, db *database.Database, log *log.Logger) {
	data := i.ApplicationCommandData()
	if len(data.Options) >= 1 {
		cmdName, ok := data.Options[0].Value.(string)
		if !ok {
			TryRespond(s, i, "Command must be a string.", log)
			return
		}
		cmd, ok := GlobalCommands[cmdName]
		if !ok {
			TryRespond(s, i, "Command does not exist.", log)
			return
		}

		if cmd.DgCommand == nil {
			log.Printf("%s: command %s lacks a dgCommand.", i.ID, cmdName)
			InternalError(s, i, log)
			return
		}

		TryRespond(s, i, fmt.Sprintf("**%s**\n\n%s", cmdName, cmd.HelpText), log)
	} else {
		out := ""

		for k, v := range GlobalCommands {
			if v.DgCommand == nil {
				log.Printf("%s: command %s lacks a dgCommand.", i.ID, k)
				continue
			}

			out += fmt.Sprintf("-%s: %s\n", k, v.DgCommand.Description)
		}

		TryRespond(s, i, out, log)
	}
}
