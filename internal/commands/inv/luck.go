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

func (i *Inv) luckCommandGroupBuilder() ken.SubCommandGroup {
	return ken.SubCommandGroup{Name: "luck", SubHandler: []ken.CommandHandler{
		ken.SubCommandHandler{Name: "add", Run: i.addLuck},
		ken.SubCommandHandler{Name: "remove", Run: i.removeLuck},
		ken.SubCommandHandler{Name: "set", Run: i.setLuck},
	}}
}
func (i *Inv) luckCommandArgBuilder() *discordgo.ApplicationCommandOption {
	return &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
		Name:        "luck",
		Description: "add/remove/set luck in an inventory",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "add",
				Description: "Add luck",
				Options: []*discordgo.ApplicationCommandOption{
					discord.IntCommandArg("luck", "Add X luck", true),
					discord.UserCommandArg(false),
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "remove",
				Description: "Remove luck",
				Options: []*discordgo.ApplicationCommandOption{
					discord.IntCommandArg("luck", "amount of luck to remove", true),
					discord.UserCommandArg(false),
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "set",
				Description: "Set the luck to X",
				Options: []*discordgo.ApplicationCommandOption{
					discord.IntCommandArg("luck", "set luck to specified amount", true),
					discord.UserCommandArg(false),
				},
			},
		},
	}
}

func (i *Inv) addLuck(ctx ken.SubCommandContext) (err error) {
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
	luckArg := ctx.Options().GetByName("luck").IntValue()
	q := models.New(i.dbPool)

	_, err = q.UpdatePlayerLuck(context.Background(), models.UpdatePlayerLuckParams{
		ID:   h.GetPlayer().ID,
		Luck: h.GetPlayer().Luck + int32(luckArg),
	})
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Failed to add luck")
	}

	return discord.SuccessfulMessage(ctx, "Luck Added", fmt.Sprintf("Added %d Luck", luckArg))
}

func (i *Inv) removeLuck(ctx ken.SubCommandContext) (err error) {
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
	luckArg := ctx.Options().GetByName("luck").IntValue()
	q := models.New(i.dbPool)

	diff := h.GetPlayer().Luck - int32(luckArg)
	if diff < 0 {
		diff = 0
	}

	_, err = q.UpdatePlayerLuck(context.Background(), models.UpdatePlayerLuckParams{
		ID:   h.GetPlayer().ID,
		Luck: diff,
	})
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Failed to remove luck")
	}

	return discord.SuccessfulMessage(ctx, "Luck Removed", fmt.Sprintf("Removed %d Luck", luckArg))
}

func (i *Inv) setLuck(ctx ken.SubCommandContext) (err error) {
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
	luckArg := ctx.Options().GetByName("luck").IntValue()
	q := models.New(i.dbPool)

	_, err = q.UpdatePlayerLuck(context.Background(), models.UpdatePlayerLuckParams{
		ID:   h.GetPlayer().ID,
		Luck: int32(luckArg),
	})
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "")
	}

	return discord.SuccessfulMessage(ctx, "Luck Set", fmt.Sprintf("Added %d Luck", luckArg))
}
