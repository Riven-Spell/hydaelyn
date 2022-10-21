package events

import (
	"github.com/Riven-Spell/hydaelyn/common"
	"github.com/Riven-Spell/hydaelyn/database"
	"github.com/Riven-Spell/hydaelyn/database/queries"
	"github.com/bwmarrin/discordgo"
	"log"
	"sort"
)

func EventGetAvailableSeries(s *discordgo.Session, i *discordgo.InteractionCreate, lcm *common.LifeCycleManager, db *database.Database, logger *log.Logger) {
	events, err := s.GuildScheduledEvents(i.GuildID, false)
	if err != nil {
		log.Printf("%s failed to get guild events: %s", i.ID, err.Error())
		return
	}

	autoEvents := make([]queries.AutoEvent, 0)
	err = db.Tx(queries.FindEvents(i.GuildID, &autoEvents))
	if err == nil {
		out := make([]*discordgo.GuildScheduledEvent, 0)
		aeMap := make(map[string]bool)

		for _, v := range autoEvents {
			aeMap[v.GuildEventID] = true
		}

		for _, v := range events {
			if _, ok := aeMap[v.ID]; ok {
				delete(aeMap, v.ID)
				out = append(out, v)
			}
		}

		events = out
	}

	sort.Sort(EventsByTime(events))

	out := make([]*discordgo.ApplicationCommandOptionChoice, common.Min(25, len(events)))

	for k, v := range events {
		out[k] = &discordgo.ApplicationCommandOptionChoice{
			Name:  v.Name,
			Value: v.ID,
		}
	}

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionApplicationCommandAutocompleteResult,
		Data: &discordgo.InteractionResponseData{
			Content: "Event IDs",
			Choices: out,
		},
	})

	if err != nil {
		log.Printf("%s: failed to send autocomplete response: %s", i.ID, err.Error())
		return
	}
	log.Printf("%s: sent autocomplete response", i.ID)
}
