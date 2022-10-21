package database

type GetTransactionOptions struct {
	SetDatabase *bool // nil = true
}

func (opt *GetTransactionOptions) GetValues() (setDatabase bool) {
	// defaults
	setDatabase = true

	// nil check
	if opt == nil {
		return
	}

	// get real values
	if opt.SetDatabase != nil {
		setDatabase = *opt.SetDatabase
	}

	return
}
