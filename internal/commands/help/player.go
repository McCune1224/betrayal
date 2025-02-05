package help

import (
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/zekrotja/ken"
)

func (h *Help) playerOverview(ctx ken.SubCommandContext) (err error) {
	if err = ctx.Defer(); err != nil {
		log.Println(err)
		return err
	}

	msg := &discordgo.MessageEmbed{
		Title:       "Player Commands Overview",
		Description: "Beatrice is your one stop shop to help with multiple betrayal game components. It will help keep track of your inventory, allow you to request doing votes and actions, and quickly fetch game information. Click a button below or do `/help player [topic]` to learn more.",
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:  "Action",
				Value: "`/action` will request an action for processing. Any ability, item, etc should be done through this command. use `/help player action` for more information.",
			},
			{
				Name:  "Inventory",
				Value: "`/inv` (short for inventory) command allows you to keep track of your inventory. Use it to keep track of your abilities, items, coins, statuses, and more. For more information, use `/help player inventory`.",
			},
			{
				Name:  "List",
				Value: "`/list` allows you to quickly fetch information about the game in list format for things like game events, current role list for the game, items, statuses, and more. For more information, use `/help player list`.",
			},
			{
				Name:  "View",
				Value: "`/view` allows you to quickly fetch information about the game including details like roles, abilities, perks, items, and more. For more information, use `/help player view`.",
			},
			{
				Name:  "Vote",
				Value: "`/vote` allows you to vote one or many players. For more information, use `/help player vote`.",
			},
		},
	}

	b := ctx.FollowUpEmbed(msg)

	// FIXME: What the actual hell
	clearAll := false
	clearAll2 := false
	b.AddComponents(func(cb *ken.ComponentBuilder) {
		cb.AddActionsRow(func(b ken.ComponentAssembler) {
			b.Add(discordgo.Button{
				CustomID: "p-inventory-help",
				Label:    "Inventory",
				Style:    discordgo.SecondaryButton,
			}, func(ctx ken.ComponentContext) bool {
				ctx.SetEphemeral(true)
				ctx.RespondEmbed(playerInventoryHelpEmbed())
				return true
			}, clearAll)
			b.Add(discordgo.Button{
				CustomID: "p-action-help",
				Style:    discordgo.SecondaryButton,
				Label:    "Action",
			}, func(ctx ken.ComponentContext) bool {
				ctx.SetEphemeral(true)
				ctx.RespondEmbed(playerActionHelpEmbed())
				return true
			}, clearAll)
		}, clearAll)
	})

	b.AddComponents(func(cb *ken.ComponentBuilder) {
		cb.AddActionsRow(func(b ken.ComponentAssembler) {
			b.Add(discordgo.Button{
				CustomID: "p-view-help",
				Style:    discordgo.SecondaryButton,
				Label:    "View",
			}, func(ctx ken.ComponentContext) bool {
				ctx.SetEphemeral(true)
				ctx.RespondEmbed(playerViewHelpEmbed())
				return true
			}, clearAll2)
			b.Add(discordgo.Button{
				CustomID: "p-list-help",
				Style:    discordgo.SecondaryButton,
				Label:    "List",
			}, func(ctx ken.ComponentContext) bool {
				ctx.SetEphemeral(true)
				ctx.RespondEmbed(playerListHelpEmbed())
				return true
			}, clearAll2)
			b.Add(discordgo.Button{
				CustomID: "p-vote-help",
				Style:    discordgo.SecondaryButton,
				Label:    "Vote",
			}, func(ctx ken.ComponentContext) bool {
				ctx.SetEphemeral(true)
				ctx.RespondEmbed(playerVoteHelpEmbed())
				return true
			}, clearAll2)
		}, clearAll2)
	})

	fum := b.Send()
	if err := fum.Error; err != nil {
		log.Println(err)
	}
	return fum.Error
}

func (*Help) playerAction(ctx ken.SubCommandContext) (err error) {
	if err := ctx.Defer(); err != nil {
		log.Println(err)
		return err
	}
	return ctx.RespondEmbed(playerActionHelpEmbed())
}

func (h *Help) playerInventory(ctx ken.SubCommandContext) (err error) {
	if err := ctx.Defer(); err != nil {
		log.Println(err)
		return err
	}
	return ctx.RespondEmbed(playerInventoryHelpEmbed())
}

func (h *Help) playerList(ctx ken.SubCommandContext) (err error) {
	if err := ctx.Defer(); err != nil {
		log.Println(err)
		return err
	}
	return ctx.RespondEmbed(playerListHelpEmbed())
}

func (h *Help) playerView(ctx ken.SubCommandContext) (err error) {
	if err := ctx.Defer(); err != nil {
		log.Println(err)
		return err
	}
	return ctx.RespondEmbed(playerViewHelpEmbed())
}

func (*Help) playerVote(ctx ken.SubCommandContext) (err error) {
	if err := ctx.Defer(); err != nil {
		log.Println(err)
		return err
	}
	return ctx.RespondEmbed(playerVoteHelpEmbed())
}
