package buy

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/mccune1224/betrayal/internal/models"
	"github.com/mccune1224/betrayal/internal/services/inventory"
	"github.com/mccune1224/betrayal/internal/util"
	"github.com/zekrotja/ken"
)

type Buy struct {
	dbPool *pgxpool.Pool
}

func (b *Buy) Initialize(pool *pgxpool.Pool) {
	b.dbPool = pool
}

var _ ken.SlashCommand = (*Buy)(nil)

// Description implements ken.SlashCommand.
func (*Buy) Description() string {
	return "Buy an item from the shop"
}

// Name implements ken.SlashCommand.
func (*Buy) Name() string {
	return discord.DebugCmd + "buy"
}

// Options implements ken.SlashCommand.
func (*Buy) Options() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{
		discord.StringCommandArg("item", "Name of the item to buy", true),
	}
}

// Run implements ken.SlashCommand.
func (b *Buy) Run(ctx ken.Context) (err error) {
	if err := ctx.Defer(); err != nil {
		log.Println(err)
		return err
	}
	event := ctx.GetEvent()
	if !discord.IsAdminRole(ctx, discord.AdminRoles...) {
		return discord.ErrorMessage(
			ctx,
			"Insufficient Permissions",
			fmt.Sprintf(
				"You must have one of the following roles: %s",
				strings.Join(discord.AdminRoles, ", "),
			),
		)
	}

	inventory, err := inventory.NewInventoryHandler(ctx, b.dbPool)
	if err != nil {
		log.Println(err)
		return discord.ErrorMessage(
			ctx,
			"Unable to find Inventory",
			"No known inventory for this channel",
		)
	}

	q := models.New(b.dbPool)
	dbCtx := context.Background()
	player := inventory.GetPlayer()

	playerConf, err := q.GetPlayerConfessional(dbCtx, player.ID)
	if util.Itoa64(playerConf.ChannelID) != event.ChannelID {
		ctx.SetEphemeral(true)
		return discord.NotConfessionalError(ctx, util.Itoa64(playerConf.ChannelID))
	}

	item, err := q.GetItemByFuzzy(dbCtx, ctx.Options().GetByName("item").StringValue())
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Unable to find Item")
	}

	if item.Cost == 0 {
		return discord.ErrorMessage(
			ctx,
			"Item is not for sale",
			fmt.Sprintf("%s cannot be purchased", item.Name),
		)
	}

	if inventory.GetPlayer().Coins < item.Cost {
		return discord.ErrorMessage(
			ctx,
			fmt.Sprintf("You cannot afford %s", item.Name),
			fmt.Sprintf("Cost: %d, Your Coins: %d", item.Cost, player.Coins),
		)
	}
	_, err = inventory.AddItem(item.Name, 1)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Failed to update Inventory with item")
	}

	_, err = q.UpdatePlayerCoins(dbCtx, models.UpdatePlayerCoinsParams{
		ID:    player.ID,
		Coins: player.Coins - item.Cost,
	})
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Failed to update player coins")
	}
	return discord.SuccessfulMessage(ctx, fmt.Sprintf("You bought %s", item.Name), fmt.Sprintf("%d -> %d", player.Coins+item.Cost, player.Coins))
}

// Version implements ken.SlashCommand.
func (*Buy) Version() string {
	return "1.0.0"
}
