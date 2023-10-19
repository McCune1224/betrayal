package luck

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/mccune1224/betrayal/internal/commands/inventory"
	"github.com/mccune1224/betrayal/internal/data"
	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/zekrotja/ken"
)

type Luck struct {
	models data.Models
}

var _ ken.SlashCommand = (*Luck)(nil)

// Description implements ken.SlashCommand.
func (*Luck) Description() string {
	return "Determine luck for a given level"
}

// Name implements ken.SlashCommand.
func (*Luck) Name() string {
	return discord.DebugCmd + "roll"
}

// Options implements ken.SlashCommand.
func (*Luck) Options() []*discordgo.ApplicationCommandOption {
	targetTypes := []string{"item", "ability"}

	options := []*discordgo.ApplicationCommandOptionChoice{}
	for _, t := range targetTypes {
		options = append(options, &discordgo.ApplicationCommandOptionChoice{
			Name:  t,
			Value: t,
		})
	}

	return []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "care_package",
			Description: "Give a random item and ability",
			Options: []*discordgo.ApplicationCommandOption{
				discord.IntCommandArg("level", "level to roll for", true),
				discord.UserCommandArg(false),
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "roll",
			Description: "Manual roll for item or ability",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "target",
					Description: "target type",
					Required:    true,
					Choices:     options,
				},
				discord.IntCommandArg("level", "level to roll for", true),
				discord.UserCommandArg(false),
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "debug",
			Description: "simulate roll and show chances",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "target",
					Description: "target type",
					Required:    true,
					Choices:     options,
				},
				discord.IntCommandArg("level", "level to roll for", true),
				discord.UserCommandArg(false),
			},
		},
	}
}

// Run implements ken.SlashCommand.
func (l *Luck) Run(ctx ken.Context) (err error) {
	return ctx.HandleSubCommands(
		ken.SubCommandHandler{Name: "roll", Run: l.luckRoll},
		ken.SubCommandHandler{Name: "debug", Run: l.luckDebug},
		ken.SubCommandHandler{Name: "care_package", Run: l.luckCarePackage},
	)
}

func (l *Luck) luckRoll(ctx ken.SubCommandContext) (err error) {
	if !discord.IsAdminRole(ctx, discord.AdminRoles...) {
		return discord.ErrorMessage(
			ctx,
			"You do not have permission to use this command",
			fmt.Sprintf(
				"You must have one of the following roles: %v",
				strings.Join(discord.AdminRoles, ","),
			),
		)
	}
	opts := ctx.Options()
	level := opts.GetByName("level").IntValue()
	target := opts.GetByName("target").StringValue()

	rng := rand.Float64()
	luckType := RollLuck(float64(level), rng)

	inv, err := inventory.Fetch(ctx, l.models)
	if err != nil {
		return discord.ErrorMessage(ctx,
			"Failed to find inventory",
			"If you're not in confessional, ensure you are in whitelist channel",
		)
	}

	if target == "item" {
		item, err := l.models.Items.GetRandomByRarity(luckType)
		if err != nil {
			log.Println(err)
			return discord.ErrorMessage(
				ctx,
				"Failed to get Random Item",
				"Alex is a bad programmer",
			)
		}
		inv.Items = append(inv.Items, item.Name)
		if len(inv.Items) > inv.ItemLimit {
			return discord.ErrorMessage(
				ctx,
				"Inventory is full",
				fmt.Sprintf("At item limit of %d/%d, please drop an item to add %s",
					inv.ItemLimit, inv.ItemLimit, item.Name,
				),
			)
		}
		l.models.Inventories.UpdateItems(inv)
		err = inventory.UpdateInventoryMessage(ctx, inv)
		if err != nil {
			log.Println(err)
			discord.SuccessfulMessage(
				ctx,
				"Failed to update inventory message",
				"Alex is a bad programmer",
			)
		}
		return discord.SuccessfulMessage(
			ctx,
			fmt.Sprintf("Got Item %s (%s)", item.Name, luckType),
			fmt.Sprintf("You now have %d/%d items", len(inv.Items), inv.ItemLimit),
		)
	}

	if target == "ability" {
		ability, err := l.models.Abilities.GetRandomByRarity(luckType)
		if err != nil {
			log.Println(err)
			return discord.ErrorMessage(
				ctx,
				"Failed to get Random Ability",
				"Alex is a bad programmer",
			)
		}
		inv.Abilities = append(inv.Abilities, ability.Name)
		err = l.models.Inventories.UpdateAbilities(inv)
		if err != nil {
			log.Println(err)
			return discord.ErrorMessage(
				ctx,
				"Failed to update inventory",
				"Alex is a bad programmer",
			)
		}
		err = inventory.UpdateInventoryMessage(ctx, inv)
		if err != nil {
			log.Println(err)
			discord.SuccessfulMessage(
				ctx,
				"Failed to update inventory message",
				"Alex is a bad programmer",
			)
		}

		desc := ""
		if ability.AnyAbility {
			desc = "IS ANY ABILITY"
		} else {
			desc = "IS NOT ANY ABILITY"
		}

		return discord.SuccessfulMessage(
			ctx,
			fmt.Sprintf("Got Ability %s (%s)", ability.Name, luckType),
			desc,
		)
	}

	return discord.ErrorMessage(ctx, "Failed to get category", "Alex is a bad programmer")
}

