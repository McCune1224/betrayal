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

func CreateChannelWithinCategory(s *discordgo.Session, e *discordgo.InteractionCreate, categoryName string, channelName string, hidden ...bool) (*discordgo.Channel, error) {
	hiddenChannel := false
	if len(hidden) > 0 {
		hiddenChannel = hidden[0]
	}

	category, err := GetGuildChannelCategory(s, e, categoryName)
	if err != nil {
		return nil, err
	}

	channel := &discordgo.Channel{}

	if hiddenChannel {
		channel, err = CreateHiddenChannel(s, e, channelName)
		if err != nil {
			return nil, err
		}
	} else {
		channel, err = s.GuildChannelCreate(BetraylGuildID, channelName, discordgo.ChannelTypeGuildText)
		if err != nil {
			return nil, err
		}
	}

	subChannel, err := s.ChannelEditComplex(channel.ID, &discordgo.ChannelEdit{
		ParentID: category.ID,
	})
	if err != nil {
		return nil, err
	}
	return subChannel, err
}

// Wrapper ontop of discordgo.GuildChannelCreate to create a hidden channel besided for the user and the admin
func CreateHiddenChannel(s *discordgo.Session, e *discordgo.InteractionCreate, channelName string, whitelistIds ...string) (*discordgo.Channel, error) {
	adminIDs := GetAdminRoleUsers(s, e, AdminRoles...)
	whiteListed := append(adminIDs, whitelistIds...)

	channel, err := s.GuildChannelCreate(e.GuildID, channelName, discordgo.ChannelTypeGuildText)
	if err != nil {
		return nil, err
	}

	// Set the default permissions for the channel to fully private
	err = s.ChannelPermissionSet(channel.ID, BetraylGuildID, discordgo.PermissionOverwriteTypeRole, 0, discordgo.PermissionViewChannel)
	if err != nil {
		return nil, err
	}

	// allow the whitelistIds to see and interact with the channel
	for _, id := range whiteListed {
		AddMemberToChannel(s, e, channel.ID, id)
	}

	//		for _, member := range guildMembers {
	//			// skip the user if they're in the whitelist
	//			if contains(whiteListed, member.User.ID) {
	//				continue
	//			}
	//			err = s.ChannelPermissionSet(channel.ID, member.User.ID, discordgo.PermissionOverwriteTypeMember, 0, discordgo.PermissionViewChannel)
	//			if err != nil {
	//				return nil, err
	//			}
	//		}
	//
	//		return channel, nil
	//	}
	return channel, nil
}

func contains(list []string, item string) bool {
	for _, i := range list {
		if i == item {
			return true
		}
	}
	return false
}

func AddMemberToChannel(s *discordgo.Session, e *discordgo.InteractionCreate, channelID string, userID string) error {
	err := s.ChannelPermissionSet(channelID, userID, discordgo.PermissionOverwriteTypeMember, discordgo.PermissionViewChannel, discordgo.PermissionViewChannel)
	if err != nil {
		return err
	}
	return nil
}
