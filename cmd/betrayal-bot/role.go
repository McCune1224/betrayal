package main

import (
	"github.com/bwmarrin/discordgo"
)

func (a *app) GetRoleCommand() SlashCommand {

	roleCommand := SlashCommand{
		Feature: discordgo.ApplicationCommand{
			Name:        "getrole",
			Description: "Get a role",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "role",
					Description: "The role to get",
					Required:    true,
				},
			},
		},
		Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			response := ""
			searchTerm := i.ApplicationCommandData().Options[0].StringValue()

			dbRole := a.models.Roles

			data, err := dbRole.GetByName(searchTerm)
			if err != nil {
				response = err.Error()
			} else {
				response = "Role: " + data.Name
			}

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: response,
				},
			})

		},
	}
	return roleCommand
}
