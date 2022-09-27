package database

import (
	"database/sql"
	"github.com/google/uuid"
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
func transact(transaction *Transaction, ops []TxOP) (err error) {
	tx := transaction.tx

	queryID := ""
	logf := func(format string, args ...any) {
		args = append([]any{queryID}, args...)
		transaction.logf("query %s: "+format, args...)
	}

	// todo: execute query
	for _, v := range ops {
		var result interface{}

		queryID = uuid.New().String()
		logf("Performing query `%s`...", v.Query)

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
			logf("QUERY FAILED: %s", queryID, err.Error())
			return err
		}

		if v.Resolver != nil {
			logf("running resolver...")
			err = v.Resolver(result)
			if err != nil {
				logf("resolver failed: %s", err.Error())
				return err
			}
		}
	}

	return nil
}
