package inventory

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/zekrotja/ken"
)

func (i *Inventory) removeAbility(ctx ken.SubCommandContext) (err error) {
	inventory, err := i.ImLazyMiddleware(ctx)
	if err != nil {
		log.Println(err)
		return err
	}
	ctx.SetEphemeral(false)
	abilityNameArg := ctx.Options().GetByName("name").StringValue()
	for k, v := range inventory.Abilities {
		abilityName := strings.Split(v, " [")[0]
		if strings.EqualFold(abilityName, abilityNameArg) {
			inventory.Abilities = append(
				inventory.Abilities[:k],
				inventory.Abilities[k+1:]...)
			err = i.models.Inventories.UpdateAbilities(inventory)
			if err != nil {
				log.Println(err)
				return discord.ErrorMessage(
					ctx,
					"Failed to remove base ability",
					"Alex is a bad programmer, and this is his fault.",
				)
			}
			err = i.updateInventoryMessage(ctx, inventory)
			if err != nil {
				return err
			}
			return discord.SuccessfulMessage(
				ctx,
				"Removed Base Ability",
				fmt.Sprintf("Removed %s from inventory.", abilityNameArg),
			)
		}
	}

	discord.ErrorMessage(
		ctx,
		"Failed to Remove Ability",
		fmt.Sprintf("Base ability %s not found in inventory.", abilityNameArg),
	)

	return err
}

func (i *Inventory) removeAnyAbility(ctx ken.SubCommandContext) (err error) {
	inventory, err := i.ImLazyMiddleware(ctx)
	if err != nil {
		log.Println(err)
		return err
	}
	ctx.SetEphemeral(false)
	abilityNameArg := ctx.Options().GetByName("name").StringValue()

	for k, v := range inventory.Abilities {
		abilityName := strings.Split(v, " [")[0]
		if strings.EqualFold(abilityName, abilityNameArg) {
			inventory.Abilities = append(
				inventory.Abilities[:k],
				inventory.Abilities[k+1:]...)
			err = i.models.Inventories.UpdateAnyAbilities(inventory)
			if err != nil {
				log.Println(err)
				return discord.ErrorMessage(
					ctx,
					"Failed to remove ability",
					"Alex is a bad programmer, and this is his fault.",
				)
			}
			err = i.updateInventoryMessage(ctx, inventory)
			if err != nil {
				return err
			}
			return discord.SuccessfulMessage(
				ctx,
				"Ability removed from inventory.",
				fmt.Sprintf("Removed %s from inventory.", abilityNameArg),
			)
		}
	}

	discord.ErrorMessage(
		ctx,
		"Failed to Remove Ability",
		fmt.Sprintf("Ability %s not found in inventory.", abilityNameArg),
	)

	return err
}

func (i *Inventory) removePerk(ctx ken.SubCommandContext) (err error) {
	inventory, err := i.ImLazyMiddleware(ctx)
	if err != nil {
		log.Println(err)
		return err
	}
	ctx.SetEphemeral(false)
	perkArg := ctx.Options().GetByName("name").StringValue()

	for k, v := range inventory.Perks {
		if strings.EqualFold(v, perkArg) {
			inventory.Perks = append(inventory.Perks[:k], inventory.Perks[k+1:]...)
			err = i.models.Inventories.UpdatePerks(inventory)
			if err != nil {
				log.Println(err)
				return discord.ErrorMessage(
					ctx,
					"Failed to remove perk",
					"Alex is a bad programmer, and this is his fault.",
				)
			}
			err = i.updateInventoryMessage(ctx, inventory)
			if err != nil {
				return err
			}
			return discord.SuccessfulMessage(
				ctx,
				"Perk removed from inventory",
				fmt.Sprintf("Removed %s from inventory.", perkArg),
			)
		}
	}

	discord.ErrorMessage(
		ctx,
		"Failed to remove Perk",
		fmt.Sprintf("Perk %s not found in inventory.", perkArg),
	)
	return err
}

func (i *Inventory) removeItem(ctx ken.SubCommandContext) (err error) {
	inventory, err := i.ImLazyMiddleware(ctx)
	if err != nil {
		log.Println(err)
		return err
	}
	ctx.SetEphemeral(false)

	itemArg := ctx.Options().GetByName("name").StringValue()

	for k, v := range inventory.Items {
		if strings.EqualFold(v, itemArg) {
			inventory.Items = append(inventory.Items[:k], inventory.Items[k+1:]...)
			err = i.models.Inventories.UpdateItems(inventory)
			if err != nil {
				log.Println(err)
				return discord.ErrorMessage(
					ctx,
					"Failed to remove item",
					"Alex is a bad programmer, and this is his fault.",
				)
			}
			err = i.updateInventoryMessage(ctx, inventory)
			if err != nil {
				return err
			}
			return discord.SuccessfulMessage(
				ctx,
				"Item removed from inventory.",
				fmt.Sprintf("Removed %s from inventory.", itemArg),
			)
		}
	}

	discord.ErrorMessage(
		ctx,
		"Failed to get Item",
		fmt.Sprintf("Item %s not found in inventory.", itemArg),
	)
	return err
}

