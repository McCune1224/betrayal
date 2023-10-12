package commands

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

	abilities, err := v.models.Roles.GetAbilities(role.ID)
	if err != nil {
		ctx.RespondError(
			fmt.Sprintf("Unable to find Abilities for Role: %s", nameArg),
			"Error Finding Abilities",
		)
		return err
	}

	perks, err := v.models.Roles.GetPerks(role.ID)
	if err != nil {
		ctx.RespondError(
			fmt.Sprintf("Unable to find Perks for Role: %s", nameArg),
			"Error Finding Perks",
		)
		return err
	}

	color := 0x000000
	switch role.Alignment {
	case "GOOD":
		color = discord.ColorThemeGreen
	case "EVIL":
		color = discord.ColorThemeRed
	case "NEUTRAL":
		color = discord.ColorThemeYellow
	}

	var embededAbilitiesFields []*discordgo.MessageEmbedField
	embededAbilitiesFields = append(embededAbilitiesFields, &discordgo.MessageEmbedField{
		Name:   "\n\n" + discord.Underline("Abilities") + "\n",
		Value:  "",
		Inline: false,
	})
	for _, ability := range abilities {
		title := ability.Name
		if !ability.AnyAbility {
			if ability.Charges == -1 {
				title = fmt.Sprintf("%s [%s]", ability.Name, infinity)
			} else {
				title = fmt.Sprintf("%s [%d]", ability.Name, ability.Charges)
			}
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
			Text: "Alignment: " + role.Alignment,
		},
	}

	err = ctx.RespondEmbed(embed)
	return err
}

func (v *View) viewAbility(ctx ken.SubCommandContext) (err error) {
	if err = ctx.Defer(); err != nil {
		return
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

	embededFields := []*discordgo.MessageEmbedField{
		{
			Value: ability.Description,
		},
	}

	associatedRole, err := v.models.Roles.GetByAbilityID(ability.ID)
	if err != nil {
		log.Println(err)
		discord.ErrorMessage(ctx,
			"Error Finding Role",
			fmt.Sprintf("Unable to find Associated Role for Ability: %s", nameArg))
	}

	abilityTitle := ability.Name
	if !ability.AnyAbility {
		abilityTitle = fmt.Sprintf("%s [%d]", ability.Name, ability.Charges)
	}

	embed := &discordgo.MessageEmbed{
		Title:  abilityTitle,
		Color:  determineColor(ability.Rarity),
		Fields: embededFields,
	}

	b := ctx.FollowUpEmbed(embed)
	var clearAll bool
	b.AddComponents(func(cb *ken.ComponentBuilder) {
		cb.AddActionsRow(func(b ken.ComponentAssembler) {
			b.Add(discordgo.Button{
				CustomID: "view-role",
				Style:    discordgo.DangerButton,
				Label:    fmt.Sprintf("View Role: %s", associatedRole.Name),
			}, func(ctx ken.ComponentContext) bool {
				ctx.RespondMessage(fmt.Sprintf("TODO: %s", associatedRole.Name))
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
	data := ctx.Options().GetByName("name").StringValue()
	perk, err := v.models.Perks.GetByName(data)
	if err != nil {
		ctx.RespondError("Unable to find Perk", err.Error())
		return err
	}

	embededFields := []*discordgo.MessageEmbedField{
		{
			Name:   perk.Name,
			Value:  perk.Description,
			Inline: false,
		},
	}

	embed := &discordgo.MessageEmbed{
		Title:  perk.Name,
		Color:  0x000000,
		Fields: embededFields,
	}

	err = ctx.RespondEmbed(embed)
	return err
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
