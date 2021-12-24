package bot

import (
	"database/sql"
	"fmt"
	"log"
	
	"github.com/bwmarrin/discordgo"
	
	"github.com/virepri/hydaelyn/common"
)

func CommandSetPrefix(s *discordgo.Session, m *discordgo.MessageCreate, args []string, db *sql.DB, logger *log.Logger, activityID string) {
	if len(args[1]) > 12 || len(args[1]) < 1 {
		TryRespond(s, m, activityID, "Prefixes must be at most 12 characters long and at least 1 character long.", logger)
		return
	}
	
	err := common.SetPrefix(db, m.Author.ID, args[1])
	
	if err != nil {
		logger.Println(activityID, "Failed to set prefix for", m.Author.ID, ":", err)
		
		owner := common.GetLifeCycleManager().Services["config"].GetSvc().(*common.Config).BotOwner
		
		TryRespond(s, m, activityID,
			fmt.Sprintf("Internal error. Contact %s with activity id `%s` and details.", owner, activityID),
			logger)
		return
	}
	
	TryRespond(s, m, activityID,
		fmt.Sprintf("Successfully changed user prefix to `%s`. NOTE: This applies globally to all servers.", args[1]),
		logger)
}
