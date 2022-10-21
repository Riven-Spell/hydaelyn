package setup

import (
	_ "embed"
	"fmt"
	"github.com/Riven-Spell/hydaelyn/common"
	"github.com/Riven-Spell/hydaelyn/database"
	"github.com/Riven-Spell/hydaelyn/database/queries"
	"github.com/Riven-Spell/hydaelyn/database/updates"
	"strings"
)

type SetupMode uint8

var setupModes = map[string]SetupMode{
	"initial":     SetupModeInitial,
	"upgrade":     SetupModeUpgrade,
	"destructive": SetupModeDestructive,
}

const (
	SetupModeInitial     SetupMode = iota // do nothing if the database already exists
	SetupModeUpgrade                      // add new tables only
	SetupModeDestructive                  // drop entire database; re-create tables
)

func tryDropDatabase(dbName string) []database.TxOP {
	return []database.TxOP{
		{
			Op:    database.OpTypeManip,
			Query: fmt.Sprintf("DROP DATABASE IF EXISTS `%s`;", dbName),
		},
	}
}

func tryCreateDatabase(dbName string) []database.TxOP {
	return []database.TxOP{
		{
			Op:    database.OpTypeManip,
			Query: fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s`;", dbName),
		},
	}
}

func useDatabase(dbName string) []database.TxOP {
	return []database.TxOP{
		{
			Op:    database.OpTypeManip,
			Query: fmt.Sprintf("USE `%s`;", dbName),
		},
	}
}

func SetupDatabase(db *database.Database, mode SetupMode) error {
	wTX, err := db.GetTransaction(&database.GetTransactionOptions{SetDatabase: common.Pointer(false)})
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			_ = wTX.Rollback()
		}
	}()

	if mode == SetupModeDestructive {
		err = wTX.Do(tryDropDatabase(db.Name()))
		if err != nil {
			return err
		}
	}

	err = wTX.Do(tryCreateDatabase(db.Name()))
	if err != nil {
		return err
	}

	err = wTX.Do(useDatabase(db.Name()))
	if err != nil {
		return err
	}

	if mode != SetupModeUpgrade {
		err = wTX.Do(updates.Init.Queries)
		if err != nil {
			return err
		}

		err = wTX.Do(queries.SetVersion(updates.Init.Version))
		if err != nil {
			return err
		}
	} else {
		// find out what update we're on
		var version uint64
		err = wTX.Do(queries.GetVersion(&version))
		if err != nil {
			if strings.Contains(err.Error(), "doesn't exist") {
				// assume we're on version 0.
				err = nil
				version = 0
			} else {
				return err
			}
		}

		wTX.Logf("Current version is %d, Latest is %d... Updating.", version, updates.Init.Version)

		toApply := updates.GetNeededUpdates(version)
		wTX.Logf("Applying %d updates...", len(toApply))

		for k, v := range toApply {
			wTX.Logf("Applying update %d of %d (version: %d)", k+1, len(toApply), v.Version)

			err = wTX.Do(v.Queries)
			if err != nil {
				return err
			}
		}
	}

	return wTX.Commit()
}

func DBSetupCorrectly(db *database.Database) error {
	extTx, err := db.GetTransaction(nil)
	if err != nil {
		return err
	}

	extTx.Logf("validating database...")

	defer extTx.Commit()

	err = extTx.Do(updates.Validate)
	if err != nil {
		extTx.Logf("failed database validation...")
		return err
	}

	extTx.Logf("finished database validation, committing...")
	return nil
}
