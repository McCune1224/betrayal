package inventory

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/mccune1224/betrayal/internal/services/inventory"
	"github.com/zekrotja/ken"
)

func (i *Inventory) removeAbility(ctx ken.SubCommandContext) (err error) {
	handler, err := FetchHandler(ctx, i.models, true)
	if err != nil {
		if errors.Is(err, ErrNotAuthorized) {
			return discord.NotAdminError(ctx)
		}
		return discord.ErrorMessage(ctx, "Failed to find inventory.", "If not in confessional, please specify a user")
	}
	ctx.SetEphemeral(false)
	abilityNameArg := ctx.Options().GetByName("name").StringValue()
	removed, err := handler.RemoveAbility(abilityNameArg)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Failed to remove ability")
	}

	err = UpdateInventoryMessage(ctx.GetSession(), handler.GetInventory())
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Failed to update inventory message")
	}

	return discord.SuccessfulMessage(ctx, fmt.Sprintf("Removed ability %s", removed), fmt.Sprintf("Removed ability %s for %s",
		abilityNameArg, discord.MentionUser(handler.GetInventory().DiscordID)))
}

func (i *Inventory) removeAnyAbility(ctx ken.SubCommandContext) (err error) {
	handler, err := FetchHandler(ctx, i.models, true)
	if err != nil {
		if errors.Is(err, ErrNotAuthorized) {
			return discord.NotAdminError(ctx)
		}
		return discord.ErrorMessage(ctx, "Failed to find inventory.", "If not in confessional, please specify a user")
	}
	ctx.SetEphemeral(false)
	abilityNameArg := ctx.Options().GetByName("name").StringValue()
	removed, err := handler.RemoveAnyAbility(abilityNameArg)
	if err != nil {
		if errors.Is(err, inventory.ErrAnyAbilityNotFound) {
			return discord.ErrorMessage(ctx, "Failed to Remove Ability", fmt.Sprintf("Ability %s not found in inventory.", abilityNameArg))
		}
		return discord.AlexError(ctx, "Failed to remove ability")
	}

	err = UpdateInventoryMessage(ctx.GetSession(), handler.GetInventory())
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Failed to update inventory message")
	}

	return discord.SuccessfulMessage(ctx, fmt.Sprintf("Removed Any Ability %s", removed),
		fmt.Sprintf("Removed Any Ability %s for %s", abilityNameArg, discord.MentionUser(handler.GetInventory().DiscordID)))
}

func (i *Inventory) removePerk(ctx ken.SubCommandContext) (err error) {
	handler, err := FetchHandler(ctx, i.models, true)
	if err != nil {
		if errors.Is(err, ErrNotAuthorized) {
			return discord.NotAdminError(ctx)
		}
		return discord.ErrorMessage(ctx, "Failed to find inventory.", "If not in confessional, please specify a user")
	}
	ctx.SetEphemeral(false)
	perkArg := ctx.Options().GetByName("name").StringValue()

	removed, err := handler.RemovePerk(perkArg)
	if err != nil {
		if errors.Is(err, inventory.ErrPerkNotFound) {
			return discord.ErrorMessage(ctx, "Failed to Remove Perk", fmt.Sprintf("Perk %s not found in inventory.", perkArg))
		}
		log.Println(err)
		return discord.AlexError(ctx, "Failed to remove perk")
	}
	err = UpdateInventoryMessage(ctx.GetSession(), handler.GetInventory())
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Failed to update inventory message")
	}

	return discord.SuccessfulMessage(ctx, fmt.Sprintf("Removed perk %s", removed),
		fmt.Sprintf("Removed perk for %s", discord.MentionUser(handler.GetInventory().DiscordID)))
}

func (i *Inventory) removeItem(ctx ken.SubCommandContext) (err error) {
	handler, err := FetchHandler(ctx, i.models, true)
	if err != nil {
		if errors.Is(err, ErrNotAuthorized) {
			return discord.NotAdminError(ctx)
		}
		return discord.ErrorMessage(ctx, "Failed to find inventory.", "If not in confessional, please specify a user")
	}
	ctx.SetEphemeral(false)
	itemArg := ctx.Options().GetByName("name").StringValue()
	item, err := handler.RemoveItem(itemArg)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Failed to remove item")
	}
	err = UpdateInventoryMessage(ctx.GetSession(), handler.GetInventory())
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Failed to update inventory message")
	}
	return discord.SuccessfulMessage(ctx, fmt.Sprintf("Removed item %s", item),
		fmt.Sprintf("Removed item %s to %s", item, discord.MentionUser(handler.GetInventory().DiscordID)))
}

func (i *Inventory) removeStatus(ctx ken.SubCommandContext) (err error) {
	inventory, err := Fetch(ctx, i.models, true)
	if err != nil {
		if errors.Is(err, ErrNotAuthorized) {
			return discord.NotAdminError(ctx)
		}
		return discord.ErrorMessage(ctx, "Failed to find inventory.", "If not in confessional, please specify a user")
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
	handler, err := FetchHandler(ctx, i.models, true)
	if err != nil {
		if errors.Is(err, ErrNotAuthorized) {
			return discord.NotAdminError(ctx)
		}
		return discord.ErrorMessage(ctx, "Failed to find inventory.", "If not in confessional, please specify a user")
	}
	ctx.SetEphemeral(false)

	immunityArg := ctx.Options().GetByName("name").StringValue()
	best, err := handler.RemoveImmunity(immunityArg)
	if err != nil {
		if errors.Is(err, inventory.ErrImmunityNotFound) {
			discord.ErrorMessage(ctx, "Failed to find immunity", fmt.Sprintf("Immunity similar to %s not found in inventory.", immunityArg))
		}
		log.Println(err)
		return discord.AlexError(ctx, "Failed to remove immunity")
	}

	err = UpdateInventoryMessage(ctx.GetSession(), handler.GetInventory())
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Failed to update inventory message")
	}
	return discord.SuccessfulMessage(ctx, fmt.Sprintf("Removed Immunity %s", best),
		fmt.Sprintf("Removed Immunity for %s", discord.MentionUser(handler.GetInventory().DiscordID)))
}

