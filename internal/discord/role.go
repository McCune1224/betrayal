package discord

import (
	"github.com/zekrotja/ken"
)

// Current roles with eleveted permissions.
var AdminRoles = []string{
	"Host",
	"Co-Host",
	"Bot Developer",
}

// Check if user who invoked command has required role
func IsAdminRole(ctx ken.Context, requiredRole ...string) bool {
	event := ctx.GetEvent()
	guildRoles, _ := ctx.GetSession().GuildRoles(event.GuildID)

	// Nothing screams "I'm a good programmer" more than a triple for loop.
	for _, rid := range event.Member.Roles {
		for _, r := range guildRoles {
			for _, ar := range requiredRole {
				if rid == r.ID && r.Name == ar {
					return true
				}
			}
		}
	}
	return false
}
