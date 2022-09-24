package database

import (
	"errors"
	"github.com/Riven-Spell/hydaelyn/common"
	_ "github.com/go-sql-driver/mysql" // imported for database/sql
)

// =========== LCM ==========

var singleLCMDatabase *Database

func getDatabase(cfg *common.ConfigDB) (*Database, error) {
	if singleLCMDatabase == nil {
		db, err := getSQLDatabase(*cfg)
		if err != nil {
			return nil, err
		}

		singleLCMDatabase = &Database{sql: db, dbName: cfg.DBName}
	}

	return singleLCMDatabase, nil
}

var LCMServiceSQLDB = common.LCMService{
	Name:         common.LCMServiceNameSQL,
	Dependencies: []string{"log", "config"},
	Startup: func() error {
		cfg := common.GetLifeCycleManager().Services[common.LCMServiceNameConfig].GetSvc().(*common.Config)
		if cfg == nil {
			return errors.New("config is empty")
		}

		_, err := getDatabase(&cfg.DB)
		return err
	},
	GetSvc: func() interface{} {
		db, _ := getDatabase(nil)
		return db
	},
	Shutdown: func() error {
		return closeSQLDatabase()
	},
}
