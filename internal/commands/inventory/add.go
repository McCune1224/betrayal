package inventory

import (
	"fmt"
	"log"

	"github.com/mccune1224/betrayal/internal/data"
	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/zekrotja/ken"
)

func (i *Inventory) addAnyAbility(ctx ken.SubCommandContext) (err error) {
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
	ctx.SetEphemeral(true)
	userArg := ctx.Options().GetByName("user").UserValue(ctx)
	perkNameArg := ctx.Options().GetByName("name").StringValue()

	inventory, err := i.models.Inventories.GetByDiscordID(userArg.ID)
	if err != nil {
		return discord.SendSilentError(
			ctx,
			fmt.Sprint("Cannot find Inventory for user: ", userArg.Username),
			"Verify if the user has an inventory. via /inventory list",
		)
	}

	if !i.authorized(ctx, inventory) {
		return discord.SendSilentError(
			ctx,
			"Unauthorized",
			"You are not authorized to use this command.",
		)
	}

	inventory.Perks = append(inventory.Perks, perkNameArg)
	err = i.models.Inventories.Update(inventory)
	if err != nil {
		log.Println(err)
		return discord.SendSilentError(
			ctx,
			"Failed to add perk",
			"Alex is a bad programmer, and this is his fault.",
		)
	}

	err = ctx.RespondMessage(fmt.Sprintf("Perk %s added to %s", perkNameArg, userArg.Username))
	return err
}

func (i *Inventory) addItem(ctx ken.SubCommandContext) (err error) {
	if err = ctx.Defer(); err != nil {
		return err
	}
	ctx.SetEphemeral(true)
	err = ctx.RespondMessage("Command not implemented go bug Alex")
	ctx.SetEphemeral(false)
	return err
}

func (i *Inventory) addStatus(ctx ken.SubCommandContext) (err error) {
	if err = ctx.Defer(); err != nil {
		return err
	}
	ctx.SetEphemeral(true)
	err = ctx.RespondMessage("Command not implemented go bug Alex")
	ctx.SetEphemeral(false)
	return err
}

func (i *Inventory) addImmunity(ctx ken.SubCommandContext) (err error) {
	if err = ctx.Defer(); err != nil {
		return err
	}
	ctx.SetEphemeral(true)
	err = ctx.RespondMessage("Command not implemented go bug Alex")
	ctx.SetEphemeral(false)
	return err
}

func (i *Inventory) addCoins(ctx ken.SubCommandContext) (err error) {
	if err = ctx.Defer(); err != nil {
		return err
	}
	ctx.SetEphemeral(true)
	err = ctx.RespondMessage("Command not implemented go bug Alex")
	ctx.SetEphemeral(false)
	return err
}

func (i *Inventory) addNote(ctx ken.SubCommandContext) (err error) {
	ctx.SetEphemeral(true)
	if !discord.IsAdminRole(ctx, discord.AdminRoles...) {
		err = discord.SendSilentError(
			ctx,
			"Unauthorized",
			"You are not authorized to use this command.",
		)
		return err
	}

	userArg := ctx.Options().GetByName("user").UserValue(ctx)
	dataArg := ctx.Options().GetByName("data").StringValue()

	inventory, err := i.models.Inventories.GetByDiscordID(userArg.ID)
	if err != nil {
		discord.SendSilentError(ctx, "Cannot find Inventory",
			"Verify if the user has an inventory. via /inventory list",
		)
		return err
	}

	inventory.Notes = append(inventory.Notes, dataArg)
	err = i.models.Inventories.Update(inventory)
	if err != nil {
		log.Println(err)
		discord.SendSilentError(
			ctx,
			"Failed to append note",
			"Alex is a bad programmer, and this is his fault.",
		)
		return err
	}

	err = ctx.RespondMessage("Note added.")
	return err
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
