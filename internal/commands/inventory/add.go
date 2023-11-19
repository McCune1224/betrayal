package inventory

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/mccune1224/betrayal/internal/data"
	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/mccune1224/betrayal/internal/services/inventory"
	"github.com/mccune1224/betrayal/internal/util"
	"github.com/zekrotja/ken"
)

func (i *Inventory) addAbility(ctx ken.SubCommandContext) (err error) {
	inventory, err := Fetch(ctx, i.models, true)
	if err != nil {
		if errors.Is(err, ErrNotAuthorized) {
			return discord.NotAdminError(ctx)
		}
		return discord.ErrorMessage(ctx, "Failed to find inventory.", "If not in confessional, please specify a user")
	}
	ctx.SetEphemeral(false)
	abilityNameArg := ctx.Options().GetByName("name").StringValue()
	chargesArg, ok := ctx.Options().GetByNameOptional("charges")
	charge := 1
	if ok {
		charge = int(chargesArg.IntValue())
	}

	ability, err := i.models.Abilities.GetByFuzzy(abilityNameArg)
	if err != nil {
		return discord.ErrorMessage(
			ctx,
			fmt.Sprint("Cannot find Ability: ", abilityNameArg),
			"Verify if the ability exists.",
		)
	}
	UpsertAbility(inventory, ability, charge)
	err = i.models.Inventories.Update(inventory)
	if err != nil {
		log.Println(err)
		return discord.ErrorMessage(
			ctx,
			"Failed to add ability",
			"Alex is a bad programmer, and this is his fault.",
		)
	}

	err = i.updateInventoryMessage(ctx, inventory)
	if err != nil {
		return err
	}

	err = discord.SuccessfulMessage(
		ctx,
		"Base Ability Added",
		fmt.Sprintf("Base Ability %s added", abilityNameArg),
	)
	return err
}

func (i *Inventory) addAnyAbility(ctx ken.SubCommandContext) (err error) {
	handler, err := FetchHandler(ctx, i.models, true)
	if err != nil {
		if errors.Is(err, ErrNotAuthorized) {
			return discord.NotAdminError(ctx)
		}
		return discord.ErrorMessage(ctx, "Failed to find inventory.", "If not in confessional, please specify a user")
	}
	abilityNameArg := ctx.Options().GetByName("name").StringValue()
	chargesArg, ok := ctx.Options().GetByNameOptional("charges")
	chargeAmount := 1
	if ok {
		chargeAmount = int(chargesArg.IntValue())
	}

	abStr, err := handler.AddAnyAbility(abilityNameArg, chargeAmount)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, fmt.Sprintf("Failed to insert any ability %s", abilityNameArg))
	}
	err = UpdateInventoryMessage(ctx.GetSession(), handler.GetInventory())
	if err != nil {
		return err
	}

	err = discord.SuccessfulMessage(ctx, fmt.Sprintf("Added Any Ability %s with %d charges total", abStr.GetName(), chargeAmount),
		fmt.Sprintf("Added for %s", discord.MentionUser(handler.GetInventory().DiscordID)))
	return err
}

