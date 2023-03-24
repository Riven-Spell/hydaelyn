package commands

import (
	"github.com/Riven-Spell/hydaelyn/common"
	"github.com/Riven-Spell/hydaelyn/database"
	"github.com/Riven-Spell/hydaelyn/database/queries"
	"github.com/bwmarrin/discordgo"
	"log"
)

func HandleEventPurge(s *discordgo.Session, i *discordgo.InteractionCreate, lcm *common.LifeCycleManager, db *database.Database, log *log.Logger) {
	out := make([]queries.AutoEvent, 0)
	err := db.Tx(queries.FindEvents(i.GuildID, &out))
	if err != nil {
		log.Printf("%s: failed to find events: %s", i.ID, err.Error())
		InternalError(s, i, log)
		return
	}

	for _, v := range out {
		_, err := s.GuildScheduledEvent(v.GuildID, v.GuildEventID, false)
		if err == nil {
			continue
		}

		err = db.Tx(queries.DeleteEvent(v.GuildID, v.GuildEventID))
		if err != nil {
			log.Printf("%s: failed to delete event: %s", i.ID, err.Error())
			InternalError(s, i, log)
			return
		}
	}

	TryRespond(s, i, "Purging non-matching events...", log)
}
