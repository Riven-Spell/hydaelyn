package database

import (
	"context"
	_ "embed"
	"fmt"
	"strings"
)

//go:embed setup.sql
var setup string

var setupQueries = func() []string {
	lines := strings.Split(setup, "\n")
	commands := make([]string, 0)

	cCommand := ""
	for _, v := range lines {
		cCommand += v
		if strings.HasSuffix(cCommand, ";") {
			commands = append(commands, cCommand)
			cCommand = ""
		}
	}

	return commands
}()

var tables = func() []string {
	out := make([]string, 0)
	for _, v := range setupQueries {
		if strings.HasPrefix(v, "CREATE TABLE ") {
			line := strings.TrimPrefix(v, "CREATE TABLE ")
			out = append(out, line[:strings.Index(line, " ")])
		}
	}

	return out
}()

//var setupQuery = `CREATE DATABASE hydaelyn;

func (db *Database) SetupDatabase() error {
	tx, err := db.sql.BeginTx(context.TODO(), nil)
	if err != nil {
		return err
	}

	_, err = tx.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS `%s`;", db.dbName))
	if err != nil {
		return err
	}

	_, err = tx.Exec(fmt.Sprintf("CREATE DATABASE `%s`;", db.dbName))
	if err != nil {
		return err
	}

	_, err = tx.Exec(fmt.Sprintf("USE `%s`;", db.dbName))
	if err != nil {
		return err
	}

	for _, v := range setupQueries {
		_, err := tx.Query(v)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (db *Database) DBSetupCorrectly() error {
	extTx, err := db.GetTransaction()
	if err != nil {
		return err
	}
	tx := extTx.tx

	for _, v := range tables {
		query := fmt.Sprintf("SELECT * FROM %s", v)
		extTx.logf("running query `%s`", query)
		row, err := tx.Query(query)
		if err != nil {
			return err
		}
		err = row.Close()
		if err != nil {
			return err
		}
	}

	extTx.logf("finished database validation, committing...")
	return tx.Commit()
}
