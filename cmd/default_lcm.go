package cmd

import (
	"github.com/Riven-Spell/hydaelyn/common"
	"github.com/Riven-Spell/hydaelyn/database"
)

var DefaultServices = []common.LCMService{
	database.LCMServiceSQLDB,
	common.LCMServiceConfig,
	common.LCMServiceLogger(),
}
