package commands

import (
	"fmt"
	"github.com/Riven-Spell/hydaelyn/common"
	"github.com/Riven-Spell/hydaelyn/database"
	"github.com/bwmarrin/discordgo"
	"log"
	"strings"
)

type Handler func(*discordgo.Session, *discordgo.InteractionCreate, *common.LifeCycleManager, *database.Database, *log.Logger)

type Command struct {
	Handler              Handler
	AutoCompleteHandlers map[string]interface{} // either another similar map or a Handler
	DgCommand            *discordgo.ApplicationCommand
	HelpText             string
}

var GlobalCommands = map[string]Command{
	RoleReactCommand.DgCommand.Name:  RoleReactCommand,
	CommandChainRoles.DgCommand.Name: CommandChainRoles,
	CommandSource.DgCommand.Name:     CommandSource,
	CommandEvents.DgCommand.Name:     CommandEvents,
}

func RegisterGlobalCommands(s *discordgo.Session, cfg common.ConfigDiscord) error {
	cmds, err := s.ApplicationCommands(cfg.BotApplicationID, "")
	if err != nil {
		return fmt.Errorf("failed to get application commands: %w", err)
	}

	// Check for existing commands and prepare to upgrade
	existingCommandMap := func() map[string]*discordgo.ApplicationCommand {
		out := make(map[string]*discordgo.ApplicationCommand)

		for _, v := range cmds {
			if v == nil {
				continue
			}

			out[v.Name] = v
		}

		return out
	}()

	errors := make([]any, 0)

	// Overwrite/create new commands
	for k, v := range GlobalCommands {
		if v.DgCommand == nil {
			errors = append(errors, fmt.Errorf("command %s does not contain a dgCommand", k))
		}

		v.DgCommand.ApplicationID = cfg.BotApplicationID
		v.DgCommand.Name = k

		if existingCmd, ok := existingCommandMap[k]; ok {
			v.DgCommand.ID = existingCmd.ID
			_, err = s.ApplicationCommandEdit(cfg.BotApplicationID, "", v.DgCommand.ID, v.DgCommand)
			if err != nil {
				errors = append(errors, fmt.Errorf("command %s failed to update: %w", k, err))
			}

			delete(existingCommandMap, k)
		} else {
			_, err = s.ApplicationCommandCreate(cfg.BotApplicationID, "", v.DgCommand)
			if err != nil {
				errors = append(errors, fmt.Errorf("command %s failed to create: %w", k, err))
			}
		}
	}

	// Delete old nonexistent commands
	if len(existingCommandMap) > 0 {
		for k, v := range existingCommandMap {
			err = s.ApplicationCommandDelete(cfg.BotApplicationID, "", v.ID)
			if err != nil {
				errors = append(errors, fmt.Errorf("failed to remove command %s: %w", k, err))
			}
		}
	}

	// Handle errors
	if len(errors) > 0 {
		errorFormat := "failed to set commands:\n" + strings.Repeat("%w\n", len(errors))
		errorFormat = strings.TrimSuffix(errorFormat, "\n")

		return fmt.Errorf(errorFormat, errors...)
	}

	return nil
}

func GetCommandHandler(lcm *common.LifeCycleManager, db *database.Database, l *log.Logger) func(s *discordgo.Session, i *discordgo.InteractionCreate) {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		cmdName := i.ApplicationCommandData().Name
		command := GlobalCommands[cmdName]

		if i.Type == discordgo.InteractionApplicationCommand {
			l.Printf("%s: %s ran by %s\n", i.ID, cmdName, TryGetUsername(i))

			err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
			})
			if err != nil {
				l.Printf("%s: failed to respond: %s", i.ID, err.Error())
			}

			if command.Handler == nil {
				l.Printf("%s: Command %s has no handler!\n", i.ID, cmdName)
				InternalError(s, i, l)
				return
			}

			command.Handler(s, i, lcm, db, l)
		} else if i.Type == discordgo.InteractionApplicationCommandAutocomplete {
			l.Printf("%s: %s polling autocomplete for %s\n", i.ID, cmdName, TryGetUsername(i))
			AutoCompleteHandler(s, i, lcm, db, l)
		}
	}
}

func TryRespond(s *discordgo.Session, i *discordgo.InteractionCreate, response string, log *log.Logger) *discordgo.Message {
	message, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &response,
	})

	if err != nil {
		log.Println(i.ID+": failed to reply to", ""+":", err)
		message = nil
	}

	return message
}

func TryGetUsername(i *discordgo.InteractionCreate) string {
	if i.User != nil {
		return i.User.String()
	} else if i.Member != nil {
		return i.Member.User.String()
	}

	return "USER_NOT_FOUND"
}

func InternalError(s *discordgo.Session, i *discordgo.InteractionCreate, log *log.Logger) *discordgo.Message {
	log.Printf("%s: Returning internal error.", i.ID)
	return TryRespond(s, i, fmt.Sprintf("Internal error, contact bot owner with %s", i.ID), log)
}
