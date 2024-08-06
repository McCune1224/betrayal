package list

import (
	"context"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/mccune1224/betrayal/internal/models"
	"github.com/zekrotja/ken"
)

var (
	DummyGoodRoles    = []string{"Agent", "Analyst", "Biker", "Cerberus", "Detective", "Fisherman", "Gunman", "Hero", "Hydra", "Judge", "Knight", "The Major", "Medium", "Nurse", "Seraph", "Terminal", "Time Traveler", "Undercover", "Wizard", "Yeti"}
	DummyNeutralRoles = []string{"Amalgamation", "Backstabber", "Bard", "Bomber", "Cheater", "Entertainer", "Empress", "Ghost", "Goliath", "Incubus", "Magician", "Masochist", "Mercenary", "Mimic", "Pathologist", "Siren", "Sidekick", "Succubus", "Villager", "Wanderer"}
	DummyEvilRoles    = []string{"Anarchist", "Arsonist", "Bartender", "Consort", "Director", "Doll", "Forsaken Angel", "Gatekeeper", "Hacker", "Highwayman", "Hunter", "Jester", "Overlord", "Parasite", "Phantom", "Psychotherapist", "Revenant", "Slaughterer", "Threatener", "Witchdoctor"}

	GameEvents = []string{
		"Care Package - Game Start - Each player starts off with a care package which contains 1 item and 1 Any Ability.",
		"Daily Bonuses - Every Day - Gain 300 coins every day, other than the first.",
		"Item Rain - Every Third Day - Everyone gains 1-3 random items (luck affects your odds).",
		"Power Drop - Day After Item Rain - Everyone gains 1 random Any Ability.",
		"Rock Paper Scissors Tournament - Day 5 Event - Everyone plays rock, paper, scissors. Winner gets a special prize.",
		"Money Heaven - Day 7 and Day 13 Event - All of the coins you earn are doubled today.",
		"Valentine's Day - Day 8 Event - Send a valentine and an anonymous message costing 50 coins to someone. You cannot receive valentines if you don't send one. Cannot send to yourself.",
		"Duels - Day 11 & 14 Event - Choose to challenge someone to a duel. Life is at stake.",
		"Ultimate Exchange - Five Player Event - Whoever is holding the Lucky Coin may convert it into 1500 coins.",
		"Double Elimination - Random Event - There will be two Elimination Phases today.",
	}
)

type List struct {
	dbPoll *pgxpool.Pool
}

func (l *List) Initialize(pool *pgxpool.Pool) {
	l.dbPoll = pool
}

var _ ken.SlashCommand = (*List)(nil)

// Description implements ken.SlashCommand.
func (*List) Description() string {
	return "Get a list of desired category"
}

// Name implements ken.SlashCommand.
func (*List) Name() string {
	return discord.DebugCmd + "list"
}

// Options implements ken.SlashCommand.
func (*List) Options() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{
		{
			Name:        "items",
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Description: "List of all items",
			Options: []*discordgo.ApplicationCommandOption{
				discord.BoolCommandArg("all", "Get all items", false),
			},
		},
		{
			Name:        "active_roles",
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Description: "List of all active roles in game.",
		},
		{
			Name:        "events",
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Description: "List of all events",
		},
		{
			Name:        "statuses",
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Description: "List of all statuses",
		},
	}
}

// Run implements ken.SlashCommand.
func (l *List) Run(ctx ken.Context) (err error) {
	return ctx.HandleSubCommands(
		ken.SubCommandHandler{Name: "items", Run: l.listItems},
		ken.SubCommandHandler{Name: "active_roles", Run: l.listActiveRoles},
		ken.SubCommandHandler{Name: "events", Run: l.listEvents},
		ken.SubCommandHandler{Name: "statuses", Run: l.listStatuses},
	)
}

func (l *List) listStatuses(ctx ken.SubCommandContext) (err error) {
	if err := ctx.Defer(); err != nil {
		log.Println(err)
		return err
	}
	q := models.New(l.dbPoll)
	statuses, err := q.ListStatus(context.Background())
	if err != nil {
		log.Println(err)
		discord.AlexError(ctx, "failed to get statuses")
	}
	fields := []*discordgo.MessageEmbedField{}
	for _, s := range statuses {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:  s.Name,
			Value: s.Description,
		})
	}
	return ctx.RespondEmbed(&discordgo.MessageEmbed{
		Title:       "Statuses",
		Description: "List of all statuses",
		Fields:      fields,
	})
}

func (l *List) listEvents(ctx ken.SubCommandContext) (err error) {
	if err := ctx.Defer(); err != nil {
		log.Println(err)
		return err
	}
	fields := []*discordgo.MessageEmbedField{}
	for _, e := range GameEvents {
		split := strings.Split(e, " -")
		name := split[0]
		desc := strings.Join(split[1:], " -")
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:  name,
			Value: desc,
		})
	}
	return ctx.RespondEmbed(&discordgo.MessageEmbed{
		Title:       "Events",
		Description: "All events in the game",
		Fields:      fields,
	})
}

func (l *List) listItems(ctx ken.SubCommandContext) (err error) {
	if err := ctx.Defer(); err != nil {
		log.Println(err)
		return err
	}
	return discord.AlexError(ctx, "not implemented")
}

func (l *List) listActiveRoles(ctx ken.SubCommandContext) (err error) {
	if err := ctx.Defer(); err != nil {
		log.Println(err)
		return err
	}
	fields := []*discordgo.MessageEmbedField{
		{
			Name:   "Good",
			Value:  strings.Join(DummyGoodRoles, "\n"),
			Inline: true,
		},
		{
			Name:   "Neutral",
			Value:  strings.Join(DummyNeutralRoles, "\n"),
			Inline: true,
		},
		{
			Name:   "Evil",
			Value:  strings.Join(DummyEvilRoles, "\n"),
			Inline: true,
		},
	}
	listEmbed := &discordgo.MessageEmbed{
		Title:       "Active Game Roles",
		Description: "All active roles for the current game",
		Fields:      fields,
	}

	return ctx.RespondEmbed(listEmbed)
}

// Version implements ken.SlashCommand.
func (*List) Version() string {
	return "1.0.0"
}
