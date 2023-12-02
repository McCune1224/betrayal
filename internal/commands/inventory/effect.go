package inventory

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/mccune1224/betrayal/internal/services/inventory"
	"github.com/mccune1224/betrayal/internal/util"
	"github.com/zekrotja/ken"
)

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
