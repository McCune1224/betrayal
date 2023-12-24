package roll

import (
	"errors"
	"fmt"
	"log"
	"math/rand"

	"github.com/bwmarrin/discordgo"
	"github.com/mccune1224/betrayal/internal/commands/inventory"
	"github.com/mccune1224/betrayal/internal/data"
	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/zekrotja/ken"
)

func (r *Roll) luckItemRain(ctx ken.SubCommandContext) (err error) {
	if err = ctx.Defer(); err != nil {
		log.Println(err)
		return err
	}
	if !discord.IsAdminRole(ctx, discord.AdminRoles...) {
		return discord.NotAdminError(ctx)
	}
	inv, err := inventory.Fetch(ctx, r.models, true)
	if err != nil {
		if errors.Is(err, inventory.ErrNotAuthorized) {
			return discord.NotAdminError(ctx)
		}
		return discord.ErrorMessage(ctx, "Failed to find inventory.", "If not in confessional, please specify a user")
	}
	luckLevel := inv.Luck
	luckArg, ok := ctx.Options().GetByNameOptional("luck")
	if ok {
		luckLevel = luckArg.IntValue()
	}
	rollAmount := rand.Intn(3) + 1
	newItems := []*data.Item{}
	for i := 0; i < rollAmount; i++ {
		roll := RollLuck(float64(luckLevel), rand.Float64())
		item, err := r.getRandomItem(roll)
		if err != nil {
			log.Println(err)
			return discord.AlexError(ctx, "Failed to get random item")
		}
		newItems = append(newItems, item)
	}

	title := fmt.Sprintf("%s Item Rain Incoming %s", discord.EmojiItem, discord.EmojiItem)
	desc := fmt.Sprintf(
		"Rolled %d Items from Item Rain!\n",
		len(newItems),
	)
	fields := []*discordgo.MessageEmbedField{}

	for _, item := range newItems {
		inv.Items = append(inv.Items, item.Name)
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   fmt.Sprintf("%s (%s)", discord.Bold(item.Name), item.Rarity),
			Value:  item.Description,
			Inline: true,
		})
	}
	footerMessage := ""
	if len(inv.Items) > inv.ItemLimit {
		footerMessage += fmt.Sprintf("\n %s inventory overflow [%d/%d] %s",
			discord.EmojiWarning,
			len(inv.Items),
			inv.ItemLimit,
			discord.EmojiWarning,
		)
	} else {
		footerMessage += fmt.Sprintf("\n %s adding %d items to inventory %s",
			discord.EmojiSuccess,
			rollAmount,
			discord.EmojiSuccess)
	}

	embdRain := &discordgo.MessageEmbed{
		Title:       title,
		Description: desc,
		Fields:      fields,
		Footer:      &discordgo.MessageEmbedFooter{Text: footerMessage},
	}

	b := ctx.FollowUpEmbed(embdRain)

	// WARNING:
	// ctx gets redeclared in button component, so need to save it here
	// Im sure this closure will come back to haunt me...Too Bad!
	sctx := ctx
	b.AddComponents(func(cb *ken.ComponentBuilder) {
		cb.AddActionsRow(func(b ken.ComponentAssembler) {
			b.Add(discordgo.Button{
				Style:    discordgo.SuccessButton,
				CustomID: "confirm-item-rain",
				Label:    "Confirm",
			}, func(ctx ken.ComponentContext) bool {
				// rare occurance where inbetween this accepting if the inventory is updated the item list is not updated
				// so re-process the current item list and add the new items
				currInv, err := inventory.Fetch(sctx, r.models, true)
				if err != nil {
					log.Println(err)
					return true
				}
				for _, item := range newItems {
					currInv.Items = append(currInv.Items, item.Name)
				}
				err = r.models.Inventories.UpdateItems(currInv)
				if err != nil {
					log.Println(err)
					return true
				}
				err = r.models.Inventories.UpdateItemLimit(currInv)
				if err != nil {
					log.Println(err)
					return true
				}
				newFooterMessage := ""
				if len(currInv.Items) > currInv.ItemLimit {
					newFooterMessage += fmt.Sprintf("\n %s inventory overflow [%d/%d] %s",
						discord.EmojiWarning,
						len(currInv.Items),
						currInv.ItemLimit,
						discord.EmojiWarning,
					)
				} else {
					newFooterMessage += fmt.Sprintf("\n %s adding %d items to inventory %s",
						discord.EmojiSuccess,
						rollAmount,
						discord.EmojiSuccess)
				}
				inventory.UpdateInventoryMessage(sctx.GetSession(), currInv)
				embdRain.Footer = &discordgo.MessageEmbedFooter{Text: newFooterMessage}
				_, err = ctx.GetSession().ChannelMessageSendEmbed(currInv.UserPinChannel, embdRain)
				if err != nil {
					log.Println(err)
					return true
				}
				discord.SuccessfulMessage(sctx, fmt.Sprintf("Item Rain Sent to %s", discord.MentionChannel(currInv.UserPinChannel)),
					fmt.Sprintf("Approved by %s", ctx.User().Username))
				return true
			}, true)
			b.Add(discordgo.Button{
				Style:    discordgo.DangerButton,
				CustomID: "decline-item-rain",
				Label:    "Decline",
			},
				func(ctx ken.ComponentContext) bool {
					discord.SuccessfulMessage(sctx, fmt.Sprintf("Declined Item Rain for %s", discord.MentionChannel(inv.UserPinChannel)),
						fmt.Sprintf("Declined by %s", ctx.User().Username))
					return true
				}, true)
		}, true).
			Condition(func(cctx ken.ComponentContext) bool {
				return true
			})
	})

	fum := b.Send()
	return fum.Error
}

