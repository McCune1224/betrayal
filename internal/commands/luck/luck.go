package roll

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

type Roll struct {
	models data.Models
}

var _ ken.SlashCommand = (*Roll)(nil)

// Description implements ken.SlashCommand.
func (*Roll) Description() string {
	return "Determine luck for a given level"
}

// Name implements ken.SlashCommand.
func (*Roll) Name() string {
	return discord.DebugCmd + "roll"
}

// Options implements ken.SlashCommand.
func (*Roll) Options() []*discordgo.ApplicationCommandOption {
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
				discord.UserCommandArg(false),
				discord.IntCommandArg("luck", "optional override of luck level", false),
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "item_rain",
			Description: "Make it rain!!",
			Options: []*discordgo.ApplicationCommandOption{
				discord.UserCommandArg(false),
				discord.IntCommandArg("luck", "optional override of luck level", false),
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "power_drop",
			Description: "give a random AA",
			Options: []*discordgo.ApplicationCommandOption{
				discord.UserCommandArg(false),
				discord.IntCommandArg("luck", "optional override of luck level", false),
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "roll",
			Description: "Manual roll for item or ability. DOES NOT ADD TO INVENTORY IMMEDIATELY",
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
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "table",
			Description: "Table view of luck calculations",
			Options: []*discordgo.ApplicationCommandOption{
				discord.IntCommandArg("low", "low range for luck", false),
				discord.IntCommandArg("high", "high range for luck", false),
			},
		},
	}
}

// Run implements ken.SlashCommand.
func (r *Roll) Run(ctx ken.Context) (err error) {
	return ctx.HandleSubCommands(
		ken.SubCommandHandler{Name: "roll", Run: r.luckRoll},
		ken.SubCommandHandler{Name: "debug", Run: r.luckDebug},
		ken.SubCommandHandler{Name: "care_package", Run: r.luckCarePackage},
		ken.SubCommandHandler{Name: "table", Run: r.luckTable},
		ken.SubCommandHandler{Name: "item_rain", Run: r.luckItemRain},
		ken.SubCommandHandler{Name: "power_drop", Run: r.luckPowerDrop},
	)
}

