package common

import (
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"
	
	_ "github.com/go-sql-driver/mysql" // imported for database/sql
)

// =========== LCM ==========

var LCMServiceSQLDB = LCMService{
	Name:         "SQL",
	Dependencies: []string{"log", "config"},
	Startup: func() error {
		_, err := GetSQLDatabase()
		return err
	},
	GetSvc: func() interface{} {
		db, _ := GetSQLDatabase() // should be ready already.
		return db
	},
	Shutdown: func() error {
		return CloseSQLDatabase()
	},
}

// ========== CORE ==========

var LargeQueryTimeoutError = errors.New("query timed out")

var singleDB *sql.DB

func GetSQLDatabase() (*sql.DB, error) {
	if singleDB != nil {
		return singleDB, nil
	}
	
	cfg, err := ReadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
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

func CloseSQLDatabase() error {
	if singleDB == nil {
		return nil
	}
	
	out := singleDB.Close()
	singleDB = nil
	
	return out
}

// ========== PRE-CANNED QUERIES ==========

type RoleReactMessage struct {
	Guild, Channel, Message string
	MaxPicks                uint
}

func AddRoleReactMessage(db *sql.DB, message RoleReactMessage) error {
	rows, err := db.Query("INSERT INTO hydaelyn.role_react_messages(guild, channel, message, maxpicks) VALUES (?, ?, ?, ?) ON DUPLICATE KEY UPDATE maxpicks = ?", message.Guild, message.Channel, message.Message, message.MaxPicks, message.MaxPicks)
	rows.Close()
	return err
}

func DeleteRoleReactMessage(db *sql.DB, message RoleReactMessage) error {
	rows, err := db.Query("DELETE FROM hydaelyn.role_react_messages WHERE guild = ? AND channel = ? AND message = ?")
	rows.Close()
	return err
}

func FindRoleReactMessage(db *sql.DB, message RoleReactMessage) (*RoleReactMessage, error) {
	rows, err := db.Query("SELECT maxpicks FROM role_react_messages WHERE guild = ? AND channel = ? AND message = ?", message.Guild, message.Channel, message.Message)
	
	if err != nil {
		return nil, err
	}
	
	maxpicks := uint(0)
	found := false
	for rows.Next() {
		if found {
			panic("Only one prefix should ever exist per user!")
		}
		
		err = rows.Scan(&maxpicks)
		if err != nil {
			return nil, err
		}
		
		found = true
	}
	
	if !found {
		return nil, nil // no error, but nothing found.
	}
	
	out := &RoleReactMessage{Guild: message.Guild, Channel: message.Channel, Message: message.Message, MaxPicks: maxpicks}
	return out, nil
}

func ListRoleReactMessages(db *sql.DB, timeout time.Duration) (<-chan RoleReactMessage, <-chan error) {
	errors := make(chan error, 1)
	
	rows, err := db.Query("SELECT guild, channel, message, maxpicks from hydaelyn.role_react_messages")
	if err != nil {
		errors <- err
		
		return nil, errors
	}
	
	out := make(chan RoleReactMessage, 50)
	
	go func() {
		defer rows.Close()
		
		for rows.Next() {
			react := RoleReactMessage{}
			
			err := rows.Scan(&react.Guild, &react.Channel, &react.Message, &react.MaxPicks)
			if err != nil {
				errors <- err
				return
			}
			
			var timer <-chan time.Time
			if timeout > 0 {
				timer = time.After(timeout)
			}
			
			select {
			case out <- react:
			case <-timer:
				errors <- LargeQueryTimeoutError
				return // Break.
			}
		}
	}()
	
	return out, errors
}

func FindRole(db *sql.DB, guild, emote string) (string, error) {
	rows, err := db.Query("SELECT role FROM hydaelyn.emotes_to_roles WHERE emote = ? AND guild = ?", emote, guild)
	defer rows.Close()
	
	if err != nil {
		return "", err
	}
	
	role := ""
	for rows.Next() {
		if role != "" {
			panic("Only one role should exist for each guild/emote combo!")
		}
		
		err = rows.Scan(&role)
		if err != nil {
			return "", err
		}
	}
	
	return role, nil
}

func SetRole(db *sql.DB, guild, emote, role string) error {
	rows, err := db.Query("INSERT INTO hydaelyn.emotes_to_roles(guild, emote, role) VALUES(?, ?, ?) ON DUPLICATE KEY UPDATE role = ?", guild, emote, role, role)
	rows.Close()
	return err
}

func FindPrefix(db *sql.DB, user string) (string, error) {
	rows, err := db.Query("SELECT prefix FROM hydaelyn.user_prefix WHERE user = ?", user)
	defer rows.Close()
	
	if err != nil {
		return "!", err
	}
	
	pfx := ""
	for rows.Next() {
		if pfx != "" {
			panic("Only one prefix should ever exist per user!")
		}
		
		err = rows.Scan(&pfx)
		if err != nil {
			return "!", err
		}
	}
	
	if pfx == "" {
		cfg, err := ReadConfig()
		if err != nil {
			return "!", err
		}
		
		return cfg.DefaultPrefix, nil
	}
	
	return pfx, nil
}

func SetPrefix(db *sql.DB, user, prefix string) error {
	rows, err := db.Query("INSERT INTO hydaelyn.user_prefix(user, prefix) VALUES(?, ?) ON DUPLICATE KEY UPDATE prefix=?", user, prefix, prefix)
	rows.Close()
	return err
}

func SetupDatabase(db *sql.DB) error {
	for _, v := range setupQueries {
		_, err := db.Query(v)
		if err != nil {
			return err
		}
	}
	
	return nil
}

func DBSetupCorrectly(db *sql.DB) error {
	row, err := db.Query("SELECT * FROM user_prefix")
	if err != nil {
		return err
	}
	err = row.Close()
	if err != nil {
		return err
	}
	
	row, err = db.Query("SELECT * FROM emotes_to_roles")
	if err != nil {
		return err
	}
	err = row.Close()
	if err != nil {
		return err
	}
	
	row, err = db.Query("SELECT * FROM role_react_messages")
	if err != nil {
		return err
	}
	err = row.Close()
	
	return err
}
