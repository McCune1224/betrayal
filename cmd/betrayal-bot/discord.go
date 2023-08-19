package main

import "github.com/bwmarrin/discordgo"

func SendChannelMessage(
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
	message string,
	eph ...bool,
) {
	if len(eph) != 0 && eph[0] {

	}
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: message,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}
