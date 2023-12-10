package inventory

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/mccune1224/betrayal/internal/services/inventory"
	"github.com/zekrotja/ken"
)

func (i *Inventory) addStatus(ctx ken.SubCommandContext) (err error) {
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
	statusNameArg := ctx.Options().GetByName("name").StringValue()
	durationArg, ok := ctx.Options().GetByNameOptional("duration")
	if ok {
		dur, err = time.ParseDuration(durationArg.StringValue())
		if err != nil {
			return discord.ErrorMessage(ctx, "Failed to parse duration", err.Error())
		}
	}
	res, err := handler.AddStatus(statusNameArg)
	if err != nil {
		if errors.Is(err, inventory.ErrStatusAlreadyExists) {
			return discord.ErrorMessage(ctx, "Status already exists", fmt.Sprintf("Status %s already exists in inventory.", statusNameArg))
		}
		return discord.ErrorMessage(ctx, "Failed to add status", "Alex is a bad programmer, and this is his fault.")
	}
	err = UpdateInventoryMessage(ctx.GetSession(), handler.GetInventory())
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Failed to update inventory message")
	}

	if dur > 0 {
		err = i.scheduler.ScheduleStatus(statusNameArg, handler.GetInventory(), dur, ctx.GetSession())
		if err != nil {
			log.Println(err)
			return discord.ErrorMessage(ctx, "Failed to schedule status", err.Error())
		}
	}

	return discord.SuccessfulMessage(ctx, fmt.Sprintf("Added Status %s", res), fmt.Sprintf("Added for %s", discord.MentionUser(handler.GetInventory().DiscordID)))
}

func (i *Inventory) removeStatus(ctx ken.SubCommandContext) (err error) {
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

	statusArg := ctx.Options().GetByName("name").StringValue()
	res, err := handler.RemoveStatus(statusArg)
	if err != nil {
		if errors.Is(err, inventory.ErrStatusNotFound) {
			return discord.ErrorMessage(ctx, "Status not found", fmt.Sprintf("Status %s not found in inventory.", statusArg))
		}
		log.Print(err)
		return discord.AlexError(ctx, "Failed to remove status")
	}

	err = UpdateInventoryMessage(ctx.GetSession(), handler.GetInventory())
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Failed to update inventory message")
	}
	return discord.SuccessfulMessage(ctx, fmt.Sprintf("Removed Status %s", res), fmt.Sprintf("Removed for %s", discord.MentionUser(handler.GetInventory().DiscordID)))
}
