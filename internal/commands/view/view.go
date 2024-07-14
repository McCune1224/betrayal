package view

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/mccune1224/betrayal/internal/models"
	"github.com/zekrotja/ken"
)

const infinity = "∞"

type View struct {
	dbPool *pgxpool.Pool
}

func (v *View) Initialize(pool *pgxpool.Pool) {
	v.dbPool = pool
}

var _ ken.SlashCommand = (*View)(nil)

// Description implements ken.SlashCommand.
func (*View) Description() string {
	return "View specified item"
}

// Name implements ken.SlashCommand.
func (*View) Name() string {
	return discord.DebugCmd + "view"
}

// Options implements ken.SlashCommand.
func (v *View) Options() []*discordgo.ApplicationCommandOption {
	statusChoices := []*discordgo.ApplicationCommandOptionChoice{}
	q := models.New(v.dbPool)
	dbCtx := context.Background()
	statuses, err := q.ListStatus(dbCtx)
	if err != nil {
		log.Println(err)
		return nil
	}
	for _, status := range statuses {
		statusChoices = append(statusChoices, &discordgo.ApplicationCommandOptionChoice{
			Name:  status.Name,
			Value: status.Name,
		})
	}
	return []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "role",
			Description: "View a role",
			Options: []*discordgo.ApplicationCommandOption{
				discord.StringCommandArg("name", "Name of the role", true),
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "ability",
			Description: "View an ability",
			Options: []*discordgo.ApplicationCommandOption{
				discord.StringCommandArg("name", "Name of the role", true),
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "perk",
			Description: "View a perk",
			Options: []*discordgo.ApplicationCommandOption{
				discord.StringCommandArg("name", "Name of the role", true),
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "item",
			Description: "View an item",
			Options: []*discordgo.ApplicationCommandOption{
				discord.StringCommandArg("name", "Name of the role", true),
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "status",
			Description: "view a status",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "name",
					Description: "Name of the status",
					Required:    true,
					Choices:     statusChoices,
				},
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "duel",
			Description: "View duel mini-game details",
		},
	}
}

// Run implements ken.SlashCommand.

func (v *View) Run(ctx ken.Context) (err error) {
	err = ctx.HandleSubCommands(
		ken.SubCommandHandler{Name: "role", Run: v.viewRole},
		ken.SubCommandHandler{Name: "ability", Run: v.viewAbility},
		ken.SubCommandHandler{Name: "perk", Run: v.viewPerk},
		ken.SubCommandHandler{Name: "item", Run: v.viewItem},
		ken.SubCommandHandler{Name: "status", Run: v.viewStatus},
		ken.SubCommandHandler{Name: "duel", Run: v.viewDuel},
	)
	return err
}

func (v *View) viewRole(ctx ken.SubCommandContext) (err error) {
	if err = ctx.Defer(); err != nil {
		log.Println(err)
		return err
	}

	nameArg := ctx.Options().GetByName("name").StringValue()
	// haha funny easter egg
	if strings.ToLower(nameArg) == "ferrari" {
		return ctx.RespondEmbed(generateFerrariRole())
	}

	q := models.New(v.dbPool)
	dbCtx := context.Background()
	role, err := q.GetRoleByFuzzy(dbCtx, nameArg)
	if err != nil {
		ctx.RespondError(
			fmt.Sprintf("Unable to find Role: %s", nameArg),
			"Error Finding Role",
		)
		return err
	}

	if role.Name == "Nephilim" || role.Name == "Nephilism - Defensive" || role.Name == "Nephilism - Offensive" {
		return v.generateNephRole(ctx, role)
	}

	roleEmbed, err := v.roleEmbed(role)
	if err != nil {
		log.Println(err)
		return discord.ErrorMessage(ctx,
			"Failed to Get Full Role Details",
			"Was not able to pull abilities and perk details for view.",
		)
	}

	return ctx.RespondEmbed(roleEmbed)
}

func (v *View) viewAbility(ctx ken.SubCommandContext) (err error) {
	if err = ctx.Defer(); err != nil {
		log.Println(err)
		return err
	}
	q := models.New(v.dbPool)
	dbCtx := context.Background()
	nameArg := ctx.Options().GetByName("name").StringValue()
	// ability, err := v.models.Abilities.GetByFuzzy(nameArg)
	ability, err := q.GetAbilityInfoByFuzzy(dbCtx, nameArg)
	if err != nil {
		discord.ErrorMessage(ctx,
			"Error Finding Ability",
			fmt.Sprintf("Unable to find Ability: %s", nameArg),
		)
		return err
	}

	associatedRoles, err := q.ListAssociatedRolesForAbility(dbCtx, ability.ID)
	if err != nil {
		log.Println(err)
		return discord.ErrorMessage(ctx,
			"Error Finding Role",
			fmt.Sprintf("Unable to find Associated Role for Ability: %s", nameArg))
	}

	dbcategories, _ := q.ListAbilityCategoryNames(dbCtx, ability.ID)
	abilityEmbed := &discordgo.MessageEmbed{
		Title:       ability.Name,
		Description: ability.Description,
		Color:       determineColor(ability.Rarity),
		// FIXME: Categories need to be queried/overhauled
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Categories",
				Value:  strings.Join(dbcategories, ", "),
				Inline: true,
			},
		},
	}
	aa, _ := q.GetAnyAbilityByFuzzy(dbCtx, ability.Name)
	msg := ""
	if aa.Rarity == models.RarityROLESPECIFIC {
		msg = fmt.Sprintf("Role Specific AA")
	} else {
		msg = fmt.Sprintf("%s AA", aa.Rarity)
	}
	abilityEmbed.Footer = &discordgo.MessageEmbedFooter{
		Text: msg,
	}

	b := ctx.FollowUpEmbed(abilityEmbed)

	b.AddComponents(func(cb *ken.ComponentBuilder) {
		cb.AddActionsRow(func(b ken.ComponentAssembler) {
			for _, associatedRole := range associatedRoles {
				b.Add(discordgo.Button{
					CustomID: fmt.Sprintf("%s-%s", associatedRole.Name, ability.Name),
					Style:    discordgo.PrimaryButton,
					Label:    associatedRole.Name,
				}, func(ctx ken.ComponentContext) bool {
					roleName := strings.Split(ctx.GetData().CustomID, "-")[0]
					// We know for sure role exists here so ignore error
					role, _ := q.GetRoleByFuzzy(dbCtx, roleName)
					roleEmbed, err := v.roleEmbed(role)
					if err != nil {
						log.Println(err)
						ctx.RespondError("Failed to Get Full Role Details", "Was not able to pull abilities and perk details for view.")
					}

					ctx.SetEphemeral(true)
					ctx.RespondEmbed(roleEmbed)
					return true
				}, false)
			}
		}, false).Condition(func(cctx ken.ComponentContext) bool {
			return true
		})
	})

	fum := b.Send()
	return fum.Error
}

