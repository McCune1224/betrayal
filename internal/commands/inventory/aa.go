package inventory

import (
	"errors"
	"fmt"
	"log"

	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/mccune1224/betrayal/internal/services/inventory"
	"github.com/zekrotja/ken"
)

func (i *Inventory) addAnyAbility(ctx ken.SubCommandContext) (err error) {
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

func (i *Inventory) removeAnyAbility(ctx ken.SubCommandContext) (err error) {
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
	abilityNameArg := ctx.Options().GetByName("name").StringValue()
	removed, err := handler.RemoveAnyAbility(abilityNameArg)
	if err != nil {
		if errors.Is(err, inventory.ErrAnyAbilityNotFound) {
			return discord.ErrorMessage(ctx, "Failed to Remove Ability", fmt.Sprintf("Ability %s not found in inventory.", abilityNameArg))
		}
		return discord.AlexError(ctx, "Failed to remove ability")
	}

	err = UpdateInventoryMessage(ctx.GetSession(), handler.GetInventory())
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Failed to update inventory message")
	}

	return discord.SuccessfulMessage(ctx, fmt.Sprintf("Removed Any Ability %s", removed),
		fmt.Sprintf("Removed Any Ability %s for %s", abilityNameArg, discord.MentionUser(handler.GetInventory().DiscordID)))
}

func (i *Inventory) setAnyAbility(ctx ken.SubCommandContext) (err error) {
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
	abilityNameArg := ctx.Options().GetByName("name").StringValue()
	chargesArg := ctx.Options().GetByName("charges").IntValue()

  err = handler.SetAnyAbilityCharges(abilityNameArg, int(chargesArg))
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Failed to set ability charges")
	}


	err = UpdateInventoryMessage(ctx.GetSession(), handler.GetInventory())
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Failed to update inventory message")
	}
  
  return discord.SuccessfulMessage(ctx, fmt.Sprintf("Updated charges for %s", abilityNameArg), fmt.Sprintf("Charges set to %d", chargesArg))

}
