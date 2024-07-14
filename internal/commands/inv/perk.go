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

func (i *Inv) perkCommandGroupBuilder() ken.SubCommandGroup {
	return ken.SubCommandGroup{Name: "perk", SubHandler: []ken.CommandHandler{
		ken.SubCommandHandler{Name: "add", Run: i.addPerk},
		ken.SubCommandHandler{Name: "remove", Run: i.removePerk},
	}}
}

func (i *Inv) perkCommandArgBuilder() *discordgo.ApplicationCommandOption {
	return &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
		Name:        "perk",
		Description: "add/remove a perk in an inventory",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "add",
				Description: "Add a perk",
				Options: []*discordgo.ApplicationCommandOption{
					discord.StringCommandArg("perk", "Perk to add", true),
					discord.UserCommandArg(false),
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "remove",
				Description: "Remove a perk",
				Options: []*discordgo.ApplicationCommandOption{
					discord.StringCommandArg("perk", "Perk to add", true),
					discord.UserCommandArg(false),
				},
			},
		},
	}
}

func (i *Inv) addPerk(ctx ken.SubCommandContext) (err error) {
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

	perkNameArg := ctx.Options().GetByName("perk").StringValue()
	q := models.New(i.dbPool)
	perk, err := q.GetPerkInfoByFuzzy(context.Background(), perkNameArg)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "")
	}

	playerPerks, _ := q.ListPlayerPerk(context.Background(), h.GetPlayer().ID)
	for _, currPerk := range playerPerks {
		if perk.ID == currPerk.ID {
			return discord.ErrorMessage(ctx, "Perk already exists", fmt.Sprintf("Player already has Perk %s", perk.Name))
		}
	}

	_, err = q.CreatePlayerPerkJoin(context.Background(), models.CreatePlayerPerkJoinParams{
		PlayerID: h.GetPlayer().ID,
		PerkID:   perk.ID,
	})
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "")
	}
	return discord.SuccessfulMessage(ctx, "Perk Added", fmt.Sprintf("Added perk %s", perk.Name))

}
func (i *Inv) removePerk(ctx ken.SubCommandContext) (err error) {
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

	q := models.New(i.dbPool)
	perkNameArg := ctx.Options().GetByName("perk").StringValue()
	perk, err := q.GetPerkInfoByFuzzy(context.Background(), perkNameArg)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "")
	}

	playerPerks, err := q.ListPlayerPerk(context.Background(), h.GetPlayer().ID)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "")
	}

	found := false
	for _, currPerk := range playerPerks {
		if perk.ID == currPerk.ID {
			found = true
			break
		}
	}

	if !found {
		return discord.ErrorMessage(ctx, "Perk not found", fmt.Sprintf("Player does not have Perk %s", perk.Name))
	}

	err = q.DeletePlayerPerk(context.Background(), models.DeletePlayerPerkParams{
		PlayerID: h.GetPlayer().ID,
		PerkID:   perk.ID,
	})
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "")
	}

	return discord.SuccessfulMessage(ctx, "Perk Removed", fmt.Sprintf("Removed perk %s", perk.Name))
}
