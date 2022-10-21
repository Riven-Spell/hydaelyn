package events

import (
	"github.com/Riven-Spell/hydaelyn/common"
	"github.com/Riven-Spell/hydaelyn/database/queries"
	"github.com/bwmarrin/discordgo"
	"time"
)

func (a *AutoSchedulerService) getEventDeleteHandler() func(session *discordgo.Session, update *discordgo.GuildScheduledEventDelete) {
	return func(session *discordgo.Session, delete *discordgo.GuildScheduledEventDelete) {
		delete.Status = discordgo.GuildScheduledEventStatusCanceled
		a.getEventUpdateHandler()(session, &discordgo.GuildScheduledEventUpdate{delete.GuildScheduledEvent})
	}
}

func (a *AutoSchedulerService) getEventUpdateHandler() func(session *discordgo.Session, update *discordgo.GuildScheduledEventUpdate) {
	return func(session *discordgo.Session, update *discordgo.GuildScheduledEventUpdate) {
		if update.GuildScheduledEvent == nil {
			a.log.Printf("event%s: event update caught, but no event included", update.ID)
		}

		if !a.IsLive() {
			a.log.Printf("event%s: update of event caught, but handler is not live.", update.ID)
			return
		}

		a.log.Printf("event%s: searching for event", update.ID)

		tx, err := a.db.GetTransaction(nil)
		if err != nil {
			a.log.Printf("event%s: failed to open transaction: %s", update.ID, err.Error())
		}

		defer tx.Commit()

		var targetEvent queries.AutoEvent
		err = tx.Do(queries.FindEvent(update.GuildID, update.ID, &targetEvent))
		if err != nil {
			a.log.Printf("event%s: could not find associated event: %s", update.ID, err.Error())
			return
		}

		switch update.Status {
		case discordgo.GuildScheduledEventStatusScheduled, discordgo.GuildScheduledEventStatusActive:
			a.log.Printf("event%s: Event does not need rescheduling")
			return // nothing to do
		case discordgo.GuildScheduledEventStatusCanceled, discordgo.GuildScheduledEventStatusCompleted:
			// schedule the next event
			nextStart := targetEvent.NextEventStart(update.ScheduledStartTime)
			var nextEnd *time.Time
			if update.ScheduledEndTime != nil {
				nextEnd = common.Pointer(targetEvent.NextEventStart(*update.ScheduledEndTime))
			}

			newEvent := &discordgo.GuildScheduledEventParams{
				Name:               update.Name,
				Description:        update.Description,
				ScheduledStartTime: &nextStart,
				ScheduledEndTime:   nextEnd,
				EntityType:         update.EntityType,
				PrivacyLevel:       update.PrivacyLevel, // copy the old privacy level for now.
			}

			if newEvent.EntityType == discordgo.GuildScheduledEventEntityTypeExternal {
				newEvent.EntityMetadata = &update.EntityMetadata
			} else {
				newEvent.ChannelID = update.ChannelID
			}

			result, err := session.GuildScheduledEventCreate(update.GuildID, newEvent)
			if err != nil {
				a.log.Printf("event%s: failed to schedule new event: %s", update.ID, err.Error())
			}

			err = tx.Do(queries.DeleteEvent(update.GuildID, targetEvent.GuildEventID))
			if err != nil {
				a.log.Printf("event%s: failed to delete old event: %s", update.ID, err.Error())
			}

			if result != nil {
				targetEvent.GuildEventID = result.ID
				err = tx.Do(queries.CreateEvent(targetEvent))
				if err != nil {
					a.log.Printf("event%s: failed to recreate new event: %s", update.ID, err.Error())
				}
			}
		}
	}
}
