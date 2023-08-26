package main

import (
	"errors"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/mccune1224/betrayal/internal/data"
)

// All slash commands for the insult feature
func (a *app) InsultCommandBundle() []SlashCommand {
	return []SlashCommand{
		a.InsultAddComand(),
		a.InsultGetCommand(),
	}
}

func (a *app) InsultAddComand() SlashCommand {
	return SlashCommand{
		Feature: discordgo.ApplicationCommand{
			Name:        "insult_add",
			Description: "new insult for mckusa",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "message",
					Description: "The message to send to mckusa",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    true,
				},
			},
		},

		Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			insult := i.ApplicationCommandData().Options[0].StringValue()

			insultEntry := data.Insult{
				Insult:   insult,
				AuthorID: i.Member.User.ID,
			}
			a.logger.Println(insultEntry)
			err := a.models.Insults.Insert(&insultEntry)
			if errors.Is(err, data.ErrRecordAlreadyExists) {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "You've already said that to mckusa, but I'll let him know again",
						Flags:   64,
					},
				})
				s.ChannelMessageSend(
					i.ChannelID,
					fmt.Sprintf("hey <@%s>, %s", mckusaID, insult),
				)
				return
			}
			if err != nil {
				a.logger.Println(err)
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "Something went wrong, ping mckusa and let him know if urgent",
						Flags:   64,
					},
				})
			}

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{},
			})
		},
	}
}

func (a *app) InsultGetCommand() SlashCommand {

	return SlashCommand{
		Feature: discordgo.ApplicationCommand{
			Name:        "insult_get",
			Description: "Get a random insult saved for Alex and let him know how bad he is",
		},

		Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			insult, err := a.models.Insults.GetRandom()
			if err != nil {
				a.logger.Println(err)
			}
			a.logger.Println(insult)
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("hey <@%s>, %s", mckusaID, insult.Insult),
				},
			},
			)
		},
	}
}
