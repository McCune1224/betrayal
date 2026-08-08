package discord

import (
	"errors"
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/zekrotja/ken"
)

// Current roles with eleveted permissions.
var AdminRoles = []string{
	"Host",
	"Co-Host",
	"Bot Developer",
}

// Check if user who invoked command has required role
func IsAdminRole(ctx ken.Context, adminRoles ...string) bool {
	admin, err := IsAdminRoleWithError(ctx, adminRoles...)
	if err != nil {
		log.Printf("discord admin role lookup failed: %v", err)
		return false
	}
	return admin
}

func IsAdminRoleWithError(ctx ken.Context, adminRoles ...string) (bool, error) {
	if ctx == nil || ctx.GetEvent() == nil || ctx.GetSession() == nil {
		return false, errors.New("discord context or session unavailable")
	}
	event := ctx.GetEvent()
	return resolveAdminRole(event.Member, func() ([]*discordgo.Role, error) {
		return ctx.GetSession().GuildRoles(event.GuildID)
	}, adminRoles...)
}

func resolveAdminRole(member *discordgo.Member, lookup func() ([]*discordgo.Role, error), adminRoles ...string) (bool, error) {
	if member == nil {
		return false, nil
	}
	guildRoles, err := lookup()
	if err != nil {
		return false, err
	}
	for _, rid := range member.Roles {
		for _, role := range guildRoles {
			for _, adminRole := range adminRoles {
				if role != nil && rid == role.ID && role.Name == adminRole {
					return true, nil
				}
			}
		}
	}
	return false, nil
}

func GetAdminRoleUsers(s *discordgo.Session, e *discordgo.InteractionCreate, adminRoles ...string) ([]string, error) {
	if s == nil || e == nil || e.Interaction == nil || e.Member == nil || e.Member.User == nil {
		return nil, errors.New("discord session or interaction member unavailable")
	}
	guildRoles, err := s.GuildRoles(e.GuildID)
	if err != nil {
		return nil, err
	}
	var users []string
	for _, rid := range e.Member.Roles {
		for _, r := range guildRoles {
			for _, ar := range adminRoles {
				if rid == r.ID && r.Name == ar {
					users = append(users, e.Member.User.ID)
				}
			}
		}
	}
	return users, nil
}
