package inv

import (
	"context"
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/mccune1224/betrayal/internal/models"
	"github.com/mccune1224/betrayal/internal/services/inventory"
	"github.com/zekrotja/ken"
)

func (i *Inv) itemCommandGroupBuilder() ken.SubCommandGroup {
	return ken.SubCommandGroup{Name: "item", SubHandler: []ken.CommandHandler{
		ken.SubCommandHandler{Name: "add", Run: i.addItem},
		ken.SubCommandHandler{Name: "delete", Run: i.deleteItem},
	}}
}

func (i *Inv) itemCommandArgBuilder() *discordgo.ApplicationCommandOption {
	return &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
		Name:        "item",
		Description: "add/remove an item(s) in an inventory",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "add",
				Description: "Add an item",
				Options: []*discordgo.ApplicationCommandOption{
					discord.StringCommandArg("item", "Item to add", true),
					discord.IntCommandArg("quantity", "amount of the item to add", false),
					discord.UserCommandArg(false),
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "delete",
				Description: "Delete an item",
				Options: []*discordgo.ApplicationCommandOption{
					discord.StringCommandArg("item", "Item to add", true),
					discord.IntCommandArg("quantity", "amount the item to add", false),
					discord.UserCommandArg(false),
				},
			},
		},
	}
}

func (i *Inv) addItem(ctx ken.SubCommandContext) (err error) {
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
	itemNameArg := ctx.Options().GetByName("item").StringValue()
	quantity := int32(1)
	if quantityArg, ok := ctx.Options().GetByNameOptional("quantity"); ok {
		quantity = int32(quantityArg.IntValue())
	}

	item, err := h.AddItem(itemNameArg, quantity)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "failed to add item")
	}

	q := models.New(i.dbPool)
	itemCount, _ := q.GetPlayerItemCount(context.Background(), h.GetPlayer().ID)
	warningMsg := ""
	if int32(itemCount.(int64)) >= h.GetPlayer().ItemLimit {
		warningMsg = fmt.Sprintf("%s %d items out of %d used slots. Use it before you lose it! %s", discord.EmojiWarning, itemCount.(int64), h.GetPlayer().ItemLimit, discord.EmojiWarning)
	}
	return discord.SuccessfulMessage(ctx, "Item Added", fmt.Sprintf("Added item %s", item.Name), warningMsg)
}

func (i *Inv) deleteItem(ctx ken.SubCommandContext) (err error) {
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
	itemNameArg := ctx.Options().GetByName("item").StringValue()

	quantity := int32(1)
	if quantityArg, ok := ctx.Options().GetByNameOptional("quantity"); ok {
		quantity = int32(quantityArg.IntValue())
	}
	item, err := h.RemoveItem(itemNameArg, quantity)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "failed to remove item")
	}

	return discord.SuccessfulMessage(ctx, "Item Removed", fmt.Sprintf("Removed item %s", item.Name))
}