func (v *View) viewPerk(ctx ken.SubCommandContext) (err error) {
	if err = ctx.Defer(); err != nil {
		return err
	}
	dbCtx := context.Background()
	q := models.New(v.dbPool)

	nameArg := ctx.Options().GetByName("name").StringValue()
	perk, err := q.GetPerkInfoByFuzzy(dbCtx, nameArg)
	if err != nil {
		ctx.RespondError("Unable to find Perk",
			fmt.Sprintf("Unable to find Perk: %s", nameArg),
		)
		return err
	}

	associatedRoles, err := q.ListAssociatedRolesForPerk(dbCtx, perk.ID)
	if err != nil {
		log.Println(err)
		discord.ErrorMessage(ctx,
			"Error Finding Role",
			fmt.Sprintf("Unable to find Associated Role for Ability: %s", nameArg))
		return err
	}

	perkEmbed := &discordgo.MessageEmbed{
		Title:       perk.Name,
		Description: perk.Description,
		Color:       discord.ColorThemeWhite,
	}

	b := ctx.FollowUpEmbed(perkEmbed)
	b.AddComponents(func(cb *ken.ComponentBuilder) {
		cb.AddActionsRow(func(b ken.ComponentAssembler) {
			for _, associatedRole := range associatedRoles {
				b.Add(discordgo.Button{
					CustomID: fmt.Sprintf("%s-%s", associatedRole.Name, perk.Name),
					Style:    discordgo.PrimaryButton,
					Label:    associatedRole.Name,
				}, func(ctx ken.ComponentContext) bool {
					roleName := strings.Split(ctx.GetData().CustomID, "-")[0]
					// We know for sure role exists here so ignore error
					role, _ := q.GetRoleByFuzzy(dbCtx, roleName)
					roleEmbed, err := v.roleEmbed(role)
					if err != nil {
						log.Println(err)
						ctx.RespondError("Failed to Get Full Role Details", "Was not able to pull abilities and perk details for view.")
					}

					ctx.SetEphemeral(true)
					ctx.RespondEmbed(roleEmbed)
					return true
				}, false)
			}
		}, false).Condition(func(cctx ken.ComponentContext) bool {
			return true
		})
	})

	fum := b.Send()
	return fum.Error
}

