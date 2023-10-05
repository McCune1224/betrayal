package inventory

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/mccune1224/betrayal/internal/data"
	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/zekrotja/ken"
)

func (i *Inventory) addAnyAbility(ctx ken.SubCommandContext) (err error) {
	ctx.SetEphemeral(true)
	inventory, err := i.imLazyMiddleware(ctx)
	if err != nil {
		return err
	}
	abilityNameArg := ctx.Options().GetByName("name").StringValue()
	charges := int64(-42069)
	chargesArg, ok := ctx.Options().GetByNameOptional("charges")
	if ok {
		charges = chargesArg.IntValue()
	}
	for _, ability := range inventory.AnyAbilities {
		abilityName := strings.Split(ability, " [")[0]
		if abilityName == abilityNameArg {
			return discord.SendSilentError(
				ctx,
				fmt.Sprintf("Ability %s already exists in inventory", abilityNameArg),
				"Did you mean to update the ability?",
			)
		}
	}

	ability, err := i.models.Abilities.GetByName(abilityNameArg)
	if err != nil {
		return discord.SendSilentError(
			ctx,
			fmt.Sprint("Cannot find Ability: ", abilityNameArg),
			"Verify if the ability exists.",
		)
	}
	if charges == -42069 {
		charges = int64(ability.Charges)
	}
	abilityString := fmt.Sprintf("%s [%d]", ability.Name, charges)
	inventory.AnyAbilities = append(inventory.AnyAbilities, abilityString)
	err = i.models.Inventories.Update(inventory)
	if err != nil {
		log.Println(err)
		return discord.SendSilentError(
			ctx,
			"Failed to add ability",
			"Alex is a bad programmer, and this is his fault.",
		)
	}

	err = i.updateInventoryMessage(ctx, inventory)
	if err != nil {
		return err
	}

	err = ctx.RespondMessage(
		fmt.Sprintf("Any Ability %s added", abilityNameArg),
	)
	return err
}

func (i *Inventory) addPerk(ctx ken.SubCommandContext) (err error) {
	inventory, err := i.imLazyMiddleware(ctx)
	if err != nil {
		return err
	}

	perkNameArg := ctx.Options().GetByName("name").StringValue()
	perk, err := i.models.Perks.GetByName(perkNameArg)
	if err != nil {
		return discord.SendSilentError(
			ctx,
			fmt.Sprint("Cannot find Perk: ", perkNameArg),
			"Verify if the perk exists.",
		)
	}

	for _, p := range inventory.Perks {
		if p == perk.Name {
			return discord.SendSilentError(
				ctx,
				fmt.Sprintf("Perk %s already exists in inventory", perkNameArg),
				"Did you mean to update the perk?",
			)
		}
	}

	inventory.Perks = append(inventory.Perks, perk.Name)
	err = i.models.Inventories.UpdatePerks(inventory)
	if err != nil {
		log.Println(err)
		return discord.SendSilentError(
			ctx,
			"Failed to add perk",
			"Alex is a bad programmer, and this is his fault.",
		)
	}
	err = i.updateInventoryMessage(ctx, inventory)
	if err != nil {
		log.Println(err)
	}

	err = ctx.RespondMessage(fmt.Sprintf("Perk %s added", perkNameArg))
	return err
}

func (i *Inventory) addItem(ctx ken.SubCommandContext) (err error) {
	inventory, err := i.imLazyMiddleware(ctx)
	if err != nil {
		return err
	}

	itemNameArg := ctx.Options().GetByName("name").StringValue()
	item, err := i.models.Items.GetByName(itemNameArg)
	if err != nil {
		return discord.SendSilentError(
			ctx,
			fmt.Sprint("Cannot find Item: ", itemNameArg),
			"Verify if the item exists.",
		)
	}

	inventory.Items = append(inventory.Items, item.Name)
	err = i.models.Inventories.UpdateItems(inventory)
	if err != nil {
		log.Println(err)
		return discord.SendSilentError(
			ctx,
			"Failed to add item",
			"Alex is a bad programmer, and this is his fault.",
		)
	}
	err = i.updateInventoryMessage(ctx, inventory)
	if err != nil {
		log.Println(err)
	}

	err = ctx.RespondMessage(fmt.Sprintf("Item %s added", itemNameArg))
	return err
}

func (i *Inventory) addStatus(ctx ken.SubCommandContext) (err error) {
	inventory, err := i.imLazyMiddleware(ctx)
	if err != nil {
		return err
	}

	statusNameArg := ctx.Options().GetByName("name").StringValue()
	status, err := i.models.Statuses.GetByName(statusNameArg)
	if err != nil {
		return discord.SendSilentError(
			ctx,
			fmt.Sprint("Cannot find Status: ", statusNameArg),
			"Verify if the status exists.",
		)
	}

	inventory.Statuses = append(inventory.Statuses, status.Name)
	err = i.models.Inventories.UpdateStatuses(inventory)
	if err != nil {
		log.Println(err)
		return discord.SendSilentError(
			ctx,
			"Failed to add status",
			"Alex is a bad programmer, and this is his fault.",
		)
	}
	err = i.updateInventoryMessage(ctx, inventory)
	if err != nil {
		log.Println(err)
	}

	err = ctx.RespondMessage(fmt.Sprintf("Status %s added", statusNameArg))
	return err
}