func (i *Inventory) removeStatus(ctx ken.SubCommandContext) (err error) {
	inventory, err := i.ImLazyMiddleware(ctx)
	if err != nil {
		log.Println(err)
		return err
	}
	ctx.SetEphemeral(false)

	statusArg := ctx.Options().GetByName("name").StringValue()

	for k, v := range inventory.Statuses {
		if strings.EqualFold(v, statusArg) {
			inventory.Statuses = append(inventory.Statuses[:k], inventory.Statuses[k+1:]...)
			err = i.models.Inventories.UpdateStatuses(inventory)
			if err != nil {
				log.Println(err)
				return discord.ErrorMessage(
					ctx,
					"Failed to remove status",
					"Alex is a bad programmer, and this is his fault.",
				)
			}
			err = i.updateInventoryMessage(ctx, inventory)
			if err != nil {
				return err
			}
			return discord.SuccessfulMessage(
				ctx,
				"Status removed from inventory",
				fmt.Sprintf("Removed %s from inventory.", statusArg),
			)
		}
	}

	discord.ErrorMessage(
		ctx,
		"Failed to find Status",
		fmt.Sprintf("Status %s not found in inventory.", statusArg),
	)
	return err
}

func (i *Inventory) removeImmunity(ctx ken.SubCommandContext) (err error) {
	inventory, err := i.ImLazyMiddleware(ctx)
	if err != nil {
		log.Println(err)
		return err
	}
	ctx.SetEphemeral(false)

	immunityArg := ctx.Options().GetByName("name").StringValue()

	for k, v := range inventory.Immunities {
		if strings.EqualFold(v, immunityArg) {
			inventory.Immunities = append(inventory.Immunities[:k], inventory.Immunities[k+1:]...)
			err = i.models.Inventories.UpdateImmunities(inventory)
			if err != nil {
				log.Println(err)
				return discord.ErrorMessage(
					ctx,
					"Failed to remove immunity",
					"Alex is a bad programmer, and this is his fault.",
				)
			}
			err = i.updateInventoryMessage(ctx, inventory)
			if err != nil {
				return err
			}
			return discord.SuccessfulMessage(
				ctx,
				"Immunity removed from inventory.",
				fmt.Sprintf("Removed %s from inventory.", immunityArg),
			)
		}
	}

	discord.ErrorMessage(
		ctx,
		"Failed to get Immunity",
		fmt.Sprintf("Immunity %s not found in inventory.", immunityArg),
	)
	return err
}

func (i *Inventory) removeEffect(ctx ken.SubCommandContext) (err error) {
	inventory, err := i.ImLazyMiddleware(ctx)
	if err != nil {
		log.Println(err)
		return err
	}
	ctx.SetEphemeral(false)

	effectArg := ctx.Options().GetByName("name").StringValue()

	for k, v := range inventory.Effects {
		if strings.EqualFold(v, effectArg) {
			inventory.Effects = append(inventory.Effects[:k], inventory.Effects[k+1:]...)
			err = i.models.Inventories.UpdateEffects(inventory)
			if err != nil {
				log.Println(err)
				return discord.ErrorMessage(
					ctx,
					"Failed to remove effect",
					"Alex is a bad programmer, and this is his fault.",
				)
			}
			err = i.updateInventoryMessage(ctx, inventory)
			if err != nil {
				return err
			}
			return discord.SuccessfulMessage(
				ctx,
				"Effect removed from inventory.",
				fmt.Sprintf("Removed %s from inventory.", effectArg),
			)
		}
	}

	discord.ErrorMessage(
		ctx,
		"Failed to get effect",
		fmt.Sprintf("Effect %s not found in inventory.", effectArg),
	)
	return err
}

func (i *Inventory) removeCoins(ctx ken.SubCommandContext) (err error) {
	ctx.SetEphemeral(true)
	inventory, err := i.ImLazyMiddleware(ctx)
	if err != nil {
		log.Println(err)
		discord.ErrorMessage(
			ctx,
			"Failed to get inventory",
			"Alex is a bad programmer, and this is his fault.",
		)
		return err
	}
	ctx.SetEphemeral(false)

	coinsArg := ctx.Options().GetByName("amount").IntValue()

	previousCoins := inventory.Coins
	inventory.Coins -= coinsArg

	if inventory.Coins < 0 {
		return discord.ErrorMessage(ctx,
			"Insufficient Funds",
			fmt.Sprintf(
				"You don't have enough coins to remove %d coins.\n %d - %d = %d",
				coinsArg,
				previousCoins,
				coinsArg,
				inventory.Coins,
			))
	}

	err = i.models.Inventories.UpdateCoins(inventory)
	if err != nil {
		log.Println(err)
		return discord.ErrorMessage(
			ctx,
			"Failed to update coins",
			"Alex is a bad programmer, and this is his fault.",
		)
	}

	err = i.updateInventoryMessage(ctx, inventory)
	if err != nil {
		log.Println(err)
	}

	err = discord.SuccessfulMessage(ctx,
		"Coins removed",
		fmt.Sprintf(
			"Removed %d coins\n %d => %d",
			coinsArg,
			previousCoins,
			inventory.Coins,
		))
	return err
}