func (v *View) viewItem(ctx ken.SubCommandContext) (err error) {
	data := ctx.Options().GetByName("name").StringValue()
	dbCtx := context.Background()
	q := models.New(v.dbPool)
	item, err := q.GetItemByFuzzy(dbCtx, data)
	if err != nil {
		discord.ErrorMessage(ctx,
			"Unable to find Item",
			fmt.Sprintf("Unable to find Item: %s", data),
		)
		return err
	}

	itemCostStr := ""
	if item.Cost == 0 {
		itemCostStr = infinity
	} else {
		itemCostStr = fmt.Sprintf("%d", item.Cost)
	}

	embededFields := []*discordgo.MessageEmbedField{
		{
			Value:  item.Description,
			Inline: false,
		},
		{
			Name:   "Cost",
			Value:  itemCostStr,
			Inline: true,
		},
		{
			Name:   "Rarity",
			Value:  string(item.Rarity),
			Inline: true,
		},
	}

	embed := &discordgo.MessageEmbed{
		Title:  item.Name,
		Color:  determineColor(item.Rarity),
		Fields: embededFields,
	}
	if item.Name == "Zingy" {
		return zingyCase(ctx, item)
	}

	if item.Cost == 0 {
		embed.Footer = &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("%s this item is not purchasable %s", discord.EmojiWarning, discord.EmojiWarning),
		}
	}
	if item.Rarity == "Unique" {
		embed.Footer = &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("%s this item is not purchasable nor obtainable from random event (like item rain) %s", discord.EmojiWarning, discord.EmojiWarning),
		}
	}

	err = ctx.RespondEmbed(embed)
	return err
}

func (v *View) viewStatus(ctx ken.SubCommandContext) (err error) {
	statusName := ctx.Options().GetByName("name").StringValue()
	q := models.New(v.dbPool)
	dbCtx := context.Background()
	status, err := q.GetStatusByFuzzy(dbCtx, statusName)
	if err != nil {
		return err
	}

	return ctx.RespondEmbed(&discordgo.MessageEmbed{
		Title:       status.Name,
		Description: status.Description,
	})
}

func (v *View) viewDuel(ctx ken.SubCommandContext) (err error) {
	gameText := []string{
		fmt.Sprintf("In %s players will present one out of nine number tiles and the player who presented the higher numbered tile wins.", discord.Bold("Black and White")),
		fmt.Sprintf("The players will each receive 9 number tiles from 0 to 8. The 9 tiles are divided into black and white colors. %s", discord.Bold("Even numbers 0, 2, 4, 6 and 8 are black. Odd numbers 1, 3, 5 and 7 are white.\n")),
		fmt.Sprintf("The starting player will first choose a number from 0 to 8 (selecting the number in their confessional), The host will announce publicly %s. The following player will then present their tile. Only hosts will see numbers used, and the player who put a higher number will win and gain one point. %s.", discord.Bold("what color was used"), discord.Bold("Used numbers will not be revealed even after the results are announced")),
		"Example: Sophia begins the game and uses a 3. The host will announce: Sophia has used a white tile. Lindsey will place a black tile, a 0. Host will announce a black tile was used. Host will announce that Sophia has won. Both tiles/numbers are taken away and a new round begins, the winner goes first in presenting the tile for the next round. Lindsey can infer very little from her loss because any white tile can beat a black 0, but Sophia will know that she used either a 0 or a 2 based on her win.",
		"The player with more points after 9th round will win, the loser will be eliminated.",
	}

	return ctx.RespondEmbed(&discordgo.MessageEmbed{
		Title:       "Game Duel - Black and White",
		Color:       discord.ColorThemePearl,
		Description: gameText[0],
		Fields: []*discordgo.MessageEmbedField{
			{
				Value: gameText[1],
			},
			{
				Value: gameText[2],
			},
			{
				Value: gameText[3],
			},
			{
				Value: gameText[4],
			},
		},
	})
}

