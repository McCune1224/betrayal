package inventory

import (
	"errors"
	"fmt"
	"log"

	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/mccune1224/betrayal/internal/services/inventory"
	"github.com/zekrotja/ken"
)

func (i *Inventory) addCoins(ctx ken.SubCommandContext) (err error) {
	ctx.SetEphemeral(false)
	handler, err := FetchHandler(ctx, i.models, true)
	if err != nil {
		if errors.Is(err, ErrNotAuthorized) {
			return discord.NotAdminError(ctx)
		}
		return discord.ErrorMessage(ctx, "Failed to find inventory.", "If not in confessional, please specify a user")
	}

	coinsArg := ctx.Options().GetByName("amount").IntValue()
	err = handler.AddCoins(coinsArg)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Failed to add coins")
	}
	err = UpdateInventoryMessage(ctx.GetSession(), handler.GetInventory())
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Failed to update inventory message")
	}

	newCoins := handler.GetInventory().Coins
	return discord.SuccessfulMessage(ctx, "Added Coins",
		fmt.Sprintf("Added %d coins\n %d => %d for %s", coinsArg, newCoins-coinsArg, newCoins, discord.MentionUser(handler.GetInventory().DiscordID)))
}

func (i *Inventory) removeCoins(ctx ken.SubCommandContext) (err error) {
	handler, err := FetchHandler(ctx, i.models, true)
	if err != nil {
		if errors.Is(err, ErrNotAuthorized) {
			return discord.NotAdminError(ctx)
		}
		return discord.ErrorMessage(ctx, "Failed to find inventory.", "If not in confessional, please specify a user")
	}

	coinsArg := ctx.Options().GetByName("amount").IntValue()
	previousCoins := handler.GetInventory().Coins
	err = handler.RemoveCoins(coinsArg)
	if err != nil {
		if errors.Is(err, inventory.ErrInsufficientCoins) {
			return discord.ErrorMessage(ctx, "Will put inventory in negative balance",
				fmt.Sprintf("Cannot remove %d coins from %d (balance will be %d)", coinsArg, previousCoins, previousCoins-coinsArg))
		}
		log.Println(err)
		return discord.AlexError(ctx, "Failed to remove coins")
	}

	err = UpdateInventoryMessage(ctx.GetSession(), handler.GetInventory())
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Failed to update inventory")
	}

	return discord.SuccessfulMessage(ctx, "Coins removed",
		fmt.Sprintf("Removed %d coins\n %d => %d for %s",
			coinsArg, previousCoins, handler.GetInventory().Coins, discord.MentionUser(handler.GetInventory().DiscordID)))
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

func (i *Inventory) addCoinBonus(ctx ken.SubCommandContext) (err error) {
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
	err = handler.AddCoinBonus(coinBonusArg)
	if err != nil {
		if errors.Is(err, inventory.ErrInvalidDecimalString) {
			return discord.ErrorMessage(ctx, "Invalid decimal string", err.Error())
		}
		log.Println(err)
		return discord.AlexError(ctx, "Failed to add coin bonus")
	}
	err = i.updateInventoryMessage(ctx, handler.GetInventory())
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Failed to update inventory message")
	}

	return discord.SuccessfulMessage(ctx, "Added Coin Bonus",
		fmt.Sprintf("%.2f => %.2f fot %s",
			float32(int(old*100))/100, float32(int(handler.GetInventory().CoinBonus*100))/100,
			discord.MentionUser(handler.GetInventory().DiscordID)))
}

func (i *Inventory) removeCoinBonus(ctx ken.SubCommandContext) (err error) {
	handler, err := FetchHandler(ctx, i.models, true)
	if err != nil {
		if errors.Is(err, ErrNotAuthorized) {
			return discord.NotAdminError(ctx)
		}
		return discord.ErrorMessage(ctx, "Failed to find inventory.", "If not in confessional, please specify a user")
	}

	coinBonusArg := ctx.Options().GetByName("amount").StringValue()
	old := handler.GetInventory().CoinBonus

	err = handler.RemoveCoinBonus(coinBonusArg)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Failed to remove coin bonus")
	}
	err = i.updateInventoryMessage(ctx, handler.GetInventory())
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "failed to update inventory message")
	}

	return discord.SuccessfulMessage(ctx, "Removed Coin Bonus",
		fmt.Sprintf("%.2f => %.2f for %s",
			float32(int(old*100))/100, float32(int(handler.GetInventory().CoinBonus*100))/100,
			discord.MentionUser(handler.GetInventory().DiscordID)))
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
