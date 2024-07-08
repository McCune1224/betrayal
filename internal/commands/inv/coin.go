package inv

import (
	"fmt"
	"log"

	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/mccune1224/betrayal/internal/services/inventory"
	"github.com/zekrotja/ken"
)

func (i *Inv) addCoin(ctx ken.SubCommandContext) (err error) {

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
	coinsArg := ctx.Options().GetByName("coin").IntValue()
	err = h.AddCoin(int32(coinsArg))
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "failed to add coins")
	}
	player := h.GetPlayer()
	return discord.SuccessfulMessage(ctx, "Coins Added", fmt.Sprintf("Added %d coins %d => %d", coinsArg, player.Coins-int32(coinsArg), player.Coins))
}
func (i *Inv) deleteCoin(ctx ken.SubCommandContext) (err error) {
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
	coinsArg := ctx.Options().GetByName("coin").IntValue()
	err = h.RemoveCoin(int32(coinsArg))
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "failed to remove coins")
	}

	player := h.GetPlayer()
	return discord.SuccessfulMessage(ctx, "Coins Removed", fmt.Sprintf("Removed %d coins %d => %d", coinsArg, player.Coins+int32(coinsArg), player.Coins))
}
func (i *Inv) setCoin(ctx ken.SubCommandContext) (err error) {

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
	coinsArg := ctx.Options().GetByName("coin").IntValue()
	err = h.SetCoin(int32(coinsArg))
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "failed to set coins")
	}
	player := h.GetPlayer()
	return discord.SuccessfulMessage(ctx, "Coins Set", fmt.Sprintf("Set %d coins %d => %d", coinsArg, player.Coins, player.Coins))
}
