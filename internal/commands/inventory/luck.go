package inventory

import (
	"errors"
	"fmt"
	"log"

	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/zekrotja/ken"
)

func (i *Inventory) addLuck(ctx ken.SubCommandContext) (err error) {
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

	if ctx.GetEvent().ChannelID == handler.GetInventory().UserPinChannel {
		ctx.SetEphemeral(true)
		return discord.ErrorMessage(ctx, "Do not use this command here", "use this in an admin only channel as listed in `/inv whitelist list`")
	}

	ctx.SetEphemeral(true)
	luckArg := ctx.Options().GetByName("amount").IntValue()
	old := handler.GetInventory().Luck
	err = handler.AddLuck(luckArg)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Failed to add luck")
	}
	err = UpdateInventoryMessage(ctx.GetSession(), handler.GetInventory())
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Failed to update inventory message")
	}
	return discord.SuccessfulMessage(ctx, fmt.Sprintf("Added %d Luck", luckArg),
		fmt.Sprintf("%d => %d for %s", old, handler.GetInventory().Luck, discord.MentionUser(handler.GetInventory().DiscordID)))
}

func (i *Inventory) removeLuck(ctx ken.SubCommandContext) (err error) {
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
	ctx.SetEphemeral(true)

	if ctx.GetEvent().ChannelID == handler.GetInventory().UserPinChannel {
		ctx.SetEphemeral(true)
		return discord.ErrorMessage(ctx, "Do not use this command here", "use this in an admin only channel as listed in `/inv whitelist list`")
	}

	luck := ctx.Options().GetByName("amount").IntValue()

	err = handler.RemoveLuck(luck)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Failed to remove luck")
	}
	err = UpdateInventoryMessage(ctx.GetSession(), handler.GetInventory())
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Failed to update inventory message")
	}
	return discord.SuccessfulMessage(ctx, fmt.Sprintf("Removed %d luck", luck), fmt.Sprintf("Removed %d luck\n %d => %d for %s",
		luck, luck, handler.GetInventory().Luck, discord.MentionUser(handler.GetInventory().DiscordID)))
}

func (i *Inventory) setLuck(ctx ken.SubCommandContext) (err error) {
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
	ctx.SetEphemeral(true)

	if ctx.GetEvent().ChannelID == handler.GetInventory().UserPinChannel {
		ctx.SetEphemeral(true)
		return discord.ErrorMessage(ctx, "Do not use this command here", "use this in an admin only channel as listed in `/inv whitelist list`")
	}

	luckLevelArg := ctx.Options().GetByName("amount").IntValue()
	oldLuck := handler.GetInventory().Luck

	err = handler.SetLuck(luckLevelArg)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Failed to set luck level")
	}
	err = UpdateInventoryMessage(ctx.GetSession(), handler.GetInventory())
	if err != nil {
		log.Println(err)
		return discord.ErrorMessage(ctx, "Failed to update inventory message", "Alex is a bad programmer, and this is his fault.")
	}

	return discord.SuccessfulMessage(ctx, fmt.Sprintf("Luck set to %d", luckLevelArg),
		fmt.Sprintf("Luck level from %d to %d for %s", oldLuck, handler.GetInventory().Luck, handler.GetInventory().DiscordID))
}
