package events

import (
	"github.com/Riven-Spell/hydaelyn/common"
	"github.com/Riven-Spell/hydaelyn/database"
	"github.com/bwmarrin/discordgo"
	log2 "log"
)

var LCMServiceAutoScheduler = common.LCMService{
	Name:         common.LCMServiceNameAutoScheduler,
	Dependencies: []string{common.LCMServiceNameSQL, common.LCMServiceNameBot, common.LCMServiceNameLog},
	GetSvc: func() interface{} {
		return singleAutoSchedulerService
	},
	Startup: func(deps []interface{}) error {
		db := deps[0].(*database.Database)
		bot := deps[1].(*discordgo.Session)
		log := deps[2].(*log2.Logger)

		singleAutoSchedulerService = &AutoSchedulerService{db: db, bot: bot, log: log, live: true}

		bot.AddHandler(singleAutoSchedulerService.getEventUpdateHandler())
		bot.AddHandler(singleAutoSchedulerService.getEventDeleteHandler())

		return nil
	},
	Shutdown: func() error {
		singleAutoSchedulerService.Shutdown()

		return nil
	},
}

var singleAutoSchedulerService *AutoSchedulerService

type AutoSchedulerService struct {
	live bool
	db   *database.Database
	log  *log2.Logger
	bot  *discordgo.Session
}

func (a *AutoSchedulerService) IsLive() bool {
	return a.live
}

func (a *AutoSchedulerService) Shutdown() {
	a.live = false
	a.bot = nil
	a.db = nil
	a.log = nil
}
