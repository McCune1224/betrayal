package search

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/mccune1224/betrayal/internal/logger"
	"github.com/mccune1224/betrayal/internal/models"
	"github.com/zekrotja/ken"
)

type AbilityWithRoles struct {
	Ability models.AbilityInfo
	Roles   []models.Role
}

type Search struct {
	dbPool *pgxpool.Pool
}

func (s *Search) Initialize(pool *pgxpool.Pool) {
	s.dbPool = pool
}

var _ ken.SlashCommand = (*Search)(nil)

// Name implements ken.SlashCommand.
func (*Search) Name() string {
	return "search"
}

// Description implements ken.SlashCommand.
func (*Search) Description() string {
	return "Search for abilities and items by keyword"
}

// Version implements ken.SlashCommand.
func (*Search) Version() string {
	return "1.0.0"
}

// Options implements ken.SlashCommand.
func (*Search) Options() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "ability",
			Description: "Search for abilities by keyword",
			Options: []*discordgo.ApplicationCommandOption{
				discord.StringCommandArg("keyword", "Search term for abilities", true),
				discord.BoolCommandArg("include_name", "Search in name and description (default: description only)", false),
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "item",
			Description: "Search for items by keyword",
			Options: []*discordgo.ApplicationCommandOption{
				discord.StringCommandArg("keyword", "Search term for items", true),
				discord.BoolCommandArg("include_name", "Search in name and description (default: description only)", false),
			},
		},
	}
}

// Run implements ken.SlashCommand.
func (s *Search) Run(ctx ken.Context) (err error) {
	defer logger.RecoverWithLog(*logger.Get())

	return ctx.HandleSubCommands(
		ken.SubCommandHandler{Name: "ability", Run: s.searchAbility},
		ken.SubCommandHandler{Name: "item", Run: s.searchItem},
	)
}

