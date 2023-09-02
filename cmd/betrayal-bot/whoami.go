package main

import "github.com/bwmarrin/discordgo"

// (#6A5ACD)

func (a *app) ChannelDetailsCommand() SlashCommand {
	return SlashCommand{
		Feature: discordgo.ApplicationCommand{
			Name:        "whoami_channel",
			Description: "Get requested channel details",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionChannel,
					Name:        "channel",
					Description: "The channel to get details from",
					Required:    true,
				},
			},
		},
		Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			channel, err := s.Channel(i.ApplicationCommandData().Options[0].ChannelValue(s).ID)
			if err != nil {
				a.logger.Println(err)
				SendEphermalChannelMessage(s, i, "Error: "+err.Error())
				return
			}
			SendChannelEmbededMessage(s, i, &discordgo.MessageEmbed{
				Title: "Channel Details",
				Color: 0x6A5ACD,
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:  "Name",
						Value: channel.Name,
					},
					{
						Name:  "ID",
						Value: channel.ID,
					},
				},
			})
		},
	}

}

func (a *app) UserDetailsCommand() SlashCommand {
	return SlashCommand{
		Feature: discordgo.ApplicationCommand{
			Name:        "whoami_user",
			Description: "Get requested details",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionUser,
					Name:        "user",
					Description: "The user to get details from",
					Required:    true,
				},
			},
		},
		Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			user, err := s.User(i.ApplicationCommandData().Options[0].UserValue(s).ID)
			if err != nil {
				a.logger.Println(err)
				SendEphermalChannelMessage(s, i, "Error: "+err.Error())
				return
			}
			SendChannelEmbededMessage(s, i, &discordgo.MessageEmbed{
				Title: "User Details",
				Color: 0x6A5ACD,
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:  "Name",
						Value: user.Username,
					},
					{
						Name:  "ID",
						Value: user.ID,
					},
				},
			})
		},
	}
}
