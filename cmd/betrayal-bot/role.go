package main

import (
	"fmt"

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
					Name:        "name",
					Description: "Name of the role",
					Required:    true,
				},
			},
		},
		Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			name := i.ApplicationCommandData().Options[0].StringValue()
			role, err := a.models.Roles.GetByName(name)
			if err != nil {
				a.logger.Println(err)
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: fmt.Sprintf("Error getting role: %s", name),
					},
				})
				return
			}

			color := 0x00ff00
			switch role.Alignment {
			case "GOOD":
				color = 0x00ff00
			case "EVIL":
				color = 0xff3300
			case "NEUTRAL":
				color = 0xffee00
			}

			embed := &discordgo.MessageEmbed{
				Title:       role.Name,
				Description: role.Description,
				Color:       color,
				Footer: &discordgo.MessageEmbedFooter{
					Text: "Alignment: " + role.Alignment,
				},
			}

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{embed},
				},
			})
		},
	}

	return role
}
