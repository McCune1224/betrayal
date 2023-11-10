package commands

import (
	"fmt"
	"log"
	"math/rand"

	"github.com/bwmarrin/discordgo"
	"github.com/go-co-op/gocron"
	"github.com/mccune1224/betrayal/internal/data"
	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/zekrotja/ken"
)

type Setup struct {
	models    data.Models
	scheduler *gocron.Scheduler
}

func (s *Setup) Initialize(models data.Models, scheduler *gocron.Scheduler) {
	s.models = models
	s.scheduler = scheduler
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
	return []*discordgo.ApplicationCommandOption{
		discord.IntCommandArg("playercount", "number of players for the game", true),
	}
}

// Run implements ken.SlashCommand.
func (s *Setup) Run(ctx ken.Context) (err error) {
	// This will prob take more than 3 seconds to run
	if err = ctx.Defer(); err != nil {
		log.Println(err)
		return err
	}
	// generate role pool
	// and make embed view

	decepts := getDeceptionist(ctx.GetSession(), ctx.GetEvent().GuildID)
	rolePool, err := generateRoleSelectPool(s.models)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx)
	}
	playerCount := int(ctx.Options().GetByName("playercount").IntValue())
	decepCount := len(decepts)
	rp := generateRolePools(rolePool, playerCount, decepCount)
	msg := roleSetupEmbed(rp)
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
	roleList := activeRolesQueue.Roles
	roles, err := m.Roles.GetBulkByName(roleList)
	if err != nil {
		return nil, err
	}

	return roles, nil
}

type rolePool struct {
	// Reserved roles for deceptionists to choose from
	deceptionOptions [][]*data.Role
	// Raw role pool for random selection
	randomPool []*data.Role
}

func generateRolePools(roles []*data.Role, playerCount, decepCount int) *rolePool {
	goodRoles, badRoles, neutralRoles := groupRoles(roles)
	rp := &rolePool{}
	gPerm := rand.Perm(len(goodRoles))
	bPerm := rand.Perm(len(badRoles))
	nPerm := rand.Perm(len(neutralRoles))
	for i := 0; i < decepCount; i++ {
		rp.deceptionOptions = append(rp.deceptionOptions, []*data.Role{goodRoles[gPerm[i]], neutralRoles[nPerm[i]], badRoles[bPerm[i]]})
	}
	rPerm := rand.Perm(len(roles))
	for i := 0; i < playerCount; i++ {
		rp.randomPool = append(rp.randomPool, roles[rPerm[i]])
	}
	return rp
}

func roleSetupEmbed(rp *rolePool) *discordgo.MessageEmbed {
	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("Role Setup (%d)", len(rp.randomPool)),
		Description: fmt.Sprintf("(%s = Deceptionist Reserved)", discord.EmojiWarning),
	}
	for i := range rp.deceptionOptions {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   fmt.Sprintf("Deceptionist %d", i+1),
			Value:  fmt.Sprintf("%s\n%s\n%s", rp.deceptionOptions[i][0].Name, rp.deceptionOptions[i][1].Name, rp.deceptionOptions[i][2].Name),
			Inline: true,
		})
	}
	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
		Name: "----- Random Roll Pool -----",
	})
	// Making a left and right column for the random pool so its not a super long list
	leftRolesField := &discordgo.MessageEmbedField{
		Inline: true,
	}
	rightRolesField := &discordgo.MessageEmbedField{
		Inline: true,
	}
	for i := range rp.randomPool {
		currSide := &discordgo.MessageEmbedField{}
		if i%2 == 0 {
			currSide = leftRolesField
		} else {
			currSide = rightRolesField
		}
		// Check if role is already within Decetionist options
		marked := false
		for j := range rp.deceptionOptions {
			if rp.randomPool[i].Name == rp.deceptionOptions[j][0].Name || rp.randomPool[i].Name == rp.deceptionOptions[j][1].Name || rp.randomPool[i].Name == rp.deceptionOptions[j][2].Name {
				// Add but add a warning
				currSide.Value += fmt.Sprintf("%s %s\n", discord.EmojiWarning, rp.randomPool[i].Name)
				marked = true
				break
			}
		}
		if !marked {
			currSide.Value += fmt.Sprintf("%s\n", rp.randomPool[i].Name)
		}
	}
	embed.Fields = append(embed.Fields, leftRolesField, rightRolesField)
	return embed
}

// Will group list of roles into sub list of roles based off alignment
func groupRoles(r []*data.Role) (goodRoles []*data.Role, badRoles []*data.Role, neutralRoles []*data.Role) {
	for i := range r {
		switch r[i].Alignment {
		case "GOOD":
			goodRoles = append(goodRoles, r[i])
		case "EVIL":
			badRoles = append(badRoles, r[i])
		case "NEUTRAL":
			neutralRoles = append(neutralRoles, r[i])
		}
	}
	// WARNING: Hahahahaha why the hell does Go have naked returns this is so goofy
	return
}

// TODO: Find me a better home :(
func randomSliceElement[T any](s []T) T {
	n := rand.Int() % len(s)
	return s[n]
}