func (l *Luck) luckDebug(ctx ken.SubCommandContext) (err error) {
	if !discord.IsAdminRole(ctx, discord.AdminRoles...) {
		return discord.ErrorMessage(
			ctx,
			"You do not have permission to use this command",
			fmt.Sprintf(
				"You must have one of the following roles: %v",
				strings.Join(discord.AdminRoles, ","),
			),
		)
	}

	rng := rand.Float64()
	level := ctx.Options().GetByName("level").IntValue()
	view := tableView(float64(level), rng)

	return ctx.RespondMessage(view)
}

func (l *Luck) luckCarePackage(ctx ken.SubCommandContext) (err error) {
	if !discord.IsAdminRole(ctx, discord.AdminRoles...) {
		return discord.ErrorMessage(
			ctx,
			"You do not have permission to use this command",
			fmt.Sprintf(
				"You must have one of the following roles: %v",
				strings.Join(discord.AdminRoles, ","),
			),
		)
	}
	level := ctx.Options().GetByName("level").IntValue()
	inv, err := inventory.Fetch(ctx, l.models)
	if err != nil {
		log.Println(err)
		return discord.ErrorMessage(ctx,
			"Failed to find inventory",
			"If you're not in confessional, ensure you are in whitelist channel",
		)

	}
	aRoll := RollLuck(float64(level), rand.Float64())
	iRoll := RollLuck(float64(level), rand.Float64())

	ability, err := l.getRandomAbility(inv.RoleName, aRoll)
	if err != nil {
		return discord.ErrorMessage(ctx, "Error getting random ability", "Alex is a bad programmer")
	}

	item, err := l.models.Items.GetRandomByRarity(iRoll)
	if err != nil {
		log.Println(err)
		log.Println(err)
		return discord.ErrorMessage(
			ctx,
			"Failed to get Random Item",
			"Alex is a bad programmer",
		)
	}

	inv.Abilities = append(inv.Abilities, ability.Name)
	inv.Items = append(inv.Items, item.Name)
	err = l.models.Inventories.UpdateItems(inv)
	if err != nil {
		log.Println(err)
		return discord.ErrorMessage(
			ctx,
			"Failed to update inventory",
			"Alex is a bad programmer",
		)
	}

	err = inventory.UpdateInventoryMessage(ctx, inv)
	if err != nil {
		log.Println(err)
		discord.SuccessfulMessage(
			ctx,
			"Failed to update inventory message",
			"Alex is a bad programmer",
		)
	}
	return ctx.RespondEmbed(&discordgo.MessageEmbed{
		Title: fmt.Sprintf("%s Care Packing Incoming %s", discord.EmojiItem, discord.EmojiItem),
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Item",
				Value:  fmt.Sprintf("%s (%s) -  %s", item.Name, iRoll, item.Description),
				Inline: true,
			},
			{
				Name:   "Ability",
				Value:  fmt.Sprintf("%s (%s) -  %s", ability.Name, aRoll, ability.Description),
				Inline: true,
			},
		},
	})
}

// Version implements ken.SlashCommand.
func (*Luck) Version() string {
	return "1.0.0"
}

func (l *Luck) SetModels(models data.Models) {
	l.models = models
}

// Helper to get a random ability. If it is not an any ability, need to check to
// make sure that the ability is the same as the user's current class roll
func (l *Luck) getRandomAbility(role string, rarity string, rec ...int) (*data.Ability, error) {
	// lil saftey net to prevent infinite recursion (hopefully)
	if len(rec) > 0 && rec[0] > 5 {
		return nil, errors.New("too many attempts to get ability")
	}
	ab, err := l.models.Abilities.GetRandomByRarity(rarity)
	if err != nil {
		return nil, err
	}
	if !ab.AnyAbility {
		associatedRole, err := l.models.Roles.GetByAbilityID(ab.ID)
		if err != nil {
			return nil, err
		}

		// Need to re-roll since they got a non-any ability that isn't their role
		if associatedRole.Name != role {
			// Haha what's the worst that could happen with a recursive function :)
			return l.getRandomAbility(role, rarity)
		}
	}
	return ab, nil
}
