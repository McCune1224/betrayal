package commands

import (
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/mccune1224/betrayal/internal/data"
	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/zekrotja/ken"
	"golang.org/x/exp/slices"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type List struct {
	models data.Models
}

func (l *List) SetModels(models data.Models) {
	l.models = models
}

var _ ken.SlashCommand = (*List)(nil)

// Description implements ken.SlashCommand.
func (*List) Description() string {
	return "Get a list of desired category"
}

// Name implements ken.SlashCommand.
func (*List) Name() string {
	return discord.DebugCmd + "list"
}

// Options implements ken.SlashCommand.
func (*List) Options() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{
		{
			Name:        "roles",
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Description: "Get a list of roles",
			Options: []*discordgo.ApplicationCommandOption{
				discord.BoolCommandArg("active", "Get active roles", false),
			},
		},
		{
			Name:        "items",
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Description: "Get a list of items",
			Options: []*discordgo.ApplicationCommandOption{
				discord.BoolCommandArg("all", "Get all items", false),
			},
		},
	}
}

// Run implements ken.SlashCommand.
func (l *List) Run(ctx ken.Context) (err error) {
	return ctx.HandleSubCommands(
		ken.SubCommandHandler{Name: "roles", Run: l.listRoles},
		ken.SubCommandHandler{Name: "items", Run: l.listItems},
	)
}

func (l *List) listRoles(ctx ken.SubCommandContext) (err error) {
	findActive := false
	activeArg, ok := ctx.Options().GetByNameOptional("active")
	if ok {
		findActive = activeArg.BoolValue()
	}
	var roles []*data.Role
	if findActive {
		inventories, err := l.models.Inventories.GetAll()
		if err != nil {
			log.Println(err)
			return discord.ErrorMessage(
				ctx,
				"Error getting active roles",
				"Alex is a bad programmer",
			)
		}
		for _, inventory := range inventories {
			role, err := l.models.Roles.GetByName(inventory.RoleName)
			if err != nil {
				log.Println(err)
				return discord.ErrorMessage(
					ctx,
					"Error getting active roles",
					"Alex is a bad programmer",
				)
			}
			roles = append(roles, role)
		}

	} else {
		var err error
		roles, err = l.models.Roles.GetAll()
		if err != nil {
			log.Println(err)
			return discord.ErrorMessage(
				ctx,
				"Error getting roles",
				"Alex is a bad programmer",
			)
		}
	}

	if len(roles) == 0 {
		return discord.ErrorMessage(ctx, "No roles found...somehow?", "Alex is a bad programmer")
	}

	//divide the role list into two columns
	fields := []*discordgo.MessageEmbedField{}
	var goodColumn []string
	var evilColumn []string
	var neutralColumn []string
	caser := cases.Title(language.AmericanEnglish)
	for _, role := range roles {
		name := caser.String(role.Name)
		switch role.Alignment {
		case "GOOD":
			goodColumn = append(goodColumn, name)
			slices.Sort(goodColumn)
			fields = append(fields, &discordgo.MessageEmbedField{
				Name:   "GOOD",
				Value:  strings.Join(goodColumn, "\n"),
				Inline: true,
			})
		case "EVIL":
			evilColumn = append(evilColumn, name)
			slices.Sort(evilColumn)
			fields = append(fields, &discordgo.MessageEmbedField{
				Name:   "EVIL",
				Value:  strings.Join(evilColumn, "\n"),
				Inline: true,
			})
		case "NEUTRAL":
			neutralColumn = append(neutralColumn, name)
			slices.Sort(goodColumn)
			fields = append(fields, &discordgo.MessageEmbedField{
				Name:   "NEUTRAL",
				Value:  strings.Join(neutralColumn, "\n"),
				Inline: true,
			})
		}
	}

	title := ""
	if findActive {
		title = discord.Underline("Active Roles")
	} else {
		title = discord.Underline("All Roles")
	}

	embed := discordgo.MessageEmbed{
		Title:  title,
		Fields: fields,
	}

	return ctx.RespondEmbed(&embed)
}

func (l *List) listItems(ctx ken.SubCommandContext) (err error) {
	items, err := l.models.Items.GetAll()
	if err != nil {
		log.Println(err)
		return discord.ErrorMessage(ctx, "Error getting items", "Alex is a bad programmer")
	}

	if len(items) == 0 {
		return discord.ErrorMessage(ctx, "No items found...somehow?", "Alex is a bad programmer")
	}

	rarityMap := make(map[string][]string)
	for _, item := range items {
		cost := ""
		if item.Cost == 0 {
			cost = "[X]"
		} else {
			cost = fmt.Sprintf("[%d]", item.Cost)
		}
		rarityMap[item.Rarity] = append(rarityMap[item.Rarity], item.Name+" - "+cost)
	}
	embed := discordgo.MessageEmbed{
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:  "COMMON",
				Value: strings.Join(rarityMap["COMMON"], "\n"),
			},
			{
				Name:  "UNCOMMON",
				Value: strings.Join(rarityMap["UNCOMMON"], "\n"),
			},
			{
				Name:  "RARE",
				Value: strings.Join(rarityMap["RARE"], "\n"),
			},
			{
				Name:  "EPIC",
				Value: strings.Join(rarityMap["EPIC"], "\n"),
			},
			{
				Name:  "LEGENDARY",
				Value: strings.Join(rarityMap["LEGENDARY"], "\n"),
			},
			{
				Name:  "MYTHICAL",
				Value: strings.Join(rarityMap["MYTHICAL"], "\n"),
			},
			{
				Name:  "UNIQUE",
				Value: strings.Join(rarityMap["UNIQUE"], "\n"),
			},
		},
	}
	embed.Title = discord.Underline("Items")
	return ctx.RespondEmbed(&embed)
}

// Version implements ken.SlashCommand.
func (*List) Version() string {
	return "1.0.0"
}
