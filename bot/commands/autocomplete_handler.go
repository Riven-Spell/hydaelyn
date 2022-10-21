package commands

import (
	"github.com/Riven-Spell/hydaelyn/common"
	"github.com/Riven-Spell/hydaelyn/database"
	"github.com/bwmarrin/discordgo"
	"log"
)

func AutoCompleteHandler(s *discordgo.Session, i *discordgo.InteractionCreate, lcm *common.LifeCycleManager, db *database.Database, l *log.Logger) {
	options := i.ApplicationCommandData().Options
	cmdName := i.ApplicationCommandData().Name
	command := GlobalCommands[cmdName]

	type queueItem struct {
		target  *discordgo.ApplicationCommandInteractionDataOption
		index   uint8
		history []uint8
	}

	queue := make([]queueItem, 0)

	for k, v := range options {
		if v == nil {
			continue
		}

		queue = append(queue, queueItem{target: v, index: uint8(k)})
	}

	found := false
	for len(queue) > 0 {
		work := queue[0]

		if work.target.Focused {
			queue = queue[:1]
			found = true
			break
		}

		newHistory := append(work.history, work.index)

		for k, v := range work.target.Options {
			queue = append(queue, queueItem{target: v, index: uint8(k), history: newHistory})
		}

		queue = queue[1:]
	}

	if !found {
		return
	}

	handlers := command.AutoCompleteHandlers
	currentLevel := &discordgo.ApplicationCommandInteractionDataOption{Options: options}
	queue[0].history = append(queue[0].history, queue[0].index)
	for k, v := range queue[0].history {
		currentLevel = currentLevel.Options[v]
		result, ok := handlers[currentLevel.Name]

		if !ok {
			l.Printf("%s: found nothing for autocorrect.")
			return
		}

		if rmap, ok := result.(map[string]interface{}); ok {
			if k == len(queue[0].history)-1 {
				l.Printf("%s: Expected to find handler, found map on top level", i.ID)
				return
			}

			handlers = rmap
		} else if h, ok := result.(func(*discordgo.Session, *discordgo.InteractionCreate, *common.LifeCycleManager, *database.Database, *log.Logger)); ok {
			if k != len(queue[0].history)-1 {
				log.Printf("%s: Expected to find map, found handler on level %d (of %d)", i.ID, k+1, len(queue[0].history))
				return
			}

			h(s, i, lcm, db, l)
		} else {
			l.Printf("%s: found unexpected type on level %d (of %d)", i.ID, k+1, len(queue[0].history))
			return
		}
	}
}
