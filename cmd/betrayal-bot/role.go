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
				SendErrorMessage(s, i, fmt.Sprintf("Error getting role: %s", name))
				return
			}
			abilities, err := a.models.Roles.GetAbilities(role.ID)
			if err != nil {
				a.logger.Println(err)
				SendErrorMessage(s, i, fmt.Sprintf("Error getting role: %s", name))
				return
			}

			perks, err := a.models.Roles.GetPerks(role.ID)
			if err != nil {
				a.logger.Println(err)
				SendErrorMessage(s, i, fmt.Sprintf("Error getting role: %s", name))
				return
			}

			color := 0x000000
			switch role.Alignment {
			case "GOOD":
				color = 0x00ff00
			case "EVIL":
				color = 0xff3300
			case "NEUTRAL":
				color = 0xffee00
			}

			var embededAbilitiesFields []*discordgo.MessageEmbedField
			embededAbilitiesFields = append(embededAbilitiesFields, &discordgo.MessageEmbedField{
				Name:   "\n\n" + Underline("Abilities") + "\n",
				Value:  "",
				Inline: false,
			})
			for _, ability := range abilities {
				embededAbilitiesFields = append(
					embededAbilitiesFields,
					&discordgo.MessageEmbedField{
						Name:   ability.Name,
						Value:  ability.Description,
						Inline: false,
					},
				)
			}
			embededAbilitiesFields = append(embededAbilitiesFields, &discordgo.MessageEmbedField{
				Name:  "\n\n",
				Value: "\n",
			})

			var embededPerksFields []*discordgo.MessageEmbedField
			embededAbilitiesFields = append(embededAbilitiesFields, &discordgo.MessageEmbedField{
				Name:   Underline("Perks"),
				Value:  "",
				Inline: false,
			})
			for _, perk := range perks {
				embededPerksFields = append(
					embededPerksFields,
					&discordgo.MessageEmbedField{
						Name:   perk.Name,
						Value:  perk.Description + "\n",
						Inline: false,
					},
				)
			}

			embed := &discordgo.MessageEmbed{
				Title:       role.Name,
				Description: role.Description,
				Color:       color,
				Fields:      append(embededAbilitiesFields, embededPerksFields...),
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
