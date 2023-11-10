package commands

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/go-co-op/gocron"
	"github.com/mccune1224/betrayal/internal/data"
	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/zekrotja/ken"
)

type Ping struct {
	models    data.Models
	scheduler *gocron.Scheduler
}

func (p *Ping) Initialize(models data.Models, scheduler *gocron.Scheduler) {
	p.models = models
	p.scheduler = scheduler
}

var _ ken.SlashCommand = (*Ping)(nil)

// Description implements ken.SlashCommand.
func (*Ping) Description() string {
	return "return timestamp for command"
}

// Name implements ken.SlashCommand.
func (*Ping) Name() string {
	return discord.DebugCmd + "ping"
}

// Options implements ken.SlashCommand.
func (*Ping) Options() []*discordgo.ApplicationCommandOption {
	return nil
}

// Run implements ken.SlashCommand.
func (p *Ping) Run(ctx ken.Context) (err error) {
	now := time.Now()

	err = ctx.Respond(&discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("Pong! %s", now.Format(time.RFC3339)),
		},
	})

	return err
}

// Version implements ken.SlashCommand.
func (*Ping) Version() string {
	return "1.0.0"
}
