package discord

import (
	"errors"

	"github.com/bwmarrin/discordgo"
)

var ErrChannelNotFound = errors.New("channel not found")

func GetGuildChannelCategory(s *discordgo.Session, e *discordgo.InteractionCreate, channelName string) (*discordgo.Channel, error) {
	channels, err := s.GuildChannels(e.GuildID)
	if err != nil {
		return nil, err
	}

	for _, c := range channels {
		if c.Type == discordgo.ChannelTypeGuildCategory && c.Name == channelName {
			return c, nil
		}
	}
	return nil, ErrChannelNotFound
}

func AddChannelWithinCategory(s *discordgo.Session, e *discordgo.InteractionCreate, categoryName string, channelName string) (*discordgo.Channel, error) {
	category, err := GetGuildChannelCategory(s, e, categoryName)
	if err != nil {
		return nil, err
	}
	channel, err := s.GuildChannelCreate(BetraylGuildID, channelName, discordgo.ChannelTypeGuildText)
	if err != nil {
		return nil, err
	}
	subChannel, err := s.ChannelEditComplex(channel.ID, &discordgo.ChannelEdit{
		ParentID: category.ID,
	})
	if err != nil {
		return nil, err
	}
	return subChannel, err
}
