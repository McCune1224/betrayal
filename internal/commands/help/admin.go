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
				Name:  "Roll",
				Value: "`/roll` allows you to roll game events as well as items/abilities on the fly. use `/help admin roll` for more information.",
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
				Name:  "Setup",
				Value: "`/setup` assists with determining roles for game creation. use `/help admin setup` for more information.",
			},
		},
	}

	b := ctx.FollowUpEmbed(msg)
	clearAll := false

	b.AddComponents(func(cb *ken.ComponentBuilder) {
		cb.AddActionsRow(func(b ken.ComponentAssembler) {
			b.Add(discordgo.Button{
				CustomID: "a-inventory-help",
				Label:    "Inventory",
				Style:    discordgo.SecondaryButton,
			}, func(ctx ken.ComponentContext) bool {
				ctx.SetEphemeral(true)
				ctx.RespondEmbed(adminInventoryEmbed())
				return true
			}, clearAll)

			b.Add(discordgo.Button{
				CustomID: "a-roll-help",
				Label:    "Roll",
				Style:    discordgo.SecondaryButton,
			}, func(ctx ken.ComponentContext) bool {
				ctx.SetEphemeral(true)
				ctx.RespondEmbed(adminRollEmbed())
				return true
			}, clearAll)

		}, clearAll)
	})

	b.AddComponents(func(cb *ken.ComponentBuilder) {
		cb.AddActionsRow(func(b ken.ComponentAssembler) {

			b.Add(discordgo.Button{
				CustomID: "a-buy-help",
				Label:    "Buy",
				Style:    discordgo.SecondaryButton,
			}, func(ctx ken.ComponentContext) bool {
				ctx.SetEphemeral(true)
				ctx.RespondEmbed(adminBuyEmbed())
				return true
			}, clearAll)

			b.Add(discordgo.Button{
				CustomID: "a-kill-help",
				Label:    "Kill/Revive",
				Style:    discordgo.SecondaryButton,
			}, func(ctx ken.ComponentContext) bool {
				ctx.SetEphemeral(true)
				ctx.RespondEmbed(adminKillEmebd())
				return true
			}, clearAll)

			b.Add(discordgo.Button{
				CustomID: "a-setup-help",
				Label:    "Setup",
				Style:    discordgo.SecondaryButton,
			}, func(ctx ken.ComponentContext) bool {
				ctx.SetEphemeral(true)
				ctx.RespondEmbed(adminSetupEmbed())
				return true
			}, clearAll)

		}, clearAll)
	})

	fum := b.Send()

	return fum.Error
}

func (h *Help) adminInventory(ctx ken.SubCommandContext) (err error) {
	if err = ctx.Defer(); err != nil {
		log.Println(err)
		return err
	}

	return ctx.RespondEmbed(adminInventoryEmbed())
}

func (h *Help) adminBuy(ctx ken.SubCommandContext) (err error) {
	if err = ctx.Defer(); err != nil {
		log.Println(err)
		return err
	}
	return ctx.RespondEmbed(adminBuyEmbed())
}

func (h *Help) adminKill(ctx ken.SubCommandContext) (err error) {
	if err = ctx.Defer(); err != nil {
		log.Println(err)
		return err
	}
	return ctx.RespondEmbed(adminKillEmebd())
}

func (h *Help) adminRoll(ctx ken.SubCommandContext) (err error) {
	if err = ctx.Defer(); err != nil {
		log.Println(err)
		return err
	}
	return ctx.RespondEmbed(adminRollEmbed())
}

func (h *Help) adminSetup(ctx ken.SubCommandContext) (err error) {
	if err = ctx.Defer(); err != nil {
		log.Println(err)
		return err
	}
	return ctx.RespondEmbed(adminSetupEmbed())
}
