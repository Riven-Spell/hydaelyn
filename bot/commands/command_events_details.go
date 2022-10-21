package commands

import (
	"fmt"
	"github.com/Riven-Spell/hydaelyn/common"
	"github.com/Riven-Spell/hydaelyn/database"
	"github.com/Riven-Spell/hydaelyn/database/queries"
	"github.com/bwmarrin/discordgo"
	"log"
)

func HandleEventDetails(s *discordgo.Session, i *discordgo.InteractionCreate, lcm *common.LifeCycleManager, db *database.Database, log *log.Logger) {
	options := i.ApplicationCommandData().Options[0].Options

	var eventID string
	var ok bool

	for _, v := range options {
		switch v.Name {
		case "series":
			eventID, ok = v.Value.(string)
			if !ok {
				log.Printf("%s: user supplied invalid `series` format", i.ID)
				TryRespond(s, i, "Invalid `series`. Can be event ID.", log)
				return
			}
		}
	}

	var targetSeries queries.AutoEvent
	err := db.Tx(queries.FindEvent(i.GuildID, eventID, &targetSeries))
	if err != nil {
		log.Printf("%s: could not find series `%s`: %s", i.ID, eventID, err.Error())
		TryRespond(s, i, fmt.Sprintf("Could not find series `%s`.", eventID), log)
		return
	}

	TryRespond(s, i, fmt.Sprintf("Event `%s` frequency: %s. Name, Description, etc. can be edited within Discord.", targetSeries.Name, targetSeries.Frequency.String()), log)
}
