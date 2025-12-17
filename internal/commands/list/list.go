package list

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/mccune1224/betrayal/internal/logger"

	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/mccune1224/betrayal/internal/models"
	"github.com/mccune1224/betrayal/internal/util"
	"github.com/zekrotja/ken"
)

var (
	DummyGoodRoles    = []string{"Agent", "Analyst", "Biker", "Cerberus", "Detective", "Fisherman", "Gunman", "Hero", "Hydra", "Judge", "Knight", "The Major", "Medium", "Nurse", "Seraph", "Terminal", "Time Traveler", "Undercover", "Wizard", "Yeti"}
	DummyNeutralRoles = []string{"Amalgamation", "Backstabber", "Bard", "Bomber", "Cheater", "Entertainer", "Empress", "Ghost", "Goliath", "Incubus", "Magician", "Masochist", "Mercenary", "Mimic", "Pathologist", "Siren", "Sidekick", "Succubus", "Villager", "Wanderer"}
	DummyEvilRoles    = []string{"Anarchist", "Arsonist", "Bartender", "Consort", "Cultist", "Juggernaut", "Doll", "Forsaken Angel", "Gatekeeper", "Hacker", "Highwayman", "Hunter", "Jester", "Overlord", "Parasite", "Phantom", "Psychotherapist", "Slaughterer", "Threatener", "Witchdoctor"}

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
	dbPool *pgxpool.Pool
}

func (l *List) Initialize(pool *pgxpool.Pool) {
	l.dbPool = pool
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
		{
			Name:        "notes",
			Description: "(Admin Only) List all current player notes",
			Type:        discordgo.ApplicationCommandOptionSubCommand,
		},
	}
}

// Run implements ken.SlashCommand.
func (l *List) Run(ctx ken.Context) (err error) {
	defer logger.RecoverWithLog(*logger.Get())

	return ctx.HandleSubCommands(
		ken.SubCommandHandler{Name: "items", Run: l.listItems},
		ken.SubCommandHandler{Name: "active_roles", Run: l.listActiveRoles},
		ken.SubCommandHandler{Name: "events", Run: l.listEvents},
		ken.SubCommandHandler{Name: "statuses", Run: l.listStatuses},
		ken.SubCommandHandler{Name: "notes", Run: l.listNotes},
	)
}

func (l *List) listStatuses(ctx ken.SubCommandContext) (err error) {
	if err := ctx.Defer(); err != nil {
		logger.Get().Error().Err(err).Msg("operation failed")
		return err
	}
	q := models.New(l.dbPool)
	statuses, err := q.ListStatus(context.Background())
	if err != nil {
		logger.Get().Error().Err(err).Msg("operation failed")
		return discord.AlexError(ctx, "failed to get statuses")
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
		logger.Get().Error().Err(err).Msg("operation failed")
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
		logger.Get().Error().Err(err).Msg("operation failed")
		return err
	}
	q := models.New(l.dbPool)
	items, err := q.ListItem(context.Background())
	if err != nil {
		logger.Get().Error().Err(err).Msg("operation failed")
		return discord.AlexError(ctx, "Failed to get items")
	}

	// Create pagination data
	event := ctx.GetEvent()
	userID := getUserID(event)
	paginationID := fmt.Sprintf("list_items_%s", userID)

	itemInterfaces := make([]any, len(items))
	for i, item := range items {
		itemInterfaces[i] = item
	}

	paginationData := &discord.PaginationData{
		Items:       itemInterfaces,
		CurrentPage: 0,
		PageSize:    discord.GetPageSize(),
		Title:       "Items",
		Description: fmt.Sprintf("All items in the game (%d total)", len(items)),
		FormatFunc:  formatItemListField,
		Color:       discord.ColorThemeGold,
	}

	discord.StorePaginationState(paginationID, paginationData)

	embed := discord.CreatePaginatedEmbed(paginationData)
	components := discord.GetPaginationComponents(paginationID, paginationData)

	return ctx.Respond(&discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{embed},
			Components: components,
		},
	})
}

func (l *List) listActiveRoles(ctx ken.SubCommandContext) (err error) {
	if err := ctx.Defer(); err != nil {
		logger.Get().Error().Err(err).Msg("operation failed")
		return err
	}

	roleGroups := [][]string{DummyGoodRoles, DummyEvilRoles, DummyNeutralRoles}

	for _, roleGroup := range roleGroups {
		slices.Sort(roleGroup)
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

func (l *List) listNotes(ctx ken.SubCommandContext) (err error) {
	if err := ctx.Defer(); err != nil {
		logger.Get().Error().Err(err).Msg("operation failed")
		return err
	}
	if !discord.IsAdminRole(ctx, discord.AdminRoles...) {
		return discord.NotAdminError(ctx)
	}
	q := models.New(l.dbPool)
	players, err := q.ListPlayer(context.Background())
	if err != nil {
		return discord.AlexError(ctx, "Failed to get players")
	}
	allNotes, err := q.ListAllPlayerNotes(context.Background())
	if err != nil {
		return discord.AlexError(ctx, "Failed to get notes")
	}
	groupedNotes := make(map[int64][]models.PlayerNote)
	for _, note := range allNotes {
		groupedNotes[note.PlayerID] = append(groupedNotes[note.PlayerID], note)
	}
	fields := []*discordgo.MessageEmbedField{}
	for playerID, notes := range groupedNotes {
		fieldName := ""
		for _, player := range players {
			if player.ID == playerID {
				discordPlayer, _ := ctx.GetSession().GuildMember(discord.BetraylGuildID, util.Itoa64(player.ID))
				fieldName = discordPlayer.DisplayName()
				break
			}
		}
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   fmt.Sprintf("%s (%d)", discord.Bold(fieldName), len(notes)),
			Value:  fmt.Sprintf("%s", discord.Code(notes[0].Info)),
			Inline: false,
		})
	}
	return ctx.RespondEmbed(&discordgo.MessageEmbed{
		Title:       "Notes",
		Description: "All notes for the game",
		Fields:      fields,
	})
}

// Version implements ken.SlashCommand.
func (*List) Version() string {
	return "1.0.0"
}

// Helper functions

func getUserID(event *discordgo.InteractionCreate) string {
	if event.Member != nil && event.Member.User != nil {
		return event.Member.User.ID
	}
	if event.User != nil {
		return event.User.ID
	}
	return "unknown"
}

func formatItemListField(item any) *discordgo.MessageEmbedField {
	itemData := item.(models.Item)
	return &discordgo.MessageEmbedField{
		Name:   fmt.Sprintf("%s (%s) $%d", itemData.Name, string(itemData.Rarity), itemData.Cost),
		Inline: true,
	}
}
