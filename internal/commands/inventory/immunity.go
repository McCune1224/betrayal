package inventory

import (
	"errors"
	"fmt"
	"log"

	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/mccune1224/betrayal/internal/services/inventory"
	"github.com/zekrotja/ken"
)

func (i *Inventory) addImmunity(ctx ken.SubCommandContext) (err error) {
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

func (i *Inventory) removeImmunity(ctx ken.SubCommandContext) (err error) {
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

	immunityArg := ctx.Options().GetByName("name").StringValue()
	best, err := handler.RemoveImmunity(immunityArg)
	if err != nil {
		if errors.Is(err, inventory.ErrImmunityNotFound) {
			discord.ErrorMessage(ctx, "Failed to find immunity", fmt.Sprintf("Immunity similar to %s not found in inventory.", immunityArg))
		}
		log.Println(err)
		return discord.AlexError(ctx, "Failed to remove immunity")
	}

	err = UpdateInventoryMessage(ctx.GetSession(), handler.GetInventory())
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Failed to update inventory message")
	}
	return discord.SuccessfulMessage(ctx, fmt.Sprintf("Removed Immunity %s", best),
		fmt.Sprintf("Removed Immunity for %s", discord.MentionUser(handler.GetInventory().DiscordID)))
}
