package bot

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	
	"github.com/bwmarrin/discordgo"
	
	"github.com/virepri/hydaelyn/common"
)

func getRoleID(role string) string {
	return strings.TrimSuffix(strings.TrimPrefix(role, "<@&"), ">")
}

// syntax: rolemap <roleID> <emote>
func CommandRoleMap(s *discordgo.Session, m *discordgo.MessageCreate, args []string, db *sql.DB, logger *log.Logger, activityID string) {
	botOwner := common.GetLifeCycleManager().Services["config"].GetSvc().(*common.Config).BotOwner
	
	// First, get the role ID out of the message.
	roleId := getRoleID(args[1])
	
	// Next, grab the emote.
	emote := args[2]
	
	// Attain the user's guild info.
	member, err := s.GuildMember(m.GuildID, m.Author.ID)
	
	if err != nil {
		logger.Println(activityID, "Failed to attain member info for guild", m.GuildID, "and user", m.Author.Username, ":", err)
		
		TryRespond(s, m, activityID,
			fmt.Sprintf("Internal error. Contact %s with activity id `%s`.", botOwner, activityID),
			logger)
		return
	}
	
	nickname := member.Nick
	
	logger.Println(activityID, "User", nickname, "attempting to register role", roleId, "to emote", emote)
	
	// Attain the roles for this server.
	roles, err := s.GuildRoles(m.GuildID)
	if err != nil {
		logger.Println(activityID, "Failed to attain role info for guild", m.GuildID, ":", err)
		
		TryRespond(s, m, activityID,
			fmt.Sprintf("Internal error. Contact %s with activity id `%s`.", botOwner, activityID),
			logger)
		return
	}
	
	// todo: cache these?
	validRoles := map[string]bool{}
	
	for _, v := range roles {
		if !v.Managed && strings.EqualFold(v.Name, "Hydaelyn") {
			validRoles[v.ID] = true
		}
	}
	
	// check if the user has the Hydaelyn role.
	hasAdmin := false
	for _, v := range member.Roles {
		if _, ok := validRoles[v]; ok {
			logger.Println(activityID, "User", nickname, "has the hydaelyn role, mapping approved.")
			hasAdmin = true
		}
	}
	
	// Deny unprivileged users.
	if !hasAdmin {
		logger.Println(activityID, "Denied user", nickname, "from attempting to add a role mapping.")
		TryRespond(s, m, activityID,
			fmt.Sprintf("I'm sorry, %s, but you do not have sufficient permissions to do this. If you are the owner or an admin of this server, add the `Hydaelyn` role to your guild and apply it to yourself.", nickname),
			logger)
		return
	}
	
	// Get a mapping of all the guild's emoji
	guildEmojis, err := s.GuildEmojis(m.GuildID)
	if err != nil {
		logger.Println(activityID, "Failed to attain emoji info for guild", m.GuildID, ":", err)
		
		TryRespond(s, m, activityID,
			fmt.Sprintf("Internal error. Contact %s with activity id `%s`.", botOwner, activityID),
			logger)
		return
	}
	
	validCustomEmoji := map[string]bool{}
	for _, v := range guildEmojis {
		validCustomEmoji[v.ID] = true
	}
	
	// Ensure the emoji is valid.
	emote = GetEmojiID(emote)
	if len(strings.Split(emote, "")) == 1 {
		// Must be a unicode emoji, which _is_ valid.
	} else {
		if _, ok := validCustomEmoji[emote]; !ok {
			logger.Println(activityID, "Cannot use emoji", emote, "because it is not from this guild. Aborting!")
			TryRespond(s, m, activityID,
				fmt.Sprintf("The emote %s is not native to this server and cannot be used.", emote),
				logger)
		}
	}
	
	// Add the mapping.
	err = common.SetRole(db, m.GuildID, emote, roleId)
	if err != nil {
		logger.Println(activityID, "Failed to set role mapping per the request of ", nickname, ":", err)
		
		owner := common.GetLifeCycleManager().Services["config"].GetSvc().(*common.Config).BotOwner
		
		TryRespond(s, m, activityID,
			fmt.Sprintf("Internal error. Contact %s with activity id `%s` and details.", owner, activityID),
			logger)
		return
	}
	
	TryRespond(s, m, activityID,
		fmt.Sprintf("Successfully mapped role %s to emote %s. You may now use this emote in role-react messages.", args[1], args[2]),
		logger)
}
