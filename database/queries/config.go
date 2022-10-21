package queries

import "github.com/Riven-Spell/hydaelyn/database"

const (
	dbConfigKeyVersion = "version"
)

func GetVersion(version *uint64) []database.TxOP {
	return []database.TxOP{
		{
			Op: database.OpTypeQueryRow,
			//language=SQL
			Query:    "SELECT cfgVal FROM config WHERE cfgKey = ?",
			Args:     database.QueryArgs(dbConfigKeyVersion),
			Resolver: database.QueryRowResolver(database.JsonResolveTarget{Target: version}),
		},
	}
}

func SetVersion(version uint64) []database.TxOP {
	return []database.TxOP{
		{
			Op: database.OpTypeManip,
			//language=SQL
			Query: "INSERT INTO config (cfgKey, cfgVal) VALUES (?, ?) ON DUPLICATE KEY UPDATE cfgVal=?",
			Args:  database.QueryArgs(dbConfigKeyVersion, version, version),
		},
	}
}
