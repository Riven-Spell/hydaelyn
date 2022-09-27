package database

import (
	"database/sql"
)

type Transaction struct {
	owner *Database
	tx    *sql.Tx
	id    string
}

func (tx *Transaction) logf(format string, inputs ...any) {
	inputs = append([]any{tx.id}, inputs...)
	tx.owner.log.Printf("tx%s: "+format, inputs...)
}

func (tx *Transaction) Commit() error {
	tx.logf("committing transaction...")
	err := tx.tx.Commit()

	if err != nil {
		tx.logf("FAILED COMMIT: %s", err.Error())
	}

	return err
}

func (tx *Transaction) Rollback() error {
	tx.logf("rolling back...")
	err := tx.tx.Rollback()

	if err != nil {
		tx.logf("FAILED ROLLBACK: %s", err.Error())
	}

	return err
}

func (tx *Transaction) Do(ops []TxOP) error {
	return transact(tx, ops)
}