func (i *Inventory) addPerk(ctx ken.SubCommandContext) (err error) {
	ctx.SetEphemeral(false)
	inventory, err := Fetch(ctx, i.models, true)
	if err != nil {
		if errors.Is(err, ErrNotAuthorized) {
			return discord.NotAdminError(ctx)
		}
		return discord.ErrorMessage(ctx, "Failed to find inventory.", "If not in confessional, please specify a user")
	}
	perkNameArg := ctx.Options().GetByName("name").StringValue()
	perk, err := i.models.Perks.GetByName(perkNameArg)
	if err != nil {
		return discord.ErrorMessage(
			ctx,
			fmt.Sprint("Cannot find Perk: ", perkNameArg),
			"Verify if the perk exists.",
		)
	}

	for _, p := range inventory.Perks {
		if p == perk.Name {
			return discord.ErrorMessage(
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
		return discord.ErrorMessage(
			ctx,
			"Failed to add perk",
			"Alex is a bad programmer, and this is his fault.",
		)
	}
	err = i.updateInventoryMessage(ctx, inventory)
	if err != nil {
		log.Println(err)
	}

	err = discord.SuccessfulMessage(ctx,
		"Perk Added",
		fmt.Sprintf("Perk %s added", perkNameArg))
	return err
}

func (i *Inventory) addItem(ctx ken.SubCommandContext) (err error) {
	handler, err := FetchHandler(ctx, i.models, true)
	if err != nil {
		if errors.Is(err, ErrNotAuthorized) {
			return discord.NotAdminError(ctx)
		}
		return discord.ErrorMessage(ctx, "Failed to find inventory.", "If not in confessional, please specify a user")
	}

	itemNameArg := ctx.Options().GetByName("name").StringValue()
	item, err := handler.AddItem(itemNameArg)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Failed to add item")
	}
	err = i.updateInventoryMessage(ctx, handler.GetInventory())
	if err != nil {
		log.Println(err)
	}

	err = discord.SuccessfulMessage(ctx, fmt.Sprintf("Added Item %s", item),
		fmt.Sprintf("Added item for %s", discord.MentionUser(handler.GetInventory().DiscordID)))
	return err
}

func (i *Inventory) addStatus(ctx ken.SubCommandContext) (err error) {
	ctx.SetEphemeral(false)
	inventory, err := Fetch(ctx, i.models, true)
	if err != nil {
		if errors.Is(err, ErrNotAuthorized) {
			return discord.NotAdminError(ctx)
		}
		return discord.ErrorMessage(ctx, "Failed to find inventory.", "If not in confessional, please specify a user")
	}
	statusNameArg := ctx.Options().GetByName("name").StringValue()
	status, err := i.models.Statuses.GetByName(statusNameArg)
	if err != nil {
		return discord.ErrorMessage(
			ctx,
			fmt.Sprint("Cannot find Status: ", statusNameArg),
			"Verify if the status exists.",
		)
	}

	inventory.Statuses = append(inventory.Statuses, status.Name)
	err = i.models.Inventories.UpdateStatuses(inventory)
	if err != nil {
		log.Println(err)
		return discord.ErrorMessage(
			ctx,
			"Failed to add status",
			"Alex is a bad programmer, and this is his fault.",
		)
	}
	err = i.updateInventoryMessage(ctx, inventory)
	if err != nil {
		log.Println(err)
	}

	err = discord.SuccessfulMessage(
		ctx,
		"Added Status",
		fmt.Sprintf("Status %s added", statusNameArg),
	)
	return err
}

func (i *Inventory) addImmunity(ctx ken.SubCommandContext) (err error) {
	ctx.SetEphemeral(false)
	handler, err := FetchHandler(ctx, i.models, true)
	if err != nil {
		if errors.Is(err, ErrNotAuthorized) {
			return discord.NotAdminError(ctx)
		}
		return discord.ErrorMessage(ctx, "Failed to find inventory.", "If not in confessional, please specify a user")
	}
	immunityNameArg := ctx.Options().GetByName("name").StringValue()

	best, err := handler.AddImmunity(immunityNameArg)
	if err != nil {
		if errors.Is(err, inventory.ErrImmunityExists) {
			return discord.ErrorMessage(ctx, "Immunity already exists", fmt.Sprintf("Error %s already in inventory", immunityNameArg))
		}
		return discord.ErrorMessage(ctx, "Immunity not found", fmt.Sprintf("%s not found", immunityNameArg))
	}
	err = UpdateInventoryMessage(ctx.GetSession(), handler.GetInventory())
	if err != nil {
		log.Println(err)
	}

	err = discord.SuccessfulMessage(ctx, fmt.Sprintf("Immunity %s Removed", best),
		fmt.Sprintf("Removed immunity for %s", discord.MentionUser(handler.GetInventory().DiscordID)))
	return err
}

func (i *Inventory) addEffect(ctx ken.SubCommandContext) (err error) {
	ctx.SetEphemeral(false)
	handler, err := FetchHandler(ctx, i.models, true)
	if err != nil {
		if errors.Is(err, ErrNotAuthorized) {
			return discord.NotAdminError(ctx)
		}
		return discord.ErrorMessage(ctx, "Failed to find inventory.", "If not in confessional, please specify a user")
	}
	dur := time.Duration(0)
	effectNameArg := ctx.Options().GetByName("name").StringValue()
	durationArg, ok := ctx.Options().GetByNameOptional("duration")
	if ok {
		dur, err = time.ParseDuration(durationArg.StringValue())
		if err != nil {
			return discord.ErrorMessage(ctx, "Failed to parse duration", err.Error())
		}
	}

	best, err := handler.AddEffect(effectNameArg)
	if err != nil {
		if errors.Is(err, inventory.ErrEffectAlreadyExists) {
			return discord.ErrorMessage(ctx, "Effect already exists", fmt.Sprintf("Error %s already in inventory", effectNameArg))
		}
		log.Println(err)
		return discord.AlexError(ctx, fmt.Sprintf("Failed to add effect %s", best))
	}

	err = i.updateInventoryMessage(ctx, handler.GetInventory())
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Failed to update inventory message")
	}

	start := util.GetEstTimeStamp()

	// FIXME: This needs to somehow be moved to the scheduler package and just get arguments from here
	// For now, I'm just going to copy the code here because I don't want to deal with import cycle issues
	if dur > 0 {
		err = i.scheduler.ScheduleEffect(effectNameArg, handler.GetInventory(), dur, func() {
			s := ctx.GetSession()
			inv, err := i.models.Inventories.GetByDiscordID(handler.GetInventory().DiscordID)
			if err != nil {
				log.Println(err)
				s.ChannelMessageSend(ctx.GetEvent().ChannelID, "Failed to find inventory for effect expiration")
				return
			}
			handler := inventory.InitInventoryHandler(i.models, inv)
			best, err := handler.RemoveEffect(effectNameArg)
			if err != nil {
				if errors.Is(err, inventory.ErrEffectNotFound) {
					s.ChannelMessageSend(ctx.GetEvent().ChannelID, fmt.Sprintf("Effect %s not found", effectNameArg))
					return
				}
				log.Println(err)
				s.ChannelMessageSend(ctx.GetEvent().ChannelID, fmt.Sprintf("Failed to remove timed effect %s", effectNameArg))
				return
			}
			msg := discordgo.MessageEmbed{
				Title:       "Effect Expired",
				Description: fmt.Sprintf("Effect %s has expired", best),
				Fields: []*discordgo.MessageEmbedField{
					{
						Value: fmt.Sprintf("Timer started at %s", start),
					},
					{
						Value: fmt.Sprintf("Timer ended at %s", util.GetEstTimeStamp()),
					},
				},
				Color:     discord.ColorThemeOrange,
				Timestamp: time.Now().Format(time.RFC3339),
			}
			_, err = s.ChannelMessageSendEmbed(ctx.GetEvent().ChannelID, &msg)
			if err != nil {
				log.Println(err)
			}
			err = UpdateInventoryMessage(s, handler.GetInventory())
			if err != nil {
				log.Println(err)
				return
			}
		})
		if err != nil {
			log.Println(err)
			return discord.ErrorMessage(ctx, "Failed to schedule effect", err.Error())
		}
	}

	err = discord.SuccessfulMessage(
		ctx,
		"Added Effect",
		fmt.Sprintf("Effect %s added", effectNameArg),
	)
	return err
}

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

