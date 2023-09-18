package commands

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/mccune1224/betrayal/internal/data"
	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/zekrotja/ken"
)

const infinity = "∞"
const black = 0x000000

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
	return debugCMD + "view"
}

// Options implements ken.SlashCommand.
func (*View) Options() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "role",
			Description: "View a role",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "name",
					Description: "Name of the role",
					Required:    true,
				},
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "ability",
			Description: "View an ability",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "name",
					Description: "Name of the ability",
					Required:    true,
				},
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "perk",
			Description: "View a perk",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "name",
					Description: "Name of the perk",
					Required:    true,
				},
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "item",
			Description: "View an item",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "name",
					Description: "Name of the item",
					Required:    true,
				},
			},
		},
	}
}

// Run implements ken.SlashCommand.

func (v *View) Run(ctx ken.Context) (err error) {
	err = ctx.HandleSubCommands(
		ken.SubCommandHandler{Name: "role", Run: v.ViewRole},
		ken.SubCommandHandler{Name: "ability", Run: v.ViewAbility},
		ken.SubCommandHandler{Name: "perk", Run: v.ViewPerk},
		ken.SubCommandHandler{Name: "item", Run: v.ViewItem},
	)
	return err
}

func (v *View) ViewRole(ctx ken.SubCommandContext) (err error) {
	if err = ctx.Defer(); err != nil {
		return err
	}
	nameArg := ctx.Options().GetByName("name").StringValue()
	role, err := v.models.Roles.GetByName(nameArg)
	if err != nil {
		ctx.SetEphemeral(true)
		ctx.RespondError(
			fmt.Sprintf("Unable to find Role: %s", nameArg),
			"Error Finding Role",
		)
		return err
	}

	abilities, err := v.models.Roles.GetAbilities(role.ID)
	if err != nil {
		ctx.SetEphemeral(true)
		ctx.RespondError(
			fmt.Sprintf("Unable to find Abilities for Role: %s", nameArg),
			"Error Finding Abilities",
		)
		return err
	}

	perks, err := v.models.Roles.GetPerks(role.ID)
	if err != nil {
		ctx.SetEphemeral(true)
		ctx.RespondError(
			fmt.Sprintf("Unable to find Perks for Role: %s", nameArg),
			"Error Finding Perks",
		)
		return err
	}

	color := 0x000000
	switch role.Alignment {
	case "GOOD":
		color = 0x00ff00
	case "EVIL":
		color = 0xff3300
	case "NEUTRAL":
		color = 0xffee00
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

	ctx.SetEphemeral(false)
	err = ctx.RespondEmbed(embed)
	return err
}

func (v *View) ViewAbility(ctx ken.SubCommandContext) (err error) {
	if err = ctx.Defer(); err != nil {
		return err
	}

	data := ctx.Options().GetByName("name").StringValue()
	ability, err := v.models.Abilities.GetByName(data)
	if err != nil {
		ctx.SetEphemeral(true)
		ctx.RespondError(
			fmt.Sprintf("Unable to find Ability: %s", data),
			"Error Finding Ability",
		)
		return err
	}

	embededFields := []*discordgo.MessageEmbedField{
		{
			Value: ability.Description,
		},
	}

	abilityTitle := ability.Name
	if !ability.AnyAbility {
		abilityTitle = fmt.Sprintf("%s [%d]", ability.Name, ability.Charges)
	}

	embed := &discordgo.MessageEmbed{
		Title:  abilityTitle,
		Color:  black,
		Fields: embededFields,
	}

	ctx.SetEphemeral(false)
	err = ctx.RespondEmbed(embed)
	return err
}

func (v *View) ViewPerk(ctx ken.SubCommandContext) (err error) {
	if err = ctx.Defer(); err != nil {
		return err
	}

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

	ctx.SetEphemeral(true)
	err = ctx.RespondEmbed(embed)
	return err
}

func (v *View) ViewItem(ctx ken.SubCommandContext) (err error) {

	if err = ctx.Defer(); err != nil {
		return err
	}

	data := ctx.Options().GetByName("name").StringValue()
	item, err := v.models.Items.GetByName(data)
	if err != nil {
		ctx.SetEphemeral(true)
		ctx.RespondError(
			fmt.Sprintf("Unable to find Item: %s", data),
			"Error Finding Item",
		)
		return err
	}

	embededFields := []*discordgo.MessageEmbedField{
		{
			Value:  item.Description,
			Inline: false,
		},
		{
			Name:   "Cost",
			Value:  fmt.Sprintf("%d", item.Cost),
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
		Color:  0x000000,
		Fields: embededFields,
	}

	ctx.SetEphemeral(false)
	err = ctx.RespondEmbed(embed)
	return err
}

// Version implements ken.SlashCommand.
func (*View) Version() string {
	return "1.0.0"
}
