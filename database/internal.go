package database

import (
	"database/sql"
	"fmt"
	"github.com/Riven-Spell/hydaelyn/common"
	"net/url"
	"strings"
	"time"
)

var singleDB *sql.DB

func getSQLDatabase(cfg common.ConfigDB) (*sql.DB, error) {
	if singleDB != nil {
		return singleDB, nil
	}

	uri, err := url.Parse(cfg.DBConnectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse SQL connection string: %w", err)
	}

	driver := uri.Scheme
	dataSource := strings.TrimPrefix(cfg.DBConnectionString, driver+"://")

	db, err := sql.Open(driver, dataSource)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetConnMaxLifetime(time.Minute * 3) // very default ass settings
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	return db, nil
}

func closeSQLDatabase() error {
	if singleDB == nil {
		return nil
	}

	out := singleDB.Close()
	singleDB = nil

	return out
}
