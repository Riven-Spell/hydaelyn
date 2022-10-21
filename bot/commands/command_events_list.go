package commands

import (
	"fmt"
	"github.com/Riven-Spell/hydaelyn/common"
	"github.com/Riven-Spell/hydaelyn/database"
	"github.com/Riven-Spell/hydaelyn/database/queries"
	"github.com/bwmarrin/discordgo"
	"log"
)

func HandleEventListing(s *discordgo.Session, i *discordgo.InteractionCreate, lcm *common.LifeCycleManager, db *database.Database, log *log.Logger) {
	out := make([]queries.AutoEvent, 0)
	err := db.Tx(queries.FindEvents(i.GuildID, &out))
	if err != nil {
		log.Printf("%s: failed to find events: %s", i.ID, err.Error())
		InternalError(s, i, log)
		return
	}

	toWrite := "Auto-scheduled Events:"

	for _, v := range out {
		toWrite += "\n"
		toWrite += fmt.Sprintf("`%s`: %s", v.Name, v.Frequency.String())
	}

	TryRespond(s, i, toWrite, log)
}
