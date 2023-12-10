package commands

import (
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/mccune1224/betrayal/internal/data"
	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/mccune1224/betrayal/internal/scheduler"
	"github.com/zekrotja/ken"
)

// TODO: Slap these in the database once game is close to starting and roles are finalized
var (
	DummyGoodRoles    = []string{"Agent", "Analyst", "Biker", "Cerberus", "Detective", "Fisherman", "Gunman", "Hero", "Hydra", "Judge", "Major", "Mecha", "Medium", "Nurse", "Seraph", "Terminal", "Time Traveler", "Undercover", "Wizard", "Yeti"}
	DummyNeutralRoles = []string{"Amalgamation", "Backstabber", "Banker", "Bomber", "Cheater", "Cyborg", "Empress", "Ghost", "Goliath", "Journalist", "Magician", "Masochist", "Mercenary", "Mimic", "Pathologist", "Salesman", "Siren", "Tinkerer", "Villager", "Wanderer"}
	DummyEvilRoles    = []string{"Anarchist", "Arsonist", "Bartender", "Consort", "Director", "Doll", "Forsaken Angel", "Gatekeeper", "Hacker", "Highwayman", "Hunter", "Imp", "Jester", "Juggernaut", "Overlord", "Phantom", "Psychotherapist", "Slaughterer", "Threatener", "Witchdoctor"}
	GameEvents        = []string{
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
	models    data.Models
	scheduler *scheduler.BetrayalScheduler
}

func (l *List) Initialize(models data.Models, scheduler *scheduler.BetrayalScheduler) {
	l.models = models
	l.scheduler = scheduler
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
			Name:        "roles",
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Description: "Get a list of roles",
			Options: []*discordgo.ApplicationCommandOption{
				discord.BoolCommandArg("active", "Get active roles", false),
			},
		},
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
	statuses, err := l.models.Statuses.GetAll()
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
	_, err = l.models.RoleLists.Get()
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "failed to get active roles")
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
