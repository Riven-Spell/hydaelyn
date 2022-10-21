package commands

import (
	"fmt"
	"github.com/Riven-Spell/hydaelyn/common"
	"github.com/Riven-Spell/hydaelyn/database"
	"github.com/Riven-Spell/hydaelyn/database/queries"
	"github.com/bwmarrin/discordgo"
	"log"
)

func HandleEventUpdate(s *discordgo.Session, i *discordgo.InteractionCreate, lcm *common.LifeCycleManager, db *database.Database, log *log.Logger) {
	options := i.ApplicationCommandData().Options[0].Options

	var eventID, warning string
	var ok bool
	var frequency queries.EventFrequency

	for _, v := range options {
		switch v.Name {
		case "series":
			eventID, ok = v.Value.(string)
			if !ok {
				log.Printf("%s: user supplied invalid `series` format", i.ID)
				TryRespond(s, i, "Invalid `series`. Can be event ID.", log)
				return
			}
		case "frequency":
			var freq string
			freq, ok = v.Value.(string)

			if !ok {
				log.Printf("%s: user supplied invalid `frequency` format", i.ID)
				TryRespond(s, i, "Invalid frequency. Valid options are: Daily, Weekly, Monthly, Yearly", log)
				return
			}
			frequency = queries.ParseFrequency(freq)

			if frequency == queries.FrequencyNone {
				log.Printf("%s: frequency %s is invalid", i.ID, v.Value.(string))
				TryRespond(s, i, "Invalid frequency '"+v.Value.(string)+"'! Valid options are: Daily, Weekly, Monthly, Yearly", log)
				return
			}

			if frequency >= queries.FrequencyMonthly {
				warning = " NOTE: Monthly/Yearly events may have inconsistent timing if they land on a nonexistent day; at which case they will be scheduled at the end of the month they would originally be in instead. e.g. Feb. 29th -> 28th on non-leap years."
			}
		}
	}

	tx, err := db.GetTransaction(nil)
	if err != nil {
		log.Printf("%s: could not open transaction: %s", i.ID, err.Error())
		InternalError(s, i, log)
		return
	}
	defer tx.Commit()

	event, err := s.GuildScheduledEvent(i.GuildID, eventID, false)
	if err != nil {
		log.Printf("%s: failed to find event: %s", i.ID, err.Error())
		InternalError(s, i, log)
		return
	}

	var targetSeries queries.AutoEvent
	err = tx.Do(queries.FindEvent(i.GuildID, eventID, &targetSeries))
	if err != nil {
		log.Printf("%s: could not find series `%s`: %s", i.ID, eventID, err.Error())
		TryRespond(s, i, fmt.Sprintf("Could not find series `%s`.", eventID), log)
		return
	}

	targetSeries.Frequency = frequency

	err = tx.Do(queries.UpdateEventData(targetSeries))
	if err != nil {
		log.Printf("%s: could not update event data: %s", i.ID, err.Error())
		InternalError(s, i, log)
		return
	}

	TryRespond(s, i, fmt.Sprintf("Updated frequency of series `%s` to %s."+warning, event.Name, frequency.String()), log)
}