func (i *Inventory) removeEffect(ctx ken.SubCommandContext) (err error) {
	handler, err := FetchHandler(ctx, i.models, true)
	if err != nil {
		if errors.Is(err, ErrNotAuthorized) {
			return discord.NotAdminError(ctx)
		}
		return discord.ErrorMessage(ctx, "Failed to find inventory.", "If not in confessional, please specify a user")
	}
	ctx.SetEphemeral(false)
	effectArg := ctx.Options().GetByName("name").StringValue()
	best, err := handler.RemoveEffect(effectArg)
	if err != nil {
		if errors.Is(err, inventory.ErrEffectNotFound) {
			return discord.ErrorMessage(ctx, "Failed to find effect", fmt.Sprintf("Effect similar to %s not found in inventory.", effectArg))
		}
		log.Println(err)
		return discord.ErrorMessage(ctx, "Failed to get effect", fmt.Sprintf("Effect %s not found in inventory.", effectArg))
	}
	return discord.SuccessfulMessage(ctx, fmt.Sprintf("Removed effect %s", best),
		fmt.Sprintf("Removed effect for %s", discord.MentionUser(handler.GetInventory().DiscordID)))
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

func (i *Inventory) removeWhitelist(ctx ken.SubCommandContext) (err error) {
	if !discord.IsAdminRole(ctx, discord.AdminRoles...) {
		return discord.NotAdminError(ctx)
	}
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
		return discord.SuccessfulMessage(ctx, "Channel removed from whitelist.", fmt.Sprintf("Removed %s from whitelist.", channelArg.Name))
	}

	return discord.ErrorMessage(ctx, "Channel not found", "This channel is not whitelisted.")
}

func (i *Inventory) removeLuck(ctx ken.SubCommandContext) (err error) {
	handler, err := FetchHandler(ctx, i.models, true)
	if err != nil {
		if errors.Is(err, ErrNotAuthorized) {
			return discord.NotAdminError(ctx)
		}
		return discord.ErrorMessage(ctx, "Failed to find inventory.", "If not in confessional, please specify a user")
	}
	ctx.SetEphemeral(true)

	luck := ctx.Options().GetByName("amount").IntValue()

	err = handler.RemoveLuck(luck)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Failed to remove luck")
	}
	err = UpdateInventoryMessage(ctx.GetSession(), handler.GetInventory())
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Failed to update inventory message")
	}
	return discord.SuccessfulMessage(ctx, fmt.Sprintf("Removed %d luck", luck), fmt.Sprintf("Removed %d luck\n %d => %d for %s",
		luck, luck, handler.GetInventory().Luck, discord.MentionUser(handler.GetInventory().DiscordID)))
}

func (i *Inventory) removeItemLimit(ctx ken.SubCommandContext) (err error) {
	inv, err := Fetch(ctx, i.models, true)
	if err != nil {
		log.Println(err)
		return discord.ErrorMessage(ctx, "Failed to find inventory.", "If not in confessional, please specify a user")
	}
	ctx.SetEphemeral(false)
	itemLimitArg := ctx.Options().GetByName("amount").IntValue()
	ih := inventory.InitInventoryHandler(i.models, inv)
	err = ih.RemoveLimit(int(itemLimitArg))
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Failed to remove item limit")
	}
	err = i.updateInventoryMessage(ctx, inv)
	if err != nil {
		log.Println(err)
		return err
	}
	return discord.SuccessfulMessage(ctx, "Item Limit Updated", fmt.Sprintf("Item limit set to %d", inv.ItemLimit))
}

func (i *Inventory) removeNote(ctx ken.SubCommandContext) (err error) {
	handler, err := FetchHandler(ctx, i.models, true)
	if err != nil {
		if errors.Is(err, ErrNotAuthorized) {
			return discord.NotAdminError(ctx)
		}
		return discord.ErrorMessage(ctx, "Failed to find inventory.", "If not in confessional, please specify a user")
	}
	ctx.SetEphemeral(true)
	if len(handler.GetInventory().Notes) == 0 {
		return discord.ErrorMessage(ctx, "No notes to remove", "Nothing to see here officer...")
	}
	// Subtract 1 to account for 0 indexing (user input is 1 indexed)
	noteArg := int(ctx.Options().GetByName("index").IntValue()) - 1
	if noteArg < 0 || noteArg > len(handler.GetInventory().Notes)-1 {
		return discord.ErrorMessage(ctx, "Invalid note index",
			fmt.Sprintf("Please enter a number between 1 and %d", len(handler.GetInventory().Notes)))
	}
	err = handler.RemoveLimit(noteArg)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Failed to remove note")
	}
	err = UpdateInventoryMessage(ctx.GetSession(), handler.GetInventory())
	if err != nil {
		return discord.AlexError(ctx, "Failed to update inventory")
	}
	return discord.SuccessfulMessage(ctx, "Note removed", fmt.Sprintf("Removed note #%d", noteArg+1))
}