// Version implements ken.SlashCommand.
func (*View) Version() string {
	return "1.0.0"
}

func determineColor(rarity models.Rarity) int {
	switch rarity {
	case models.RarityCOMMON:
		return discord.ColorItemCommon
	case models.RarityUNCOMMON:
		return discord.ColorItemUncommon
	case models.RarityRARE:
		return discord.ColorItemRare
	case models.RarityEPIC:
		return discord.ColorItemEpic
	case models.RarityLEGENDARY:
		return discord.ColorItemLegendary
	case models.RarityMYTHICAL:
		return discord.ColorItemMythical
	case models.RarityUNIQUE:
		return discord.ColorItemUnique

	default:
		return discord.ColorThemeWhite
	}
}

func zingyCase(ctx ken.SubCommandContext, zingy models.Item) (err error) {
	if err := ctx.Defer(); err != nil {
		log.Println(err)
		return err
	}
	// attach zingy image to embed

	// split zingy message into 2 based of \n
	split := strings.Split(zingy.Description, "\n")

	initMsg := &discordgo.MessageEmbed{
		Title:       zingy.Name,
		Description: split[0],
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: "https://www.fairy-tales-inc.com/images/thumbs/0058033_bashful-zingy-bunny-medium-by-jellycat_550.jpeg",
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("%s this item is not purchasable nor obtainable from random event (like item rain) %s", discord.EmojiWarning, discord.EmojiWarning),
		},
	}

	secondaryMsg := &discordgo.MessageEmbed{
		Title:       zingy.Name,
		Description: strings.Join(split[1:], "\n"),
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: "https://www.fairy-tales-inc.com/images/thumbs/0058033_bashful-zingy-bunny-medium-by-jellycat_550.jpeg",
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("%s this item is not purchasable nor obtainable from random event (like item rain) %s", discord.EmojiWarning, discord.EmojiWarning),
		},
	}

	b := ctx.FollowUpEmbed(initMsg)

	b.AddComponents(func(cb *ken.ComponentBuilder) {
		cb.AddActionsRow(func(b ken.ComponentAssembler) {
			b.Add(discordgo.Button{
				CustomID: "outcomes",
				Style:    discordgo.PrimaryButton,
				Label:    "Roll Outcomes",
			}, func(ctx ken.ComponentContext) bool {
				ctx.SetEphemeral(true)
				ctx.RespondEmbed(secondaryMsg)
				return true
			}, false)
		}, false).Condition(func(cctx ken.ComponentContext) bool {
			return true
		})
	})

	return b.Send().Error
}

