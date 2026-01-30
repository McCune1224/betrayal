package help

import (
	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mccune1224/betrayal/internal/logger"
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
					Name:        "search",
					Description: "how to use search commands",
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "tarot",
					Description: "how to use the tarot command",
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
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "channels",
					Description: "how to configure and manage game channels",
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "cycle",
					Description: "how to manage game phases and cycles",
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "healthcheck",
					Description: "how to verify game setup and infrastructure",
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "tarot",
					Description: "how to use tarot admin options",
				},
			},
		},
	}
}

// Run implements ken.SlashCommand.
func (h *Help) Run(ctx ken.Context) (err error) {
	defer logger.RecoverWithLog(*logger.Get())

	return ctx.HandleSubCommands(
		ken.SubCommandGroup{Name: "player", SubHandler: []ken.CommandHandler{
			ken.SubCommandHandler{Name: "overview", Run: h.playerOverview},
			ken.SubCommandHandler{Name: "inventory", Run: h.playerInventory},
			ken.SubCommandHandler{Name: "action", Run: h.playerAction},
			ken.SubCommandHandler{Name: "vote", Run: h.playerVote},
			ken.SubCommandHandler{Name: "view", Run: h.playerView},
			ken.SubCommandHandler{Name: "search", Run: h.playerSearch},
			ken.SubCommandHandler{Name: "tarot", Run: h.playerTarot},
			// ken.SubCommandHandler{Name: "list", Run: h.playerList},
		}},
		ken.SubCommandGroup{Name: "admin", SubHandler: []ken.CommandHandler{
			ken.SubCommandHandler{Name: "overview", Run: h.adminOverview},
			ken.SubCommandHandler{Name: "inventory", Run: h.adminInventory},
			ken.SubCommandHandler{Name: "alliance", Run: h.adminAlliance},
			ken.SubCommandHandler{Name: "buy", Run: h.adminBuy},
			ken.SubCommandHandler{Name: "kill", Run: h.adminKill},
			ken.SubCommandHandler{Name: "roll", Run: h.adminRoll},
			ken.SubCommandHandler{Name: "setup", Run: h.adminSetup},
			ken.SubCommandHandler{Name: "channels", Run: h.adminChannels},
			ken.SubCommandHandler{Name: "cycle", Run: h.adminCycle},
			ken.SubCommandHandler{Name: "healthcheck", Run: h.adminHealthcheck},
			ken.SubCommandHandler{Name: "tarot", Run: h.adminTarot},
		}},
	)
}