func (i *Inventory) addWhitelist(ctx ken.SubCommandContext) (err error) {
	ctx.SetEphemeral(false)
	if !discord.IsAdminRole(ctx, discord.AdminRoles...) {
		discord.NotAdminError(ctx)
	}

	channelArg := ctx.Options().GetByName("channel").ChannelValue(ctx)

	whitelistChannels, err := i.models.Whitelists.GetAll()
	if err != nil {
		discord.ErrorMessage(ctx, "Cannot find any whitelisted channels",
			"Verify if there are any whitelisted channels. via /inventory whitelist list",
		)

		return err
	}

	for _, wc := range whitelistChannels {
		if wc.ChannelID == channelArg.ID {
			err = discord.ErrorMessage(
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
		err = discord.ErrorMessage(
			ctx,
			"Failed to add channel to whitelist",
			"Alex is a bad programmer",
		)
		return err
	}

	err = discord.SuccessfulMessage(
		ctx,
		"Added Channel",
		fmt.Sprintf("Added %s to whitelist", discord.MentionChannel(channelArg.ID)),
	)
	return err
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

func (i *Inventory) addItemLimit(ctx ken.SubCommandContext) (err error) {
	handler, err := FetchHandler(ctx, i.models, true)
	if err != nil {
		log.Println(err)
		return discord.ErrorMessage(ctx, "Failed to find inventory.", "If not in confessional, please specify a user")
	}
	ctx.SetEphemeral(false)
	itemLimitArg := ctx.Options().GetByName("amount").IntValue()

	err = handler.AddLimit(int(itemLimitArg))
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Failed to add item limit")
	}
	err = i.updateInventoryMessage(ctx, handler.GetInventory())
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Failed to update inventory message")
	}
	return discord.SuccessfulMessage(ctx, "Item Limit Updated", fmt.Sprintf("Item limit set to %d", handler.GetInventory().ItemLimit))
}

func (i *Inventory) addLuck(ctx ken.SubCommandContext) (err error) {
	handler, err := FetchHandler(ctx, i.models, true)
	if err != nil {
		if errors.Is(err, ErrNotAuthorized) {
			return discord.NotAdminError(ctx)
		}
		return discord.ErrorMessage(ctx, "Failed to find inventory.", "If not in confessional, please specify a user")
	}

	ctx.SetEphemeral(true)
	luckArg := ctx.Options().GetByName("amount").IntValue()
	old := handler.GetInventory().Luck
	err = handler.AddLuck(luckArg)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Failed to add luck")
	}
	err = i.updateInventoryMessage(ctx, handler.GetInventory())
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Failed to update inventory message")
	}
	return discord.SuccessfulMessage(ctx, fmt.Sprintf("Added %d Luck", luckArg),
		fmt.Sprintf("%d => %d for ", old, handler.GetInventory().Luck))
}

func (i *Inventory) addNote(ctx ken.SubCommandContext) (err error) {
	handler, err := FetchHandler(ctx, i.models, true)
	if err != nil {
		if errors.Is(err, ErrNotAuthorized) {
			return discord.NotAdminError(ctx)
		}
		return discord.ErrorMessage(ctx, "Failed to find inventory.", "If not in confessional, please specify a user")
	}
	noteArg := ctx.Options().GetByName("message").StringValue()
	ctx.SetEphemeral(true)

	err = handler.AddNote(noteArg)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Failed to add note")
	}
	err = UpdateInventoryMessage(ctx.GetSession(), handler.GetInventory())
	if err != nil {
		log.Println(err)
	}

	return discord.SuccessfulMessage(ctx, "Added Note", fmt.Sprintf("Added note %s", noteArg))
}
