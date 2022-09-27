package commands

import (
	_ "embed"
	"fmt"
	"github.com/Riven-Spell/hydaelyn/bot/rolereact"
	"github.com/Riven-Spell/hydaelyn/common"
	"github.com/Riven-Spell/hydaelyn/database"
	"github.com/Riven-Spell/hydaelyn/database/queries"
	"github.com/bwmarrin/discordgo"
	"github.com/forPelevin/gomoji"
	"log"
	"regexp"
	"strings"
)

//go:embed helptext/rolemessage_long.txt
var roleReactHelptext string

var emojiRegex = regexp.MustCompile("<:(.+:)?(\\d)+>")

func GetEmojiID(emoji string) string {
	base := strings.TrimSuffix(strings.TrimPrefix(emoji, "<"), ">")

	if base[0] == ':' {
		base = base[strings.LastIndex(base, ":")+1:]
	}

	return base
}

var RoleReactCommand = Command{
	Handler: RoleReactCommandHandler,
	DgCommand: &discordgo.ApplicationCommand{
		Type:                     discordgo.ChatApplicationCommand,
		Name:                     "rolereact",
		DMPermission:             common.Pointer(false),
		Description:              "Creates a role react message for unprivileged users. /setup required.",
		DefaultMemberPermissions: common.Pointer(int64(discordgo.PermissionManageRoles)),
		Options: func() []*discordgo.ApplicationCommandOption {
			// Generate a bunch of role react options
			out := make([]*discordgo.ApplicationCommandOption, 24)
			for k, _ := range out {
				out[k] = &discordgo.ApplicationCommandOption{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        fmt.Sprintf("rolereact%d", k),
					Description: "Role react format: `emoji:roleid,roleid...`",
					Required:    false,
				}
			}

			textOption := &discordgo.ApplicationCommandOption{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "description",
				Description: "Describe this role react message!",
				Required:    true,
			}

			out = append([]*discordgo.ApplicationCommandOption{textOption}, out...)

			return out
		}(),
	},
	HelpText: roleReactHelptext,
}

func RoleReactCommandHandler(s *discordgo.Session, i *discordgo.InteractionCreate, lcm *common.LifeCycleManager, db *database.Database, log *log.Logger) {
	if i.Member.Permissions&discordgo.PermissionManageRoles != discordgo.PermissionManageRoles {
		log.Printf("%s: user has insufficient privileges to execute rolereact.", i.ID)
		TryRespond(s, i, "You have insufficient privileges to execute this command.", log)
		return
	}

	data := i.ApplicationCommandData()

	reacts := make(queries.RoleReacts, 0)
	necessaryRoles := map[string]bool{}

	msgDescription := ""

	for _, v := range data.Options {
		str, ok := v.Value.(string)

		if !ok {
			log.Printf("%s: user supplied invalid format input", i.ID)
			TryRespond(s, i, "The format of inputs are emoji:roleid", log)
			return
		}

		if v.Name == "description" {
			msgDescription = str
			continue
		}

		reacts = append(reacts, rolereact.ParseRoleReact(str))
		log.Printf("%s: rolereact roles: %s", i.ID, reacts[len(reacts)-1])

		for _, v := range reacts[len(reacts)-1].Roles {
			necessaryRoles[v] = true
		}
	}

	// validate all roles first
	guildRoles, err := s.GuildRoles(i.GuildID)
	if err != nil {
		log.Printf("%s: failed to obtain guild roles: %s", i.ID, err.Error())
		InternalError(s, i, log)
		return
	}

	roleNames := map[string]string{}

	for _, v := range guildRoles {
		if v == nil {
			continue
		}

		if _, ok := necessaryRoles[v.ID]; ok {
			roleNames[v.ID] = v.Name
			delete(necessaryRoles, v.ID)
		}
	}

	if len(necessaryRoles) > 0 {
		badRoles := ""

		for k := range necessaryRoles {
			badRoles += k + ", "
		}

		badRoles = strings.TrimSuffix(badRoles, ", ")

		TryRespond(s, i, fmt.Sprintf("The role(s) %s are invalid or were not found in this guild.", badRoles), log)
		return
	}

	log.Printf("%s: All roles are valid.", i.ID)

	// validate all emoji originate from this guild
	for _, v := range reacts {
		if !emojiRegex.MatchString(v.Emoji) {
			emoji := gomoji.CollectAll(v.Emoji)
			if len(emoji) != 1 {
				TryRespond(s, i, fmt.Sprintf("Emoji input %s is invalid. Only one emoji may be used.", v.Emoji), log)
				return
			}

			continue // probably normal emoji
		}

		_, err := s.GuildEmoji(i.GuildID, GetEmojiID(v.Emoji))
		if err != nil {
			log.Printf("Cannot find emoji %s on guild %s", v.Emoji, i.GuildID)
			TryRespond(s, i, fmt.Sprintf("Emoji %s is not valid or not from this server.", v.Emoji), log)
			return
		}
	}

	log.Printf("%s: All emoji are valid.", i.ID)

	msg := msgDescription + "\n\n"

	getRoleString := func(react queries.RoleReact) string {
		out := ""

		for _, v := range react.Roles {
			out += roleNames[v] + ", "
		}

		out = strings.TrimSuffix(out, ", ")

		return out
	}

	for _, v := range reacts {
		msg += fmt.Sprintf("%s - %s\n", v.Emoji, getRoleString(v))
	}

	message := TryRespond(s, i, msg, log)
	if message == nil {
		log.Printf("%s: failed to create role react message", i.ID)
		return
	}

	for _, v := range reacts {
		err = s.MessageReactionAdd(i.ChannelID, message.ID, strings.TrimSuffix(strings.TrimPrefix(v.Emoji, "<:"), ">"))
		if err != nil {
			log.Printf("%s: failed to react with %s: %s", i.ID, v.Emoji, err)
			InternalError(s, i, log)
			return
		}
	}

	rrs := lcm.Services[common.LCMServiceNameRoleReact].GetSvc().(*rolereact.RoleReactService)

	err = rrs.RegisterRoleReactMessage(message, reacts)
	if err != nil {
		log.Printf("%s: failed to register role react message: %s", i.ID, err)
	}

	log.Printf("%s: Created role react message (channelID: %s messageID: %s)", i.ID, i.ChannelID, message.ID)
}
