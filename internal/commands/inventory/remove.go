package inventory

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/zekrotja/ken"
)

func (i *Inventory) removeAnyAbility(ctx ken.SubCommandContext) (err error) {
	inventory, err := i.imLazyMiddleware(ctx)
	if err != nil {
		log.Println(err)
		return err
	}
	abilityNameArg := ctx.Options().GetByName("name").StringValue()

	log.Println(inventory.AnyAbilities)
	for k, v := range inventory.AnyAbilities {
		abilityName := strings.Split(v, " [")[0]
		if strings.EqualFold(abilityName, abilityNameArg) {
			log.Println("Removing ability from inventory")
			log.Println(inventory.AnyAbilities)
			inventory.AnyAbilities = append(
				inventory.AnyAbilities[:k],
				inventory.AnyAbilities[k+1:]...)
			log.Println(inventory.AnyAbilities)
			err = i.models.Inventories.UpdateAnyAbilities(inventory)
			if err != nil {
				log.Println(err)
				return discord.SendSilentError(
					ctx,
					"Failed to remove ability",
					"Alex is a bad programmer, and this is his fault.",
				)
			}
			err = i.updateInventoryMessage(ctx, inventory)
			if err != nil {
				return err
			}
			return ctx.RespondMessage("Ability removed from inventory.")
		}
	}

	ctx.RespondMessage(fmt.Sprintf("Ability %s not found in inventory.", abilityNameArg))

	return err
}

func (i *Inventory) removePerk(ctx ken.SubCommandContext) (err error) {
	inventory, err := i.imLazyMiddleware(ctx)
	if err != nil {
		log.Println(err)
		return err
	}
	perkArg := ctx.Options().GetByName("name").StringValue()

	for k, v := range inventory.Perks {
		if strings.EqualFold(v, perkArg) {
			inventory.Perks = append(inventory.Perks[:k], inventory.Perks[k+1:]...)
			err = i.models.Inventories.UpdatePerks(inventory)
			if err != nil {
				log.Println(err)
				return discord.SendSilentError(
					ctx,
					"Failed to remove perk",
					"Alex is a bad programmer, and this is his fault.",
				)
			}
			err = i.updateInventoryMessage(ctx, inventory)
			if err != nil {
				return err
			}
			return ctx.RespondMessage("Perk removed from inventory.")
		}
	}

	ctx.RespondMessage(fmt.Sprintf("Perk %s not found in inventory.", perkArg))
	return err
}

func (i *Inventory) removeItem(ctx ken.SubCommandContext) (err error) {
	inventory, err := i.imLazyMiddleware(ctx)
	if err != nil {
		log.Println(err)
		return err
	}

	itemArg := ctx.Options().GetByName("name").StringValue()

	for k, v := range inventory.Items {
		if strings.EqualFold(v, itemArg) {
			inventory.Items = append(inventory.Items[:k], inventory.Items[k+1:]...)
			err = i.models.Inventories.UpdateItems(inventory)
			if err != nil {
				log.Println(err)
				return discord.SendSilentError(
					ctx,
					"Failed to remove item",
					"Alex is a bad programmer, and this is his fault.",
				)
			}
			err = i.updateInventoryMessage(ctx, inventory)
			if err != nil {
				return err
			}
			return ctx.RespondMessage("Item removed from inventory.")
		}
	}

	ctx.RespondMessage(fmt.Sprintf("Item %s not found in inventory.", itemArg))
	return err
}

func (i *Inventory) removeStatus(ctx ken.SubCommandContext) (err error) {
	inventory, err := i.imLazyMiddleware(ctx)
	if err != nil {
		log.Println(err)
		return err
	}

	statusArg := ctx.Options().GetByName("name").StringValue()

	for k, v := range inventory.Statuses {
		if strings.EqualFold(v, statusArg) {
			inventory.Statuses = append(inventory.Statuses[:k], inventory.Statuses[k+1:]...)
			err = i.models.Inventories.UpdateStatuses(inventory)
			if err != nil {
				log.Println(err)
				return discord.SendSilentError(
					ctx,
					"Failed to remove status",
					"Alex is a bad programmer, and this is his fault.",
				)
			}
			err = i.updateInventoryMessage(ctx, inventory)
			if err != nil {
				return err
			}
			return ctx.RespondMessage("Status removed from inventory.")
		}
	}

	ctx.RespondMessage(fmt.Sprintf("Status %s not found in inventory.", statusArg))
	return err
}

