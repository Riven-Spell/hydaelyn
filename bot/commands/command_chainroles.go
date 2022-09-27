package commands

import (
	_ "embed"
	"fmt"
	"github.com/Riven-Spell/hydaelyn/common"
	"github.com/Riven-Spell/hydaelyn/database"
	"github.com/bwmarrin/discordgo"
	"log"
	"math/big"
)

//go:embed helptext/chainroles_long.txt
var chainRolesHelptext string

const (
	baseRoleName = "base_role"
)

var CommandChainRoles = Command{
	Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate, lcm *common.LifeCycleManager, db *database.Database, log *log.Logger) {
		data := i.ApplicationCommandData()

		baseRoleID := ""
		newRoles := make([]*discordgo.Role, 0)
		for _, v := range data.Options {
			if v == nil {
				continue
			}

			if v.Name == baseRoleName {
				baseRoleID = v.RoleValue(s, i.GuildID).ID
				continue
			}

			newRoles = append(newRoles, v.RoleValue(s, i.GuildID))
		}

		roleSetFailure := false
		var highestID big.Int
		for {
			members, err := s.GuildMembers(i.GuildID, highestID.String(), 1000)
			if err != nil {
				log.Printf("%s: failed to list guild members: %s", i.ID, err)
				InternalError(s, i, log)
				return
			}

			log.Printf("Found %d users... processing...", len(members))

			for _, v := range members {
				if v == nil {
					log.Printf("%s: nil guild member in query")
					continue
				}

				var id big.Int
				id.SetString(v.User.ID, 10)
				if id.Cmp(&highestID) == 1 {
					highestID = id
				}

				foundRole := false
				for _, v := range v.Roles {
					if v == baseRoleID {
						foundRole = true
						break
					}
				}

				if foundRole {
					log.Printf("%s: Assigning new roles to user %s", i.ID, v.User.String())

					for _, newRole := range newRoles {
						err = s.GuildMemberRoleAdd(i.GuildID, v.User.ID, newRole.ID)
						if err != nil {
							roleSetFailure = true
							log.Printf("%s: role addition failed (user: %s role: %s): %err", i.ID, v.User.String(), newRole.Name, err)
						}
					}
				}
			}

			if len(members) < 1000 {
				break
			}
		}

		if !roleSetFailure {
			log.Printf("%s: Assigned all new roles to users.", i.ID)
			TryRespond(s, i, "All new roles were assigned.", log)
		} else {
			log.Printf("%s: Some roles failed to assign.", i.ID)
			TryRespond(s, i, fmt.Sprintf("Some roles failed to assign. Contact bot owner with %s.", i.ID), log)
		}
	},
	DgCommand: &discordgo.ApplicationCommand{
		Name:                     "chainroles",
		Description:              "Assign multiple roles to members with existing roles",
		DefaultMemberPermissions: common.Pointer(int64(discordgo.PermissionManageRoles)),
		DMPermission:             common.Pointer(false),
		Options: func() []*discordgo.ApplicationCommandOption {
			out := make([]*discordgo.ApplicationCommandOption, 24)

			for k := range out {
				out[k] = &discordgo.ApplicationCommandOption{
					Type:        discordgo.ApplicationCommandOptionRole,
					Name:        fmt.Sprintf("new_role%d", k),
					Description: "The initial required role to receive additional roles.",
					Required:    false,
				}
			}

			required := []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionRole,
					Name:        baseRoleName,
					Description: "The initial required role to receive additional roles.",
					Required:    true,
				},
			}

			return append(required, out...)
		}(),
	},
	HelpText: chainRolesHelptext,
}
