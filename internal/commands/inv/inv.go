package inv

import (
	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/zekrotja/ken"
)

type Inv struct {
	dbPool *pgxpool.Pool
}

func (i *Inv) Initialize(pool *pgxpool.Pool) {
	i.dbPool = pool
}

// Description implements ken.SlashCommand.
func (i *Inv) Description() string {
	return "Player Inventory"
}

// Name implements ken.SlashCommand.
func (i *Inv) Name() string {
	return "inv"
}

// Options implements ken.SlashCommand.
func (i *Inv) Options() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
			Name:        "ability",
			Description: "create/update/delete an ability in an inventory",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "add",
					Description: "Add an ability",
					Options: []*discordgo.ApplicationCommandOption{
						discord.StringCommandArg("ability", "Ability to add", true),
						discord.IntCommandArg("quantity", "amount of charges to add", false),
						discord.UserCommandArg(false),
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "delete",
					Description: "Delete an ability",
					Options: []*discordgo.ApplicationCommandOption{
						discord.StringCommandArg("ability", "Ability to add", true),
						discord.UserCommandArg(false),
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "set",
					Description: "set the quantity of an ability",
					Options: []*discordgo.ApplicationCommandOption{
						discord.StringCommandArg("ability", "Ability to add", true),
						discord.IntCommandArg("quantity", "amount of charges to st", true),
						discord.UserCommandArg(false),
					},
				},
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
			Name:        "coin",
			Description: "create/update/delete an coin in an inventory",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "add",
					Description: "Add coins",
					Options: []*discordgo.ApplicationCommandOption{
						discord.IntCommandArg("coin", "Add X coins", true),
						discord.UserCommandArg(false),
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "remove",
					Description: "Remove X coins",
					Options: []*discordgo.ApplicationCommandOption{
						discord.IntCommandArg("coin", "amount of coins to remove", true),
						discord.UserCommandArg(false),
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "set",
					Description: "Set the coins to X",
					Options: []*discordgo.ApplicationCommandOption{
						discord.IntCommandArg("coin", "set coins to specified amount", true),
						discord.UserCommandArg(false),
					},
				},
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "create",
			Description: "Create a player's inventory",
			Options: []*discordgo.ApplicationCommandOption{
				discord.StringCommandArg("role", "Role to create inventory for", true),
				discord.UserCommandArg(true),
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "delete",
			Description: "Delete a player's inventory",
			Options: []*discordgo.ApplicationCommandOption{
				discord.UserCommandArg(true),
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "get",
			Description: "Get player's inventory",
			Options: []*discordgo.ApplicationCommandOption{
				discord.UserCommandArg(true),
			},
		},
	}
}

// Run implements ken.SlashCommand.
func (i *Inv) Run(ctx ken.Context) (err error) {
	return ctx.HandleSubCommands(
		ken.SubCommandHandler{Name: "create", Run: i.create},
		ken.SubCommandHandler{Name: "delete", Run: i.delete},
		ken.SubCommandHandler{Name: "get", Run: i.get},
		ken.SubCommandGroup{Name: "ability", SubHandler: []ken.CommandHandler{
			ken.SubCommandHandler{Name: "add", Run: i.addAbility},
			ken.SubCommandHandler{Name: "delete", Run: i.deleteAbility},
			ken.SubCommandHandler{Name: "set", Run: i.setAbility},
		}},
		ken.SubCommandGroup{Name: "coin", SubHandler: []ken.CommandHandler{
			ken.SubCommandHandler{Name: "add", Run: i.addCoin},
			ken.SubCommandHandler{Name: "remove", Run: i.deleteCoin},
			ken.SubCommandHandler{Name: "set", Run: i.setCoin},
		}},
	)
}

// Version implements ken.SlashCommand.
func (i *Inv) Version() string {
	return "1.0.0"
}

var _ ken.SlashCommand = (*Inv)(nil)
