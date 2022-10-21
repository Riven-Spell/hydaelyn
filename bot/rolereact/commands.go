package rolereact

import (
	"github.com/Riven-Spell/hydaelyn/database/queries"
	"github.com/bwmarrin/discordgo"
)

func (r *RoleReactService) RegisterRoleReactMessage(m *discordgo.Message, roles queries.RoleReacts) error {
	tx, err := r.db.GetTransaction(nil)
	if err != nil {
		return err
	}

	err = tx.Do(queries.CreateRoleReactQuery(m, roles))
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (r *RoleReactService) GetRoleReactMessage(m *discordgo.Message, roles *queries.RoleReacts) error {
	tx, err := r.db.GetTransaction(nil)
	if err != nil {
		return err
	}

	err = tx.Do(queries.GetRoleReactQuery(m, roles))
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (r *RoleReactService) DeleteRoleReactMessage(m *discordgo.Message) error {
	tx, err := r.db.GetTransaction(nil)
	if err != nil {
		return err
	}

	err = tx.Do(queries.DeleteRoleReactQuery(m))
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}
