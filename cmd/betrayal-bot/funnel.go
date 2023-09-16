package main

// Alex-action-funnel (TODO: make this reassignable/modular)
const funnelChannelID = "1144040897617612920"

//
// func (a *app) FunnelCommand() SlashCommand {
// 	options := []*discordgo.ApplicationCommandOption{
// 		{
// 			Type:        discordgo.ApplicationCommandOptionString,
// 			Name:        "action",
// 			Description: "what it be what it do",
// 			Required:    true,
// 		},
// 		{
// 			Type:        discordgo.ApplicationCommandOptionString,
// 			Name:        "who",
// 			Description: "Who it be what they get",
// 			Required:    true,
// 		},
// 	}
//
// 	return SlashCommand{
//
// 		Feature: discordgo.ApplicationCommand{
// 			Name:        "dothis",
// 			Description: "Do this to that",
// 			Options:     options,
// 		},
// 		Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
// 			commandInvoke := time.Now().Format(time.UnixDate)
// 			//Human-readable timestamp
// 			action := i.ApplicationCommandData().Options[0].StringValue()
// 			who := i.ApplicationCommandData().Options[1].StringValue()
//
// 			//Message sent to embeded channel
// 			funnelEmbeded := &discordgo.MessageEmbed{
// 				Title:       "LE EPIC FUNNEL TIME",
// 				Description: "Funnel command",
// 				Fields: []*discordgo.MessageEmbedField{
// 					{
// 						Name:   "Action",
// 						Value:  action,
// 						Inline: true,
// 					},
// 					{
// 						Name:   "Who",
// 						Value:  who,
// 						Inline: true,
// 					},
// 					{
// 						Name:  "Timestamp",
// 						Value: commandInvoke,
// 					},
// 					{
// 						Name:  "User",
// 						Value: i.Member.User.Username,
// 					},
// 					{
// 						Name: "Plain text",
// 						Value: fmt.Sprintf(
// 							"Action: %s\nWho: %s\nTimestamp: %s\nUser: %s",
// 							action, who, commandInvoke, i.Member.User.Username,
// 						),
// 					},
// 				},
// 			}
// 			//Send action to funnel channel and ensure it is sent
// 			channel, err := s.Channel(funnelChannelID)
// 			if err != nil {
// 				SendChannelMessage(s, i, "Error: Could not find funnel channel")
// 				return
// 			}
// 			_, err = s.ChannelMessageSendEmbed(channel.ID, funnelEmbeded)
// 			if err != nil {
// 				SendChannelMessage(s, i, "Failed to send embeded message")
// 				return
// 			}
// 			//Send confirmation message to user
// 			SendChannelMessage(s, i, "Action sent to funnel channel")
// 		},
// 	}
//
// }
