package main

import "github.com/bwmarrin/discordgo"

const mckusaID = "206268866714796032"
// Give the option to allow this command to be ephemeral (hidden to other users)
func EphermalOptional() *discordgo.ApplicationCommandOption {
	return &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionString,
		Name:        "hidden",
		Description: "The role to get",
		Required:    false,
	}

}

func SendChannelMessage(
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
	message string,
) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: message,
		},
	})
}

func SendChannelEmbededMessage(
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
	embeded *discordgo.MessageEmbed,
) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embeded},
		},
	})
}
