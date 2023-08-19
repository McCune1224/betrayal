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
		},
		Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Pong!",
				},
			})
		},
	}

	return ping
}
