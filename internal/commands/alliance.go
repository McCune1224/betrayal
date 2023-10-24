package commands

import (
	"github.com/bwmarrin/discordgo"
	"github.com/mccune1224/betrayal/internal/data"
	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/zekrotja/ken"
)

type Alliance struct {
	models data.Models
}

func (a *Alliance) SetModels(models data.Models) {
	a.models = models
}

var _ ken.SlashCommand = (*Alliance)(nil)

// Description implements ken.SlashCommand.
func (*Alliance) Description() string {
	return "Create and join alliances."
}

// Name implements ken.SlashCommand.
func (*Alliance) Name() string {
	return discord.DebugCmd + "alliance"
}

// Options implements ken.SlashCommand.
func (*Alliance) Options() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "request",
			Description: "request to create an alliance.",
			Options: []*discordgo.ApplicationCommandOption{
				discord.StringCommandArg("name", "name of alliance", true),
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "inbox",
			Description: "List alliance invites.",
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "invite",
			Description: "Invite player[s] to your alliance.",
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "accept",
			Description: "Accept an alliance invite.",
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
			Name:        "admin",
			Description: "Manage alliance requests.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "pending",
					Description: "list pending alliance requests.",
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "approve",
					Description: "Approve an alliance request.",
					Options: []*discordgo.ApplicationCommandOption{
						discord.StringCommandArg("name", "name of alliance", true),
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "decline",
					Description: "Decline an alliance request.",
					Options: []*discordgo.ApplicationCommandOption{
						discord.StringCommandArg("name", "name of alliance", true),
					},
				},
			},
		},
	}
}

// Run implements ken.SlashCommand.
func (a *Alliance) Run(ctx ken.Context) (err error) {
	return ctx.HandleSubCommands(
		ken.SubCommandHandler{Name: "request", Run: a.request},
		ken.SubCommandHandler{Name: "inbox", Run: a.inbox},
		ken.SubCommandHandler{Name: "invite", Run: a.invite},
		ken.SubCommandHandler{Name: "accept", Run: a.accept},
		ken.SubCommandGroup{
			Name: "admin", SubHandler: []ken.CommandHandler{
				ken.SubCommandHandler{Name: "pending", Run: a.adminPending},
				ken.SubCommandHandler{Name: "approve", Run: a.adminApprove},
				ken.SubCommandHandler{Name: "decline", Run: a.adminDecline},
			},
		},
	)
}

func (a *Alliance) request(ctx ken.SubCommandContext) (err error) {
	return discord.AlexError(ctx)
}

func (a *Alliance) inbox(ctx ken.SubCommandContext) (err error) {
	return discord.AlexError(ctx)
}

func (a *Alliance) invite(ctx ken.SubCommandContext) (err error) {
	return discord.AlexError(ctx)
}

func (a *Alliance) accept(ctx ken.SubCommandContext) (err error) {
	return discord.AlexError(ctx)
}

func (a *Alliance) adminPending(ctx ken.SubCommandContext) (err error) {
	return discord.AlexError(ctx)
}

func (a *Alliance) adminApprove(ctx ken.SubCommandContext) (err error) {
	return discord.AlexError(ctx)
}

func (a *Alliance) adminDecline(ctx ken.SubCommandContext) (err error) {
	return discord.AlexError(ctx)
}

// Version implements ken.SlashCommand.
func (*Alliance) Version() string {
	return "1.0.0"
}
