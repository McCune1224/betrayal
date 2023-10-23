package roll

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/mccune1224/betrayal/internal/commands/inventory"
	"github.com/mccune1224/betrayal/internal/data"
	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/zekrotja/ken"
)

func (r *Roll) luckItemRain(ctx ken.SubCommandContext) (err error) {
	if err = ctx.Defer(); err != nil {
		return err
	}
	inv, err := inventory.Fetch(ctx, r.models, true)
	if err != nil {
		if errors.Is(err, inventory.ErrNotAuthorized) {
			return discord.NotAuthorizedError(ctx)
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
		item, err := r.models.Items.GetRandomByRarity(RollLuck(float64(luckLevel), rand.Float64()))
		if err != nil {
			discord.ErrorMessage(ctx, "Failed to get item", "Alex is a bad programmer")
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
	if len(inv.Items)+len(newItems) > inv.ItemLimit {
		footerMessage += fmt.Sprintf("\n %s inventory overflow [%d/%d] %s",
			discord.EmojiWarning,
			len(inv.Items)+len(newItems),
			inv.ItemLimit,
			discord.EmojiWarning,
		)
	} else {
		footerMessage += fmt.Sprintf("\n %s added %d items to inventory %s",
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
						len(currInv.Items)-1,
						currInv.ItemLimit,
						discord.EmojiWarning,
					)
				} else {
					newFooterMessage += fmt.Sprintf("\n %s added %d items to inventory %s",
						discord.EmojiSuccess,
						rollAmount,
						discord.EmojiSuccess)
				}
				inventory.UpdateInventoryMessage(sctx, currInv)
				embdRain.Footer = &discordgo.MessageEmbedFooter{Text: newFooterMessage}
				_, err = ctx.GetSession().ChannelMessageSendEmbed(currInv.UserPinChannel, embdRain)
				if err != nil {
					log.Println(err)
					return true
				}
				discord.SuccessfulMessage(sctx, "Processed Item Rain", fmt.Sprintf("Sent to %s", discord.MentionChannel(currInv.UserPinChannel)))
				return true
			}, true)
			b.Add(discordgo.Button{
				Style:    discordgo.DangerButton,
				CustomID: "decline-item-rain",
				Label:    "Decline",
			},
				func(ctx ken.ComponentContext) bool {
					discord.SuccessfulMessage(sctx, fmt.Sprintf("Declined Item Rain for %s", discord.MentionChannel(inv.UserPinChannel)), fmt.Sprintf("declined by %s", ctx.User().Username))
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
	rarityType := RollLuck(float64(luckLevel), rand.Float64())
	ability, err := r.getRandomAbility(inv.RoleName, rarityType)
	if err != nil {
		return discord.ErrorMessage(
			ctx,
			"Failed to get ability",
			"Alex is a bad programmer",
		)
	}

	if !ability.AnyAbility {
		// If its not an any ability, instead find the ability in base abilities and update charges instead
		for k, v := range inv.Abilities {
			currInvName := strings.Split(v, " [")[0]
			left := strings.Index(v, "[") + 1
			right := strings.Index(v, "]")
			charge, _ := strconv.Atoi(v[left:right])
			if strings.EqualFold(currInvName, ability.Name) {
				inv.Abilities[k] = fmt.Sprintf("%s [%d]", currInvName, charge)
				err = r.models.Inventories.UpdateAbilities(inv)
				if err != nil {
					log.Println(err)
					return discord.ErrorMessage(
						ctx,
						"Failed to update ability",
						"Alex is a bad programmer, and this is his fault.",
					)
				}
				err = inventory.UpdateInventoryMessage(ctx, inv)
				if err != nil {
					return err
				}
				return ctx.RespondMessage("Ability updated in inventory.")
			}
		}

		return discord.ErrorMessage(ctx, "Failed to find Role Specific Ability...???", "Alex made a major fucky wucky here somehow")
	}

	inv.AnyAbilities = append(inv.AnyAbilities, ability.Name)
	err = r.models.Inventories.UpdateAbilities(inv)
	if err != nil {
		return discord.ErrorMessage(
			ctx,
			"Failed to update inventory",
			"Alex is a bad programmer")
	}

	err = inventory.UpdateInventoryMessage(ctx, inv)
	if err != nil {
		return discord.ErrorMessage(
			ctx,
			"Failed to update inventory message",
			"Alex is a bad programmer")
	}

	return ctx.RespondEmbed(&discordgo.MessageEmbed{
		Title: fmt.Sprintf("%s Power Drop Incoming %s", discord.EmojiAbility, discord.EmojiAbility),
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Ability",
				Value:  fmt.Sprintf("%s (%s) -  %s", ability.Name, rarityType, ability.Description),
				Inline: true,
			},
		},
	})
}
