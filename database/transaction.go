package database

import (
	"database/sql"
	"fmt"
)

type Transaction struct {
	owner *Database
	tx    *sql.Tx
	id    string
}

func (tx *Transaction) Logf(format string, inputs ...any) {
	tx.LogfCalldepth(3, format, inputs...)
}

func (tx *Transaction) LogfCalldepth(calldepth int, format string, inputs ...any) {
	inputs = append([]any{tx.id}, inputs...)
	_ = tx.owner.log.Output(calldepth, fmt.Sprintf("tx%s: "+format, inputs...))
}

func (tx *Transaction) Commit() error {
	tx.Logf("committing transaction...")
	err := tx.tx.Commit()

	if err != nil {
		tx.Logf("FAILED COMMIT: %s", err.Error())
	}

	return err
}

func (tx *Transaction) Rollback() error {
	tx.Logf("rolling back...")
	err := tx.tx.Rollback()

	if err != nil {
		tx.Logf("FAILED ROLLBACK: %s", err.Error())
	}

	return err
}

func (tx *Transaction) Do(ops []TxOP) error {
	return transact(tx, ops)
}
