package commands

import (
	"github.com/bwmarrin/discordgo"
	"github.com/mccune1224/betrayal/internal/data"
	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/zekrotja/ken"
)

// Base luck types (at level 0)
var (
	commonLuck    = float32(0.8)
	uncommonLuck  = float32(0.15)
	rareLuck      = float32(0.02)
	epicLuck      = float32(0.015)
	legendaryLuck = float32(0.001)
	MythicalLuck  = float32(0.005)
)

func commonLuckChance(level int) float32 {
	return commonLuck + (float32(level) * 0.01)
}

type Luck struct {
	models data.Models
}

var (
	_ ken.SlashCommand = (*Luck)(nil)
)

// Description implements ken.SlashCommand.
func (*Luck) Description() string {
	return "Determine luck for a given level"
}

// Name implements ken.SlashCommand.
func (*Luck) Name() string {
	return discord.DebugCmd + "luck"
}

// Options implements ken.SlashCommand.
func (*Luck) Options() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{
		discord.IntCommandArg("level", "The level to calculate luck for", true),
	}
}

// Run implements ken.SlashCommand.
func (*Luck) Run(ctx ken.Context) (err error) {
	return discord.SuccessfulMessage(ctx, "TODO", "TODO LUCK")

}

// Version implements ken.SlashCommand.
func (*Luck) Version() string {
	return "1.0.0"
}

func (l *Luck) SetModels(models data.Models) {
	l.models = models
}
