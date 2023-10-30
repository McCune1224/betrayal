package inventory

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/zekrotja/ken"
)

func (i *Inventory) setAbility(ctx ken.SubCommandContext) (err error) {
	inv, err := Fetch(ctx, i.models, true)
	if err != nil {
		if errors.Is(err, ErrNotAuthorized) {
			return discord.NotAuthorizedError(ctx)
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
			err = UpdateInventoryMessage(ctx, inv)
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
	inv, err := Fetch(ctx, i.models, true)
	if err != nil {
		if errors.Is(err, ErrNotAuthorized) {
			return discord.NotAuthorizedError(ctx)
		}
		return discord.ErrorMessage(ctx, "Failed to find inventory.", "If not in confessional, please specify a user")
	}
	ctx.SetEphemeral(false)
	abilityNameArg := ctx.Options().GetByName("name").StringValue()
	chargesArg := ctx.Options().GetByName("charges").IntValue()

	for k, v := range inv.AnyAbilities {
		abilityName := strings.Split(v, " [")[0]
		if strings.EqualFold(abilityName, abilityNameArg) {
			inv.AnyAbilities[k] = fmt.Sprintf("%s [%d]", abilityName, chargesArg)
			err = i.models.Inventories.UpdateAnyAbilities(inv)
			if err != nil {
				log.Println(err)
				return discord.ErrorMessage(
					ctx,
					"Failed to update ability",
					"Alex is a bad programmer, and this is his fault.",
				)
			}
			err = UpdateInventoryMessage(ctx, inv)
			if err != nil {
				log.Println(err)
				return err
			}
			return discord.SuccessfulMessage(ctx, "Ability updated", fmt.Sprintf("Set %s to %d charges", abilityName, chargesArg))
		}
	}

	return discord.ErrorMessage(ctx, "Unable to Set Ability Charge", fmt.Sprintf("Ability %s not found in inventory.", abilityNameArg))
}

func (i *Inventory) setCoins(ctx ken.SubCommandContext) (err error) {
	inv, err := Fetch(ctx, i.models, true)
	if err != nil {
		if errors.Is(err, ErrNotAuthorized) {
			return discord.NotAuthorizedError(ctx)
		}
		return discord.ErrorMessage(ctx, "Failed to find inventory.", "If not in confessional, please specify a user")
	}
	ctx.SetEphemeral(false)
	coinsArg := ctx.Options().GetByName("amount").IntValue()
	inv.Coins = coinsArg
	err = i.models.Inventories.UpdateCoins(inv)
	if err != nil {
		log.Println(err)
		return discord.ErrorMessage(
			ctx,
			"Failed to update coins",
			"Alex is a bad programmer, and this is his fault.",
		)
	}
	err = UpdateInventoryMessage(ctx, inv)
	if err != nil {
		log.Println(err)
		return err
	}
	return discord.SuccessfulMessage(
		ctx,
		"Coins updated",
		fmt.Sprintf("Set coins from %d to %d", coinsArg, inv.Coins),
	)
}

func (i Inventory) setCoinBonus(ctx ken.SubCommandContext) (err error) {
	inv, err := Fetch(ctx, i.models, true)
	if err != nil {
		if errors.Is(err, ErrNotAuthorized) {
			return discord.NotAuthorizedError(ctx)
		}
		return discord.ErrorMessage(ctx, "Failed to find inventory.", "If not in confessional, please specify a user")
	}
	ctx.SetEphemeral(false)

	coinBonusArg := ctx.Options().GetByName("amount").StringValue()
	old := inv.CoinBonus
	fCoinBonusArg, err := strconv.ParseFloat(coinBonusArg, 32)
	if err != nil {
		log.Println(err)
		return discord.ErrorMessage(
			ctx,
			"Failed to parse coin bonus",
			"Unable to parse coin bonus")
	}
	// FIXME: Remove round down
	inv.CoinBonus = (float32(fCoinBonusArg) / 100)

	err = i.models.Inventories.UpdateProperty(inv, "coin_bonus", inv.CoinBonus)
	if err != nil {
		log.Println(err)
		return discord.ErrorMessage(
			ctx,
			"Failed to update coin bonus",
			"Alex is a bad programmer, and this is his fault.",
		)
	}

	err = UpdateInventoryMessage(ctx, inv)
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
	inv, err := Fetch(ctx, i.models, true)
	if err != nil {
		if errors.Is(err, ErrNotAuthorized) {
			return discord.NotAuthorizedError(ctx)
		}
		return discord.ErrorMessage(ctx, "Failed to find inventory.", "If not in confessional, please specify a user")
	}
	ctx.SetEphemeral(false)
	itemsLimitArg := ctx.Options().GetByName("size").IntValue()
	inv.ItemLimit = int(itemsLimitArg)
	err = i.models.Inventories.UpdateProperty(inv, "item_limit", itemsLimitArg)
	if err != nil {
		log.Println(err)
		return discord.ErrorMessage(
			ctx,
			"Failed to update items limit",
			"Alex is a bad programmer, and this is his fault.",
		)
	}
	err = UpdateInventoryMessage(ctx, inv)
	if err != nil {
		log.Println(err)
		return err
	}

	return discord.SuccessfulMessage(
		ctx,
		"Items limit updated",
		fmt.Sprintf(
			"Items limit set to %d",
			inv.ItemLimit,
		),
	)
}

func (i *Inventory) setLuckLevel(ctx ken.SubCommandContext) (err error) {
	inv, err := Fetch(ctx, i.models, true)
	if err != nil {
		if errors.Is(err, ErrNotAuthorized) {
			return discord.NotAuthorizedError(ctx)
		}
		return discord.ErrorMessage(ctx, "Failed to find inventory.", "If not in confessional, please specify a user")
	}
	ctx.SetEphemeral(true)
	luckLevelArg := ctx.Options().GetByName("level").IntValue()
	oldLuck := inv.Luck
	inv.Luck = luckLevelArg

	err = i.models.Inventories.UpdateLuck(inv)
	if err != nil {
		log.Println(err)
		return discord.ErrorMessage(
			ctx,
			"Failed to update luck level",
			"Alex is a bad programmer, and this is his fault.",
		)
	}
	err = UpdateInventoryMessage(ctx, inv)
	if err != nil {
		log.Println(err)
		return discord.ErrorMessage(
			ctx,
			"Failed to update inventory message",
			"Alex is a bad programmer, and this is his fault.",
		)
	}

	return ctx.RespondEmbed(&discordgo.MessageEmbed{
		Title:       "Luck level updated",
		Description: fmt.Sprintf("Luck level from %d to %d", oldLuck, inv.Luck),
		Color:       discord.ColorThemeGreen,
	})
}
