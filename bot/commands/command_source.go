package commands

import (
	"github.com/Riven-Spell/hydaelyn/common"
	"github.com/Riven-Spell/hydaelyn/database"
	"github.com/bwmarrin/discordgo"
	"log"
)

var CommandSource = Command{
	Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate, lcm *common.LifeCycleManager, db *database.Database, log *log.Logger) {
		TryRespond(s, i, "My source code is available at: https://github.com/Riven-Spell/hydaelyn", log)
	},
	DgCommand: &discordgo.ApplicationCommand{
		Name:        "source",
		Description: "Links to the source code.",
		Type:        discordgo.ChatApplicationCommand,
	},
	HelpText: "Links to the source code.",
}