// Helper to build the embed for a role
// Will pull abilities and perks from the database
func (v *View) roleEmbed(role models.Role) (*discordgo.MessageEmbed, error) {
	color := 0x000000
	switch role.Alignment {
	case models.AlignmentGOOD:
		color = discord.ColorThemeGreen
	case models.AlignmentEVIL:
		color = discord.ColorThemeRed
	case models.AlignmentNEUTRAL:
		color = discord.ColorThemeYellow
	}
	q := models.New(v.dbPool)
	dbCtx := context.Background()
	abilities, err := q.ListRoleAbilityForRole(dbCtx, role.ID)
	if err != nil {
		return nil, err
	}

	perks, err := q.ListRolePerkForRole(dbCtx, role.ID)
	if err != nil {
		return nil, err
	}

	var embededAbilitiesFields []*discordgo.MessageEmbedField
	embededAbilitiesFields = append(embededAbilitiesFields, &discordgo.MessageEmbedField{
		Name:   "\n\n" + discord.Underline("Abilities") + "\n",
		Value:  "",
		Inline: false,
	})
	for _, ability := range abilities {
		title := ability.Name
		fStr := "%s [%d] - %s"
		dbcategories, _ := q.ListAbilityCategoryNames(dbCtx, ability.ID)
		categories := strings.Join(dbcategories, ", ")
		if ability.DefaultCharges == -1 {
			// title = fmt.Sprintf("%s [%s]", ability.Name, infinity)
			title = fmt.Sprintf("%s [%s] - %s", ability.Name, infinity, categories)
		} else {
			title = fmt.Sprintf(fStr, ability.Name, ability.DefaultCharges, categories)
			// title = fmt.Sprintf(fStr, ability.Name, ability.DefaultCharges)
		}
		embededAbilitiesFields = append(
			embededAbilitiesFields,
			&discordgo.MessageEmbedField{
				Name:   title,
				Value:  ability.Description,
				Inline: false,
			},
		)
	}
	embededAbilitiesFields = append(embededAbilitiesFields, &discordgo.MessageEmbedField{
		Name:  "\n\n",
		Value: "\n",
	})

	var embededPerksFields []*discordgo.MessageEmbedField
	embededAbilitiesFields = append(embededAbilitiesFields, &discordgo.MessageEmbedField{
		Name:   discord.Underline("Perks"),
		Value:  "",
		Inline: false,
	})

	for _, perk := range perks {
		embededPerksFields = append(
			embededPerksFields,
			&discordgo.MessageEmbedField{
				Name:   perk.Name,
				Value:  perk.Description + "\n",
				Inline: false,
			},
		)
	}

	embed := &discordgo.MessageEmbed{
		Title:       role.Name,
		Description: role.Description,
		Color:       color,
		Fields:      append(embededAbilitiesFields, embededPerksFields...),
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Alignment: " + string(role.Alignment),
		},
	}
	return embed, nil
}

// Given three options, return the two that are not the given role
func missing(role string) []string {
	options := []string{"Nephilim", "Nephilim - Offensive", "Nephilim - Defensive"}
	missing := []string{}
	for _, option := range options {
		if option != role {
			missing = append(missing, option)
		}
	}
	return missing
}

// Outlier role that has stances to it (aka 3 roles in one).
// Really just need to attach button components to this to pull up the other two roles
func (v *View) generateNephRole(ctx ken.Context, role models.Role) (err error) {
	base, err := v.roleEmbed(role)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, fmt.Sprintf("Failed to generate embeded message for role %s", role.Name))
	}

	missing := missing(role.Name)
	firstMissing := missing[0]
	secondMissing := missing[1]
	q := models.New(v.dbPool)
	dbCtx := context.Background()

	firstMissingRole, err := q.GetRoleByFuzzy(dbCtx, firstMissing)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, fmt.Sprintf("Failed to generate embeded message for role %s", firstMissing))
	}

	secondMissingRole, err := q.GetRoleByName(dbCtx, secondMissing)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, fmt.Sprintf("Failed to generate embeded message for role %s", secondMissing))
	}

	missingRoles := []models.Role{firstMissingRole, secondMissingRole}

	b := ctx.FollowUpEmbed(base)
	b.AddComponents(func(cb *ken.ComponentBuilder) {
		cb.AddActionsRow(func(b ken.ComponentAssembler) {
			b.Add(discordgo.Button{
				CustomID: missingRoles[0].Name,
				Style:    discordgo.PrimaryButton,
				Label:    missingRoles[0].Name,
			}, func(ctx ken.ComponentContext) bool {
				roleEmbed, err := v.roleEmbed(missingRoles[0])
				if err != nil {
					ctx.RespondMessage("Idek neph is stupid to format lol xd")
					return true
				}
				ctx.SetEphemeral(true)
				ctx.RespondEmbed(roleEmbed)
				return true
			}, false)
			b.Add(discordgo.Button{
				CustomID: missingRoles[1].Name,
				Style:    discordgo.PrimaryButton,
				Label:    missingRoles[1].Name,
			}, func(ctx ken.ComponentContext) bool {
				roleEmbed, err := v.roleEmbed(missingRoles[1])
				if err != nil {
					ctx.RespondMessage("Idek neph is stupid to format lol xd")
					return true
				}
				ctx.SetEphemeral(true)
				ctx.RespondEmbed(roleEmbed)
				return true
			}, false)
		}, false).Condition(func(cctx ken.ComponentContext) bool {
			return true
		})
	})

	fum := b.Send()
	return fum.Error
}

