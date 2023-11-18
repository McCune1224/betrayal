package inventory

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/mccune1224/betrayal/internal/data"
	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/mccune1224/betrayal/internal/services/inventory"
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
	inventory, err := Fetch(ctx, i.models, true)
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

	ability, err := i.models.Abilities.GetAnyAbilitybyFuzzy(abilityNameArg)
	if err != nil {
		return discord.ErrorMessage(
			ctx,
			fmt.Sprint("Cannot find Ability: ", abilityNameArg),
			"Verify if the ability exists.",
		)
	}
	UpsertAA(inventory, ability, chargeAmount)
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
		"Any Ability Added",
		fmt.Sprintf("Any Ability %s added", abilityNameArg),
	)
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
	inventory, err := Fetch(ctx, i.models, true)
	if err != nil {
		if errors.Is(err, ErrNotAuthorized) {
			return discord.NotAdminError(ctx)
		}
		return discord.ErrorMessage(ctx, "Failed to find inventory.", "If not in confessional, please specify a user")
	}

	itemNameArg := ctx.Options().GetByName("name").StringValue()
	item, err := i.models.Items.GetByFuzzy(itemNameArg)
	if err != nil {
		return discord.ErrorMessage(
			ctx,
			fmt.Sprint("Cannot find Item: ", itemNameArg),
			"Verify if the item exists.",
		)
	}

	inventory.Items = append(inventory.Items, item.Name)
	err = i.models.Inventories.UpdateItems(inventory)
	if err != nil {
		log.Println(err)
		return discord.ErrorMessage(
			ctx,
			"Failed to add item",
			"Alex is a bad programmer, and this is his fault.",
		)
	}
	err = i.updateInventoryMessage(ctx, inventory)
	if err != nil {
		log.Println(err)
	}

	err = discord.SuccessfulMessage(ctx, "Added Item", fmt.Sprintf("Item %s added", itemNameArg))
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
	inventory, err := Fetch(ctx, i.models, true)
	if err != nil {
		if errors.Is(err, ErrNotAuthorized) {
			return discord.NotAdminError(ctx)
		}
		return discord.ErrorMessage(ctx, "Failed to find inventory.", "If not in confessional, please specify a user")
	}
	immunityNameArg := ctx.Options().GetByName("name").StringValue()

	for _, v := range inventory.Immunities {
		if strings.EqualFold(v, immunityNameArg) {
			return discord.ErrorMessage(
				ctx,
				fmt.Sprintf("Immunity %s already exists in inventory", immunityNameArg),
				"Did you mean to remove the immunity?")
		}
	}

	inventory.Immunities = append(inventory.Immunities, immunityNameArg)
	err = i.models.Inventories.UpdateImmunities(inventory)
	if err != nil {
		log.Println(err)
		return discord.ErrorMessage(
			ctx,
			"Failed to add immunity",
			"Alex is a bad programmer, and this is his fault.",
		)
	}

	err = i.updateInventoryMessage(ctx, inventory)
	if err != nil {
		log.Println(err)
	}

	err = discord.SuccessfulMessage(
		ctx,
		"Added Immunity",
		fmt.Sprintf("Immunity %s added", immunityNameArg),
	)
	return err
}

func (i *Inventory) addEffect(ctx ken.SubCommandContext) (err error) {
	ctx.SetEphemeral(false)
	inventory, err := Fetch(ctx, i.models, true)
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

	for _, v := range inventory.Effects {
		if strings.EqualFold(v, effectNameArg) {
			return discord.ErrorMessage(
				ctx,
				fmt.Sprintf("Effect %s already exists in inventory", effectNameArg),
				"Did you mean to remove the effect?")
		}
	}

	inventory.Effects = append(inventory.Effects, effectNameArg)
	err = i.models.Inventories.UpdateEffects(inventory)
	if err != nil {
		log.Println(err)
		return discord.ErrorMessage(
			ctx,
			"Failed to add effect",
			"Alex is a bad programmer, and this is his fault.",
		)
	}

	err = i.updateInventoryMessage(ctx, inventory)
	if err != nil {
		log.Println(err)
	}

	if dur > 0 {
		s := ctx.GetSession()
		err = i.scheduler.ScheduleEffect(effectNameArg, inventory, dur, func() {
			for k, v := range inventory.Effects {
				if strings.EqualFold(effectNameArg, v) {
					inventory.Effects = append(inventory.Effects[:k], inventory.Effects[k+1:]...)
					err = i.models.Inventories.UpdateEffects(inventory)
					if err != nil {
						log.Println(err)
						msg := &discordgo.MessageEmbed{
							Title:       "Failed to remove effect",
							Description: fmt.Sprintf("Failed to remove effect %s from inventory", effectNameArg),
							Fields: []*discordgo.MessageEmbedField{
								{
									Name:  "Error",
									Value: err.Error(),
								},
							},
						}
						_, err := s.ChannelMessageSendEmbed(ctx.GetEvent().ChannelID, msg)
						if err != nil {
							log.Println(err)
							return
						}
					}
				}
			}

			msg := discordgo.MessageEmbed{
				Title:       "Effect Expired",
				Description: fmt.Sprintf("Effect %s has expired", effectNameArg),
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:  "Target Inventory",
						Value: fmt.Sprintf("<@%s>", inventory.DiscordID),
					},
					{
						Name:  "Target Channel",
						Value: fmt.Sprintf("<#%s>", inventory.UserPinChannel),
					},
				},
				Color:     0x00ff00,
				Timestamp: time.Now().Format(time.RFC3339),
			}
			_, err := s.ChannelMessageSendEmbed(ctx.GetEvent().ChannelID, &msg)
			if err != nil {
				log.Println(err)
			}
		})
		if err != nil {
			log.Println(err)
			return discord.ErrorMessage(ctx, "Failed to schedule effect", err.Error())
		}
		i.updateInventoryMessage(ctx, inventory)
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
	return discord.SuccessfulMessage(ctx, "Added Coins", fmt.Sprintf("Added %d coins\n %d => %d", coinsArg, newCoins-coinsArg, newCoins))
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
		fmt.Sprintf("%.2f => %.2f", float32(int(old*100))/100, float32(int(handler.GetInventory().CoinBonus*100))/100))
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
	return discord.SuccessfulMessage(ctx, "Added Luck", fmt.Sprintf("%d => %d", old, handler.GetInventory().Luck))
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
