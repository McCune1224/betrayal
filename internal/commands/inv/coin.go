package inv

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/mccune1224/betrayal/internal/services/inventory"
	"github.com/zekrotja/ken"
)

func (i *Inv) coinCommandGroupBuilder() ken.SubCommandGroup {
	return ken.SubCommandGroup{Name: "coin", SubHandler: []ken.CommandHandler{
		ken.SubCommandHandler{Name: "add", Run: i.addCoin},
		ken.SubCommandHandler{Name: "remove", Run: i.deleteCoin},
		ken.SubCommandHandler{Name: "set", Run: i.setCoin},
		ken.SubCommandHandler{Name: "bonus", Run: i.setBonus},
	}}
}

func (i *Inv) coinCommandArgBuilder() *discordgo.ApplicationCommandOption {
	return &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
		Name:        "coin",
		Description: "create/update/delete an coin in an inventory",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "add",
				Description: "Add coins",
				Options: []*discordgo.ApplicationCommandOption{
					discord.IntCommandArg("coin", "Add X coins", true),
					discord.UserCommandArg(false),
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "remove",
				Description: "Remove X coins",
				Options: []*discordgo.ApplicationCommandOption{
					discord.IntCommandArg("coin", "amount of coins to remove", true),
					discord.UserCommandArg(false),
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "set",
				Description: "Set the coins to X",
				Options: []*discordgo.ApplicationCommandOption{
					discord.IntCommandArg("coin", "set coins to specified amount", true),
					discord.UserCommandArg(false),
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "bonus",
				Description: "Set the coin bonus to X",
				Options: []*discordgo.ApplicationCommandOption{
					discord.StringCommandArg("bonus", "update coin bonus to specified amount (i.e 12.25 = %12.25, 50 = %50.00 , 300.25 = %300.25)", true),
					discord.UserCommandArg(false),
				},
			},
		},
	}
}

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
	player := h.SyncPlayer()
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

	player := h.SyncPlayer()
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
	previousCoins := h.SyncPlayer().Coins
	err = h.SetCoin(int32(coinsArg))
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "failed to set coins")
	}
	player := h.SyncPlayer()
	return discord.SuccessfulMessage(ctx, "Coins Set", fmt.Sprintf("Set %d coins %d => %d", coinsArg, previousCoins, player.Coins))
}

func (i *Inv) setBonus(ctx ken.SubCommandContext) (err error) {
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
	bonusArg := ctx.Options().GetByName("bonus").StringValue()
	err = h.UpdateCoinBonus(bonusArg)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "")
	}
	return discord.SuccessfulMessage(ctx, "Coin Bonus Updated", fmt.Sprintf("Updated coin bonus to %s", bonusArg))
}
