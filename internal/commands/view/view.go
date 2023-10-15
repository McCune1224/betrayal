package view

import (
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/mccune1224/betrayal/internal/data"
	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/zekrotja/ken"
)

const infinity = "∞"

type View struct {
	models data.Models
}

func (v *View) SetModels(models data.Models) {
	v.models = models
}

var (
	_ ken.SlashCommand = (*View)(nil)
)

// Description implements ken.SlashCommand.
func (*View) Description() string {
	return "View specified item"
}

// Name implements ken.SlashCommand.
func (*View) Name() string {
	return discord.DebugCmd + "view"
}

// Options implements ken.SlashCommand.
func (*View) Options() []*discordgo.ApplicationCommandOption {
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
	}
}

// Run implements ken.SlashCommand.

func (v *View) Run(ctx ken.Context) (err error) {
	err = ctx.HandleSubCommands(
		ken.SubCommandHandler{Name: "role", Run: v.viewRole},
		ken.SubCommandHandler{Name: "ability", Run: v.viewAbility},
		ken.SubCommandHandler{Name: "perk", Run: v.viewPerk},
		ken.SubCommandHandler{Name: "item", Run: v.viewItem},
	)
	return err
}

func (v *View) viewRole(ctx ken.SubCommandContext) (err error) {
	nameArg := ctx.Options().GetByName("name").StringValue()
	role, err := v.models.Roles.GetByName(nameArg)
	if err != nil {
		ctx.RespondError(
			fmt.Sprintf("Unable to find Role: %s", nameArg),
			"Error Finding Role",
		)
		return err
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
		return err
	}
	nameArg := ctx.Options().GetByName("name").StringValue()
	ability, err := v.models.Abilities.GetByName(nameArg)
	if err != nil {
		discord.SuccessfulMessage(ctx,
			"Error Finding Ability",
			fmt.Sprintf("Unable to find Ability: %s", nameArg),
		)
		return err
	}

	associatedRole, err := v.models.Roles.GetByAbilityID(ability.ID)
	if err != nil {
		log.Println(err)
		return discord.ErrorMessage(ctx,
			"Error Finding Role",
			fmt.Sprintf("Unable to find Associated Role for Ability: %s", nameArg))
	}

	title := ability.Name
	fStr := "%s [%d] - %s"
	categories := strings.Join(ability.Categories, ", ")
	if ability.Charges == -1 {
		title = fmt.Sprintf(fStr, ability.Name, infinity, categories)
	} else {
		title = fmt.Sprintf(fStr, ability.Name, ability.Charges, categories)
	}

	embed := &discordgo.MessageEmbed{
		Title:       title,
		Description: ability.Description,
		Color:       determineColor(ability.Rarity),
	}

	b := ctx.FollowUpEmbed(embed)
	var clearAll bool

	b.AddComponents(func(cb *ken.ComponentBuilder) {
		cb.AddActionsRow(func(b ken.ComponentAssembler) {
			b.Add(discordgo.Button{
				CustomID: "ability-view",
				Style:    discordgo.PrimaryButton,
				Label:    fmt.Sprintf("%s", associatedRole.Name),
			}, func(ctx ken.ComponentContext) bool {
				roleEmbed, err := v.roleEmbed(associatedRole)
				if err != nil {
					log.Println(err)
					ctx.RespondError(
						"Error Finding Role",
						fmt.Sprintf("Unable to find Role: %s", associatedRole.Name),
					)
					return false
				}
				ctx.RespondEmbed(roleEmbed)
				return true
			}, !clearAll)
		}, clearAll).
			Condition(func(cctx ken.ComponentContext) bool {
				return cctx.User().ID == ctx.User().ID
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
	perk, err := v.models.Perks.GetByName(nameArg)
	if err != nil {
		ctx.RespondError("Unable to find Perk",
			fmt.Sprintf("Unable to find Perk: %s", nameArg),
		)
		return err
	}

	associatedRole, err := v.models.Roles.GetByPerkID(perk.ID)
	if err != nil {
		log.Println(err)
		discord.ErrorMessage(ctx,
			"Error Finding Role",
			fmt.Sprintf("Unable to find Associated Role for Ability: %s", nameArg))
		return err
	}

	embed := &discordgo.MessageEmbed{
		Title:  perk.Name,
		Color:  discord.ColorThemeWhite,
		Fields: []*discordgo.MessageEmbedField{{Value: perk.Description}},
	}

	b := ctx.FollowUpEmbed(embed)
	var clearAll bool

	b.AddComponents(func(cb *ken.ComponentBuilder) {
		cb.AddActionsRow(func(b ken.ComponentAssembler) {
			b.Add(discordgo.Button{
				CustomID: "perk-view",
				Style:    discordgo.PrimaryButton,
				Label:    fmt.Sprintf("%s", associatedRole.Name),
			}, func(ctx ken.ComponentContext) bool {
				roleEmbed, err := v.roleEmbed(associatedRole)
				if err != nil {
					log.Println(err)
					ctx.RespondError(
						"Error Finding Role",
						fmt.Sprintf("Unable to find Role: %s", associatedRole.Name),
					)
					return false
				}
				ctx.RespondEmbed(roleEmbed)
				return true
			}, !clearAll)
		}, clearAll).
			Condition(func(cctx ken.ComponentContext) bool {
				return cctx.User().ID == ctx.User().ID
			})
	})

	fum := b.Send()
	return fum.Error
}

func (v *View) viewItem(ctx ken.SubCommandContext) (err error) {

	data := ctx.Options().GetByName("name").StringValue()
	item, err := v.models.Items.GetByName(data)
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

	err = ctx.RespondEmbed(embed)
	return err
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
