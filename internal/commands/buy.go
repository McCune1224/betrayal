package commands

import (
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
	cmdInv "github.com/mccune1224/betrayal/internal/commands/inventory"
	"github.com/mccune1224/betrayal/internal/cron"
	"github.com/mccune1224/betrayal/internal/data"
	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/zekrotja/ken"
)

type Buy struct {
	models    data.Models
	scheduler *cron.BetrayalScheduler
}

func (b *Buy) Initialize(models data.Models, scheduler *cron.BetrayalScheduler) {
	b.models = models
	b.scheduler = scheduler
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

	inventory, err := b.models.Inventories.GetByPinChannel(event.ChannelID)
	if err != nil {
		log.Println(err)
		return discord.ErrorMessage(
			ctx,
			"Unable to find Inventory",
			"No known inventory for this channel",
		)
	}
	if inventory.UserPinChannel != event.ChannelID {
		ctx.SetEphemeral(true)
		return discord.NotConfessionalError(ctx, inventory.UserPinChannel)
	}

	itemName := ctx.Options().GetByName("item").StringValue()
	item, err := b.models.Items.GetByFuzzy(itemName)
	if err != nil {
		log.Println(err)
		return discord.ErrorMessage(
			ctx,
			"Unable to find Item",
			"No known item of name "+itemName,
		)
	}

	if item.Cost == 0 {
		return discord.ErrorMessage(
			ctx,
			"Item is not for sale",
			fmt.Sprintf("%s cannot be purchased", item.Name),
		)
	}

	if inventory.Coins < item.Cost {
		return discord.ErrorMessage(
			ctx,
			fmt.Sprintf("You cannot afford %s", item.Name),
			fmt.Sprintf("Cost: %d, Your Coins: %d", item.Cost, inventory.Coins),
		)
	}

	if inventory.ItemLimit == len(inventory.Items) {
		return discord.ErrorMessage(
			ctx,
			fmt.Sprintf("You cannot buy %s", item.Name),
			fmt.Sprintf("You have reached your item limit of %d", inventory.ItemLimit),
		)
	}

	inventory.Items = append(inventory.Items, item.Name)
	err = b.models.Inventories.Update(inventory)
	if err != nil {
		log.Println(err)
		return discord.ErrorMessage(
			ctx,
			"Failed to update Inventory",
			"Alex is a bad programmer.",
		)
	}
	inventory.Coins = inventory.Coins - item.Cost
	// FIXME: Right now catch all update command not changing coins (int values)
	// For now just manually update coins with its own function
	err = b.models.Inventories.UpdateCoins(inventory)
	if err != nil {
		log.Println(err)
		return discord.ErrorMessage(ctx, "Failed to update coins", "Alex is a bad programmer.")
	}

	err = cmdInv.UpdateInventoryMessage(ctx.GetSession(), inventory)
	if err != nil {
		log.Println(err)
		return discord.ErrorMessage(
			ctx,
			"Failed to update inventory message",
			"Alex is a bad programmer.",
		)
	}

	return discord.SuccessfulMessage(ctx,
		fmt.Sprintf("You bought %s", item.Name),
		fmt.Sprintf("%d -> %d", inventory.Coins+item.Cost, inventory.Coins),
	)
}

// Version implements ken.SlashCommand.
func (*Buy) Version() string {
	return "1.0.0"
}
