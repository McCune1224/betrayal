package main

import "github.com/bwmarrin/discordgo"

type SlashCommand struct {
	Feature discordgo.ApplicationCommand
	Handler func(s *discordgo.Session, i *discordgo.InteractionCreate)
}
