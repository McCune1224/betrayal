package view

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/mccune1224/betrayal/internal/data"
	"github.com/mccune1224/betrayal/internal/discord"
)

// Helper to build the embed for a role
// Will pull abilities and perks from the database
func (v *View) roleEmbed(role *data.Role) (*discordgo.MessageEmbed, error) {
	color := 0x000000
	switch role.Alignment {
	case "GOOD":
		color = discord.ColorThemeGreen
	case "EVIL":
		color = discord.ColorThemeRed
	case "NEUTRAL":
		color = discord.ColorThemeYellow
	}
	abilities, err := v.models.Roles.GetAbilities(role.ID)
	if err != nil {
		return nil, err
	}
	perks, err := v.models.Roles.GetPerks(role.ID)
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

		categories := strings.Join(ability.Categories, ", ")
		if ability.Charges == -1 {
			title = fmt.Sprintf(fStr, ability.Name, infinity, categories)
		} else {
			title = fmt.Sprintf(fStr, ability.Name, ability.Charges, categories)
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
	return embed, nil
}
