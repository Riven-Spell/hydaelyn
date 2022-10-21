package commands

import (
	"fmt"
	"github.com/Riven-Spell/hydaelyn/common"
	"github.com/Riven-Spell/hydaelyn/database"
	"github.com/Riven-Spell/hydaelyn/database/queries"
	"github.com/bwmarrin/discordgo"
	"log"
)

func HandleEventDelete(s *discordgo.Session, i *discordgo.InteractionCreate, lcm *common.LifeCycleManager, db *database.Database, log *log.Logger) {
	options := i.ApplicationCommandData().Options[0].Options

	var eventID string
	var deleteFinale, ok bool

	deleteFinale = true // default

	for _, v := range options {
		switch v.Name {
		case "delete_finale":
			deleteFinale, ok = v.Value.(bool)
			if !ok {
				log.Printf("%s: user supplied invalid `delete_finale` format", i.ID)
				TryRespond(s, i, "Invalid `delete_finale`. Can be a boolean (true/false).", log)
				return
			}
		case "series":
			eventID, ok = v.Value.(string)
			if !ok {
				log.Printf("%s: user supplied invalid `series` format", i.ID)
				TryRespond(s, i, "Invalid `series`. Can be event ID.", log)
				return
			}

			return
		}
	}

	tx, err := db.GetTransaction(nil)
	if err != nil {
		log.Printf("%s: failed to open db : %s", i.ID, err.Error())
		InternalError(s, i, log)
		return
	}

	err = tx.Do(queries.DeleteEvent(i.GuildID, eventID))
	if err != nil {
		log.Printf("%s: failed to delete event from database: %s", i.ID, err.Error())
		InternalError(s, i, log)
		return
	}

	if deleteFinale {
		err := s.GuildScheduledEventDelete(i.GuildID, eventID)
		if err != nil {
			log.Printf("%s: failed to delete event: %s", i.ID, err.Error())
		}
	}

	log.Printf("%s: finished event delete")
	TryRespond(s, i, fmt.Sprintf("Deleted autoscheduled event series."), log)
}
