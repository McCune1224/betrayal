package roll

import (
	"context"
	"fmt"
	"github.com/mccune1224/betrayal/internal/logger"
	"math/rand"

	"github.com/bwmarrin/discordgo"
	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/mccune1224/betrayal/internal/models"
	"github.com/mccune1224/betrayal/internal/services/inventory"
	"github.com/mccune1224/betrayal/internal/util"
	"github.com/zekrotja/ken"
)

func (r *Roll) luckItemRain(ctx ken.SubCommandContext) (err error) {
	if err = ctx.Defer(); err != nil {
		logger.Get().Error().Err(err).Msg("operation failed")
		return err
	}
	if !discord.IsAdminRole(ctx, discord.AdminRoles...) {
		return discord.NotAdminError(ctx)
	}
	inv, err := inventory.NewInventoryHandler(ctx, r.dbPool)
	if err != nil {
		return discord.ErrorMessage(ctx, "Failed to find inventory.", "If not in confessional, please specify a user")
	}
	player := inv.GetPlayer()
	luckLevel := player.Luck
	luckArg, ok := ctx.Options().GetByNameOptional("luck")
	if ok {
		luckLevel = int32(luckArg.IntValue())
	}
	rollAmount := rand.Intn(3) + 1

	q := models.New(r.dbPool)
	dbCtx := context.Background()

	newItems := []models.Item{}
	for i := 0; i < rollAmount; i++ {
		rollRarity := RollRarityLevel(float64(luckLevel), rand.Float64())
		item, err := q.GetRandomItemByRarity(dbCtx, rollRarity)
		if err != nil {
			logger.Get().Error().Err(err).Msg("operation failed")
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
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   fmt.Sprintf("%s (%s)", discord.Bold(item.Name), item.Rarity),
			Value:  item.Description,
			Inline: true,
		})
	}
	currPlayeritems, _ := q.ListPlayerItem(dbCtx, player.ID)
	newPlayerItemCount := int32(len(currPlayeritems) + len(newItems))

	footerMessage := ""
	if newPlayerItemCount > player.ItemLimit {
		footerMessage += fmt.Sprintf("\n %s inventory overflow [%d/%d] %s",
			discord.EmojiWarning,
			newPlayerItemCount,
			player.ItemLimit,
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
			}, logger.WrapKenComponent(func(ctx ken.ComponentContext) bool {
				// rare occurance where inbetween this accepting if the inventory is updated the item list is not updated
				// so re-process the current item list and add the new items
				currInv, err := inventory.NewInventoryHandler(sctx, r.dbPool)
				if err != nil {
					logger.Get().Error().Err(err).Msg("operation failed")
					return true
				}
				for _, item := range newItems {
					_, err = currInv.AddItem(item.Name, 1)
					if err != nil {
						logger.Get().Error().Err(err).Msg("operation failed")
						return true
					}
				}
				newFooterMessage := ""
				if newPlayerItemCount > player.ItemLimit {
					newFooterMessage += fmt.Sprintf("\n %s inventory overflow [%d/%d] %s",
						discord.EmojiWarning,
						newPlayerItemCount,
						player.ItemLimit,
						discord.EmojiWarning,
					)
				} else {
					newFooterMessage += fmt.Sprintf("\n %s adding %d items to inventory %s",
						discord.EmojiSuccess,
						rollAmount,
						discord.EmojiSuccess)
				}
				currInv.UpdateInventoryMessage(sctx.GetSession())
				playerChan, _ := q.GetPlayerConfessional(dbCtx, player.ID)
				embdRain.Footer = &discordgo.MessageEmbedFooter{Text: newFooterMessage}
				_, err = ctx.GetSession().ChannelMessageSendEmbed(util.Itoa64(playerChan.ChannelID), embdRain)
				if err != nil {
					logger.Get().Error().Err(err).Msg("operation failed")
					return true
				}
				// discord.SuccessfulMessage(sctx, fmt.Sprintf("Item Rain Sent to %s", discord.MentionChannel(util.Itoa64(playerChan.ChannelID))),
				// 	fmt.Sprintf("Approved by %s", ctx.User().Username))
				embdRain.Title = fmt.Sprintf("Item Rain Sent to %s (approved by %s)", discord.MentionChannel(util.Itoa64(playerChan.ChannelID)), ctx.User().Username)
				sctx.RespondEmbed(embdRain)
				return true
			}), true)
			b.Add(discordgo.Button{
				Style:    discordgo.DangerButton,
				CustomID: "decline-item-rain",
				Label:    "Decline",
			},
				logger.WrapKenComponent(func(ctx ken.ComponentContext) bool {
					playerChan, _ := q.GetPlayerConfessional(dbCtx, player.ID)
					discord.SuccessfulMessage(sctx, fmt.Sprintf("Declined Item Rain for %s", discord.MentionChannel(util.Itoa64(playerChan.ChannelID))),
						fmt.Sprintf("Declined by %s", ctx.User().Username))
					return true
				}), true)
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
		logger.Get().Error().Err(err).Msg("operation failed")
		return err
	}

	if !discord.IsAdminRole(ctx, discord.AdminRoles...) {
		return discord.NotAdminError(ctx)
	}

	inv, err := inventory.NewInventoryHandler(ctx, r.dbPool)
	if err != nil {
		return discord.ErrorMessage(
			ctx,
			"Failed to get inventory",
			"Are you in a whitelist or confessional channel?",
		)
	}

	q := models.New(r.dbPool)
	dbCtx := context.Background()

	player := inv.GetPlayer()
	luckLevel := player.Luck
	luckArg, ok := ctx.Options().GetByNameOptional("luck")
	if ok {
		luckLevel = int32(luckArg.IntValue())
	}
	rollRarity := RollRarityLevel(float64(luckLevel), rand.Float64())
	aa, err := q.GetRandomAnyAbilityIncludingRoleSpecific(dbCtx, models.GetRandomAnyAbilityIncludingRoleSpecificParams{
		Rarity: rollRarity,
		RoleID: player.RoleID.Int32,
	})
	if err != nil {
		logger.Get().Error().Err(err).Msg("operation failed")
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
		confChan, _ := q.GetPlayerConfessional(dbCtx, player.ID)
		currInv, err := inventory.NewInventoryHandler(sctx, r.dbPool)
		cb.AddActionsRow(func(b ken.ComponentAssembler) {
			b.Add(discordgo.Button{
				Style:    discordgo.SuccessButton,
				CustomID: "confirm-power-drop",
				Label:    "Confirm",
			}, logger.WrapKenComponent(func(ctx ken.ComponentContext) bool {
				// rare occurance where inbetween this accepting if the inventory is updated the item list is not updated
				// so re-process the current item list and add the new items
				if err != nil {
					logger.Get().Error().Err(err).Msg("operation failed")
					return true
				}
				_, err = currInv.AddAbility(aa.Name, 1)
				if err != nil {
					if err.Error() == "ability already added" {
						currInv.UpdateAbility(aa.Name, 1)
					} else {
						logger.Get().Error().Err(err).Msg("operation failed")
						// Don't respond here, will respond at the end
						return true
					}
				}

				currInv.UpdateInventoryMessage(sctx.GetSession())
				_, err = ctx.GetSession().ChannelMessageSendEmbed(util.Itoa64(confChan.ChannelID), embedPowerDrop)
				if err != nil {
					logger.Get().Error().Err(err).Msg("operation failed")
					// Don't respond here, will respond at the end
					return true
				}
				_, err = ctx.GetSession().ChannelMessageSendEmbed(ctx.GetEvent().ChannelID, embedPowerDrop)
				if err != nil {
					logger.Get().Error().Err(err).Msg("operation failed")
					logger.Get().Error().Msg("Failed to send message: Could not find user confessional")
					return true
				}
				// discord.SuccessfulMessage(sctx, fmt.Sprintf("Power Drop Sent to %s", discord.MentionChannel(util.Itoa64(confChan.ChannelID))),
				// 	fmt.Sprintf("Approved by %s", ctx.User().Username))
				embedPowerDrop.Title = fmt.Sprintf("Power Drop Sent to %s (approved by %s)", discord.MentionChannel(util.Itoa64(confChan.ChannelID)), ctx.User().Username)
				sctx.RespondEmbed(embedPowerDrop)
				return true
			}), true)
			b.Add(discordgo.Button{
				Style:    discordgo.DangerButton,
				CustomID: "decline-power-drop",
				Label:    "Decline",
			},
				logger.WrapKenComponent(func(ctx ken.ComponentContext) bool {
					discord.SuccessfulMessage(sctx, fmt.Sprintf("Declined Power Drop for %s", discord.MentionChannel(util.Itoa64(confChan.ChannelID))), fmt.Sprintf("Declined by %s", ctx.User().Username))
					return true
				}), true)
		}, true).
			Condition(func(cctx ken.ComponentContext) bool {
				return true
			})
	})
	fum := b.Send()
	return fum.Error
}
func (r *Roll) luckCarePackage(ctx ken.SubCommandContext) (err error) {
	if err = ctx.Defer(); err != nil {
		logger.Get().Error().Err(err).Msg("operation failed")
		return err
	}

	if !discord.IsAdminRole(ctx, discord.AdminRoles...) {
		return discord.NotAdminError(ctx)
	}

	inv, err := inventory.NewInventoryHandler(ctx, r.dbPool)
	if err != nil {
		return discord.ErrorMessage(ctx, "Failed to find inventory.", "If not in confessional, please specify a user")
	}
	player := inv.GetPlayer()
	luckLevel := player.Luck
	luckArg, ok := ctx.Options().GetByNameOptional("luck")
	if ok {
		luckLevel = int32(luckArg.IntValue())
	}

	aRoll := RollRarityLevel(float64(luckLevel), rand.Float64())
	iRoll := RollRarityLevel(float64(luckLevel), rand.Float64())

	q := models.New(r.dbPool)
	dbCtx := context.Background()

	aa, err := q.GetRandomAnyAbilityIncludingRoleSpecific(dbCtx, models.GetRandomAnyAbilityIncludingRoleSpecificParams{
		Rarity: aRoll,
		RoleID: player.RoleID.Int32,
	})
	if err != nil {
		return discord.ErrorMessage(ctx, "Error getting random ability", "Alex is a bad programmer")
	}

	item, err := q.GetRandomItemByRarity(dbCtx, iRoll)
	if err != nil {
		logger.Get().Error().Err(err).Msg("operation failed")
		return discord.ErrorMessage(ctx, "Failed to get Random Item", "Alex is a bad programmer")
	}

	err = inv.UpdateInventoryMessage(ctx.GetSession())
	if err != nil {
		logger.Get().Error().Err(err).Msg("operation failed")
		discord.SuccessfulMessage(ctx, "Failed to update inventory message", "Alex is a bad programmer")
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
		confChan, _ := q.GetPlayerConfessional(dbCtx, player.ID)
		cb.AddActionsRow(func(b ken.ComponentAssembler) {
			b.Add(discordgo.Button{
				Style:    discordgo.SuccessButton,
				CustomID: "confirm-care-package",
				Label:    "Confirm",
			}, logger.WrapKenComponent(func(ctx ken.ComponentContext) bool {
				// rare occurance where inbetween this accepting if the inventory is updated the item list is not updated
				// so re-process the current item list and add the new items
				currInv, err := inventory.NewInventoryHandler(sctx, r.dbPool)
				if err != nil {
					logger.Get().Error().Err(err).Msg("operation failed")
					return true
				}
				_, err = currInv.AddAbility(aa.Name, 1)
				if err != nil {
					if err.Error() == "ability already added" {
						currInv.UpdateAbility(aa.Name, 1)
					} else {
						logger.Get().Error().Err(err).Msg("operation failed")
						return true
					}
				}
				_, err = currInv.AddItem(item.Name, 1)
				if err != nil {
					logger.Get().Error().Err(err).Msg("operation failed")
					// Don't respond here, will respond at the end
					return true
				}

				currInv.UpdateInventoryMessage(sctx.GetSession())

				_, err = ctx.GetSession().ChannelMessageSendEmbed(util.Itoa64(confChan.ChannelID), embedCarePackage)
				if err != nil {
					logger.Get().Error().Err(err).Msg("operation failed")
					// Don't respond here, will respond at the end
					return true
				}
				// discord.SuccessfulMessage(sctx, fmt.Sprintf("Care Package Sent to %s", discord.MentionChannel(util.Itoa64(confChan.ChannelID))),
				// 	fmt.Sprintf("Approved by %s", ctx.User().Username))
				embedCarePackage.Title = fmt.Sprintf("Care Package Sent to %s (approved by %s)", discord.MentionChannel(util.Itoa64(confChan.ChannelID)), ctx.User().Username)
				sctx.RespondEmbed(embedCarePackage)
				return true
			}), true)
			b.Add(discordgo.Button{
				Style:    discordgo.DangerButton,
				CustomID: "decline-care-package",
				Label:    "Decline",
			},
				logger.WrapKenComponent(func(ctx ken.ComponentContext) bool {
					discord.SuccessfulMessage(sctx, fmt.Sprintf("Declined Power Drop for %s", discord.MentionChannel(util.Itoa64(confChan.ChannelID))), fmt.Sprintf("Declined by %s", ctx.User().Username))
					return true
				}), true)
		}, true).
			Condition(func(cctx ken.ComponentContext) bool {
				return true
			})
	})

	fum := b.Send()
	return fum.Error
}