func (i *Inventory) removeImmunity(ctx ken.SubCommandContext) (err error) {
	inventory, err := i.imLazyMiddleware(ctx)
	if err != nil {
		log.Println(err)
		return err
	}

	immunityArg := ctx.Options().GetByName("name").StringValue()

	for k, v := range inventory.Immunities {
		if strings.EqualFold(v, immunityArg) {
			inventory.Immunities = append(inventory.Immunities[:k], inventory.Immunities[k+1:]...)
			err = i.models.Inventories.UpdateImmunities(inventory)
			if err != nil {
				log.Println(err)
				return discord.SendSilentError(
					ctx,
					"Failed to remove immunity",
					"Alex is a bad programmer, and this is his fault.",
				)
			}
			err = i.updateInventoryMessage(ctx, inventory)
			if err != nil {
				return err
			}
			return ctx.RespondMessage("Immunity removed from inventory.")
		}
	}

	ctx.RespondMessage(fmt.Sprintf("Immunity %s not found in inventory.", immunityArg))
	return err
}

func (i *Inventory) removeEffect(ctx ken.SubCommandContext) (err error) {
	inventory, err := i.imLazyMiddleware(ctx)
	if err != nil {
		log.Println(err)
		return err
	}

	effectArg := ctx.Options().GetByName("name").StringValue()

	for k, v := range inventory.Effects {
		if strings.EqualFold(v, effectArg) {
			inventory.Effects = append(inventory.Effects[:k], inventory.Effects[k+1:]...)
			err = i.models.Inventories.UpdateEffects(inventory)
			if err != nil {
				log.Println(err)
				return discord.SendSilentError(
					ctx,
					"Failed to remove effect",
					"Alex is a bad programmer, and this is his fault.",
				)
			}
			err = i.updateInventoryMessage(ctx, inventory)
			if err != nil {
				return err
			}
			return ctx.RespondMessage("Effect removed from inventory.")
		}
	}

	ctx.RespondMessage(fmt.Sprintf("Effect %s not found in inventory.", effectArg))
	return err
}

func (i *Inventory) removeCoins(ctx ken.SubCommandContext) (err error) {
	ctx.SetEphemeral(true)
	inventory, err := i.imLazyMiddleware(ctx)
	if err != nil {
		log.Println(err)
		discord.SendSilentError(
			ctx,
			"Failed to get inventory",
			"Alex is a bad programmer, and this is his fault.",
		)
		return err
	}

	coinsArg := ctx.Options().GetByName("amount").IntValue()

	previousCoins := inventory.Coins
	inventory.Coins -= coinsArg

	if inventory.Coins < 0 {
		return ctx.RespondMessage(fmt.Sprintf(
			"You don't have enough coins to remove %d coins.\n %d - %d = %d",
			coinsArg,
			previousCoins,
			coinsArg,
			inventory.Coins,
		))
	}

	err = i.models.Inventories.Update(inventory)
	if err != nil {
		log.Println(err)
		return discord.SendSilentError(
			ctx,
			"Failed to update coins",
			"Alex is a bad programmer, and this is his fault.",
		)
	}

	err = i.updateInventoryMessage(ctx, inventory)
	if err != nil {
		log.Println(err)
	}

	err = ctx.RespondMessage(fmt.Sprintf(
		"Removed %d coins\n %d => %d",
		coinsArg,
		previousCoins,
		inventory.Coins,
	))
	return err
}

func (i *Inventory) removeCoinBonus(ctx ken.SubCommandContext) (err error) {
	ctx.SetEphemeral(true)
	inventory, err := i.imLazyMiddleware(ctx)
	if err != nil {
		log.Println(err)
		discord.SendSilentError(
			ctx,
			"Failed to get inventory",
			"Alex is a bad programmer, and this is his fault.",
		)
		return err
	}

	coinBonusArg := ctx.Options().GetByName("amount").StringValue()
	fCoinBonusArg, err := strconv.ParseFloat(coinBonusArg, 32)
	if err != nil {
		log.Println(err)
		return discord.SendSilentError(
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
		return discord.SendSilentError(
			ctx,
			"Failed to update coin bonus",
			"Alex is a bad programmer, and this is his fault.",
		)
	}

	err = i.updateInventoryMessage(ctx, inventory)
	if err != nil {
		log.Println(err)
	}

	err = ctx.RespondMessage(fmt.Sprintf(
		"Removed %s coin bonus\n %f => %f",
		coinBonusArg,
		inventory.CoinBonus+float32(fCoinBonusArg),
		inventory.CoinBonus,
	))
	return err
}

func (i *Inventory) removeWhitelist(ctx ken.SubCommandContext) (err error) {
	ctx.SetEphemeral(true)
	channelArg := ctx.Options().GetByName("channel").ChannelValue(ctx)

	whitelists, _ := i.models.Whitelists.GetAll()
	if len(whitelists) == 0 {
		err = discord.SendSilentError(ctx, "No whitelisted channels", "Nothing here...")
		return err
	}

	for _, v := range whitelists {
		if v.ChannelID == channelArg.ID {
			i.models.Whitelists.Delete(v)
		}
		err = ctx.RespondMessage("Channel removed from whitelist.")
		return err
	}

	err = discord.SendSilentError(ctx, "Channel not found", "This channel is not whitelisted.")
	return err

}
