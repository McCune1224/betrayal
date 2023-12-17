package view

import (
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/mccune1224/betrayal/internal/data"
	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/mccune1224/betrayal/internal/scheduler"
	"github.com/zekrotja/ken"
)

const infinity = "∞"

type View struct {
	models    data.Models
	scheduler *scheduler.BetrayalScheduler
}

func (v *View) Initialize(models data.Models, scheduler *scheduler.BetrayalScheduler) {
	v.models = models
	v.scheduler = scheduler
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
	statuses, err := v.models.Statuses.GetAll()
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
	role, err := v.models.Roles.GetByFuzzy(nameArg)
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
	nameArg := ctx.Options().GetByName("name").StringValue()
	ability, err := v.models.Abilities.GetByFuzzy(nameArg)
	if err != nil {
		discord.ErrorMessage(ctx,
			"Error Finding Ability",
			fmt.Sprintf("Unable to find Ability: %s", nameArg),
		)
		return err
	}

	associatedRoles, err := v.models.Roles.GetAllByAbilityID(ability.ID)
	if err != nil {
		log.Println(err)
		return discord.ErrorMessage(ctx,
			"Error Finding Role",
			fmt.Sprintf("Unable to find Associated Role for Ability: %s", nameArg))
	}

	abilityEmbed := &discordgo.MessageEmbed{
		Title:       ability.Name,
		Description: ability.Description,
		Color:       determineColor(ability.Rarity),
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Categories",
				Value:  strings.Join(ability.Categories, ", "),
				Inline: true,
			},
		},
	}
	aa, _ := v.models.Abilities.GetAnyAbilitybyFuzzy(ability.Name)
	if aa != nil {
		msg := ""
		if aa.RoleSpecific != "" {
			msg = fmt.Sprintf("Role Specific AA - %s", aa.RoleSpecific)
		} else {
			msg = fmt.Sprintf("%s AA", aa.Rarity)
		}
		abilityEmbed.Footer = &discordgo.MessageEmbedFooter{
			Text: msg,
		}
	} else {
		abilityEmbed.Footer = &discordgo.MessageEmbedFooter{
			Text: "Only base ability, not an AA",
		}
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
					role, _ := v.models.Roles.GetByFuzzy(roleName)
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
	nameArg := ctx.Options().GetByName("name").StringValue()
	perk, err := v.models.Perks.GetByFuzzy(nameArg)
	if err != nil {
		ctx.RespondError("Unable to find Perk",
			fmt.Sprintf("Unable to find Perk: %s", nameArg),
		)
		return err
	}

	associatedRoles, err := v.models.Roles.GetAllByPerkID(perk)
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
					role, _ := v.models.Roles.GetByFuzzy(roleName)
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
	item, err := v.models.Items.GetByFuzzy(data)
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
			Value:  item.Rarity,
			Inline: true,
		},
	}

	embed := &discordgo.MessageEmbed{
		Title:  item.Name,
		Color:  determineColor(item.Rarity),
		Fields: embededFields,
	}
	if item.Name == "Zingy" {
		// attach zingy image to embed
		embed.Thumbnail = &discordgo.MessageEmbedThumbnail{
			URL: "https://www.fairy-tales-inc.com/images/thumbs/0058033_bashful-zingy-bunny-medium-by-jellycat_550.jpeg",
		}
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
	status, err := v.models.Statuses.GetByName(statusName)
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

func determineColor(rarity string) int {
	rarity = strings.ToUpper(rarity)
	switch rarity {
	case "COMMON":
		return discord.ColorItemCommon
	case "UNCOMMON":
		return discord.ColorItemUncommon
	case "RARE":
		return discord.ColorItemRare
	case "EPIC":
		return discord.ColorItemEpic
	case "LEGENDARY":
		return discord.ColorItemLegendary
	case "MYTHICAL":
		return discord.ColorItemMythical
	case "UNIQUE":
		return discord.ColorItemUnique

	default:
		return discord.ColorThemeWhite
	}
}
