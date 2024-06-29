package help

import (
	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zekrotja/ken"
)

type Help struct {
	dbPool *pgxpool.Pool
}

func (h *Help) Initialize(pool *pgxpool.Pool) {
	h.dbPool = pool
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

// Version implements ken.SlashCommand.
func (*Help) Version() string {
	return "1.0.0"
}

// Options implements ken.SlashCommand.
func (*Help) Options() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
			Name:        "player",
			Description: "Get help with player commands.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "overview",
					Description: "Get an overview of all player commands.",
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "alliance",
					Description: "how to use alliance commands",
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "action",
					Description: "how to use action commands",
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "inventory",
					Description: "how to use inventory commands",
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "vote",
					Description: "how to use vote commands",
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
			},
		},

		{
			Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
			Name:        "admin",
			Description: "Get help with admin commands.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "overview",
					Description: "get an overview of all admin commands",
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "inventory",
					Description: "how to use inventory commands",
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "alliance",
					Description: "how to use alliance commands",
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "buy",
					Description: "how to use buy commands",
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "kill",
					Description: "how to use kill and revive commands",
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "roll",
					Description: "how to use roll commands",
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "setup",
					Description: "how to use setup commands",
				},
			},
		},
	}
}

// Run implements ken.SlashCommand.
func (h *Help) Run(ctx ken.Context) (err error) {
	return ctx.HandleSubCommands(
		ken.SubCommandGroup{Name: "player", SubHandler: []ken.CommandHandler{
			ken.SubCommandHandler{Name: "overview", Run: h.playerOverview},
			ken.SubCommandHandler{Name: "inventory", Run: h.playerInventory},
			ken.SubCommandHandler{Name: "action", Run: h.playerAction},
			ken.SubCommandHandler{Name: "alliance", Run: h.playerAlliance},
			ken.SubCommandHandler{Name: "vote", Run: h.playerVote},
			ken.SubCommandHandler{Name: "view", Run: h.playerView},
			ken.SubCommandHandler{Name: "list", Run: h.playerList},
		}},
		ken.SubCommandGroup{Name: "admin", SubHandler: []ken.CommandHandler{
			ken.SubCommandHandler{Name: "overview", Run: h.adminOverview},
			ken.SubCommandHandler{Name: "inventory", Run: h.adminInventory},
			ken.SubCommandHandler{Name: "alliance", Run: h.adminAlliance},
			ken.SubCommandHandler{Name: "buy", Run: h.adminBuy},
			ken.SubCommandHandler{Name: "kill", Run: h.adminKill},
			ken.SubCommandHandler{Name: "roll", Run: h.adminRoll},
			ken.SubCommandHandler{Name: "setup", Run: h.adminSetup},
		}},
	)
}
