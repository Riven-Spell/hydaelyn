package database

import "database/sql"

type Transaction struct {
	owner *Database
	tx    *sql.Tx
}

func (tx *Transaction) Commit() error {
	return tx.tx.Commit()
}

func (tx *Transaction) Rollback() error {
	return tx.tx.Rollback()
}

func (tx *Transaction) Do(ops []TxOP) error {
	return transact(tx.tx, ops)
}