func (s *Search) searchAbility(ctx ken.SubCommandContext) error {
	if err := ctx.Defer(); err != nil {
		logger.Get().Error().Err(err).Msg("operation failed")
		return err
	}

	keyword := ctx.Options().GetByName("keyword").StringValue()
	includeName := false
	if nameOpt, ok := ctx.Options().GetByNameOptional("include_name"); ok {
		includeName = nameOpt.BoolValue()
	}

	q := models.New(s.dbPool)

	// Apply fuzzy matching with light tolerance for typos
	keywordLower := fuzzyKeyword(keyword)
	searchTerm := "%" + keywordLower + "%"

	var abilities []models.AbilityInfo
	var err error

	if includeName {
		abilities, err = q.SearchAbilityByKeyword(context.Background(), searchTerm)
	} else {
		abilities, err = q.SearchAbilityByDescription(context.Background(), searchTerm)
	}

	if err != nil {
		logger.Get().Error().Err(err).Msg("operation failed")
		return discord.AlexError(ctx, "Failed to search abilities")
	}

	if len(abilities) == 0 {
		return discord.AlexError(ctx, fmt.Sprintf("No abilities found matching '%s'. Try `/list abilities` for all options!", keyword))
	}

	// Fetch associated roles for each ability
	abilitiesWithRoles := make([]AbilityWithRoles, 0, len(abilities))
	for _, ability := range abilities {
		roles, err := q.ListAssociatedRolesForAbility(context.Background(), ability.ID)
		if err != nil {
			logger.Get().Error().Err(err).Msg("failed to fetch roles for ability")
			// Continue without roles if fetch fails
			roles = []models.Role{}
		}
		abilitiesWithRoles = append(abilitiesWithRoles, AbilityWithRoles{
			Ability: ability,
			Roles:   roles,
		})
	}

	// Create pagination data
	event := ctx.GetEvent()
	userID := getUserID(event)
	paginationID := fmt.Sprintf("search_ability_%s_%d", userID, time.Now().UnixNano())

	paginationData := &discord.PaginationData{
		Items:       convertAbilitiesWithRolesToInterface(abilitiesWithRoles),
		CurrentPage: 0,
		PageSize:    discord.GetPageSize(),
		Title:       fmt.Sprintf("Ability Search: %s", keyword),
		Description: fmt.Sprintf("Found %d ability/abilities matching '%s'", len(abilities), keyword),
		FormatFunc:  formatAbilityField,
		Color:       discord.ColorThemeBlue,
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

func (s *Search) searchItem(ctx ken.SubCommandContext) error {
	if err := ctx.Defer(); err != nil {
		logger.Get().Error().Err(err).Msg("operation failed")
		return err
	}

	keyword := ctx.Options().GetByName("keyword").StringValue()
	includeName := false
	if nameOpt, ok := ctx.Options().GetByNameOptional("include_name"); ok {
		includeName = nameOpt.BoolValue()
	}

	q := models.New(s.dbPool)

	// Apply fuzzy matching with light tolerance for typos
	keywordLower := fuzzyKeyword(keyword)
	searchTerm := "%" + keywordLower + "%"

	var items []models.Item
	var err error

	if includeName {
		items, err = q.SearchItemByKeyword(context.Background(), searchTerm)
	} else {
		items, err = q.SearchItemByDescription(context.Background(), searchTerm)
	}

	if err != nil {
		logger.Get().Error().Err(err).Msg("operation failed")
		return discord.AlexError(ctx, "Failed to search items")
	}

	if len(items) == 0 {
		return discord.AlexError(ctx, fmt.Sprintf("No items found matching '%s'. Try `/list items` for all options!", keyword))
	}

	// Create pagination data
	event := ctx.GetEvent()
	userID := getUserID(event)
	paginationID := fmt.Sprintf("search_item_%s_%d", userID, time.Now().UnixNano())
	paginationData := &discord.PaginationData{
		Items:       convertItemsToInterface(items),
		CurrentPage: 0,
		PageSize:    discord.GetPageSize(),
		Title:       fmt.Sprintf("Item Search: %s", keyword),
		Description: fmt.Sprintf("Found %d item/items matching '%s'", len(items), keyword),
		FormatFunc:  formatItemField,
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

func fuzzyKeyword(keyword string) string {
	// Simple fuzzy matching - just convert to lowercase and trim
	// For more sophisticated fuzzy matching, could use Levenshtein distance here
	return strings.TrimSpace(strings.ToLower(keyword))
}

func formatAbilityField(item any) *discordgo.MessageEmbedField {
	abilityWithRoles := item.(AbilityWithRoles)
	ability := abilityWithRoles.Ability

	// Build name with asterisk for non-role-specific abilities (potential any-abilities)
	abilityName := ability.Name
	if ability.Rarity != models.RarityROLESPECIFIC {
		abilityName = ability.Name + " *"
	}

	// Build roles string
	rolesStr := ""
	if len(abilityWithRoles.Roles) > 0 {
		roleNames := make([]string, 0, len(abilityWithRoles.Roles))
		for _, role := range abilityWithRoles.Roles {
			roleNames = append(roleNames, role.Name)
		}
		rolesStr = fmt.Sprintf("\n**Roles:** %s", strings.Join(roleNames, ", "))
	} else {
		rolesStr = "\n**Roles:** Any"
	}

	return &discordgo.MessageEmbedField{
		Name:  fmt.Sprintf("%s [%s]", abilityName, ability.Rarity),
		Value: fmt.Sprintf("%s\n*Default Charges: %d*%s", ability.Description, ability.DefaultCharges, rolesStr),
	}
}

func formatItemField(item any) *discordgo.MessageEmbedField {
	itemData := item.(models.Item)
	return &discordgo.MessageEmbedField{
		Name:  fmt.Sprintf("%s [%s] $%d", itemData.Name, itemData.Rarity, itemData.Cost),
		Value: itemData.Description,
	}
}

func convertItemsToInterface(items []models.Item) []any {
	result := make([]any, len(items))
	for i, item := range items {
		result[i] = item
	}
	return result
}

func convertAbilitiesWithRolesToInterface(abilitiesWithRoles []AbilityWithRoles) []any {
	result := make([]any, len(abilitiesWithRoles))
	for i, awr := range abilitiesWithRoles {
		result[i] = awr
	}
	return result
}
