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

	if !discord.IsAdminRole(ctx, discord.AdminRoles...) && ctx.GetEvent().Member.User.ID != targetPlayer.ID {
		return discord.ErrorMessage(ctx, "Unauthorized", "You are not authorized to use this command.")
	}
	pId, err := util.Atoi64(targetPlayer.ID)
	if err != nil {
		return discord.AlexError(ctx, "WHY DOES DISCORD STORE THEIR PLAYER IDS AS STRINGS LULW")
	}
	h, err := inventory.NewInventoryHandler(pId, i.dbPool, false)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "failed to init inv handler")
	}
	log.Println("Hit inv init")
	//FIXME: The sqlc query here def is not working...
	inv, err := h.FetchInventory()
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "failed to fetch player inv")
	}
	log.Println("hit inv create")

	log.Println(inv)
	return discord.SuccessfulMessage(ctx, "cool", "epic")
}
