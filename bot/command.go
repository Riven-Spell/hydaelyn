package bot

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	
	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"
	
	"github.com/virepri/hydaelyn/common"
)

type Command struct {
	f     func(s *discordgo.Session, m *discordgo.MessageCreate, args []string, db *sql.DB, logger *log.Logger, activityID string)
	argRq []string
}

var Commands = map[string]Command{
	"setprefix": {CommandSetPrefix, []string{"prefix"}},
	"rolemap":   {CommandRoleMap, []string{"role", "emote"}},
	"rolereact": {CommandRoleMessage, []string{"`\"message\"`", "maxpicks", "emote(s)..."}},
}

func SplitArgs(input string) []string {
	out := make([]string, 0)
	
	inString := false
	strStart := ""
	currTok := ""
	
	for k, v := range input {
		if !inString {
			if v == ' ' {
				if currTok != "" {
					out = append(out, currTok)
					currTok = ""
				}
			} else if currTok == "" && (v == '`' || v == '"') {
				longStart := input[k : k+3] // capture the next two
				if longStart == "```" {
					strStart = longStart
				} else {
					strStart = string(v)
				}
				
				inString = true
			} else {
				currTok += string(v)
			}
		} else {
			currTok += string(v)
			
			if strings.HasSuffix(currTok, strStart) {
				currTok = strings.TrimSuffix(currTok, strStart)
				
				if strStart == "```" {
					currTok = currTok[2:]
				}
				
				out = append(out, currTok)
				currTok = ""
				strStart = ""
				inString = false
			}
		}
	}
	
	if currTok != "" {
		out = append(out, currTok)
	}
	
	return out
}

func HandleCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Generate an activity ID.
	activityId := uuid.New().String()
	
	// Get the basics.
	lcm := common.GetLifeCycleManager()
	db := lcm.Services["SQL"].GetSvc().(*sql.DB)
	logger := lcm.Services["log"].GetSvc().(*log.Logger)
	
	// Log the command.
	logger.Println("Received user command:", m.Content)
	
	// Get the user's prefix.
	pfx, err := common.FindPrefix(db, m.Author.ID)
	
	if err != nil {
		logger.Println(activityId, "WARNING: Failed to obtain user prefix: ", err)
	}
	
	if !strings.HasPrefix(m.Content, pfx) {
		return // Nothing to do.
	}
	
	// Parse & Execute the command.
	// todo: this is bad parsing.
	args := SplitArgs(strings.TrimPrefix(m.Content, pfx))
	
	if len(args) < 1 {
		return // Nothing to do.
	}
	
	if cmd, ok := Commands[args[0]]; ok && cmd.f != nil {
		_ = s.ChannelTyping(m.ChannelID) // start typing.
		
		if len(args) < len(cmd.argRq)+1 {
			TryRespond(s, m, activityId,
				fmt.Sprintf("`%s` requires at least `%d` parameters. (%s)", args[0], len(cmd.argRq), strings.Join(cmd.argRq, " ")),
				logger)
			
			return // Nothing to do.
		}
		
		cmd.f(s, m, args, db, logger, activityId)
	}
}

func TryRespond(s *discordgo.Session, m *discordgo.MessageCreate, activityID, response string, logger *log.Logger) *discordgo.Message {
	message, err := s.ChannelMessageSendReply(m.ChannelID, response, m.Reference())
	
	if err != nil {
		logger.Println(activityID, "Failed to send message in response to", m.ID, "from user", m.Author.ID, ":", err)
	}
	
	return message
}
