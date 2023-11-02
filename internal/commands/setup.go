package commands

import (
	"fmt"
	"math/rand"

	"github.com/bwmarrin/discordgo"
	"github.com/mccune1224/betrayal/internal/data"
	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/zekrotja/ken"
	"golang.org/x/exp/slices"
)

type Setup struct {
	models data.Models
}

func (s *Setup) SetModels(models data.Models) {
	s.models = models
}

var _ ken.SlashCommand = (*Setup)(nil)

// Description implements ken.SlashCommand.
func (*Setup) Description() string {
	return "Helper commands to prepare for game"
}

// Name implements ken.SlashCommand.
func (*Setup) Name() string {
	return "setup"
}

// Options implements ken.SlashCommand.
func (*Setup) Options() []*discordgo.ApplicationCommandOption {
	return nil
}

// Run implements ken.SlashCommand.
func (s *Setup) Run(ctx ken.Context) (err error) {
	// This will prob take more than 3 seconds to run
	if err = ctx.Defer(); err != nil {
		return err
	}
	// generate role pool
	// and make embed view

	decepts := getDeceptionist(ctx.GetSession(), ctx.GetEvent().GuildID)
	rolePool, err := generateRoleSelectPool(s.models)
	if err != nil {
		return discord.AlexError(ctx)
	}
	msg := rolePreviewEmbed(rolePool, len(decepts))

	return ctx.RespondEmbed(msg)
}

// Version implements ken.SlashCommand.
func (*Setup) Version() string {
	return "1.0.0"
}

// Finds all players within the server that contain Deceiptionist role
func getDeceptionist(s *discordgo.Session, gID string) []*discordgo.User {
	guildRoles, _ := s.GuildRoles(gID)
	var decRole *discordgo.Role
	for _, r := range guildRoles {
		if r.Name == "Deceptionist" {
			decRole = r
			break
		}
	}
	var deceptionists []*discordgo.User
	members, _ := s.GuildMembers(gID, "", 1000)
	for _, m := range members {
		for _, r := range m.Roles {
			if r == decRole.ID {
				deceptionists = append(deceptionists, m.User)
				break
			}
		}
	}
	return deceptionists
}

// Will find all active roles listed for the game
func generateRoleSelectPool(m data.Models) ([]*data.Role, error) {
	activeRolesQueue, err := m.RoleLists.Get()
	if err != nil {
		return nil, err
	}
	var roles []*data.Role
	roleList := activeRolesQueue.Roles
	for i := range roleList {
		r, err := m.Roles.GetByName(roleList[i])
		if err != nil {
			return nil, err
		}
		roles = append(roles, r)
	}
	return roles, err
}

// TODO: Find a better home to locate this function... really don't want a 5000 line file called "utils.go"...
func randomSliceElement[T any](s []T) T {
	return s[rand.Intn(len(s))]
}

func rolePreviewEmbed(roles []*data.Role, decepCount int) *discordgo.MessageEmbed {
	goodRoles, badRoles, neutralRoles := groupRoles(roles)
	takenRoles := []*data.Role{}
	deceptionistsChoices := [][]*data.Role{}
	for i := 0; i < decepCount; i++ {
		// Keep selecting random roles until we find one that isn't already reserved
		g := randomSliceElement(goodRoles)
		b := randomSliceElement(badRoles)
		n := randomSliceElement(neutralRoles)
		for slices.Contains(takenRoles, g) {
			g = randomSliceElement(goodRoles)
		}
		for slices.Contains(takenRoles, b) {
			b = randomSliceElement(badRoles)
		}
		for slices.Contains(takenRoles, n) {
			n = randomSliceElement(neutralRoles)
		}
		deceptionistsChoices = append(deceptionistsChoices, []*data.Role{g, b, n})
		takenRoles = append(takenRoles, g, b, n)
	}
	deceptFields := []*discordgo.MessageEmbedField{}
	for i := range deceptionistsChoices {
		deceptFields = append(deceptFields, &discordgo.MessageEmbedField{
			Name:   "Deceptionist " + fmt.Sprintf("%d", i+1),
			Value:  deceptionistsChoices[i][0].Name + "\n" + deceptionistsChoices[i][1].Name + "\n" + deceptionistsChoices[i][2].Name,
			Inline: true,
		})
	}
	deceptFields = append(deceptFields, &discordgo.MessageEmbedField{
		Name:   "Remaining Roles",
		Value:  "Remaining roles will be randomly assigned to players",
		Inline: false,
	})
	remainderRoleFields := []*discordgo.MessageEmbedField{}
	for i := range roles {
		if !slices.Contains(takenRoles, roles[i]) {
			remainderRoleFields = append(remainderRoleFields, &discordgo.MessageEmbedField{
				Name:   roles[i].Name,
				Value:  roles[i].Description,
				Inline: true,
			})
		}
	}
	msg := &discordgo.MessageEmbed{
		Title:  "Role Preview",
		Color:  0x00ff00,
		Fields: append(deceptFields, remainderRoleFields...),
	}
	return msg
}

// Will group list of roles into sub list of roles based off alignment
func groupRoles(r []*data.Role) (goodRoles []*data.Role, badRoles []*data.Role, neutralRoles []*data.Role) {
	for i := range r {
		switch r[i].Alignment {
		case "GOOD":
			goodRoles = append(goodRoles, r[i])
		case "BAD":
			badRoles = append(badRoles, r[i])
		case "NEUTRAL":
			neutralRoles = append(neutralRoles, r[i])
		}
	}
	// WARNING: Hahahahaha why the hell does Go have naked returns this is so goofy
	return
}
