package commands

import (
	"github.com/bwmarrin/discordgo"
	"github.com/mccune1224/betrayal/internal/data"
	"github.com/zekrotja/ken"
)

type Duel struct {
	models data.Models
}

func (b *Duel) SetModels(models data.Models) {
	b.models = models
}

var _ ken.SlashCommand = (*Duel)(nil)

// Description implements ken.SlashCommand.
func (*Duel) Description() string {
	return "Duel Game Mode"
}

// Name implements ken.SlashCommand.
func (*Duel) Name() string {
	return "duel"
}

// Options implements ken.SlashCommand.
func (*Duel) Options() []*discordgo.ApplicationCommandOption {
	return nil
}

// Run implements ken.SlashCommand.
func (*Duel) Run(ctx ken.Context) (err error) {
	return ctx.RespondEmbed(&discordgo.MessageEmbed{
		Title:       "TODO",
		Description: "Alex is lazy and hasn't made this yet",
	})
}

// Version implements ken.SlashCommand.
func (*Duel) Version() string {
	return "1.0.0"
}
