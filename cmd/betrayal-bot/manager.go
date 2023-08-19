package main

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

type SlashCommandManager struct {
	MappedCommands map[string]SlashCommand
	CommandIDs     map[string]string
}

func (a *app) NewSlashCommandManager() *SlashCommandManager {
	return &SlashCommandManager{
		// TODO: move this elsewhere so we don't have to pass it around, works for now at least ;)
		MappedCommands: make(map[string]SlashCommand),
		CommandIDs:     make(map[string]string),
	}
}

func (scm *SlashCommandManager) MapCommand(sc SlashCommand) {
	scm.MappedCommands[sc.Feature.Name] = sc
}

func (scm *SlashCommandManager) RegisterCommands(session *discordgo.Session) int {

	// Pass the function to the handler so long as the command is registered [has a key in MappedCommand]
	session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		// Likely will want to add a check for the command's guild ID here at some point...
		log.Printf("%s invoked by %s", i.ApplicationCommandData().Name, i.Member.User.Username)

		if cmd, ok := scm.MappedCommands[i.ApplicationCommandData().Name]; ok {
			cmd.Handler(s, i)
		} else {
			log.Printf("Failed to invoke command %s", i.ApplicationCommandData().Name)
		}
	})

	// Register the Slash Commands with Discord
	totalAddedCommands := 0
	for _, slashCmd := range scm.MappedCommands {
		rcmd, err := session.ApplicationCommandCreate(session.State.User.ID, "", &slashCmd.Feature)
		if err != nil {
			log.Printf("Failed to add command %s\nError: %s", slashCmd.Feature.Name, err.Error())
			continue
		}
		scm.CommandIDs[rcmd.ID] = rcmd.Name
		totalAddedCommands++
	}

	return totalAddedCommands
}
