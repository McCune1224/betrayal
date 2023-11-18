package inventory

import (
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/zekrotja/ken"
)

func (i *Inventory) Options() []*discordgo.ApplicationCommandOption {
	statusChoices := []*discordgo.ApplicationCommandOptionChoice{}
	statuses, err := i.models.Statuses.GetAll()
	if err != nil {
		log.Println(err)
		return nil
	}

	for _, status := range statuses {
		statusChoices = append(statusChoices, &discordgo.ApplicationCommandOptionChoice{
			Name:  status.Name,
			Value: status.Name,
		})
	}

	return []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "get",
			Description: "get player's inventory",
			Options: []*discordgo.ApplicationCommandOption{
				discord.UserCommandArg(true),
				discord.BoolCommandArg("show", "show inventory message (admin view)", false),
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "create",
			Description: "create a new player",
			Options: []*discordgo.ApplicationCommandOption{
				discord.UserCommandArg(true),
				discord.StringCommandArg("role", "Role to assign to player", true),
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "delete",
			Description: "delete inventory",
			Options: []*discordgo.ApplicationCommandOption{
				discord.UserCommandArg(true),
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
			Name:        "whitelist",
			Description: "whitelist channel for inventory",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "add",
					Description: "add whitelist channel",
					Options: []*discordgo.ApplicationCommandOption{
						discord.ChannelCommandArg(true),
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "remove",
					Description: "remove whitelist channel",
					Options: []*discordgo.ApplicationCommandOption{
						discord.ChannelCommandArg(true),
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "list",
					Description: "list whitelist channels",
				},
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
			Name:        "ability",
			Description: "add, remove, or set an ability",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "add",
					Description: "add a base ability (will increase charge if already in inventory)",
					Options: []*discordgo.ApplicationCommandOption{
						discord.StringCommandArg("name", "Name of the ability", true),
						discord.IntCommandArg("charges", "number of charges", false),
						discord.UserCommandArg(false),
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "remove",
					Description: "remove a base ability",
					Options: []*discordgo.ApplicationCommandOption{
						discord.StringCommandArg("name", "Name of the ability", true),
						discord.UserCommandArg(false),
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "set",
					Description: "set charges for a base ability already in inventory",
					Options: []*discordgo.ApplicationCommandOption{
						discord.StringCommandArg("name", "Name of the ability", true),
						discord.IntCommandArg("charges", "Number of charges", true),
						discord.UserCommandArg(false),
					},
				},
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
			Name:        "aa",
			Description: "add, remove, or set an any ability",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "add",
					Description: "add an any ability (will increase charge if already in inventory)",
					Options: []*discordgo.ApplicationCommandOption{
						discord.StringCommandArg("name", "Name of the ability", true),
						discord.IntCommandArg("charges", "number of charges", false),
						discord.UserCommandArg(false),
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "remove",
					Description: "remove an any ability",
					Options: []*discordgo.ApplicationCommandOption{
						discord.StringCommandArg("name", "Name of the ability", true),
						discord.UserCommandArg(false),
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "set",
					Description: "set charges for an any ability already in inventory",
					Options: []*discordgo.ApplicationCommandOption{
						discord.StringCommandArg("name", "Name of the ability", true),
						discord.IntCommandArg("charges", "Number of charges", true),
						discord.UserCommandArg(false),
					},
				},
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
			Name:        "perk",
			Description: "add or remove a perk",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "add",
					Description: "add a perk",
					Options: []*discordgo.ApplicationCommandOption{
						discord.StringCommandArg("name", "Name of the perk", true),
						discord.UserCommandArg(false),
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "remove",
					Description: "remove a perk",
					Options: []*discordgo.ApplicationCommandOption{
						discord.StringCommandArg("name", "Name of the perk", true),
						discord.UserCommandArg(false),
					},
				},
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
			Name:        "item",
			Description: "add or remove an item",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "add",
					Description: "add an item",
					Options: []*discordgo.ApplicationCommandOption{
						discord.StringCommandArg("name", "Name of the perk", true),
						discord.UserCommandArg(false),
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "remove",
					Description: "remove an item",
					Options: []*discordgo.ApplicationCommandOption{
						discord.StringCommandArg("name", "Name of the perk", true),
						discord.UserCommandArg(false),
					},
				},
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
			Name:        "status",
			Description: "add or remove status",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "add",
					Description: "add a status",
					Options: []*discordgo.ApplicationCommandOption{
						discord.StringCommandArg("name", "add a status", true),
						discord.UserCommandArg(false),
					},
					Choices: statusChoices,
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "remove",
					Description: "remove status",
					Options: []*discordgo.ApplicationCommandOption{
						discord.StringCommandArg("name", "Name of the status", true),
						discord.UserCommandArg(false),
					},
				},
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
			Name:        "immunity",
			Description: "add or remove immunity",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "add",
					Description: "add a immunity",
					Options: []*discordgo.ApplicationCommandOption{
						discord.StringCommandArg("name", "add a immunity", true),
						discord.UserCommandArg(false),
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "remove",
					Description: "remove immunity",
					Options: []*discordgo.ApplicationCommandOption{
						discord.StringCommandArg("name", "Name of the immunity", true),
						discord.UserCommandArg(false),
					},
				},
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
			Name:        "effect",
			Description: "add or remove effect",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "add",
					Description: "add an effect",
					Options: []*discordgo.ApplicationCommandOption{
						discord.StringCommandArg("name", "add an effect", true),
						discord.StringCommandArg("duration", "how long should duration last (12h, 1d...)", false),
						discord.UserCommandArg(false),
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "remove",
					Description: "remove effect",
					Options: []*discordgo.ApplicationCommandOption{
						discord.StringCommandArg("name", "Name of the effect", true),
						discord.UserCommandArg(false),
					},
				},
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
			Name:        "coins",
			Description: "add or remove coins",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "add",
					Description: "add coins",
					Options: []*discordgo.ApplicationCommandOption{
						discord.IntCommandArg("amount", "Amount of coins to add", true),
						discord.UserCommandArg(false),
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "remove",
					Description: "remove coins",
					Options: []*discordgo.ApplicationCommandOption{
						discord.IntCommandArg("amount", "Amount of coins to remove", true),
						discord.UserCommandArg(false),
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "set",
					Description: "set coins",
					Options: []*discordgo.ApplicationCommandOption{
						discord.IntCommandArg("amount", "Amount of coins to set", true),
						discord.UserCommandArg(false),
					},
				},
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
			Name:        "bonus",
			Description: "add or remove coin bonus",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "add",
					Description: "add bonus",
					Options: []*discordgo.ApplicationCommandOption{
						// Discord is fucking stupid and doesn't take decimals...need to use string arg
						discord.StringCommandArg("amount", "Amount of coin bonus to set", true),
						discord.UserCommandArg(false),
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "remove",
					Description: "remove bonus",
					Options: []*discordgo.ApplicationCommandOption{
						// Discord is fucking stupid and doesn't take decimals...need to use string arg
						discord.StringCommandArg("amount", "Amount of coin bonus to set", true),
						discord.UserCommandArg(false),
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "set",
					Description: "set bonus",
					Options: []*discordgo.ApplicationCommandOption{
						// Discord is fucking stupid and doesn't take decimals...need to use string arg
						discord.StringCommandArg("amount", "Amount of coin bonus to set", true),
						discord.UserCommandArg(false),
					},
				},
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
			Name:        "luck",
			Description: "add, remove, or set luck",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "add",
					Description: "Add onto of the current luck level",
					Options: []*discordgo.ApplicationCommandOption{
						discord.IntCommandArg("amount", "amount of luck to add", true),
						discord.UserCommandArg(false),
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "remove",
					Description: "remove from the current luck level",
					Options: []*discordgo.ApplicationCommandOption{
						discord.IntCommandArg("amount", "amount of luck to add", true),
						discord.UserCommandArg(false),
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "set",
					Description: "set luck level for player",
					Options: []*discordgo.ApplicationCommandOption{
						discord.IntCommandArg("amount", "amount of luck to add", true),
						discord.UserCommandArg(false),
					},
				},
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
			Name:        "alignment",
			Description: "set the alignment of a player",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "set",
					Description: "set the alignment of a player",
					Options: []*discordgo.ApplicationCommandOption{
						discord.StringCommandArg("name", "alignment type", true),
					},
				},
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
			Name:        "limit",
			Description: "change inventory item limit",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "add",
					Description: "add to inventory limit",
					Options: []*discordgo.ApplicationCommandOption{
						discord.IntCommandArg("amount", "increase the limit by specified amount", true),
						discord.UserCommandArg(false),
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "remove",
					Description: "remove from inventory limit",
					Options: []*discordgo.ApplicationCommandOption{
						discord.IntCommandArg("amount", "decrease the limit by specified amount", true),
						discord.UserCommandArg(false),
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "set",
					Description: "set inventory limit",
					Options: []*discordgo.ApplicationCommandOption{
						discord.IntCommandArg("size", "How many items the inventory should carry", true),
						discord.UserCommandArg(false),
					},
				},
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
			Name:        "note",
			Description: "add or remove a note",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "add",
					Description: "add a note",
					Options: []*discordgo.ApplicationCommandOption{
						discord.StringCommandArg("message", "Note to add", true),
						discord.UserCommandArg(false),
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "remove",
					Description: "remove a note by index number",
					Options: []*discordgo.ApplicationCommandOption{
						discord.IntCommandArg("index", "Index # to remove", true),
						discord.UserCommandArg(false),
					},
				},
			},
		},
	}
}

// Run implements ken.SlashCommand.
func (i *Inventory) Run(ctx ken.Context) (err error) {
	return ctx.HandleSubCommands(
		ken.SubCommandHandler{Name: "get", Run: i.get},
		ken.SubCommandHandler{Name: "create", Run: i.create},
		ken.SubCommandHandler{Name: "delete", Run: i.delete},
		ken.SubCommandGroup{Name: "whitelist", SubHandler: []ken.CommandHandler{
			ken.SubCommandHandler{Name: "add", Run: i.addWhitelist},
			ken.SubCommandHandler{Name: "remove", Run: i.removeWhitelist},
			ken.SubCommandHandler{Name: "list", Run: i.listWhitelist},
		}},
		ken.SubCommandGroup{Name: "ability", SubHandler: []ken.CommandHandler{
			ken.SubCommandHandler{Name: "add", Run: i.addAbility},
			ken.SubCommandHandler{Name: "remove", Run: i.removeAbility},
			ken.SubCommandHandler{Name: "set", Run: i.setAbility},
		}},
		ken.SubCommandGroup{Name: "aa", SubHandler: []ken.CommandHandler{
			ken.SubCommandHandler{Name: "add", Run: i.addAnyAbility},
			ken.SubCommandHandler{Name: "remove", Run: i.removeAnyAbility},
			ken.SubCommandHandler{Name: "set", Run: i.setAnyAbility},
		}},
		ken.SubCommandGroup{Name: "perk", SubHandler: []ken.CommandHandler{
			ken.SubCommandHandler{Name: "add", Run: i.addPerk},
			ken.SubCommandHandler{Name: "remove", Run: i.removePerk},
		}},
		ken.SubCommandGroup{Name: "item", SubHandler: []ken.CommandHandler{
			ken.SubCommandHandler{Name: "add", Run: i.addItem},
			ken.SubCommandHandler{Name: "remove", Run: i.removeItem},
		}},
		ken.SubCommandGroup{Name: "status", SubHandler: []ken.CommandHandler{
			ken.SubCommandHandler{Name: "add", Run: i.addStatus},
			ken.SubCommandHandler{Name: "remove", Run: i.removeStatus},
		}},
		ken.SubCommandGroup{Name: "immunity", SubHandler: []ken.CommandHandler{
			ken.SubCommandHandler{Name: "add", Run: i.addImmunity},
			ken.SubCommandHandler{Name: "remove", Run: i.removeImmunity},
		}},
		ken.SubCommandGroup{Name: "effect", SubHandler: []ken.CommandHandler{
			ken.SubCommandHandler{Name: "add", Run: i.addEffect},
			ken.SubCommandHandler{Name: "remove", Run: i.removeEffect},
		}},
		ken.SubCommandGroup{Name: "coins", SubHandler: []ken.CommandHandler{
			ken.SubCommandHandler{Name: "add", Run: i.addCoins},
			ken.SubCommandHandler{Name: "remove", Run: i.removeCoins},
			ken.SubCommandHandler{Name: "set", Run: i.setCoins},
		}},
		ken.SubCommandGroup{Name: "bonus", SubHandler: []ken.CommandHandler{
			ken.SubCommandHandler{Name: "add", Run: i.addCoinBonus},
			ken.SubCommandHandler{Name: "remove", Run: i.removeCoinBonus},
			ken.SubCommandHandler{Name: "set", Run: i.setCoinBonus},
		}},
		ken.SubCommandGroup{Name: "luck", SubHandler: []ken.CommandHandler{
			ken.SubCommandHandler{Name: "add", Run: i.addLuck},
			ken.SubCommandHandler{Name: "remove", Run: i.removeLuck},
			ken.SubCommandHandler{Name: "set", Run: i.setLuck},
		}},
		ken.SubCommandGroup{Name: "limit", SubHandler: []ken.CommandHandler{
			ken.SubCommandHandler{Name: "add", Run: i.addItemLimit},
			ken.SubCommandHandler{Name: "remove", Run: i.removeItemLimit},
			ken.SubCommandHandler{Name: "set", Run: i.setItemsLimit},
		}},
		ken.SubCommandGroup{Name: "note", SubHandler: []ken.CommandHandler{
			ken.SubCommandHandler{Name: "add", Run: i.addNote},
			ken.SubCommandHandler{Name: "remove", Run: i.removeNote},
		}},
		ken.SubCommandGroup{Name: "alignment", SubHandler: []ken.CommandHandler{
			ken.SubCommandHandler{Name: "set", Run: i.setAlignment},
		}},
	)
}
