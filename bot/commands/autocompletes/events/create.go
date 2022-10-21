package events

import (
	"github.com/Riven-Spell/hydaelyn/common"
	"github.com/Riven-Spell/hydaelyn/database"
	"github.com/bwmarrin/discordgo"
	"log"
	"sort"
)

type EventsByTime []*discordgo.GuildScheduledEvent

func (e EventsByTime) Len() int      { return len(e) }
func (e EventsByTime) Swap(i, j int) { e[i], e[j] = e[j], e[i] }
func (e EventsByTime) Less(i, j int) bool {
	return e[i].ScheduledStartTime.Unix() < e[j].ScheduledStartTime.Unix()
}

func CreateEventInitialEvent(s *discordgo.Session, i *discordgo.InteractionCreate, lcm *common.LifeCycleManager, db *database.Database, logger *log.Logger) {
	events, err := s.GuildScheduledEvents(i.GuildID, false)
	if err != nil {
		log.Printf("%s failed to get guild events: %s", i.ID, err.Error())
		return
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

func CreateEventFrequency(s *discordgo.Session, i *discordgo.InteractionCreate, lcm *common.LifeCycleManager, db *database.Database, logger *log.Logger) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionApplicationCommandAutocompleteResult,
		Data: &discordgo.InteractionResponseData{
			Choices: []*discordgo.ApplicationCommandOptionChoice{
				{
					Name:  "Daily",
					Value: "daily",
				},
				{
					Name:  "Weekly",
					Value: "weekly",
				},
				{
					Name:  "Monthly",
					Value: "monthly",
				},
				{
					Name:  "Yearly",
					Value: "yearly",
				},
			},
		},
	})

	if err != nil {
		log.Printf("%s: failed to send autocomplete response: %s", i.ID, err.Error())
		return
	}
	log.Printf("%s: sent autocomplete response", i.ID)
}
