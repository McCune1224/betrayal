package inventory

import (
	"errors"
	"fmt"
	"log"

	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/mccune1224/betrayal/internal/services/inventory"
	"github.com/zekrotja/ken"
)

func (i *Inventory) addPerk(ctx ken.SubCommandContext) (err error) {
	handler, err := FetchHandler(ctx, i.models, true)
	if err != nil {
		if errors.Is(err, ErrNotAuthorized) {
			return discord.NotAdminError(ctx)
		}
		return discord.ErrorMessage(ctx, "Failed to find inventory.", "If not in confessional, please specify a user")
	}
	ctx.SetEphemeral(false)
	perkNameArg := ctx.Options().GetByName("name").StringValue()

	add, err := handler.AddPerk(perkNameArg)
	if err != nil {
		if errors.Is(err, inventory.ErrPerkAlreadyExists) {
			return discord.ErrorMessage(ctx, "Perk already exists", fmt.Sprintf("Error %s already in inventory", perkNameArg))
		}
		return discord.ErrorMessage(ctx, "Perk not found", fmt.Sprintf("%s not found", perkNameArg))
	}

	err = UpdateInventoryMessage(ctx.GetSession(), handler.GetInventory())
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Failed to update inventory message")
	}

	return discord.SuccessfulMessage(ctx, fmt.Sprintf("Added Perk %s", add),
		fmt.Sprintf("Added for %s", discord.MentionUser(handler.GetInventory().DiscordID)))
}

func (i *Inventory) removePerk(ctx ken.SubCommandContext) (err error) {
	handler, err := FetchHandler(ctx, i.models, true)
	if err != nil {
		if errors.Is(err, ErrNotAuthorized) {
			return discord.NotAdminError(ctx)
		}
		return discord.ErrorMessage(ctx, "Failed to find inventory.", "If not in confessional, please specify a user")
	}
	ctx.SetEphemeral(false)
	perkArg := ctx.Options().GetByName("name").StringValue()

	removed, err := handler.RemovePerk(perkArg)
	if err != nil {
		if errors.Is(err, inventory.ErrPerkNotFound) {
			return discord.ErrorMessage(ctx, "Failed to Remove Perk", fmt.Sprintf("Perk %s not found in inventory.", perkArg))
		}
		log.Println(err)
		return discord.AlexError(ctx, "Failed to remove perk")
	}
	err = UpdateInventoryMessage(ctx.GetSession(), handler.GetInventory())
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Failed to update inventory message")
	}

	return discord.SuccessfulMessage(ctx, fmt.Sprintf("Removed perk %s", removed),
		fmt.Sprintf("Removed perk for %s", discord.MentionUser(handler.GetInventory().DiscordID)))
}
