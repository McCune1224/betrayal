package inventory

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/zekrotja/ken"
)

func (i *Inventory) setAbility(ctx ken.SubCommandContext) (err error) {
	inventory, err := i.imLazyMiddleware(ctx)
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
	abilityNameArg := ctx.Options().GetByName("name").StringValue()
	chargesArg := ctx.Options().GetByName("charges").IntValue()

	for k, v := range inventory.Abilities {
		abilityName := strings.Split(v, " [")[0]
		if strings.EqualFold(abilityName, abilityNameArg) {
			inventory.Abilities[k] = fmt.Sprintf("%s [%d]", abilityName, chargesArg)
			err = i.models.Inventories.UpdateAbilities(inventory)
			if err != nil {
				log.Println(err)
				return discord.ErrorMessage(
					ctx,
					"Failed to update ability",
					"Alex is a bad programmer, and this is his fault.",
				)
			}
			err = i.updateInventoryMessage(ctx, inventory)
			if err != nil {
				return err
			}
			return ctx.RespondMessage("Ability updated in inventory.")
		}
	}

	ctx.RespondMessage(fmt.Sprintf("Ability %s not found in inventory.", abilityNameArg))
	return err
}

func (i *Inventory) setAnyAbility(ctx ken.SubCommandContext) (err error) {
	inventory, err := i.imLazyMiddleware(ctx)
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
	abilityNameArg := ctx.Options().GetByName("name").StringValue()
	chargesArg := ctx.Options().GetByName("charges").IntValue()

	for k, v := range inventory.AnyAbilities {
		abilityName := strings.Split(v, " [")[0]
		if strings.EqualFold(abilityName, abilityNameArg) {
			inventory.AnyAbilities[k] = fmt.Sprintf("%s [%d]", abilityName, chargesArg)
			err = i.models.Inventories.UpdateAnyAbilities(inventory)
			if err != nil {
				log.Println(err)
				return discord.ErrorMessage(
					ctx,
					"Failed to update ability",
					"Alex is a bad programmer, and this is his fault.",
				)
			}
			err = i.updateInventoryMessage(ctx, inventory)
			if err != nil {
				return err
			}
			return ctx.RespondMessage("Ability updated in inventory.")
		}
	}

	ctx.RespondMessage(fmt.Sprintf("Ability %s not found in inventory.", abilityNameArg))
	return err
}

func (i *Inventory) setCoins(ctx ken.SubCommandContext) (err error) {
	inventory, err := i.imLazyMiddleware(ctx)
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
	inventory.Coins = coinsArg
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
		return err
	}
	return discord.SuccessfulMessage(
		ctx,
		"Coins updated",
		fmt.Sprintf("Set coins from %d to %d", coinsArg, inventory.Coins),
	)
}

func (i Inventory) setCoinBonus(ctx ken.SubCommandContext) (err error) {
	inventory, err := i.imLazyMiddleware(ctx)
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
			"Unable to parse coin bonus")
	}
    //FIXME: Remove round down 
	inventory.CoinBonus = (float32(fCoinBonusArg) / 100)

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
		return discord.ErrorMessage(
			ctx,
			"Failed to update inventory message",
			"Alex is a bad programmer, and this is his fault.",
		)
	}

	return discord.SuccessfulMessage(
		ctx,
		"Coin bonus updated",
		fmt.Sprintf(
			"Coin bonus set to %s%% (was %s%%)",
			coinBonusArg,
			strconv.FormatFloat(float64(old*100), 'f', 2, 32),
		))
}

func (i *Inventory) setItemsLimit(ctx ken.SubCommandContext) (err error) {
	inventory, err := i.imLazyMiddleware(ctx)
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
	itemsLimitArg := ctx.Options().GetByName("size").IntValue()
	inventory.ItemLimit = int(itemsLimitArg)
	err = i.models.Inventories.UpdateProperty(inventory, "item_limit", itemsLimitArg)
	if err != nil {
		log.Println(err)
		return discord.ErrorMessage(
			ctx,
			"Failed to update items limit",
			"Alex is a bad programmer, and this is his fault.",
		)
	}
	err = i.updateInventoryMessage(ctx, inventory)
	if err != nil {
		return err
	}

	return discord.SuccessfulMessage(
		ctx,
		"Items limit updated",
		fmt.Sprintf(
			"Items limit set to %d",
			inventory.ItemLimit,
		),
	)
}
