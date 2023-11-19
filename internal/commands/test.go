package commands

import (
	"github.com/bwmarrin/discordgo"
	"github.com/mccune1224/betrayal/internal/data"
	"github.com/mccune1224/betrayal/internal/scheduler"
	"github.com/zekrotja/ken"
)

type Test struct {
	models    data.Models
	scheduler *scheduler.BetrayalScheduler
}

var _ ken.SlashCommand = (*Test)(nil)

// Initialize implements BetrayalCommand.
func (t *Test) Initialize(m data.Models, s *scheduler.BetrayalScheduler) {
	t.models = m
	t.scheduler = s
}

// Description implements ken.SlashCommand.
func (*Test) Description() string {
	return "Dev Sandbox for commands"
}

// Name implements ken.SlashCommand.
func (*Test) Name() string {
	return "t"
}

// Options implements ken.SlashCommand.
func (*Test) Options() []*discordgo.ApplicationCommandOption {
	return nil
}

// Run implements ken.SlashCommand.
func (t *Test) Run(ctx ken.Context) (err error) {
	return ctx.RespondMessage("Nothing to see here officer...")
}

// Version implements ken.SlashCommand.
func (*Test) Version() string {
	return "1.0.0"
}