// Joke role
func generateFerrariRole() *discordgo.MessageEmbed {
	// -EVIL
	// Ferrari
	// We are checking...
	//
	// Abilities:
	// Perfect Strategy (x3) - Roll a d10, if 2-10, you crash into the wall. If you roll a 1, you can still play the game.
	//
	// Ignore Race Engineer (x1) - Explicitly ignore your engineer and strategist and attempt to win the race. If you are not the most voted at the next elimination, gain immunity for the one following it. If you are the most voted, both you and the other Ferrari driver will be perma-dead and cannot be revived.
	//
	// Perks:
	// Tifosi tears - You are despaired in game as well as irl, this cannot be cured.
	//
	// Forced Contract - At game start, another player will be informed they are Ferrari's second driver. They will likely despise you for this as they have no real benefits, and only pain associated with this.
	//
	// Mattia's curse - You will be told special information once per day phase from the hosts, this information will either be unhelpful or a lie.
	var embededAbilitiesFields []*discordgo.MessageEmbedField
	embededAbilitiesFields = append(embededAbilitiesFields, &discordgo.MessageEmbedField{
		Name:   "\n\n" + discord.Underline("Abilities") + "\n",
		Value:  "",
		Inline: false,
	})
	// Im too tired for this shit
	embededAbilitiesFields = append(embededAbilitiesFields, &discordgo.MessageEmbedField{
		Name:   "Perfect Strategy [3]",
		Value:  "Roll a d10, if 2-10, you crash into the wall. If you roll a 1, you can still play the game.",
		Inline: false,
	})
	embededAbilitiesFields = append(embededAbilitiesFields, &discordgo.MessageEmbedField{
		Name:   "Ignore Race Engineer [1]",
		Value:  "Explicitly ignore your engineer and strategist and attempt to win the race. If you are not the most voted at the next elimination, gain immunity for the one following it. If you are the most voted, both you and the other Ferrari driver will be perma-dead and cannot be revived.",
		Inline: false,
	})
	embededAbilitiesFields = append(embededAbilitiesFields, &discordgo.MessageEmbedField{
		Name:  "\n\n",
		Value: "\n",
	})

	var embededPerksFields []*discordgo.MessageEmbedField
	embededAbilitiesFields = append(embededAbilitiesFields, &discordgo.MessageEmbedField{
		Name:   discord.Underline("Perks"),
		Value:  "",
		Inline: false,
	})

	embededPerksFields = append(embededPerksFields, &discordgo.MessageEmbedField{
		Name:   "Tifosi tears",
		Value:  "You are despaired in game as well as irl, this cannot be cured.",
		Inline: false,
	})
	embededPerksFields = append(embededPerksFields, &discordgo.MessageEmbedField{
		Name:   "Forced Contract",
		Value:  "At game start, another player will be informed they are Ferrari's second driver. They will likely despise you for this as they have no real benefits, and only pain associated with this.",
		Inline: false,
	})
	embededPerksFields = append(embededPerksFields, &discordgo.MessageEmbedField{
		Name:   "Mattia's curse",
		Value:  "You will be told special information once per day phase from the hosts, this information will either be unhelpful or a lie.",
		Inline: false,
	})

	embed := &discordgo.MessageEmbed{
		Title:       "Ferrari",
		Description: "We are checking...",
		Color:       discord.ColorThemeRed,
		Fields:      append(embededAbilitiesFields, embededPerksFields...),
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Alignment: EVIL    %s not actual role...%s.", discord.EmojiError, discord.EmojiError),
		},
	}
	return embed
}
