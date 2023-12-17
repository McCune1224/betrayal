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
func (v *View) generateNephRole(ctx ken.Context, role *data.Role) (err error) {
  base, err := v.roleEmbed(role)
  if err != nil {
    log.Println(err)
    return discord.AlexError(ctx, fmt.Sprintf("Failed to generate embeded message for role %s", role.Name))
  }

  missing := missing(role.Name)
  firstMissing := missing[0]
  secondMissing := missing[1]


  firstMissingRole, err := v.models.Roles.GetByName(firstMissing)
  if err != nil {
    log.Println(err)
    return discord.AlexError(ctx, fmt.Sprintf("Failed to generate embeded message for role %s", firstMissing))
  }

  secondMissingRole, err := v.models.Roles.GetByName(secondMissing)
  if err != nil {
    log.Println(err)
    return discord.AlexError(ctx, fmt.Sprintf("Failed to generate embeded message for role %s", secondMissing))
  }

  missingRoles := []*data.Role{firstMissingRole, secondMissingRole}


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