func (r *Roll) luckPowerDrop(ctx ken.SubCommandContext) (err error) {
	if err = ctx.Defer(); err != nil {
		log.Println(err)
		return err
	}

	if !discord.IsAdminRole(ctx, discord.AdminRoles...) {
		return discord.NotAdminError(ctx)
	}

	inv, err := inventory.Fetch(ctx, r.models, true)
	if err != nil {
		return discord.ErrorMessage(
			ctx,
			"Failed to get inventory",
			"Are you in a whitelist or confessional channel?",
		)
	}
	luckLevel := inv.Luck
	luckArg, ok := ctx.Options().GetByNameOptional("luck")
	if ok {
		luckLevel = luckArg.IntValue()
	}
	rarity := RollLuck(float64(luckLevel), rand.Float64())
	aa, err := r.getRandomAnyAbility(inv.RoleName, rarity)
	if err != nil {
		log.Println(err)
		return discord.ErrorMessage(
			ctx,
			"Failed to get random any ability",
			"Alex is a bad programmer",
		)
	}

	embedPowerDrop := &discordgo.MessageEmbed{
		Title: fmt.Sprintf("%s Power Drop Incoming %s", discord.EmojiItem, discord.EmojiItem),
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   fmt.Sprintf("%s (%s)", aa.Name, aa.Rarity),
				Value:  aa.Description,
				Inline: true,
			},
		},
	}
	b := ctx.FollowUpEmbed(embedPowerDrop)

	// WARNING:
	// ctx gets redeclared in button component, so need to save it here
	// Im sure this closure will come back to haunt me...Too Bad!
	sctx := ctx
	b.AddComponents(func(cb *ken.ComponentBuilder) {
		cb.AddActionsRow(func(b ken.ComponentAssembler) {
			b.Add(discordgo.Button{
				Style:    discordgo.SuccessButton,
				CustomID: "confirm-power-drop",
				Label:    "Confirm",
			}, func(ctx ken.ComponentContext) bool {
				// rare occurance where inbetween this accepting if the inventory is updated the item list is not updated
				// so re-process the current item list and add the new items
				currInv, err := inventory.Fetch(sctx, r.models, true)
				if err != nil {
					log.Println(err)
					return true
				}
				if aa.RoleSpecific == currInv.RoleName {
					ab, err := r.models.Abilities.GetByFuzzy(aa.RoleSpecific)
					if err != nil {
						log.Println(err)
						return true
					}
					inventory.UpsertAbility(currInv, ab)
					err = r.models.Inventories.UpdateAbilities(currInv)
					if err != nil {
						log.Println(err)
						discord.AlexError(sctx, "Failed to update abilities")
						return true
					}
				} else {
					inventory.UpsertAA(currInv, aa)
					err = r.models.Inventories.UpdateAnyAbilities(currInv)
					if err != nil {
						log.Println(err)
						discord.AlexError(sctx, "Failed to update any abilities")
						return true
					}
				}
				inventory.UpdateInventoryMessage(sctx.GetSession(), currInv)
				_, err = ctx.GetSession().ChannelMessageSendEmbed(currInv.UserPinChannel, embedPowerDrop)
				if err != nil {
					log.Println(err)
					discord.AlexError(sctx, "Failed to send message")
					return true
				}
				_, err = ctx.GetSession().ChannelMessageSendEmbed(inv.UserPinChannel, embedPowerDrop)
				if err != nil {
					log.Println(err)
					discord.ErrorMessage(sctx, "Failed to send message", "Could not find user confessional")
					return true
				}

				discord.SuccessfulMessage(sctx, fmt.Sprintf("Power Drop Sent to %s", discord.MentionChannel(currInv.UserPinChannel)),
					fmt.Sprintf("Approved by %s", ctx.User().Username))
				return true
			}, true)
			b.Add(discordgo.Button{
				Style:    discordgo.DangerButton,
				CustomID: "decline-power-drop",
				Label:    "Decline",
			},
				func(ctx ken.ComponentContext) bool {
					discord.SuccessfulMessage(sctx, fmt.Sprintf("Declined Power Drop for %s", discord.MentionChannel(inv.UserPinChannel)), fmt.Sprintf("Declined by %s", ctx.User().Username))
					return true
				}, true)
		}, true).
			Condition(func(cctx ken.ComponentContext) bool {
				return true
			})
	})
	fum := b.Send()
	return fum.Error
}

