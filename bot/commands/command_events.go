package commands

import (
	_ "embed"
	"fmt"
	autocomplete "github.com/Riven-Spell/hydaelyn/bot/commands/autocompletes/events"
	"github.com/Riven-Spell/hydaelyn/common"
	"github.com/Riven-Spell/hydaelyn/database"
	"github.com/bwmarrin/discordgo"
	"log"
)

//go:embed helptext/events_long.txt
var EventsHelptext string

var CommandEvents = Command{
	Handler: CommandEventsHandler,
	AutoCompleteHandlers: map[string]interface{}{
		"create": map[string]interface{}{
			"frequency":     autocomplete.CreateEventFrequency,
			"initial_event": autocomplete.CreateEventInitialEvent,
		},
		"details": map[string]interface{}{
			"series": autocomplete.EventGetAvailableSeries,
		},
		"delete": map[string]interface{}{
			"series": autocomplete.EventGetAvailableSeries,
		},
		"update": map[string]interface{}{
			"series":    autocomplete.EventGetAvailableSeries, // todo:
			"frequency": autocomplete.CreateEventFrequency,
		},
	},
	DgCommand: &discordgo.ApplicationCommand{
		Type:                     discordgo.ChatApplicationCommand,
		Name:                     "events",
		Description:              "Wrangle auto-scheduled events.",
		DefaultMemberPermissions: common.Pointer(int64(discordgo.PermissionManageEvents)),
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "create",
				Description: "Create a new auto-scheduled event. Can clone existing events.",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:         discordgo.ApplicationCommandOptionString,
						Name:         "initial_event",
						Description:  "Initial event. Autocorrect presents the 25 soonest non-series events.",
						Required:     true,
						Autocomplete: true,
					},
					{
						Type:         discordgo.ApplicationCommandOptionString,
						Name:         "frequency",
						Description:  "How frequently the event should exist",
						Required:     true,
						Autocomplete: true,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "details",
				Description: "Check frequency of auto-scheduled events.",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:         discordgo.ApplicationCommandOptionString,
						Name:         "series",
						Description:  "Target series. Autocorrect prevents the 25 series with the soonest events.",
						Required:     true,
						Autocomplete: true,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "list",
				Description: "List auto-scheduled events.",
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "delete",
				Description: "Delete auto-scheduled events.",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:         discordgo.ApplicationCommandOptionString,
						Name:         "series",
						Description:  "Target series. Autocorrect prevents the 25 series with the soonest events.",
						Required:     true,
						Autocomplete: true,
					},
					{
						Type:        discordgo.ApplicationCommandOptionBoolean,
						Name:        "delete_finale",
						Description: "Delete the final event? (default: true)",
						Required:    false,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "update",
				Description: "Adjust the frequency of an event.",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:         discordgo.ApplicationCommandOptionString,
						Name:         "series",
						Description:  "Target series. Autocorrect prevents the 25 series with the soonest events.",
						Required:     true,
						Autocomplete: true,
					},
					{
						Type:         discordgo.ApplicationCommandOptionString,
						Name:         "frequency",
						Description:  "How frequently the event should exist",
						Required:     true,
						Autocomplete: true,
					},
				},
			},
		},
	},
	HelpText: EventsHelptext,
}

func CommandEventsHandler(s *discordgo.Session, i *discordgo.InteractionCreate, lcm *common.LifeCycleManager, db *database.Database, log *log.Logger) {
	switch i.ApplicationCommandData().Options[0].Name {
	case "create":
		HandleEventCreate(s, i, lcm, db, log)
	case "details":
		HandleEventDetails(s, i, lcm, db, log)
	case "list":
		HandleEventListing(s, i, lcm, db, log)
	case "delete":
		HandleEventDelete(s, i, lcm, db, log)
	case "update":
		HandleEventUpdate(s, i, lcm, db, log)
	default:
		fmt.Println("dafuk")
	}
}
