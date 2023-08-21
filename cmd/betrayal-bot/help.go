package main

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

func (a *app) HelpCommand() SlashCommand {
	help := SlashCommand{
		Feature: discordgo.ApplicationCommand{
			Name:        "help",
			Description: "Get help with all commands available",
		},
		Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			availableCommands := a.commandHandler.GetCommands()
			embededFields := []*discordgo.MessageEmbedField{}
			for _, cmd := range availableCommands {
				commandName := cmd.Feature.Name
				commandDescription := cmd.Feature.Description
				commandOptions := cmd.Feature.Options
				commandOptionstring := ""
				for _, commandOpt := range commandOptions {
					commandOptionstring += fmt.Sprintf(
						"%s: %s\n",
						commandOpt.Name,
						commandOpt.Description,
					)
				}
				embededDescription := fmt.Sprintf(
					"%s\n%s\n%s",
					commandName,
					commandDescription,
					discordgo.ApplicationCommandOptionString,
				)
				embededFields = append(embededFields, &discordgo.MessageEmbedField{
					Name:   cmd.Feature.Name,
					Value:  embededDescription,
					Inline: false,
				})
			}
			embed := discordgo.MessageEmbed{
				Title:       "Help",
				Description: "All commands available",
				Fields:      embededFields,
			}

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{&embed},
				},
			})

		},
	}
	return help
}
