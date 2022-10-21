package database

import (
	"errors"
	"github.com/Riven-Spell/hydaelyn/common"
	_ "github.com/go-sql-driver/mysql" // imported for database/sql
	"io"
	log2 "log"
)

// =========== LCM ==========

var singleLCMDatabase *Database

func getDatabase(cfg *common.ConfigDB, log *log2.Logger) (*Database, error) {
	if log == nil {
		log = log2.New(io.Discard, "", 0)
	}

	if singleLCMDatabase == nil {
		log.Print("Setting up SQL database connection...")

		db, err := getSQLDatabase(*cfg)
		if err != nil {
			return nil, err
		}

		log.Print("Successful connection.")
		singleLCMDatabase = &Database{sql: db, dbName: cfg.DBName, log: log}
	}

	return singleLCMDatabase, nil
}

var LCMServiceSQLDB = common.LCMService{
	Name:         common.LCMServiceNameSQL,
	Dependencies: []string{"log", "config"},
	Startup: func(deps []interface{}) error {
		log := deps[0].(*log2.Logger)
		cfg := deps[1].(*common.Config)
		if cfg == nil {
			return errors.New("config is empty")
		}

		_, err := getDatabase(&cfg.DB, log)
		return err
	},
	GetSvc: func() interface{} {
		db, _ := getDatabase(nil, nil)
		return db
	},
	Shutdown: func() error {
		return closeSQLDatabase()
	},
}