// Get 1 Random Item and 1 Random AA
func (r *Roll) luckCarePackage(ctx ken.SubCommandContext) (err error) {
	if err = ctx.Defer(); err != nil {
		log.Println(err)
		return err
	}

	if !discord.IsAdminRole(ctx, discord.AdminRoles...) {
		return discord.NotAdminError(ctx)
	}

	inv, err := inventory.Fetch(ctx, r.models, true)
	if err != nil {
		if errors.Is(err, inventory.ErrNotAuthorized) {
			return discord.NotAdminError(ctx)
		}
		return discord.ErrorMessage(ctx, "Failed to find inventory.", "If not in confessional, please specify a user")
	}
	luckLevel := inv.Luck
	luckArg, ok := ctx.Options().GetByNameOptional("luck")
	if ok {
		luckLevel = luckArg.IntValue()
	}

	aRoll := RollLuck(float64(luckLevel), rand.Float64())
	iRoll := RollLuck(float64(luckLevel), rand.Float64())

	aa, err := r.getRandomAnyAbility(inv.RoleName, aRoll)
	if err != nil {
		return discord.ErrorMessage(ctx, "Error getting random ability", "Alex is a bad programmer")
	}

	item, err := r.models.Items.GetRandomByRarity(iRoll)
	if err != nil {
		log.Println(err)
		log.Println(err)
		return discord.ErrorMessage(
			ctx,
			"Failed to get Random Item",
			"Alex is a bad programmer",
		)
	}

	err = inventory.UpdateInventoryMessage(ctx.GetSession(), inv)
	if err != nil {
		log.Println(err)
		discord.SuccessfulMessage(
			ctx,
			"Failed to update inventory message",
			"Alex is a bad programmer",
		)
	}

	embedCarePackage := &discordgo.MessageEmbed{
		Title: fmt.Sprintf("%s Care Package Incoming %s", discord.EmojiItem, discord.EmojiItem),
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   fmt.Sprintf("Item: %s (%s)", item.Name, item.Rarity),
				Value:  item.Description,
				Inline: true,
			},
			{
				Name:   fmt.Sprintf("Any Ability: %s (%s)", aa.Name, aa.Rarity),
				Value:  aa.Description,
				Inline: true,
			},
		},
	}
	b := ctx.FollowUpEmbed(embedCarePackage)

	// WARNING:
	// ctx gets redeclared in button component, so need to save it here
	// Im sure this closure will come back to haunt me...Too Bad!
	sctx := ctx
	b.AddComponents(func(cb *ken.ComponentBuilder) {
		cb.AddActionsRow(func(b ken.ComponentAssembler) {
			b.Add(discordgo.Button{
				Style:    discordgo.SuccessButton,
				CustomID: "confirm-care-package",
				Label:    "Confirm",
			}, func(ctx ken.ComponentContext) bool {
				// rare occurance where inbetween this accepting if the inventory is updated the item list is not updated
				// so re-process the current item list and add the new items
				currInv, err := inventory.Fetch(sctx, r.models, true)
				if err != nil {
					log.Println(err)
					return true
				}
				if aa.RoleSpecific == currInv.RoleName {
					ab, err := r.models.Abilities.GetByFuzzy(aa.RoleSpecific)
					if err != nil {
						log.Println(err)
						discord.AlexError(sctx, "Failed to get ability")
						return true
					}
					inventory.UpsertAbility(currInv, ab)
					err = r.models.Inventories.UpdateAbilities(currInv)
					if err != nil {
						log.Println(err)
						discord.AlexError(sctx, "Failed to update abilities")
						return true
					}

				} else {
					inventory.UpsertAA(currInv, aa)
					err = r.models.Inventories.UpdateAnyAbilities(currInv)
					if err != nil {
						log.Println(err)
						discord.AlexError(sctx, "Failed to update any abilities")
						return true
					}
				}

				currInv.Items = append(currInv.Items, item.Name)
				err = r.models.Inventories.UpdateItems(currInv)
				if err != nil {
					log.Println(err)
					discord.AlexError(sctx, "Failed to update items")
					return true
				}

				inventory.UpdateInventoryMessage(sctx.GetSession(), currInv)
				_, err = ctx.GetSession().ChannelMessageSendEmbed(inv.UserPinChannel, embedCarePackage)
				if err != nil {
					log.Println(err)
					discord.ErrorMessage(sctx, "Failed to send message", "Could not find user confessional")
					return true
				}

				discord.SuccessfulMessage(sctx, fmt.Sprintf("Care Package Sent to %s", discord.MentionChannel(currInv.UserPinChannel)),
					fmt.Sprintf("Approved by %s", ctx.User().Username))
				return true
			}, true)
			b.Add(discordgo.Button{
				Style:    discordgo.DangerButton,
				CustomID: "decline-care-package",
				Label:    "Decline",
			},
				func(ctx ken.ComponentContext) bool {
					discord.SuccessfulMessage(sctx, fmt.Sprintf("Declined Power Drop for %s", discord.MentionChannel(inv.UserPinChannel)), fmt.Sprintf("Declined by %s", ctx.User().Username))
					return true
				}, true)
		}, true).
			Condition(func(cctx ken.ComponentContext) bool {
				return true
			})
	})

	fum := b.Send()
	return fum.Error
}
