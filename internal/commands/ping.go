package commands

import (
	"fmt"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/mccune1224/betrayal/internal/scheduler"
	"github.com/mccune1224/betrayal/pkg/data"
	"github.com/zekrotja/ken"
)

type Ping struct {
	models    data.Models
	scheduler *scheduler.BetrayalScheduler
}

func (p *Ping) Initialize(models data.Models, scheduler *scheduler.BetrayalScheduler) {
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
	if err := ctx.Defer(); err != nil {
		log.Println(err)
		return err
	}
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
