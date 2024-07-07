package inv

import (
	"fmt"
	"log"

	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/mccune1224/betrayal/internal/services/inventory"
	"github.com/zekrotja/ken"
)

func (i *Inv) addAbility(ctx ken.SubCommandContext) (err error) {
	if err = ctx.Defer(); err != nil {
		return err
	}
	if !discord.IsAdminRole(ctx, discord.AdminRoles...) {
		return discord.NotAdminError(ctx)
	}
	h, err := inventory.NewInventoryHandler(ctx, i.dbPool)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "failed to init inv handler")
	}
	defer h.UpdateInventoryMessage(ctx.GetSession())

	abilityNameArg := ctx.Options().GetByName("ability").StringValue()
	quantityArg := int32(ctx.Options().GetByName("quantity").IntValue())
	ability, err := h.AddAbility(abilityNameArg, quantityArg)
	if err != nil {
		return discord.AlexError(ctx, "failed to add ability")
	}
	return discord.SuccessfulMessage(ctx, "Ability Added", fmt.Sprintf("Added ability %s", ability.Name))
}

func (i *Inv) deleteAbility(ctx ken.SubCommandContext) (err error) {
	if err = ctx.Defer(); err != nil {
		return err
	}
	if !discord.IsAdminRole(ctx, discord.AdminRoles...) {
		return discord.NotAdminError(ctx)
	}
	h, err := inventory.NewInventoryHandler(ctx, i.dbPool)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "failed to init inv handler")
	}
	defer h.UpdateInventoryMessage(ctx.GetSession())

	abilityNameArg := ctx.Options().GetByName("ability").StringValue()
	ability, err := h.RemoveAbility(abilityNameArg)
	if err != nil {
		return discord.AlexError(ctx, "failed to remove ability")
	}
	return discord.SuccessfulMessage(ctx, "Ability Removed", fmt.Sprintf("Removed ability %s", ability.Name))
}

func (i *Inv) setAbility(ctx ken.SubCommandContext) (err error) {
	if err = ctx.Defer(); err != nil {
		log.Println(err)
		return err
	}
	if !discord.IsAdminRole(ctx, discord.AdminRoles...) {
		return discord.NotAdminError(ctx)
	}
	h, err := inventory.NewInventoryHandler(ctx, i.dbPool)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "failed to init inv handler")
	}
	defer h.UpdateInventoryMessage(ctx.GetSession())

	abilityNameArg := ctx.Options().GetByName("ability").StringValue()
	quantityArg := int(ctx.Options().GetByName("quantity").IntValue())
	ability, err := h.UpdateAbility(abilityNameArg, quantityArg)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "failed to remove ability")
	}
	return discord.SuccessfulMessage(ctx, "Ability Updated Charges", fmt.Sprintf("ability %s set to %d", ability.Name, quantityArg))
}
