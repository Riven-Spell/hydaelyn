package bot

import (
	"database/sql"
	"log"
	
	"github.com/bwmarrin/discordgo"
	
	"github.com/virepri/hydaelyn/common"
)

func HandleRoleReactAdd(s *discordgo.Session, r *discordgo.MessageReactionAdd) {
	lcm := common.GetLifeCycleManager()
	db := lcm.Services["SQL"].GetSvc().(*sql.DB)
	logger := lcm.Services["log"].GetSvc().(*log.Logger)
	
	m, err := s.ChannelMessage(r.ChannelID, r.MessageID)
	
	if err != nil {
		logger.Println("Failed to retrieve message.")
		return
	}
	if r.UserID == m.Author.ID {
		return // no need to give ourselves roles.
	}
	
	react, err := common.FindRoleReactMessage(db, common.RoleReactMessage{
		Guild:   r.GuildID,
		Channel: r.ChannelID,
		Message: r.MessageID,
	})
	
	if err != nil {
		logger.Println("Failed to find role react message :", err)
		return
	}
	
	if react == nil {
		logger.Println("Did not find role react message.")
		return // nothing to do, it wasn't a role react.
	}
	
	role, err := common.FindRole(db, r.GuildID, common.TernaryString(r.Emoji.ID == "", r.Emoji.Name, r.Emoji.ID))
	
	if err != nil || role == "" {
		logger.Println("Failed to find role match for emoji", r.Emoji.ID, ":", err)
		return
	}
	
	err = s.GuildMemberRoleAdd(r.GuildID, r.UserID, role)
	
	if err != nil {
		logger.Println("Failed to add role", role, "to user", r.UserID, ":", err)
		return
	}
	
	logger.Println("Added role", role, "to user", r.UserID)
}

func HandleRoleReactRemove(s *discordgo.Session, r *discordgo.MessageReactionRemove) {
	lcm := common.GetLifeCycleManager()
	db := lcm.Services["SQL"].GetSvc().(*sql.DB)
	logger := lcm.Services["log"].GetSvc().(*log.Logger)
	
	m, err := s.ChannelMessage(r.ChannelID, r.MessageID)
	
	if err != nil {
		logger.Println("Failed to retrieve message.")
		return
	}
	if r.UserID == m.Author.ID {
		return // no need to give ourselves roles.
	}
	
	react, err := common.FindRoleReactMessage(db, common.RoleReactMessage{
		Guild:   r.GuildID,
		Channel: r.ChannelID,
		Message: r.MessageID,
	})
	
	if err != nil {
		logger.Println("Failed to find role react message :", err)
		return
	}
	
	if react == nil {
		logger.Println("Did not find role react message.")
		return // nothing to do, it wasn't a role react.
	}
	
	role, err := common.FindRole(db, r.GuildID, common.TernaryString(r.Emoji.ID == "", r.Emoji.Name, r.Emoji.ID))
	
	if err != nil {
		logger.Println("Failed to find role match :", err)
		return
	}
	
	err = s.GuildMemberRoleRemove(r.GuildID, r.UserID, role)
	
	if err != nil {
		logger.Println("Failed to add role", role, "to user", r.UserID)
		return
	}
	
	logger.Println("Added role", role, "to user", r.UserID)
}
