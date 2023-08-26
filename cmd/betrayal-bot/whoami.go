package main

import "github.com/bwmarrin/discordgo"

// (#6A5ACD)

func (a *app) ChannelDetailsCommand() SlashCommand {
	return SlashCommand{
		Feature: discordgo.ApplicationCommand{
			Name:        "channel",
			Description: "Get Channel Details",
			Options:     []*discordgo.ApplicationCommandOption{},
		},
		Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			embeded := &discordgo.MessageEmbed{
				Title: "Channel Details",
				Color: 0x6A5ACD,
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:   "Channel ID",
						Value:  i.ChannelID,
						Inline: true,
					}, {
						Name:   "Guild ID",
						Value:  i.GuildID,
						Inline: true,
					},
				},
			}
			SendChannelEmbededMessage(s, i, embeded)
		},
	}
}

func (a *app) UserDetailsCommand() SlashCommand {
	return SlashCommand{
		Feature: discordgo.ApplicationCommand{
			Name:        "user",
			Description: "Get User Details",
			Options:     []*discordgo.ApplicationCommandOption{},
		},
		Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			embeded := &discordgo.MessageEmbed{
				Title: "User Details",
				Color: 0x6A5ACD,
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:   "User ID",
						Value:  i.Member.User.ID,
						Inline: true,
					}, {
						Name:   "Guild ID",
						Value:  i.GuildID,
						Inline: true,
					},
				},
			}
			SendChannelEmbededMessage(s, i, embeded)
		},
	}
}
