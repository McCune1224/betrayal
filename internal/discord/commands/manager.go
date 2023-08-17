package commands

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

//Centeralized location for managing commands

type SlashCommand struct {
	Feature discordgo.ApplicationCommand
	Handler func(s *discordgo.Session, i *discordgo.InteractionCreate)
}

type SlashCommandManager struct {
	MappedCommands map[string]SlashCommand
}

func NewSlashCommandManager() *SlashCommandManager {
	return &SlashCommandManager{
		make(map[string]SlashCommand),
	}
}

func (scm *SlashCommandManager) AddCommand(sc SlashCommand) {
	scm.MappedCommands[sc.Feature.Name] = sc
}

// From the map of SlashCommands stored in the SCM, make them available for the current session.
// Requires Discord Websocket Connection to be open first before calling this function

// Returns a tally of successful applications created for the session
func (scm *SlashCommandManager) RegisterCommands(session *discordgo.Session) int {

	session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		// Likely will want to add a check for the command's guild ID here at some point...
		log.Printf("%s invoked by %s", i.ApplicationCommandData().Name, i.Member.User.Username)

		if cmd, ok := scm.MappedCommands[i.ApplicationCommandData().Name]; ok {
			cmd.Handler(s, i)
		} else {
			log.Printf("Failed to invoke command %s", i.ApplicationCommandData().Name)
		}
	})

	tally := 0
	for _, slashcmd := range scm.MappedCommands {
		_, err := session.ApplicationCommandCreate(session.State.User.ID, "", &slashcmd.Feature)
		if err != nil {
			log.Printf("Failed to add command %s\nError: %s", slashcmd.Feature.Name, err.Error())
			continue
		}
		tally++
	}

	return tally
}