func (i *Inventory) addImmunity(ctx ken.SubCommandContext) (err error) {
	inventory, err := i.imLazyMiddleware(ctx)
	if err != nil {
		log.Println(err)
		return err
	}

	immunityNameArg := ctx.Options().GetByName("name").StringValue()

	for _, v := range inventory.Immunities {
		if strings.EqualFold(v, immunityNameArg) {
			return discord.SendSilentError(
				ctx,
				fmt.Sprintf("Immunity %s already exists in inventory", immunityNameArg),
				"Did you mean to remove the immunity?")
		}
	}

	inventory.Immunities = append(inventory.Immunities, immunityNameArg)
	err = i.models.Inventories.UpdateImmunities(inventory)
	if err != nil {
		log.Println(err)
		return discord.SendSilentError(
			ctx,
			"Failed to add immunity",
			"Alex is a bad programmer, and this is his fault.",
		)
	}

	err = i.updateInventoryMessage(ctx, inventory)
	if err != nil {
		log.Println(err)
	}

	err = ctx.RespondMessage(fmt.Sprintf("Immunity %s added", immunityNameArg))
	return err
}

func (i *Inventory) addEffect(ctx ken.SubCommandContext) (err error) {
	inventory, err := i.imLazyMiddleware(ctx)
	if err != nil {
		log.Println(err)
		return err
	}

	effectNameArg := ctx.Options().GetByName("name").StringValue()

	for _, v := range inventory.Effects {
		if strings.EqualFold(v, effectNameArg) {
			return discord.SendSilentError(
				ctx,
				fmt.Sprintf("Effect %s already exists in inventory", effectNameArg),
				"Did you mean to remove the effect?")
		}
	}

	inventory.Effects = append(inventory.Effects, effectNameArg)
	err = i.models.Inventories.UpdateEffects(inventory)
	if err != nil {
		log.Println(err)
		return discord.SendSilentError(
			ctx,
			"Failed to add effect",
			"Alex is a bad programmer, and this is his fault.",
		)
	}

	err = i.updateInventoryMessage(ctx, inventory)
	if err != nil {
		log.Println(err)
	}

	err = ctx.RespondMessage(fmt.Sprintf("Effect %s added", effectNameArg))
	return err
}

func (i *Inventory) addCoins(ctx ken.SubCommandContext) (err error) {
	inventory, err := i.imLazyMiddleware(ctx)
	if err != nil {
		log.Println(err)
		return err
	}

	coinsArg := ctx.Options().GetByName("amount").IntValue()

	inventory.Coins = inventory.Coins + coinsArg
	err = i.models.Inventories.Update(inventory)
	if err != nil {
		log.Println(err)
		return discord.SendSilentError(
			ctx,
			"Failed to add coins",
			"Alex is a bad programmer, and this is his fault.",
		)
	}

	err = i.updateInventoryMessage(ctx, inventory)
	if err != nil {
		log.Println(err)
	}

	return ctx.RespondMessage(
		fmt.Sprintf(
			"Added %d coins\n %d => %d",
			coinsArg,
			inventory.Coins-coinsArg,
			inventory.Coins,
		),
	)
}

func (i *Inventory) addWhitelist(ctx ken.SubCommandContext) (err error) {
	ctx.SetEphemeral(true)
	if !discord.IsAdminRole(ctx, discord.AdminRoles...) {
		err = discord.SendSilentError(
			ctx,
			"Unauthorized",
			"You are not authorized to use this command.",
		)
		return err
	}

	channelArg := ctx.Options().GetByName("channel").ChannelValue(ctx)

	whitelistChannels, err := i.models.Whitelists.GetAll()
	if err != nil {
		discord.SendSilentError(ctx, "Cannot find any whitelisted channels",
			"Verify if there are any whitelisted channels. via /inventory whitelist list",
		)

		return err
	}

	for _, wc := range whitelistChannels {
		if wc.ChannelID == channelArg.ID {
			err = discord.SendSilentError(
				ctx,
				"Error Updating Whitelists",
				"Channel already whitelisted",
			)
			return err
		}
	}

	err = i.models.Whitelists.Insert(&data.Whitelist{
		ChannelID:   channelArg.ID,
		GuildID:     ctx.GetEvent().GuildID,
		ChannelName: channelArg.Name,
	})
	if err != nil {
		err = discord.SendSilentError(
			ctx,
			"Failed to add channel to whitelist",
			"Alex is a bad programmer",
		)
		return err
	}

	err = ctx.RespondMessage("Channel added to whitelist.")
	return err
}

func (i *Inventory) addCoinBonus(ctx ken.SubCommandContext) (err error) {
	inventory, err := i.imLazyMiddleware(ctx)
	if err != nil {
		log.Println(err)
		return err
	}

	coinBonusArg := ctx.Options().GetByName("amount").StringValue()
	fCoinBonusArg, err := strconv.ParseFloat(coinBonusArg, 32)
	if err != nil {
		log.Println(err)
		return discord.SendSilentError(
			ctx,
			"Failed to add coin bonus",
			"Unable to parse coin bonus",
		)
	}

	// 2.5 -> 0.025
	inventory.CoinBonus += (float32(fCoinBonusArg) / 100)
	err = i.models.Inventories.Update(inventory)
	if err != nil {
		log.Println(err)
		return discord.SendSilentError(
			ctx,
			"Failed to add coin bonus",
			"Alex is a bad programmer, and this is his fault.",
		)
	}
	err = i.updateInventoryMessage(ctx, inventory)
	if err != nil {
		log.Println(err)
	}

	err = ctx.RespondMessage(
		fmt.Sprintf(
			"Added %s coin bonus\n %f => %f",
			coinBonusArg,
			inventory.CoinBonus-float32(fCoinBonusArg),
			inventory.CoinBonus,
		),
	)
	return err
}
