package rolereact

import (
	"github.com/Riven-Spell/hydaelyn/database/queries"
	"github.com/bwmarrin/discordgo"
)

func (r *RoleReactService) SetupHandlers() {
	r.dgSess.AddHandler(r.GetReactAddHandler())
	r.dgSess.AddHandler(r.GetReactRemoveHandler())
	r.dgSess.AddHandler(r.GetMessageDeleteHandler())
}

func (r *RoleReactService) GetReactAddHandler() func(s *discordgo.Session, r *discordgo.MessageReactionAdd) {
	return func(s *discordgo.Session, react *discordgo.MessageReactionAdd) {
		if !r.live {
			return
		}

		m, err := s.ChannelMessage(react.ChannelID, react.MessageID)
		if err != nil {
			r.log.Printf("Failed to get message for reaction (channelID: %s messageID: %s): %s", react.ChannelID, react.MessageID, err)
			return
		}

		if react.UserID == m.Author.ID {
			return // Don't give ourselves roles for the initial reacts
		}

		var roleReacts = make(queries.RoleReacts, 0)

		err = r.GetRoleReactMessage(m, &roleReacts)
		if err != nil {
			r.log.Printf("Failed to get role reacts for message (channelID: %s messageID: %s): %s", react.ChannelID, react.MessageID, err)
			return
		}

		var roleReact queries.RoleReact
		for _, v := range roleReacts {
			if v.Emoji == react.Emoji.MessageFormat() {
				roleReact = v
				break
			}
		}

		if len(roleReact.Roles) == 0 {
			r.log.Printf("No roles found for emoji %s on message (channelID: %s messageID: %s)", react.Emoji.MessageFormat(), react.ChannelID, react.MessageID)
		}

		for _, v := range roleReact.Roles {
			err = s.GuildMemberRoleAdd(react.GuildID, react.UserID, v)
			if err != nil {
				r.log.Printf("Failed to add role to user %s when reacting to message (channelID: %s messageID: %s): %s", react.Member.User.String(), react.ChannelID, react.MessageID, err)
			}

			r.log.Printf("Added role %s to user %s", v, react.Member.User.String())
		}
	}
}

func (r *RoleReactService) GetReactRemoveHandler() func(s *discordgo.Session, r *discordgo.MessageReactionRemove) {
	return func(s *discordgo.Session, react *discordgo.MessageReactionRemove) {
		if !r.live {
			return
		}

		m, err := s.ChannelMessage(react.ChannelID, react.MessageID)
		if err != nil {
			r.log.Printf("Failed to get message for reaction (channelID: %s messageID: %s): %s", react.ChannelID, react.MessageID, err)
			return
		}

		var roleReacts = make(queries.RoleReacts, 0)

		err = r.GetRoleReactMessage(m, &roleReacts)
		if err != nil {
			r.log.Printf("Failed to get role reacts for message (channelID: %s messageID: %s): %s", react.ChannelID, react.MessageID, err)
			return
		}

		var roleReact queries.RoleReact
		for _, v := range roleReacts {
			if v.Emoji == react.Emoji.MessageFormat() {
				roleReact = v
				break
			}
		}

		if len(roleReact.Roles) == 0 {
			r.log.Printf("No roles found for emoji %s on message (channelID: %s messageID: %s)", react.Emoji.MessageFormat(), react.ChannelID, react.MessageID)
		}

		for _, v := range roleReact.Roles {
			err = s.GuildMemberRoleRemove(react.GuildID, react.UserID, v)
			if err != nil {
				r.log.Printf("Failed to remove role from user %s when reacting to message (channelID: %s messageID: %s): %s", react.UserID, react.ChannelID, react.MessageID, err)
			}

			r.log.Printf("Removed role %s from user %s", v, react.UserID)
		}
	}
}

func (r *RoleReactService) GetMessageDeleteHandler() func(s *discordgo.Session, messageDelete *discordgo.MessageDelete) {
	return func(s *discordgo.Session, d *discordgo.MessageDelete) {
		if !r.live {
			return
		}

		m := d.Message

		r.log.Printf("Attempting to remove role react message (channelID: %s, messageID: %s)...", m.ChannelID, m.ID)

		err := r.DeleteRoleReactMessage(m)
		if err != nil {
			r.log.Printf("Failed to delete role react message (channelID: %s, messageID: %s): %s", m.ChannelID, m.ID, err)
		}

		r.log.Printf("Removed message successfully (channelID: %s, messageID: %s)", m.ChannelID, m.ID)
	}
}
