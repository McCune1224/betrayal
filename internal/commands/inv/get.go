package inv

import (
	"log"

	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/mccune1224/betrayal/internal/services/inventory"
	"github.com/mccune1224/betrayal/internal/util"
	"github.com/zekrotja/ken"
)

func (i *Inv) get(ctx ken.SubCommandContext) (err error) {
	if err = ctx.Defer(); err != nil {
		log.Println(err)
		return err
	}

	targetPlayer := ctx.Options().GetByName("user").UserValue(ctx)

	pId, err := util.Atoi64(targetPlayer.ID)
	if err != nil {
		return discord.AlexError(ctx, "WHY DOES DISCORD STORE THEIR PLAYER IDS AS STRINGS LULW")
	}

	h, err := inventory.NewInventoryHandler(pId, i.dbPool, false)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "failed to init inv handler")
	}

	authorized, err := h.InventoryAuthorized(ctx)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "failed auth check")
	}

	if !authorized {
		return discord.ErrorMessage(ctx, "Unauthorized Inventory Get", "This should only be done in	a player's confessional channel or a whitelisted channel.")
	}

	inv, err := h.FetchInventory()
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "failed to fetch player inv")
	}

	msg := h.InventoryEmbedBuilder(inv, false)

	return ctx.RespondEmbed(msg)
}
