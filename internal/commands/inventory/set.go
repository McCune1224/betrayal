package inventory

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/zekrotja/ken"
)

func (i *Inventory) setAnyAbility(ctx ken.SubCommandContext) (err error) {
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
	abilityNameArg := ctx.Options().GetByName("name").StringValue()
	chargesArg := ctx.Options().GetByName("charges").IntValue()

	for k, v := range inventory.AnyAbilities {
		abilityName := strings.Split(v, " [")[0]
		if strings.EqualFold(abilityName, abilityNameArg) {
			inventory.AnyAbilities[k] = fmt.Sprintf("%s [%d]", abilityName, chargesArg)
			err = i.models.Inventories.UpdateAnyAbilities(inventory)
			if err != nil {
				log.Println(err)
				return discord.SendSilentError(
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
	inventory.Coins = coinsArg
	err = i.models.Inventories.UpdateProperty(inventory, "coins", coinsArg)
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
		return err
	}
	return ctx.RespondMessage(
		fmt.Sprintf(
			"Coins set to %d",
			inventory.Coins,
		),
	)
}

func (i Inventory) setCoinBonus(ctx ken.SubCommandContext) (err error) {
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
			"Unable to parse coin bonus")
	}
	inventory.CoinBonus = (float32(fCoinBonusArg) / 100)

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
		return discord.SendSilentError(
			ctx,
			"Failed to update inventory message",
			"Alex is a bad programmer, and this is his fault.",
		)
	}

	return ctx.RespondMessage(
		fmt.Sprintf(
			"Coin bonus set to %f from %f",
			inventory.CoinBonus,
			fCoinBonusArg,
		))
}
