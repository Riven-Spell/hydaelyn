package database

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	log2 "log"
)

type Database struct {
	sql    *sql.DB
	log    *log2.Logger
	dbName string
}

func (db *Database) Name() string {
	return db.dbName
}

func (db *Database) GetTransaction(opt *GetTransactionOptions) (*Transaction, error) {
	txID := uuid.New().String()

	setDB := opt.GetValues()

	db.log.Printf("tx%s: Opening transaction", txID)
	tx, err := db.sql.BeginTx(context.TODO(), nil)
	if err != nil {
		return nil, err
	}

	if setDB {
		db.log.Printf("tx%s: Setting database", txID)
		_, err = tx.Exec(fmt.Sprintf("USE `%s`;", db.dbName))
		if err != nil {
			return nil, err
		}
	}

	return &Transaction{tx: tx, owner: db, id: txID}, nil
}

func (db *Database) Tx(ops []TxOP) error {
	tx, err := db.GetTransaction(nil)

	err = tx.Do(ops)
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
