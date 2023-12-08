package help

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/mccune1224/betrayal/internal/data"
	"github.com/mccune1224/betrayal/internal/scheduler"
	"github.com/zekrotja/ken"
)

type Help struct {
	models    data.Models
	scheduler *scheduler.BetrayalScheduler
}

func (h *Help) Initialize(models data.Models, scheduler *scheduler.BetrayalScheduler) {
	h.models = models
	h.scheduler = scheduler
}

var _ ken.SlashCommand = (*Help)(nil)

// Description implements ken.SlashCommand.
func (*Help) Description() string {
	return "Get help with commands and how to use them."
}

// Name implements ken.SlashCommand.
func (*Help) Name() string {
	return "help"
}

// Options implements ken.SlashCommand.
func (*Help) Options() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "overview",
			Description: "Get an overview of all commands.",
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "alliance",
			Description: "how to use alliance commands",
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "inventory",
			Description: "how to use inventory commands",
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "view",
			Description: "how to use view commands",
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "list",
			Description: "how to use list commands",
		},
	}
}

// Run implements ken.SlashCommand.
func (h *Help) Run(ctx ken.Context) (err error) {
	return ctx.HandleSubCommands(
		ken.SubCommandHandler{Name: "overview", Run: h.overview},
		ken.SubCommandHandler{Name: "inventory", Run: h.inventory},
		ken.SubCommandHandler{Name: "action", Run: h.action},
		ken.SubCommandHandler{Name: "alliance", Run: h.alliance},
		ken.SubCommandHandler{Name: "view", Run: h.view},
		ken.SubCommandHandler{Name: "list", Run: h.list},
	)
}

func (h *Help) overview(ctx ken.SubCommandContext) (err error) {
	if err = ctx.Defer(); err != nil {
		log.Println(err)
		return err
	}

	msg := &discordgo.MessageEmbed{
		Title:       "Betrayal Bot Overview",
		Description: "Lexibot is a Discord bot that helps you play Betrayal. It can help you keep track of your inventory, your alliances, and quickly fetch game information. Click a button below or do `/help [topic]` to learn more about a specific topic and commands.",
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:  "Inventory",
				Value: "The `/inventory` command allows you to keep track of your inventory. Use it to keep track of your abilities, items, coins, statuses, and more. For more information, use `/help inventory`.",
			},
			{
				Name:  "Action",
				Value: "The `/action` will send an action for processing. Any ability, item, etc should be done through this command.",
			},
			{
				Name:  "Alliances",
				Value: "The `/alliance` command allows you request creating, joining, and creating alliances. For more information, use `/help alliance`.",
			},
			{
				Name:  "View",
				Value: "The `/view` and `/list` commands allow you to quickly fetch information about the game including details like roles, abilities, perks, items, and more. For more information, use `/help view`.",
			},
			{
				Name:  "List",
				Value: "The `/view` and `/list` commands allow you to quickly fetch information about the game in list format, including details like [lol idk what to put here]. For more information, use `/help list`.",
			},
		},
	}

	b := ctx.FollowUpEmbed(msg)

	clearAll := false
	b.AddComponents(func(cb *ken.ComponentBuilder) {
		cb.AddActionsRow(func(b ken.ComponentAssembler) {
			b.Add(discordgo.Button{
				CustomID: "inventory-help",
				Label:    "Inventory",
				Style:    discordgo.SecondaryButton,
			}, func(ctx ken.ComponentContext) bool {
				ctx.SetEphemeral(true)
				ctx.RespondEmbed(InventoryHelpEmbed())
				return true
			}, clearAll)
			b.Add(discordgo.Button{
				CustomID: "action-help",
				Style:    discordgo.SecondaryButton,
				Label:    "Action",
			}, func(ctx ken.ComponentContext) bool {
				ctx.SetEphemeral(true)
				ctx.RespondEmbed(&discordgo.MessageEmbed{
					Description: fmt.Sprintf("Responded to %s", ctx.GetData().CustomID),
				})
				return true
			}, clearAll)
			b.Add(discordgo.Button{
				CustomID: "alliance-help",
				Style:    discordgo.SecondaryButton,
				Label:    "Alliance",
			}, func(ctx ken.ComponentContext) bool {
				ctx.SetEphemeral(true)
				ctx.RespondEmbed(AllianceHelpEmbed())
				return true
			}, clearAll)
			b.Add(discordgo.Button{
				CustomID: "view-help",
				Style:    discordgo.SecondaryButton,
				Label:    "View",
			}, func(ctx ken.ComponentContext) bool {
				ctx.SetEphemeral(true)
				ctx.RespondEmbed(&discordgo.MessageEmbed{
					Description: fmt.Sprintf("Responded to %s", ctx.GetData().CustomID),
				})
				return true
			}, clearAll)
			b.Add(discordgo.Button{
				CustomID: "list-help",
				Style:    discordgo.SecondaryButton,
				Label:    "List",
			}, func(ctx ken.ComponentContext) bool {
				ctx.SetEphemeral(true)
				ctx.RespondEmbed(&discordgo.MessageEmbed{
					Description: fmt.Sprintf("Responded to %s", ctx.GetData().CustomID),
				})
				return true
			}, clearAll)
		}, clearAll)
	})

	fum := b.Send()
	if err := fum.Error; err != nil {
		log.Println(err)
	}
	return fum.Error
}

func (h *Help) inventory(ctx ken.SubCommandContext) (err error) {
	return ctx.RespondEmbed(InventoryHelpEmbed())
}

func (h *Help) alliance(ctx ken.SubCommandContext) (err error) {
	return ctx.RespondEmbed(AllianceHelpEmbed())
}

func (h *Help) view(ctx ken.SubCommandContext) (err error) {
	return ctx.RespondMessage("todo")
}

func (h *Help) list(ctx ken.SubCommandContext) (err error) {
	return ctx.RespondMessage("todo")
}

// Version implements ken.SlashCommand.
func (*Help) Version() string {
	return "1.0.0"
}

func (*Help) action(ctx ken.SubCommandContext) (err error) {
	return ctx.RespondMessage("todo")
}
