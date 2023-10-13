package commands

import (
	"github.com/bwmarrin/discordgo"
	"github.com/mccune1224/betrayal/internal/data"
	"github.com/zekrotja/ken"
)

type Help struct {
	models data.Models
}

// Description implements ken.SlashCommand.
func (*Help) Description() string {
	panic("unimplemented")
}

// Name implements ken.SlashCommand.
func (*Help) Name() string {
	panic("unimplemented")
}

// Options implements ken.SlashCommand.
func (*Help) Options() []*discordgo.ApplicationCommandOption {
	panic("unimplemented")
}

// Run implements ken.SlashCommand.
func (*Help) Run(ctx ken.Context) (err error) {
	panic("unimplemented")
}

// Version implements ken.SlashCommand.
func (*Help) Version() string {
	panic("unimplemented")
}

func (h *Help) SetModels(models data.Models) {
	h.models = models
}

var _ ken.SlashCommand = (*Help)(nil)
