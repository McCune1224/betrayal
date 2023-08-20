package main

import "github.com/bwmarrin/discordgo"

// Slash Command to return the slash commnad option back to the user
func (a *app) EchoCommand() SlashCommand {
	return SlashCommand{
		Feature: discordgo.ApplicationCommand{
			Name:        "echo",
			Description: "Echoes back the message sent to it",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "echo",
					Description: "message to repeat back",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    true,
				},
			},
		},

		Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			message := ""
			if len(i.ApplicationCommandData().Options) != 0 {
				message = i.ApplicationCommandData().Options[0].StringValue()
			} else {
				message = "No message provided"
			}
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: message,
				},
			})
		},
	}
}
