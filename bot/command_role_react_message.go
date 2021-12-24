package bot

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"
	
	"github.com/bwmarrin/discordgo"
	
	"github.com/virepri/hydaelyn/common"
)

func GetEmojiID(emoji string) string {
	base := strings.TrimSuffix(strings.TrimPrefix(emoji, "<"), ">")
	
	if base[0] == ':' {
		base = base[strings.LastIndex(base, ":")+1:]
	}
	
	return base
}

func GetEmojiName(emoji string) string {
	base := strings.TrimSuffix(strings.TrimPrefix(emoji, "<:"), ">")
	
	if len(strings.Split(base, "")) == 1 {
		return base
	}
	
	return base[:strings.LastIndex(base, ":")]
}

// syntax !rolereact `message` maxcount <emotes>
func CommandRoleMessage(s *discordgo.Session, m *discordgo.MessageCreate, args []string, db *sql.DB, logger *log.Logger, activityID string) {
	botOwner := common.GetLifeCycleManager().Services["config"].GetSvc().(*common.Config).BotOwner
	
	message := args[1]
	count := args[2]
	emotes := args[3:]
	
	num, err := strconv.ParseUint(count, 10, 64)
	if err != nil {
		logger.Println("Failed to parse maxcount", count, ":", err)
		TryRespond(s, m, activityID, fmt.Sprintf("Maximum picks (%s) must be a number.", count), logger)
	}
	
	// Attain the user's guild info.
	member, err := s.GuildMember(m.GuildID, m.Author.ID)
	
	if err != nil {
		logger.Println("Failed to attain member info for guild", m.GuildID, "and user", m.Author.Username, ":", err)
		
		TryRespond(s, m, activityID,
			fmt.Sprintf("Internal error. Contact %s with activity id `%s`.", botOwner, activityID),
			logger)
		return
	}
	
	nickname := member.Nick
	
	logger.Println("User", nickname, "attempting to create role message")
	
	// Attain the roles for this server.
	roles, err := s.GuildRoles(m.GuildID)
	if err != nil {
		logger.Println("Failed to attain role info for guild", m.GuildID, ":", err)
		
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
			logger.Println("User", nickname, "has the hydaelyn role, mapping approved.")
			hasAdmin = true
		}
	}
	
	// Deny unprivileged users.
	if !hasAdmin {
		logger.Println("Denied user", nickname, "from attempting to add a role react message.")
		TryRespond(s, m, activityID,
			fmt.Sprintf("I'm sorry, %s, but you do not have sufficient permissions to do this. If you are the owner or an admin of this server, add the `Hydaelyn` role to your guild and apply it to yourself.", nickname),
			logger)
		return
	}
	
	// Get a mapping of all the guild's emoji
	guildEmojis, err := s.GuildEmojis(m.GuildID)
	if err != nil {
		logger.Println("Failed to attain emoji info for guild", m.GuildID, ":", err)
		
		TryRespond(s, m, activityID,
			fmt.Sprintf("Internal error. Contact %s with activity id `%s`.", botOwner, activityID),
			logger)
		return
	}
	
	validCustomEmoji := map[string]bool{}
	for _, v := range guildEmojis {
		fmt.Println(v.ID, v.Name)
		validCustomEmoji[v.ID] = true
	}
	
	// Check all emojis are valid
	for _, v := range emotes {
		_, ok := validCustomEmoji[GetEmojiID(v)]
		if !ok {
			ok = len(strings.Split(GetEmojiID(v), "")) == 1
		}
		var role string
		if ok {
			role, err = common.FindRole(db, m.GuildID, GetEmojiID(v))
		}
		
		if !ok || err != nil || role == "" {
			logger.Println("Could not find emoji", v, "for guild", m.GuildID)
			TryRespond(s, m, activityID,
				fmt.Sprintf("I'm sorry, %s, but I couldn't find a role registered to %s.", nickname, v),
				logger)
		}
	}
	
	// Create the role-react message.
	sent := TryRespond(s, m, activityID, message, logger)
	
	// Add the emotes.
	for _, v := range emotes {
		err := s.MessageReactionAdd(m.ChannelID, sent.ID, GetEmojiName(v)+":"+GetEmojiID(v))
		
		if err != nil {
			fmt.Println(GetEmojiID(v))
			logger.Println("Failed to add reaction", GetEmojiName(v)+":"+GetEmojiID(v), "to message ", sent.ID, ":", err)
			
			TryRespond(s, m, activityID,
				fmt.Sprintf("Internal error. Contact %s with activity id `%s`.", botOwner, activityID),
				logger)
			return
		}
	}
	
	// Register the role-react message.
	err = common.AddRoleReactMessage(db, common.RoleReactMessage{
		Guild:    m.GuildID,
		Channel:  m.ChannelID,
		Message:  sent.ID,
		MaxPicks: uint(num),
	})
	
	if err != nil {
		logger.Println("Failed to add role reaction to database :", err)
		
		TryRespond(s, m, activityID,
			fmt.Sprintf("Internal error. Contact %s with activity id `%s`.", botOwner, activityID),
			logger)
		return
	}
	
	_ = s.ChannelMessageDelete(m.ChannelID, m.ID)
}
