package events

import (
	"encoding/json"
	"fmt"
	"github.com/Riven-Spell/hydaelyn/database/queries"
)

func (a *AutoSchedulerService) ScheduleNewEvent(event queries.AutoEvent) error {
	if !event.Ready() {
		buf, _ := json.Marshal(event)
		return fmt.Errorf("cannot create an event without more details %s", string(buf))
	}

	tx, err := a.db.GetTransaction(nil)
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	if err != nil {
		return err
	}

	err = tx.Do(queries.CreateEvent(event))
	if err != nil {
		return err
	}

	return tx.Commit()
}
