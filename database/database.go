package database

import (
	"context"
	"database/sql"
	"fmt"
)

type Database struct {
	sql    *sql.DB
	dbName string
}

func (db *Database) GetTransaction() (*Transaction, error) {
	tx, err := db.sql.BeginTx(context.TODO(), nil)
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(fmt.Sprintf("USE `%s`;", db.dbName))
	if err != nil {
		return nil, err
	}

	return &Transaction{tx: tx, owner: db}, nil
}

func (db *Database) Tx(ops []TxOP) error {
	tx, err := db.sql.BeginTx(context.TODO(), nil)
	if err != nil {
		return err
	}

	err = transact(tx, ops)
	if err != nil {
		rbErr := tx.Rollback()

		if rbErr != nil {
			return fmt.Errorf("failed to transact: %w\nfailed to rollback: %w", err, rbErr)
		}
		return fmt.Errorf("failed to transact: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
