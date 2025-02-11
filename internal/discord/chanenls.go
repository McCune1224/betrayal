package discord

import (
	"errors"
	"fmt"
	"strings"

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

func GetChannelsWithinCategory(s *discordgo.Session, e *discordgo.InteractionCreate, categoryName string) (*[]discordgo.Channel, error) {
	category, err := GetGuildChannelCategory(s, e, categoryName)
	if err != nil {
		return nil, err
	}

	channels, err := s.GuildChannels(e.GuildID)
	if err != nil {
		return nil, err
	}

	var matchingChannels []discordgo.Channel
	for _, c := range channels {
		if c.ParentID == category.ID {
			matchingChannels = append(matchingChannels, *c)
		}
	}

	if len(matchingChannels) == 0 {
		return nil, ErrChannelNotFound
	}

	return &matchingChannels, nil
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
		AddMemberToChannel(s, channel.ID, id)
	}
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

func AddMemberToChannel(s *discordgo.Session, channelID string, userID string) error {
	err := s.ChannelPermissionSet(channelID, userID, discordgo.PermissionOverwriteTypeMember, discordgo.PermissionViewChannel, discordgo.PermissionViewChannel)
	if err != nil {
		return err
	}
	return nil
}

func GetChannelByName(s *discordgo.Session, e *discordgo.InteractionCreate, name string) (*discordgo.Channel, error) {
	guildID := e.GuildID
	channels, err := s.GuildChannels(guildID)
	if err != nil {
		return nil, err
	}
	target := &discordgo.Channel{}

	for _, c := range channels {
		if strings.EqualFold(c.Name, name) {
			target = c
			break
		}
	}

	if target.ID == "" {
		return nil, ErrChannelNotFound
	}

	return target, nil
}

func GetChannelMembers(s *discordgo.Session, e *discordgo.InteractionCreate, channelID string) ([]*discordgo.Member, error) {
	members, err := s.GuildMembers(BetraylGuildID, "", 1000)
	if err != nil {
		return nil, err
	}

	permissions, err := s.State.UserChannelPermissions(s.State.User.ID, channelID)
	if err != nil {
		fmt.Println("Error getting user permissions,", err)
		return nil, err
	}

	allowedMembers := []*discordgo.Member{}
	guildRoles, _ := s.GuildRoles(e.GuildID)
	for _, member := range members {
		if permissions&discordgo.PermissionViewChannel == discordgo.PermissionViewChannel {
			for _, rid := range member.Roles {
				for _, r := range guildRoles {
					if rid == r.ID && r.Name == "Participant" {
						allowedMembers = append(allowedMembers, member)
					}
				}
			}
		}
	}
	return allowedMembers, nil
}
