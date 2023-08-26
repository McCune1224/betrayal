package main

import (
	"github.com/bwmarrin/discordgo"
)

func (a *app) GetRoleCommand() SlashCommand {

	role := SlashCommand{
		Feature: discordgo.ApplicationCommand{
			Name:        "role",
			Description: "Get a role",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "role2",
					Description: "The role to get new and improved",
					Required:    true,
				},
			},
		},
		Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			SendChannelMessage(s, i, "wip")
		},
	}

	return role
}
