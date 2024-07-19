package inv

import (
	"context"
	"fmt"
	"log"

	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/mccune1224/betrayal/internal/models"
	"github.com/mccune1224/betrayal/internal/services/inventory"
	"github.com/mccune1224/betrayal/internal/util"
	"github.com/zekrotja/ken"
)

func (i *Inv) get(ctx ken.SubCommandContext) (err error) {
	ctx.SetEphemeral(false)
	if err = ctx.Defer(); err != nil {
		log.Println(err)
		return err
	}

	if !discord.IsAdminRole(ctx, discord.AdminRoles...) {
		return discord.ErrorMessage(ctx, "Unauthorized", "If you're a player use '/inv me'.")
	}

	h, err := inventory.NewInventoryHandler(ctx, i.dbPool)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, err.Error())
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

	q := models.New(i.dbPool)
	if showArg, ok := ctx.Options().GetByNameOptional("show"); ok {
		show := showArg.BoolValue()
		if show {
			currChannelID, _ := util.Atoi64(ctx.GetEvent().ChannelID)
			playerConf, _ := q.GetPlayerConfessional(context.Background(), h.GetPlayer().ID)
			if currChannelID != playerConf.ChannelID {
				return discord.ErrorMessage(ctx, "Unauthorized", fmt.Sprintf("You can only show this in %s", discord.MentionChannel(util.Itoa64(playerConf.ChannelID))))
			}
			return ctx.RespondEmbed(h.InventoryEmbedBuilder(inv, false))
		}
	}

	return ctx.RespondEmbed(h.InventoryEmbedBuilder(inv, true))
}

func (i *Inv) me(ctx ken.SubCommandContext) (err error) {
	ctx.SetEphemeral(true)
	if err = ctx.Defer(); err != nil {
		log.Println(err)
		return err
	}
	q := models.New(i.dbPool)
	//WARNING: This is a one off hack. Need to manually create this instead of using the NewInventoryHandler
	// as this breaks the two checks for inventory authorization but is still *technically* correct
	targetPlayerID, _ := util.Atoi64(ctx.GetEvent().Member.User.ID)
	player, err := q.GetPlayer(context.Background(), targetPlayerID)
	if err != nil {
		return discord.ErrorMessage(ctx, "Player not found", "Unable to find you as a player")
	}
	h := inventory.Jank(player, i.dbPool)
	inv, err := h.FetchInventory()
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "failed to fetch player inv")
	}

	msg := h.InventoryEmbedBuilder(inv, false)
	if showArg, ok := ctx.Options().GetByNameOptional("show"); ok {
		show := showArg.BoolValue()
		currChannelID, _ := util.Atoi64(ctx.GetEvent().ChannelID)
		conf, _ := q.GetPlayerConfessional(context.Background(), h.GetPlayer().ID)
		if show {
			if conf.ChannelID != currChannelID {
				return discord.ErrorMessage(ctx, "Unauthorized", fmt.Sprintf("You can only show this in %s", discord.MentionChannel(util.Itoa64(conf.ChannelID))))
			}
			ctx.SetEphemeral(false)
			return ctx.RespondEmbed(msg)
		}
	}
	return ctx.RespondEmbed(msg)
}
