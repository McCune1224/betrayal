package commands

import "github.com/bwmarrin/discordgo"

var Whoami = SlashCommand{
	Feature: discordgo.ApplicationCommand{
		Name:        "whoami",
		Description: "Responds with your user ID and your username",
	},
	Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Your user ID is " + i.Member.User.ID + " and your username is " + i.Member.User.Username,
			},
		})
	},
}
