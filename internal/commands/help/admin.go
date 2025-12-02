package help

import (
	"github.com/mccune1224/betrayal/internal/logger"

	"github.com/bwmarrin/discordgo"
	"github.com/zekrotja/ken"
)

func (h *Help) adminOverview(ctx ken.SubCommandContext) (err error) {
	if err = ctx.Defer(); err != nil {
		logger.Get().Error().Err(err).Msg("operation failed")
		return err
	}
	msg := &discordgo.MessageEmbed{
		Title:       "Admin Commands Overview",
		Description: "Comprehensive guide to all admin commands for managing the game.",
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:  "Inventory",
				Value: "`/inv` command allows you to manage a player's inventory within their confessional channel or within a whitelisted channel. Use `/help admin inventory` for more details.",
			},
			{
				Name:  "Alliance",
				Value: "`/alliance admin` handles approval of alliance creation, invites, and management. Use `/help admin alliance` for more information.",
			},
			{
				Name:  "Channels",
				Value: "`/channel` manages game infrastructure like admin channels, vote channels, action channels, and lifeboards. Use `/help admin channels` for setup details.",
			},
			{
				Name:  "Cycle",
				Value: "`/cycle` controls game phases and progression through Day/Elimination cycles. Use `/help admin cycle` for phase management.",
			},
			{
				Name:  "Roll",
				Value: "`/roll` allows you to roll game events, items, and abilities on the fly. Use `/help admin roll` for more information.",
			},
			{
				Name:  "Buy",
				Value: "`/buy` allows you to purchase items on behalf of a player. Use `/help admin buy` for more information.",
			},
			{
				Name:  "Kill/Revive",
				Value: "`/kill` and `/revive` manage player death states and status boards. Use `/help admin kill` for more information.",
			},
			{
				Name:  "Setup",
				Value: "`/setup` assists with generating role lists for game creation. Use `/help admin setup` for more information.",
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
			}, logger.WrapKenComponent(func(ctx ken.ComponentContext) bool {
				ctx.SetEphemeral(true)
				ctx.RespondEmbed(adminInventoryEmbed())
				return true
			}), clearAll)

			b.Add(discordgo.Button{
				CustomID: "a-alliance-help",
				Label:    "Alliance",
				Style:    discordgo.SecondaryButton,
			}, logger.WrapKenComponent(func(ctx ken.ComponentContext) bool {
				ctx.SetEphemeral(true)
				ctx.RespondEmbed(adminAllianceEmbed())
				return true
			}), clearAll)

			b.Add(discordgo.Button{
				CustomID: "a-channels-help",
				Label:    "Channels",
				Style:    discordgo.SecondaryButton,
			}, logger.WrapKenComponent(func(ctx ken.ComponentContext) bool {
				ctx.SetEphemeral(true)
				ctx.RespondEmbed(adminChannelsEmbed())
				return true
			}), clearAll)

			b.Add(discordgo.Button{
				CustomID: "a-cycle-help",
				Label:    "Cycle",
				Style:    discordgo.SecondaryButton,
			}, logger.WrapKenComponent(func(ctx ken.ComponentContext) bool {
				ctx.SetEphemeral(true)
				ctx.RespondEmbed(adminCycleEmbed())
				return true
			}), clearAll)

		}, clearAll)
	})

	b.AddComponents(func(cb *ken.ComponentBuilder) {
		cb.AddActionsRow(func(b ken.ComponentAssembler) {

			b.Add(discordgo.Button{
				CustomID: "a-roll-help",
				Label:    "Roll",
				Style:    discordgo.SecondaryButton,
			}, logger.WrapKenComponent(func(ctx ken.ComponentContext) bool {
				ctx.SetEphemeral(true)
				ctx.RespondEmbed(adminRollEmbed())
				return true
			}), clearAll)

			b.Add(discordgo.Button{
				CustomID: "a-buy-help",
				Label:    "Buy",
				Style:    discordgo.SecondaryButton,
			}, logger.WrapKenComponent(func(ctx ken.ComponentContext) bool {
				ctx.SetEphemeral(true)
				ctx.RespondEmbed(adminBuyEmbed())
				return true
			}), clearAll)

			b.Add(discordgo.Button{
				CustomID: "a-kill-help",
				Label:    "Kill/Revive",
				Style:    discordgo.SecondaryButton,
			}, logger.WrapKenComponent(func(ctx ken.ComponentContext) bool {
				ctx.SetEphemeral(true)
				ctx.RespondEmbed(adminKillEmebd())
				return true
			}), clearAll)

			b.Add(discordgo.Button{
				CustomID: "a-setup-help",
				Label:    "Setup",
				Style:    discordgo.SecondaryButton,
			}, logger.WrapKenComponent(func(ctx ken.ComponentContext) bool {
				ctx.SetEphemeral(true)
				ctx.RespondEmbed(adminSetupEmbed())
				return true
			}), clearAll)

		}, clearAll)
	})

	fum := b.Send()

	return fum.Error
}

func (h *Help) adminInventory(ctx ken.SubCommandContext) (err error) {
	if err = ctx.Defer(); err != nil {
		logger.Get().Error().Err(err).Msg("operation failed")
		return err
	}

	return ctx.RespondEmbed(adminInventoryEmbed())
}

func (h *Help) adminBuy(ctx ken.SubCommandContext) (err error) {
	if err = ctx.Defer(); err != nil {
		logger.Get().Error().Err(err).Msg("operation failed")
		return err
	}
	return ctx.RespondEmbed(adminBuyEmbed())
}

func (h *Help) adminKill(ctx ken.SubCommandContext) (err error) {
	if err = ctx.Defer(); err != nil {
		logger.Get().Error().Err(err).Msg("operation failed")
		return err
	}
	return ctx.RespondEmbed(adminKillEmebd())
}

func (h *Help) adminRoll(ctx ken.SubCommandContext) (err error) {
	if err = ctx.Defer(); err != nil {
		logger.Get().Error().Err(err).Msg("operation failed")
		return err
	}
	return ctx.RespondEmbed(adminRollEmbed())
}

func (h *Help) adminSetup(ctx ken.SubCommandContext) (err error) {
	if err = ctx.Defer(); err != nil {
		logger.Get().Error().Err(err).Msg("operation failed")
		return err
	}
	return ctx.RespondEmbed(adminSetupEmbed())
}

func (h *Help) adminAlliance(ctx ken.SubCommandContext) (err error) {
	if err = ctx.Defer(); err != nil {
		logger.Get().Error().Err(err).Msg("operation failed")
		return err
	}
	return ctx.RespondEmbed(adminAllianceEmbed())
}

func (h *Help) adminChannels(ctx ken.SubCommandContext) (err error) {
	if err = ctx.Defer(); err != nil {
		logger.Get().Error().Err(err).Msg("operation failed")
		return err
	}
	return ctx.RespondEmbed(adminChannelsEmbed())
}

func (h *Help) adminCycle(ctx ken.SubCommandContext) (err error) {
	if err = ctx.Defer(); err != nil {
		logger.Get().Error().Err(err).Msg("operation failed")
		return err
	}
	return ctx.RespondEmbed(adminCycleEmbed())
}
