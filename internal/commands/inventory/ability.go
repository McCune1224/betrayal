package inventory

import (
	"errors"
	"fmt"
	"log"

	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/zekrotja/ken"
)

func (i *Inventory) addAbility(ctx ken.SubCommandContext) (err error) {
	handler, err := FetchHandler(ctx, i.models, true)
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

	abStr, err := handler.AddAbility(abilityNameArg, charge)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Failed to add ability")
	}
	err = UpdateInventoryMessage(ctx.GetSession(), handler.GetInventory())
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Failed to update inventory message")
	}

	return discord.SuccessfulMessage(ctx, fmt.Sprintf("Added Ability %s", abStr.GetName()), fmt.Sprintf("Added for %s", discord.MentionUser(handler.GetInventory().DiscordID)))
}

func (i *Inventory) removeAbility(ctx ken.SubCommandContext) (err error) {
	handler, err := FetchHandler(ctx, i.models, true)
	if err != nil {
		if errors.Is(err, ErrNotAuthorized) {
			return discord.NotAdminError(ctx)
		}
		return discord.ErrorMessage(ctx, "Failed to find inventory.", "If not in confessional, please specify a user")
	}
	ctx.SetEphemeral(false)
	abilityNameArg := ctx.Options().GetByName("name").StringValue()
	removed, err := handler.RemoveAbility(abilityNameArg)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Failed to remove ability")
	}

	err = UpdateInventoryMessage(ctx.GetSession(), handler.GetInventory())
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Failed to update inventory message")
	}

	return discord.SuccessfulMessage(ctx, fmt.Sprintf("Removed ability %s", removed), fmt.Sprintf("Removed ability %s for %s",
		abilityNameArg, discord.MentionUser(handler.GetInventory().DiscordID)))
}

func (i *Inventory) setAbility(ctx ken.SubCommandContext) (err error) {
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

	err = handler.SetAbilityCharges(abilityNameArg, int(chargesArg))
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Failed to set ability charge")
	}

	err = UpdateInventoryMessage(ctx.GetSession(), handler.GetInventory())
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Failed to update inventory message")
	}

	return discord.SuccessfulMessage(ctx, "Ability Charge Set",
		fmt.Sprintf("Set %s charges to %d", abilityNameArg, chargesArg))
}
