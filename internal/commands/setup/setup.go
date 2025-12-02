package setup

import (
	"context"
	"fmt"
	"github.com/mccune1224/betrayal/internal/logger"
	"math/rand"

	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/mccune1224/betrayal/internal/models"
	"github.com/zekrotja/ken"
)

var activeRoleList = []string{
	"Agent", "Amalgamation", "Anarchist",
	"Analyst", "Backstabber", "Arsonist",
	"Biker", "Bard", "Bartender",
	"Cerberus", "Bomber", "Consort",
	"Detective", "Cheater", "Director",
	"Fisherman", "Entertainer", "Doll",
	"Gunman", "Empress", "Forsaken Angel",
	"Hero", "Ghost", "Gatekeeper",
	"Hydra", "Goliath", "Hacker",
	"Judge", "Incubus", "Highwayman",
	"Knight", "Magician", "Hunter",
	"The Major", "Masochist", "Jester",
	"Medium", "Mercenary", "Juggernaut",
	"Nurse", "Mimic", "Overlord",
	"Seraph", "Pathologist", "Parasite",
	"Terminal", "Salesman", "Phantom",
	"Time Traveler", "Siren", "Psychotherapist",
	"Undercover", "Sidekick", "Slaughterer",
	"Wizard", "Villager", "Threatener",
	"Yeti", "Wanderer", "Witchdoctor",
}

type Setup struct {
	dbPool *pgxpool.Pool
}

func (s *Setup) Initialize(pool *pgxpool.Pool) {
	s.dbPool = pool
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
		discord.IntCommandArg("player_count", "number of players for the game", true),
		discord.IntCommandArg("decept_count", "number of deceptionists for the game", false),
	}
}

// Run implements ken.SlashCommand.
func (s *Setup) Run(ctx ken.Context) (err error) {
	defer logger.RecoverWithLog(*logger.Get())

	// This will prob take more than 3 seconds to run
	if err = ctx.Defer(); err != nil {
		logger.Get().Error().Err(err).Msg("operation failed")
		return err
	}
	// generate role pool
	// and make embed view

	rolePool, err := generateRoleSelectPool(s.dbPool)
	if err != nil {
		logger.Get().Error().Err(err).Msg("operation failed")
		return discord.AlexError(ctx, "failed to generate role pool")
	}

	playerCount := int(ctx.Options().GetByName("player_count").IntValue())

	// Validate player count against available roles
	if playerCount > len(rolePool) {
		return discord.ErrorMessage(ctx, "Invalid Player Count",
			fmt.Sprintf("Player count (%d) cannot exceed available roles (%d)", playerCount, len(rolePool)))
	}

	// Default grab all deceptionists from server if not specified
	decepCount := 0
	if decepArg, ok := ctx.Options().GetByNameOptional("decept_count"); ok {
		decepCount = int(decepArg.IntValue())
		// Validate deceptionist count
		if decepCount > len(rolePool) {
			return discord.ErrorMessage(ctx, "Invalid Deceptionist Count",
				fmt.Sprintf("Deceptionist count (%d) cannot exceed available roles (%d)", decepCount, len(rolePool)))
		}
	} else {
		decepts := getDeceptionist(ctx.GetSession(), ctx.GetEvent().GuildID)
		decepCount = len(decepts)
	}

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
func generateRoleSelectPool(m *pgxpool.Pool) ([]models.Role, error) {
	q := models.New(m)
	dbCtx := context.Background()
	roles, err := q.ListRolesByName(dbCtx, activeRoleList)
	if err != nil {
		return nil, err
	}

	return roles, nil
}

type rolePool struct {
	// Reserved roles for deceptionists to choose from
	deceptionOptions [][]models.Role
	// Raw role pool for random selection
	randomPool []models.Role
}

func generateRolePools(roles []models.Role, playerCount, decepCount int) *rolePool {
	goodRoles, badRoles, neutralRoles := groupRoles(roles)
	rp := &rolePool{}
	gPerm := rand.Perm(len(goodRoles))
	bPerm := rand.Perm(len(badRoles))
	nPerm := rand.Perm(len(neutralRoles))
	for i := 0; i < decepCount; i++ {
		rp.deceptionOptions = append(rp.deceptionOptions, []models.Role{goodRoles[gPerm[i]], neutralRoles[nPerm[i]], badRoles[bPerm[i]]})
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
func groupRoles(r []models.Role) (goodRoles []models.Role, badRoles []models.Role, neutralRoles []models.Role) {
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
