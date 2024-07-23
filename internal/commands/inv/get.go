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

	inv, err := h.FetchInventory()
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "failed to fetch player inv")
	}

	q := models.New(i.dbPool)
	dbCtx := context.Background()

	adminChannels, _ := q.ListAdminChannel(dbCtx)
	isAdminChannel := false
	for _, adminChannel := range adminChannels {
		if adminChannel == ctx.GetEvent().ChannelID {
			isAdminChannel = true
			break
		}
	}

	if isAdminChannel {
		return ctx.RespondEmbed(h.InventoryEmbedBuilder(inv, true))
	}

	return ctx.RespondEmbed(h.InventoryEmbedBuilder(inv, false))
}

func (i *Inv) me(ctx ken.SubCommandContext) (err error) {
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
		} else {
			ctx.SetEphemeral(true)
			return ctx.RespondEmbed(msg)

		}
	}
	ctx.SetEphemeral(true)
	return ctx.RespondEmbed(msg)
}
