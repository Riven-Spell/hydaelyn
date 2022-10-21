package commands

import (
	"fmt"
	"github.com/Riven-Spell/hydaelyn/bot/commands/events"
	"github.com/Riven-Spell/hydaelyn/common"
	"github.com/Riven-Spell/hydaelyn/database"
	"github.com/Riven-Spell/hydaelyn/database/queries"
	"github.com/bwmarrin/discordgo"
	"log"
)

func HandleEventCreate(s *discordgo.Session, i *discordgo.InteractionCreate, lcm *common.LifeCycleManager, db *database.Database, log *log.Logger) {
	options := i.ApplicationCommandData().Options[0].Options
	input := queries.AutoEvent{
		GuildID: i.GuildID,
	}

	var ok bool
	var warning string

	for _, v := range options {
		switch v.Name {
		case "frequency":
			var freq string
			freq, ok = v.Value.(string)

			if !ok {
				log.Printf("%s: user supplied invalid `frequency` format", i.ID)
				TryRespond(s, i, "Invalid frequency. Valid options are: Daily, Weekly, Monthly, Yearly", log)
				return
			}
			input.Frequency = queries.ParseFrequency(freq)

			if input.Frequency == queries.FrequencyNone {
				log.Printf("%s: frequency %s is invalid", i.ID, v.Value.(string))
				TryRespond(s, i, "Invalid frequency '"+v.Value.(string)+"'! Valid options are: Daily, Weekly, Monthly, Yearly", log)
				return
			}

			if input.Frequency >= queries.FrequencyMonthly {
				warning = " NOTE: Monthly/Yearly events may have inconsistent timing if they land on a nonexistent day; at which case they will be scheduled at the end of the month they would originally be in instead. e.g. Feb. 29th -> 28th on non-leap years."
			}
		case "initial_event":
			eventID, ok := v.Value.(string)

			if !ok {
				log.Printf("%s: user supplied invalid `initial_event` format", i.ID)
				TryRespond(s, i, "Invalid `initial_event`. Can be event ID.", log)
				return
			}

			log.Printf("%s: finding event for ID %s", i.ID, eventID)
			targetEvent, err := s.GuildScheduledEvent(i.GuildID, eventID, false)
			if err != nil || targetEvent == nil {
				log.Printf("%s: event not found: %s", i.ID, eventID)
				TryRespond(s, i, "Event ID "+eventID+" not found.", log)
				return
			}

			input = queries.AutoEvent{
				Name:         targetEvent.Name,
				Description:  targetEvent.Description,
				GuildID:      i.GuildID,
				GuildEventID: targetEvent.ID,
				Location:     targetEvent.ChannelID,
				EntityType:   targetEvent.EntityType,
				PrivacyLevel: targetEvent.PrivacyLevel,
				Frequency:    input.Frequency,
				CurrentEvent: targetEvent.ScheduledStartTime,
			}
		}
	}

	scheduler := lcm.Services[common.LCMServiceNameAutoScheduler].GetSvc().(*events.AutoSchedulerService)

	if !scheduler.IsLive() {
		log.Printf("%s: event scheduler is not live", i.ID)
		InternalError(s, i, log)
		return
	}

	err := scheduler.ScheduleNewEvent(input)
	if err != nil {
		log.Printf("%s: failed to create new event: %s", i.ID, err.Error())
		InternalError(s, i, log)
	} else {
		log.Printf("%s: Successfully created event `%s`", i.ID, input.Name)
		TryRespond(s, i, fmt.Sprintf("Successfully created event `%s`."+warning, input.Name), log)
	}
}
