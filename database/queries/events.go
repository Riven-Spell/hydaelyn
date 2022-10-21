package queries

import (
	"github.com/Riven-Spell/hydaelyn/common"
	"github.com/Riven-Spell/hydaelyn/database"
	"github.com/bwmarrin/discordgo"
	"strings"
	"time"
)

type AutoEvent struct {
	Name         string
	Description  string
	GuildID      string
	GuildEventID string // Discord ID
	Location     string
	BaseDay      int // Sometimes, we have to schedule at the end of the month. What's the real day we should be working with?
	EntityType   discordgo.GuildScheduledEventEntityType
	PrivacyLevel discordgo.GuildScheduledEventPrivacyLevel
	Frequency    EventFrequency
	CurrentEvent time.Time
}

func (e AutoEvent) Ready() bool {
	return !(e.Name == "" || e.GuildID == "" || e.GuildEventID == "" || e.Location == "" || e.EntityType == 0 || e.Frequency == FrequencyNone || e.CurrentEvent.IsZero())
}

func (e AutoEvent) NextEventStart(lastStartTime time.Time) time.Time {
	var months time.Month = 1

	getDays := func(year int, month time.Month) time.Time {
		return time.Date(year, month+1, 0, 0, 0, 0, 0, lastStartTime.Location())
	}

	switch e.Frequency {
	case FrequencyDaily:
		return lastStartTime.Add(time.Hour * 24)
	case FrequencyWeekly:
		return lastStartTime.Add(time.Hour * 24 * 7)
	case FrequencyYearly:
		months = 12
		fallthrough
	case FrequencyMonthly:
		remainingDays := getDays(lastStartTime.Year(), lastStartTime.Month()+months+1).Day()
		targetDay := common.Ternary(e.BaseDay > remainingDays, remainingDays, e.BaseDay)
		return time.Date(lastStartTime.Year(), lastStartTime.Month()+months, targetDay, lastStartTime.Hour(), lastStartTime.Minute(), lastStartTime.Second(), lastStartTime.Nanosecond(), lastStartTime.Location())
	}

	return lastStartTime // event should have a frequency
}

type EventFrequency uint8

func (f EventFrequency) String() string {
	switch f {
	case FrequencyNone:
		return ""
	case FrequencyDaily:
		return "Daily"
	case FrequencyWeekly:
		return "Weekly"
	case FrequencyMonthly:
		return "Monthly"
	case FrequencyYearly:
		return "Yearly"
	}

	return ""
}

func ParseFrequency(input string) EventFrequency {
	switch strings.ToLower(input) {
	case "daily":
		return FrequencyDaily
	case "weekly":
		return FrequencyWeekly
	case "monthly":
		return FrequencyMonthly
	case "yearly":
		return FrequencyYearly
	default:
		return FrequencyNone
	}
}

const (
	FrequencyNone EventFrequency = iota
	FrequencyDaily
	FrequencyWeekly
	FrequencyMonthly
	FrequencyYearly
)

func CreateEvent(event AutoEvent) []database.TxOP {
	return []database.TxOP{
		{
			Op: database.OpTypeManip,
			//language=SQL
			Query: "INSERT INTO events (guildID, guildEventID, eventData) VALUES (?, ?, ?)",
			Args:  database.QueryArgs(event.GuildID, event.GuildEventID, &database.JsonResolveTarget{Target: event}),
		},
	}
}

func DeleteEvent(GuildID, eventID string) []database.TxOP {
	return []database.TxOP{
		{
			Op: database.OpTypeManip,
			//language=SQL
			Query: "DELETE FROM events WHERE guildID = ? AND guildEventID = ?",
			Args:  database.QueryArgs(GuildID, eventID),
		},
	}
}

func UpdateEventData(event AutoEvent) []database.TxOP {
	return []database.TxOP{
		{
			Op: database.OpTypeManip,
			//language=SQL
			Query: "UPDATE events SET eventData = ? WHERE guildEventID = ? AND guildID = ?",
			Args:  database.QueryArgs(&database.JsonResolveTarget{Target: event}, event.GuildEventID, event.GuildID),
		},
	}
}

func FindEvents(GuildID string, events *[]AutoEvent) []database.TxOP {
	var result AutoEvent

	return []database.TxOP{
		{
			Op: database.OpTypeQuery,
			//language=SQL
			Query: "SELECT eventData FROM events WHERE guildID = ?",
			Args:  database.QueryArgs(GuildID),
			Resolver: database.QueryRowsResolver(func() error {
				*events = append(*events, result)
				return nil
			}, &database.JsonResolveTarget{Target: &result}),
		},
	}
}

func FindEvent(GuildID, EventID string, event *AutoEvent) []database.TxOP {
	return []database.TxOP{
		{
			Op: database.OpTypeQueryRow,
			//language=SQL
			Query:    "SELECT eventData FROM events WHERE guildID = ? AND guildEventID = ?",
			Args:     database.QueryArgs(GuildID, EventID),
			Resolver: database.QueryRowResolver(&database.JsonResolveTarget{Target: event}),
		},
	}
}
