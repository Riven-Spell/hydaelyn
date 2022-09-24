package database

import (
	"database/sql"
)

type OpType uint

const (
	OpTypeManip OpType = iota // Only manipulating data, not reading.
	OpTypeQuery               // Reading data
	OpTypeQueryRow
)

type TxOP struct {
	Op       OpType
	Query    string
	Args     []interface{}
	Resolver func(i interface{}) error
}

func QueryArgs(args ...interface{}) []interface{} {
	return args
}

// Tx automagically fires off requests and returns the Rows, Row, or Result in Resolvers.
// if Tx fails, Tx automatically rolls everything back.
func transact(tx *sql.Tx, ops []TxOP) (err error) {
	// todo: execute query
	for _, v := range ops {
		var result interface{}
		switch v.Op {
		case OpTypeManip:
			result, err = tx.Exec(v.Query, v.Args...)
		case OpTypeQuery:
			result, err = tx.Query(v.Query, v.Args...)
		case OpTypeQueryRow:
			result = tx.QueryRow(v.Query, v.Args...)
			err = result.(*sql.Row).Err()
		}

		if err != nil {
			return err
		}

		if v.Resolver != nil {
			err = v.Resolver(result)
			return err
		}
	}

	return nil
}

func QueryRowResolver(i ...interface{}) func(interface{}) error {
	return nil // todo
}
