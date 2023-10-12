package commands

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/mccune1224/betrayal/internal/data"
	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/zekrotja/ken"
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
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "category",
			Description: "desired category to list",
			Required:    true,
			Options:     []*discordgo.ApplicationCommandOption{},
			Choices: []*discordgo.ApplicationCommandOptionChoice{
				{
					Name:  "items",
					Value: "items",
				},
				{
					Name:  "roles",
					Value: "roles",
				},
				{
					Name:  "abilities",
					Value: "abilities",
				},
				{
					Name:  "perks",
					Value: "perks",
				},
			},
		},
	}
}

// Run implements ken.SlashCommand.
func (*List) Run(ctx ken.Context) (err error) {
	options := ctx.Options()
	return ctx.RespondMessage(
		fmt.Sprintf("Got options %s", options.Get(0).StringValue()),
	)
}

// Version implements ken.SlashCommand.
func (*List) Version() string {
	return "1.0.0"
}
