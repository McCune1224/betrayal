package commands

import (
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/mccune1224/betrayal/pkg/data"
	"github.com/zekrotja/ken"
)

type RoleGet struct {
	models data.Models
}

func (rg *RoleGet) SetModels(models data.Models) {
	rg.models = models
}

var _ ken.SlashCommand = (*RoleGet)(nil)

// Description implements ken.SlashCommand.
func (*RoleGet) Description() string {
	return "Get a role"
}

// Name implements ken.SlashCommand.
func (*RoleGet) Name() string {
	return discord.DebugCmd + "role_get"
}

// Options implements ken.SlashCommand.
func (*RoleGet) Options() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "name",
			Description: "Name of the role",
		},
	}
}

// Run implements ken.SlashCommand.
func (rg *RoleGet) Run(ctx ken.Context) (err error) {
	if err := ctx.Defer(); err != nil {
		log.Println(err)
		return err
	}
	ctx.SetEphemeral(true)
	data := ctx.Options().GetByName("name").StringValue()
	role, err := rg.models.Roles.GetByFuzzy(data)
	if err != nil {
		discord.ErrorMessage(ctx, "Unable to find Role", "No known role of name "+data)
		return err
	}

	abilities, err := rg.models.Roles.GetAbilities(role.ID)
	if err != nil {
		return discord.ErrorMessage(ctx, "Failed to Find base Abilities",
			"Unable to find Abilities for Role "+role.Name,
		)
	}

	perks, err := rg.models.Roles.GetPerks(role.ID)
	if err != nil {
		discord.ErrorMessage(ctx, "Failed to find role perks",
			"Unable to find Perks for Role "+role.Name,
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
		embededAbilitiesFields = append(
			embededAbilitiesFields,
			&discordgo.MessageEmbedField{
				Name:   ability.Name,
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

// Version implements ken.SlashCommand.
func (*RoleGet) Version() string {
	return "1.0.0"
}
