package inventory

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/zekrotja/ken"
)

func (i *Inventory) setAbility(ctx ken.SubCommandContext) (err error) {
	inv, err := Fetch(ctx, i.models, true)
	if err != nil {
		if errors.Is(err, ErrNotAuthorized) {
			return discord.NotAdminError(ctx)
		}
		return discord.ErrorMessage(ctx, "Failed to find inventory.", "If not in confessional, please specify a user")
	}
	ctx.SetEphemeral(false)
	abilityNameArg := ctx.Options().GetByName("name").StringValue()
	chargesArg := ctx.Options().GetByName("charges").IntValue()

	for k, v := range inv.Abilities {
		abilityName := strings.Split(v, " [")[0]
		if strings.EqualFold(abilityName, abilityNameArg) {
			inv.Abilities[k] = fmt.Sprintf("%s [%d]", abilityName, chargesArg)
			err = i.models.Inventories.UpdateAbilities(inv)
			if err != nil {
				log.Println(err)
				return discord.ErrorMessage(
					ctx,
					"Failed to update ability",
					"Alex is a bad programmer, and this is his fault.",
				)
			}
			err = UpdateInventoryMessage(ctx.GetSession(), inv)
			if err != nil {
				log.Println(err)
				return err
			}
			return discord.SuccessfulMessage(ctx, "Ability updated", fmt.Sprintf("Set %s to %d charges", abilityName, chargesArg))
		}
	}

	return discord.ErrorMessage(ctx, "Unable to Set Ability Charge", fmt.Sprintf("Ability %s not found in inventory.", abilityNameArg))
}

func (i *Inventory) setAnyAbility(ctx ken.SubCommandContext) (err error) {
	handler, err := FetchHandler(ctx, i.models, true)
	if err != nil {
		if errors.Is(err, ErrNotAuthorized) {
			return discord.NotAdminError(ctx)
		}
		return discord.ErrorMessage(ctx, "Failed to find inventory.", "If not in confessional, please specify a user")
	}
	ctx.SetEphemeral(false)
	abilityNameArg := ctx.Options().GetByName("name").StringValue()
	chargesArg := ctx.Options().GetByName("charges").IntValue()

	handler.SetAnyAbilityCharges(abilityNameArg, int(chargesArg))

	return discord.ErrorMessage(ctx, "Unable to Set Ability Charge", fmt.Sprintf("Ability %s not found in inventory.", abilityNameArg))
}

func (i *Inventory) setCoins(ctx ken.SubCommandContext) (err error) {
	handler, err := FetchHandler(ctx, i.models, true)
	if err != nil {
		if errors.Is(err, ErrNotAuthorized) {
			return discord.NotAdminError(ctx)
		}
		return discord.ErrorMessage(ctx, "Failed to find inventory.", "If not in confessional, please specify a user")
	}
	ctx.SetEphemeral(false)
	coinsArg := ctx.Options().GetByName("amount").IntValue()
	oldCoins := handler.GetInventory().Coins
	err = handler.SetCoins(coinsArg)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Failed to set coins")
	}
	err = UpdateInventoryMessage(ctx.GetSession(), handler.GetInventory())
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Failed to update inventory message")
	}
	return discord.SuccessfulMessage(ctx, "Coins Set", fmt.Sprintf("Set coins from %d to %d for %s",
		oldCoins, handler.GetInventory().Coins, handler.GetInventory().DiscordID))
}

func (i Inventory) setCoinBonus(ctx ken.SubCommandContext) (err error) {
	handler, err := FetchHandler(ctx, i.models, true)
	if err != nil {
		if errors.Is(err, ErrNotAuthorized) {
			return discord.NotAdminError(ctx)
		}
		return discord.ErrorMessage(ctx, "Failed to find inventory.", "If not in confessional, please specify a user")
	}
	ctx.SetEphemeral(false)
	coinBonusArg := ctx.Options().GetByName("amount").StringValue()
	old := handler.GetInventory().CoinBonus
	err = handler.SetCoinBonus(coinBonusArg)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Failed to set coin bonus")
	}
	err = UpdateInventoryMessage(ctx.GetSession(), handler.GetInventory())
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Failed to update inventory message")
	}
	return discord.SuccessfulMessage(ctx, "Set Coin Bonus",
		fmt.Sprintf("%.2f => %.2f for %s",
			float32(int(old*100))/100, float32(int(handler.GetInventory().CoinBonus*100))/100,
			handler.GetInventory().DiscordID))
}

func (i *Inventory) setItemsLimit(ctx ken.SubCommandContext) (err error) {
	handler, err := FetchHandler(ctx, i.models, true)
	if err != nil {
		if errors.Is(err, ErrNotAuthorized) {
			return discord.NotAdminError(ctx)
		}
		return discord.ErrorMessage(ctx, "Failed to find inventory.", "If not in confessional, please specify a user")
	}
	ctx.SetEphemeral(false)
	itemsLimitArg := ctx.Options().GetByName("size").IntValue()

	err = handler.SetLimit(int(itemsLimitArg))
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Failed to set item limit")
	}
	err = UpdateInventoryMessage(ctx.GetSession(), handler.GetInventory())
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Failed to update inventory message")
	}

	return discord.SuccessfulMessage(
		ctx, "Items Limit updated", fmt.Sprintf("Items Limit set to %d", handler.GetInventory().ItemLimit),
	)
}

func (i *Inventory) setLuck(ctx ken.SubCommandContext) (err error) {
	handler, err := FetchHandler(ctx, i.models, true)
	if err != nil {
		if errors.Is(err, ErrNotAuthorized) {
			return discord.NotAdminError(ctx)
		}
		return discord.ErrorMessage(ctx, "Failed to find inventory.", "If not in confessional, please specify a user")
	}
	ctx.SetEphemeral(true)
	luckLevelArg := ctx.Options().GetByName("amount").IntValue()
	oldLuck := handler.GetInventory().Luck

	err = handler.SetLuck(luckLevelArg)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Failed to set luck level")
	}
	err = UpdateInventoryMessage(ctx.GetSession(), handler.GetInventory())
	if err != nil {
		log.Println(err)
		return discord.ErrorMessage(ctx, "Failed to update inventory message", "Alex is a bad programmer, and this is his fault.")
	}

	return discord.SuccessfulMessage(ctx, fmt.Sprintf("Luck set to %d", luckLevelArg),
		fmt.Sprintf("Luck level from %d to %d for %s", oldLuck, handler.GetInventory().Luck, handler.GetInventory().DiscordID))
}

func (i *Inventory) setAlignment(ctx ken.SubCommandContext) (err error) {
	alignmentArg := ctx.Options().GetByName("name").StringValue()
	handler, err := FetchHandler(ctx, i.models, true)
	if err != nil {
		if errors.Is(err, ErrNotAuthorized) {
			return discord.NotAdminError(ctx)
		}
		return discord.ErrorMessage(ctx, "Failed to find inventory.", "If not in confessional, please specify a user")
	}
	ctx.SetEphemeral(false)
	err = handler.SetAlignment(alignmentArg)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Failed to set alignment")
	}
	err = UpdateInventoryMessage(ctx.GetSession(), handler.GetInventory())
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Failed to update inventory message")
	}

	return discord.SuccessfulMessage(ctx, fmt.Sprintf("Alignment set to %s", handler.GetInventory().Alignment),
		fmt.Sprintf("Upated alignment for %s", handler.GetInventory().DiscordID))
}
