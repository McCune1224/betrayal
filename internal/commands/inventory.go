package commands

import (
	"github.com/bwmarrin/discordgo"
	"github.com/mccune1224/betrayal/internal/data"
	"github.com/zekrotja/ken"
)

type Inventory struct {
	models data.Models
}

func (i *Inventory) SetModels(models data.Models) {
	i.models = models
}

// Description implements ken.SlashCommand.
func (*Inventory) Description() string {
	return "Command for managing inventory"
}

// Name implements ken.SlashCommand.
func (*Inventory) Name() string {
	return "inventory"
}

// Options implements ken.SlashCommand.
func (*Inventory) Options() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{
		{},
	}
}

// Run implements ken.SlashCommand.
func (*Inventory) Run(ctx ken.Context) (err error) {
	panic("unimplemented")
}

// Version implements ken.SlashCommand.
func (*Inventory) Version() string {
	panic("unimplemented")
}

var _ ken.SlashCommand = (*Inventory)(nil)