func (r *Roll) luckRoll(ctx ken.SubCommandContext) (err error) {
	inv, err := inventory.Fetch(ctx, r.models, true)
	if err != nil {
		if errors.Is(err, inventory.ErrNotAuthorized) {
			return discord.NotAuthorizedError(ctx)
		}
	}
	opts := ctx.Options()
	target := opts.GetByName("target").StringValue()
	level := opts.GetByName("level").IntValue()

	rng := rand.Float64()
	luckType := RollLuck(float64(level), rng)

	if target == "item" {
		item, err := r.models.Items.GetRandomByRarity(luckType)
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
		r.models.Inventories.UpdateItems(inv)
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
		ability, err := r.models.Abilities.GetRandomByRarity(luckType)
		if err != nil {
			log.Println(err)
			return discord.ErrorMessage(
				ctx,
				"Failed to get Random Ability",
				"Alex is a bad programmer",
			)
		}
		inv.Abilities = append(inv.Abilities, ability.Name)
		err = r.models.Inventories.UpdateAbilities(inv)
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

func (r *Roll) luckDebug(ctx ken.SubCommandContext) (err error) {
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

// func (r *Roll) luckPowerDrop(ctx ken.SubCommandContext) (err error) {
// 	inv, err := inventory.Fetch(ctx, r.models, true)
// 	if err != nil {
// 		return discord.ErrorMessage(
// 			ctx,
// 			"Failed to get inventory",
// 			"Are you in a whitelist or confessional channel?",
// 		)
// 	}
// 	luckLevel := inv.Luck
// 	luckArg, ok := ctx.Options().GetByNameOptional("luck")
// 	if ok {
// 		luckLevel = luckArg.IntValue()
// 	}
// 	rarityType := RollLuck(float64(luckLevel), rand.Float64())
// 	ability, err := r.getRandomAbility(inv.RoleName, rarityType)
// 	if err != nil {
// 		return discord.ErrorMessage(
// 			ctx,
// 			"Failed to get ability",
// 			"Alex is a bad programmer",
// 		)
// 	}
//
// 	if !ability.AnyAbility {
// 		// If its not an any ability, instead find the ability in base abilities and update charges instead
// 		for k, v := range inv.Abilities {
// 			currInvName := strings.Split(v, " [")[0]
// 			left := strings.Index(v, "[") + 1
// 			right := strings.Index(v, "]")
// 			charge, _ := strconv.Atoi(v[left:right])
// 			if strings.EqualFold(currInvName, ability.Name) {
// 				inv.Abilities[k] = fmt.Sprintf("%s [%d]", currInvName, charge)
// 				err = r.models.Inventories.UpdateAbilities(inv)
// 				if err != nil {
// 					log.Println(err)
// 					return discord.ErrorMessage(
// 						ctx,
// 						"Failed to update ability",
// 						"Alex is a bad programmer, and this is his fault.",
// 					)
// 				}
// 				err = inventory.UpdateInventoryMessage(ctx, inv)
// 				if err != nil {
// 					return err
// 				}
// 				return ctx.RespondMessage("Ability updated in inventory.")
// 			}
// 		}
//
// 		return discord.ErrorMessage(ctx, "Failed to find Role Specific Ability...???", "Alex made a major fucky wucky here somehow")
// 	}
//
// 	inv.AnyAbilities = append(inv.AnyAbilities, ability.Name)
// 	err = r.models.Inventories.UpdateAbilities(inv)
// 	if err != nil {
// 		return discord.ErrorMessage(
// 			ctx,
// 			"Failed to update inventory",
// 			"Alex is a bad programmer")
// 	}
//
// 	err = inventory.UpdateInventoryMessage(ctx, inv)
// 	if err != nil {
// 		return discord.ErrorMessage(
// 			ctx,
// 			"Failed to update inventory message",
// 			"Alex is a bad programmer")
// 	}
//
// 	return ctx.RespondEmbed(&discordgo.MessageEmbed{
// 		Title: fmt.Sprintf("%s Power Drop Incoming %s", discord.EmojiAbility, discord.EmojiAbility),
// 		Fields: []*discordgo.MessageEmbedField{
// 			{
// 				Name:   "Ability",
// 				Value:  fmt.Sprintf("%s (%s) -  %s", ability.Name, rarityType, ability.Description),
// 				Inline: true,
// 			},
// 		},
// 	})
// }

func (r *Roll) luckCarePackage(ctx ken.SubCommandContext) (err error) {
	inv, err := inventory.Fetch(ctx, r.models, true)
	if err != nil {
		if errors.Is(err, inventory.ErrNotAuthorized) {
			return discord.NotAuthorizedError(ctx)
		}
		return discord.ErrorMessage(ctx, "Failed to find inventory.", "If not in confessional, please specify a user")
	}
	luckLevel := inv.Luck
	luckArg, ok := ctx.Options().GetByNameOptional("luck")
	if ok {
		luckLevel = luckArg.IntValue()
	}

	aRoll := RollLuck(float64(luckLevel), rand.Float64())
	iRoll := RollLuck(float64(luckLevel), rand.Float64())

	ability, err := r.getRandomAbility(inv.RoleName, aRoll)
	if err != nil {
		return discord.ErrorMessage(ctx, "Error getting random ability", "Alex is a bad programmer")
	}

	item, err := r.models.Items.GetRandomByRarity(iRoll)
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
	err = r.models.Inventories.UpdateItems(inv)
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
		Title: fmt.Sprintf("%s Care Package Incoming %s", discord.EmojiItem, discord.EmojiItem),
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

func (r *Roll) luckTable(ctx ken.SubCommandContext) (err error) {
	// Setting to eph for now to avoid flooding channels with bulky messages
	ctx.SetEphemeral(true)
	low := 0
	high := 100
	lArg, lOk := ctx.Options().GetByNameOptional("low")
	hArg, hOk := ctx.Options().GetByNameOptional("high")
	if lOk {
		low = int(lArg.IntValue())
		if !hOk {
			high = int(lArg.IntValue()) + 10
		}
	}

	if hOk {
		high = int(hArg.IntValue())
	}

	if low < 0 || high < 0 {
		return discord.ErrorMessage(ctx, "Invalid Range", "Please provide a non-negative number")
	}

	tMsg := ""

	for level := float64(low); level < float64(high); level++ {
		currChances := []float64{
			commonLuckChance(level) * 100,
			uncommonLuckChance(level) * 100,
			rareLuckChance(level) * 100,
			epicLuckChance(level) * 100,
			legendaryLuckChance(level) * 100,
			mythicalLuckChance(level) * 100,
		}

		tMsg += fmt.Sprintf("%d - ,", int(level))
		for i := range currChances {
			tMsg += fmt.Sprintf("%.2f%%\t", currChances[i])
		}
		tMsg += "\n"
	}

	return ctx.RespondMessage(discord.Code(tMsg))
}

// func (r *Roll) luckItemRain(ctx ken.SubCommandContext) (err error) {
// 	inv, err := inventory.Fetch(ctx, r.models, true)
// 	if err != nil {
// 		if errors.Is(err, inventory.ErrNotAuthorized) {
// 			return discord.NotAuthorizedError(ctx)
// 		}
// 		return discord.ErrorMessage(ctx, "Failed to find inventory.", "If not in confessional, please specify a user")
// 	}
// 	luckLevel := inv.Luck
// 	luckArg, ok := ctx.Options().GetByNameOptional("luck")
// 	if ok {
// 		luckLevel = luckArg.IntValue()
// 	}
// 	numItems := rand.Intn(3) + 1
// 	newItems := []*data.Item{}
// 	for i := 0; i < numItems; i++ {
// 		item, err := r.models.Items.GetRandomByRarity(RollLuck(float64(luckLevel), rand.Float64()))
// 		if err != nil {
// 			discord.ErrorMessage(ctx, "Failed to get item", "Alex is a bad programmer")
// 		}
// 		newItems = append(newItems, item)
// 	}
//
// 	newItemsListing := []string{}
// 	for _, item := range newItems {
// 		newItemsListing = append(
// 			newItemsListing,
// 			fmt.Sprintf("%s (%s) - %s", discord.Bold(item.Name), item.Rarity, item.Description),
// 		)
// 	}
//
// 	title := fmt.Sprintf("%s Item Rain Incoming %s", discord.EmojiItem, discord.EmojiItem)
// 	desc := fmt.Sprintf(
// 		"Rolled %d Items from Item Rain!\n %s",
// 		len(newItems),
// 		strings.Join(newItemsListing, "\n\n"),
// 	)
// 	if len(inv.Items)+len(newItems) > inv.ItemLimit {
// 		desc += fmt.Sprintf(
// 			"\n %s inventory overflow [%d/%d] %s",
// 			discord.EmojiWarning,
// 			len(inv.Items)+len(newItems),
// 			inv.ItemLimit,
// 			discord.EmojiWarning,
// 		)
// 	}
// 	fields := []*discordgo.MessageEmbedField{}
//
// 	for _, item := range newItems {
// 		inv.Items = append(inv.Items, item.Name)
// 		fields = append(fields, &discordgo.MessageEmbedField{
// 			Name:   discord.Bold(item.Name),
// 			Value:  item.Description,
// 			Inline: true,
// 		})
// 	}
// 	err = r.models.Inventories.UpdateItems(inv)
// 	if err != nil {
// 		return discord.ErrorMessage(ctx, "Failed to update inventory", "Alex is a bad programmer")
// 	}
// 	inventory.UpdateInventoryMessage(ctx, inv)
//
// 	return ctx.RespondEmbed(&discordgo.MessageEmbed{
// 		Title:       title,
// 		Description: desc,
// 		Fields:      fields,
// 	})
// }

// Version implements ken.SlashCommand.
func (*Roll) Version() string {
	return "1.0.0"
}

func (r *Roll) SetModels(models data.Models) {
	r.models = models
}

// Helper to get a random ability. If it is not an any ability, need to check to
// make sure that the ability is the same as the user's current class roll
func (r *Roll) getRandomAbility(role string, rarity string, rec ...int) (*data.Ability, error) {
	// lil saftey net to prevent infinite recursion (hopefully)

	rec = append(rec, 1)
	if len(rec) > 0 && rec[0] > 10 {
		return nil, errors.New("too many attempts to get ability")
	}
	ab, err := r.models.Abilities.GetRandomByRarity(rarity)
	if err != nil {
		return nil, err
	}
	if !ab.AnyAbility {
		associatedRole, err := r.models.Roles.GetByAbilityID(ab.ID)
		if err != nil {
			return nil, err
		}
		// Need to re-roll since they got a non-any ability that isn't their role
		if associatedRole.Name != role {
			// FIXME: Every time a recursive call is made an angel loses its wings
			return r.getRandomAbility(role, rarity, rec[0]+1)
		}
	}
	return ab, nil
}
