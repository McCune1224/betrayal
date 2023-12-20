package inventory

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/mccune1224/betrayal/internal/services/inventory"
	"github.com/zekrotja/ken"
)

func (i *Inventory) addEffect(ctx ken.SubCommandContext) (err error) {
	if err := ctx.Defer(); err != nil {
		log.Println(err)
		return err
	}
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

	err = UpdateInventoryMessage(ctx.GetSession(), handler.GetInventory())
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Failed to update inventory message")
	}

	if dur > 0 {
		err = i.scheduler.ScheduleEffect(effectNameArg, handler.GetInventory(), dur, ctx.GetSession())
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
	if err := ctx.Defer(); err != nil {
		log.Println(err)
		return err
	}
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

	potentialJobID := fmt.Sprintf("%s-%s-%s-%s", handler.GetInventory().DiscordID, "effect", "remove", strings.ToLower(best))
	exists := i.scheduler.JobExists(potentialJobID)
	log.Println(exists, potentialJobID)
	if exists {
		err := i.scheduler.DeleteJob(potentialJobID)
		if err != nil {
			log.Println(err)
			ctx.GetSession().ChannelMessageSendEmbed(ctx.GetEvent().ChannelID,
				&discordgo.MessageEmbed{
					Title:       "FAILED TO REMOVE TIMING FOR EFFECT",
					Description: fmt.Sprintf("Was able to remove effect, but failed to remove scheduled job for removal. Please contact %s to fix.", discord.MentionUser(discord.McKusaID)),
				})
		}
	}
	return discord.SuccessfulMessage(ctx, fmt.Sprintf("Removed effect %s", best),
		fmt.Sprintf("Removed effect for %s", discord.MentionUser(handler.GetInventory().DiscordID)))
}
