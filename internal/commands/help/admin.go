package help

import (
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/zekrotja/ken"
)

func (h *Help) adminOverview(ctx ken.SubCommandContext) (err error) {
	if err = ctx.Defer(); err != nil {
		log.Println(err)
		return err
	}
	msg := &discordgo.MessageEmbed{
		Title:       "Admin Commands Overview",
		Description: "All Admin based commands. Helps ",
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:  "Inventory",
				Value: "`/inv` (short for inventory) command allows you to manage a player's inventory within their confessional channel or within a whiltelisted channel. Use it to /help admin inventory`.",
			},
			{
				Name:  "Alliance",
				Value: "`/alliance admin` assists with approving/declining player requests to create alliances as well as accept a pending invite for an alliance. use `/help admin alliance` for more information.",
			},
			{
				Name:  "Buy",
				Value: "`/buy` allows you to buy an item on behalf of a player. use `/help admin buy` for more information.",
			},
			{
				Name:  "Kill/Revive",
				Value: "`/kill` and `/revive` allows you to kill and revive players. use `/help admin kill` for more information.",
			},
			{
				Name:  "Roll",
				Value: "`/roll` allows you to roll game events as well as items/abilities on the fly. use `/help admin roll` for more information.",
			},
			{
				Name:  "Setup",
				Value: "`/setup` assists with determining roles for game creation. use `/help admin setup` for more information.",
			},
		},
	}

	return ctx.RespondEmbed(msg)
}

func (h *Help) adminInventory(ctx ken.SubCommandContext) (err error) {
	if err = ctx.Defer(); err != nil {
		log.Println(err)
		return err
	}

	return ctx.RespondEmbed(adminInventoryEmbed())
}

func (h *Help) adminAlliance(ctx ken.SubCommandContext) (err error) {
	if err = ctx.Defer(); err != nil {
		log.Println(err)
		return err
	}
	return ctx.RespondEmbed(adminAllianceEmbed())
}

func (h *Help) adminBuy(ctx ken.SubCommandContext) (err error) {
	if err = ctx.Defer(); err != nil {
		log.Println(err)
		return err
	}
	return ctx.RespondMessage("TODO")
}

func (h *Help) adminKill(ctx ken.SubCommandContext) (err error) {
	if err = ctx.Defer(); err != nil {
		log.Println(err)
		return err
	}
	return ctx.RespondMessage("TODO")
}

func (h *Help) adminRoll(ctx ken.SubCommandContext) (err error) {
	if err = ctx.Defer(); err != nil {
		log.Println(err)
		return err
	}
	return ctx.RespondMessage("TODO")
}

func (h *Help) adminSetup(ctx ken.SubCommandContext) (err error) {
	if err = ctx.Defer(); err != nil {
		log.Println(err)
		return err
	}
	return ctx.RespondMessage("TODO")
}
