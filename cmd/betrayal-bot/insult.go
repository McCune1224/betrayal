package main

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/mccune1224/betrayal/internal/data"
)

func (a *app) InsultCommand() SlashCommand {
	return SlashCommand{
		Feature: discordgo.ApplicationCommand{
			Name:        "insult",
			Description: "Let mckusa know he's a bad programmer",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "message",
					Description: "Give him your most brutal insult",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    true,
				},
			},
		},

		Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			insult := i.ApplicationCommandData().Options[0].StringValue()
			message := ""
			foo := a.models.Insults.DB.Create(&data.Insult{Insult: insult})
			if foo.Error != nil {
				message = "Alex is a REALLY bad programmer. Can't even make a way to complain about his bad programming work, ironic."
			} else {
				message = "Thanks for the insult. I'll make sure he sees it. use /insult_random to see what you others have said at random."
			}
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: message,
				},
			})
		},
	}
}

func (a *app) RandomInsultCommand() SlashCommand {
	return SlashCommand{
		Feature: discordgo.ApplicationCommand{
			Name:        "insult_random",
			Description: "Get a random insult saved for Alex and let him know how bad he is",
		},
		Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			message := ""
			insult, err := a.models.Insults.GetRandomInsult()
			if err != nil {
				message = "Alex is a REALLY bad programmer. Can't even make a way to complain about his bad programming work, ironic."
			} else {
				mentionUser := fmt.Sprintf("<@%s>", "206268866714796032")
				message = fmt.Sprintf("Hey %s, %s", mentionUser, insult.Insult)
			}
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: message,
				},
			},
			)
		},
	}
}
