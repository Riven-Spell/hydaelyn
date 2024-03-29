package rolereact

import (
	"github.com/Riven-Spell/hydaelyn/common"
	"github.com/Riven-Spell/hydaelyn/database"
	"github.com/bwmarrin/discordgo"
	log2 "log"
)

var singleRoleReactService *RoleReactService

var LCMServiceRoleReact = common.LCMService{
	Name:         common.LCMServiceNameRoleReact,
	Dependencies: []string{"bot", "log", "SQL"},
	Startup: func(deps []interface{}) error {
		dgSess := deps[0].(*discordgo.Session)
		log := deps[1].(*log2.Logger)
		db := deps[2].(*database.Database)

		singleRoleReactService = &RoleReactService{live: true, dgSess: dgSess, log: log, db: db}

		singleRoleReactService.SetupHandlers()

		return nil
	},
	GetSvc: func() interface{} {
		return singleRoleReactService
	},
	Shutdown: func() error {
		return nil
	},
}

type RoleReactService struct {
	live   bool
	dgSess *discordgo.Session
	db     *database.Database
	log    *log2.Logger
}

func (r *RoleReactService) IsLive() bool {
	return r.live
}

func (r *RoleReactService) Shutdown() {
	r.live = false
	r.dgSess = nil
	r.db = nil
	r.log = nil
}
