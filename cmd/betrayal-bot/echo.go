package main

import "github.com/bwmarrin/discordgo"

// Slash Command to return the slash commnad option back to the user
func (a *app) EchoCommand() SlashCommand {
	return SlashCommand{
		Feature: discordgo.ApplicationCommand{
			Name:        "echo",
			Description: "Echoes back the message sent to it",
		},
		Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: i.ApplicationCommandData().Options[0].StringValue(),
				},
			})
		},
	}
}
