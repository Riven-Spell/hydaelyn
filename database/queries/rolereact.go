package queries

import (
	"encoding/json"
	"github.com/Riven-Spell/hydaelyn/database"
	"github.com/bwmarrin/discordgo"
)

type RoleReacts []RoleReact

type RoleReact struct {
	Emoji string // APIName
	Roles []string
}

func DeleteRoleReactQuery(m *discordgo.Message) []database.TxOP {
	return []database.TxOP{
		{
			Op: database.OpTypeManip,
			//language=SQL
			Query: "DELETE FROM rolereacts WHERE channelID = ? AND messageID = ?",
			Args:  database.QueryArgs(m.ChannelID, m.ID),
		},
	}
}

func CreateRoleReactQuery(m *discordgo.Message, roles RoleReacts) []database.TxOP {
	buf, _ := json.Marshal(roles)

	return []database.TxOP{
		{
			Op: database.OpTypeManip,
			//language=SQL
			Query: "INSERT INTO rolereacts (channelID, messageID, roles) VALUES (?, ?, ?)",
			Args:  database.QueryArgs(m.ChannelID, m.ID, string(buf)),
		},
	}
}

func GetRoleReactQuery(m *discordgo.Message, roles *RoleReacts) []database.TxOP {
	return []database.TxOP{
		{
			Op: database.OpTypeQueryRow,
			//language=SQL
			Query:    "SELECT roles FROM rolereacts WHERE channelID = ? AND messageID = ?",
			Args:     database.QueryArgs(m.ChannelID, m.ID),
			Resolver: database.QueryRowResolver(&database.JsonResolveTarget{Target: roles}),
		},
	}
}
