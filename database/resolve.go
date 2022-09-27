package database

import (
	"database/sql"
)

type SpecialResolveTarget interface {
	Substitute() any // create a false Target that row.Scan can use
	Resolve() error  // truly resolve it
}

type scannable interface {
	Scan(...any) error
	Err() error
}

func QueryRowResolver(results ...any) func(any) error {
	return func(i any) error {
		row := i.(scannable)

		innerResults := make([]any, len(results))

		for k, v := range results {
			if resolver, ok := v.(SpecialResolveTarget); ok {
				innerResults[k] = resolver.Substitute()
			} else {
				innerResults[k] = v
			}
		}

		if err := row.Err(); err != nil {
			return err
		}

		err := row.Scan(innerResults...)
		if err != nil {
			return err
		}

		for _, v := range results {
			if resolver, ok := v.(SpecialResolveTarget); ok {
				if err = resolver.Resolve(); err != nil {
					return err
				}
			}
		}

		return nil
	}
}

// QueryRowsResolver takes in a set of column targets
func QueryRowsResolver(processRow func() error, targets ...any) func(any) error {
	return func(i any) error {
		rows := i.(*sql.Rows)

		for rows.Next() {
			err := QueryRowResolver(targets)(rows)
			if err != nil {
				return err
			}

			if err = processRow(); err != nil {
				return err
			}
		}

		return rows.Err()
	}
}