func (i *Inventory) removeCoinBonus(ctx ken.SubCommandContext) (err error) {
	inventory, err := i.ImLazyMiddleware(ctx)
	if err != nil {
		log.Println(err)
		discord.ErrorMessage(
			ctx,
			"Failed to get inventory",
			"Alex is a bad programmer, and this is his fault.",
		)
		return err
	}
	ctx.SetEphemeral(false)

	coinBonusArg := ctx.Options().GetByName("amount").StringValue()
	old := inventory.CoinBonus
	fCoinBonusArg, err := strconv.ParseFloat(coinBonusArg, 32)
	if err != nil {
		log.Println(err)
		return discord.ErrorMessage(
			ctx,
			"Failed to parse coin bonus",
			"Alex is a bad programmer, and this is his fault.",
		)
	}

	// 2.5 -> 0.025

	inventory.CoinBonus -= (float32(fCoinBonusArg) / 100)

	err = i.models.Inventories.UpdateProperty(inventory, "coin_bonus", inventory.CoinBonus)
	if err != nil {
		log.Println(err)
		return discord.ErrorMessage(
			ctx,
			"Failed to update coin bonus",
			"Alex is a bad programmer, and this is his fault.",
		)
	}

	err = i.updateInventoryMessage(ctx, inventory)
	if err != nil {
		log.Println(err)
	}

	err = discord.SuccessfulMessage(ctx,
		"Coin bonus removed",
		fmt.Sprintf(
			"Removed %s%% coin bonus\n %s%% => %s%%",
			strconv.FormatFloat(float64(fCoinBonusArg), 'f', 2, 32),
			strconv.FormatFloat(float64(old*100), 'f', 2, 32),
			strconv.FormatFloat(float64(inventory.CoinBonus*100), 'f', 2, 32),
		))
	return err
}

func (i *Inventory) removeWhitelist(ctx ken.SubCommandContext) (err error) {
	ctx.SetEphemeral(true)
	channelArg := ctx.Options().GetByName("channel").ChannelValue(ctx)

	whitelists, _ := i.models.Whitelists.GetAll()
	if len(whitelists) == 0 {
		err = discord.ErrorMessage(ctx, "No whitelisted channels", "Nothing here...")
		return err
	}

	for _, v := range whitelists {
		if v.ChannelID == channelArg.ID {
			i.models.Whitelists.Delete(v)
		}
		err = discord.SuccessfulMessage(ctx,
			"Channel removed from whitelist.",
			fmt.Sprintf("Removed %s from whitelist.", channelArg.Name))
		return err
	}

	err = discord.ErrorMessage(ctx, "Channel not found", "This channel is not whitelisted.")
	return err

}

func (i *Inventory) removeNote(ctx ken.SubCommandContext) (err error) {
	if !discord.IsAdminRole(ctx, discord.AdminRoles...) {
		return discord.ErrorMessage(
			ctx,
			"Unauthorized",
			"You are not authorized to use this command.",
		)
	}

	inventory, err := i.ImLazyMiddleware(ctx)
	if err != nil {
		log.Println(err)
		discord.ErrorMessage(
			ctx,
			"Failed to get inventory",
			"Alex is a bad programmer, and this is his fault.",
		)
		return err
	}

	noteArg := int(ctx.Options().GetByName("index").IntValue())
	// Subtract 1 to account for 0 indexing
	noteArg -= 1

	if noteArg < 0 || noteArg > len(inventory.Notes)-1 {
		return discord.ErrorMessage(ctx,
			"Invalid index",
			fmt.Sprintf("Index must be between 1 and %d", len(inventory.Notes)))
	}

	removedNote := inventory.Notes[noteArg]
	inventory.Notes = append(inventory.Notes[:noteArg], inventory.Notes[noteArg+1:]...)
	err = i.models.Inventories.UpdateNotes(inventory)
	if err != nil {
		log.Println(err)
		return discord.ErrorMessage(
			ctx,
			"Failed to remove note",
			"Alex is a bad programmer, and this is his fault.",
		)
	}

	err = discord.SuccessfulMessage(
		ctx,
		"Note removed",
		fmt.Sprintf("Removed note:\n %s", removedNote))

	return err
}
