package main

import (
	"github.com/bwmarrin/discordgo"
)

// Responde
func (a *app) PingCommand() SlashCommand {
	ping := SlashCommand{
		Feature: discordgo.ApplicationCommand{
			Name:        "ping",
			Description: "Bing Database",
			Options: []*discordgo.ApplicationCommandOption{
				EphermalOptional(),
			},
		},
		Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			SendChannelMessage(s, i, "Pong!")
		},
	}

	return ping
}
