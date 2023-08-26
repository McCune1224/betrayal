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
			query := i.ApplicationCommandData().Options[0].StringValue()
			role, err := a.models.Roles.GetByName(query)
			if err != nil {
				SendChannelMessage(s, i, "Role not found")
				return
			}
			SendChannelEmbededMessage(s, i,
				&discordgo.MessageEmbed{
					Title:       role.Name,
					Description: role.Description,
					Color:       0x6A5ACD,
				},
			)
		},
	}

	return role
}
