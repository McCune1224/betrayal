package commands

import (
	"fmt"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/go-co-op/gocron"
	"github.com/mccune1224/betrayal/internal/data"
	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/zekrotja/ken"
)

type Test struct {
	models    data.Models
	scheduler *gocron.Scheduler
}

var _ ken.SlashCommand = (*Test)(nil)

// Initialize implements BetrayalCommand.
func (t *Test) Initialize(m data.Models, s *gocron.Scheduler) {
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
	return []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "timer",
			Description: "test cron job timer",
			Options: []*discordgo.ApplicationCommandOption{
				discord.StringCommandArg("duration", "duration of the timer", true),
				discord.StringCommandArg("task", "Task to do", true),
			},
		},
	}
}

// Run implements ken.SlashCommand.
func (t *Test) Run(ctx ken.Context) (err error) {
	return ctx.HandleSubCommands(
		ken.SubCommandHandler{Name: "timer", Run: t.testTimer},
	)
}

// Version implements ken.SlashCommand.
func (*Test) Version() string {
	return "1.0.0"
}

func (t *Test) testTimer(ctx ken.SubCommandContext) (err error) {
	dur := ctx.Options().GetByName("duration").StringValue()
	timeDur, err := time.ParseDuration(dur)
	if err != nil {
		return discord.ErrorMessage(ctx, "Failed to Parse Duration Argument", `A duration string is a possibly signed sequence of decimal numbers, each with optional fraction and a unit suffix, such as "300ms", "-1.5h" or "2h45m". Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".`)
	}
	taskArg := ctx.Options().GetByName("task").StringValue()

	// WARNING: Is this the correct way to do this? I don't know
	// Do care? No
	// Anon functions are fun
	job := func() { sendMessageTask(ctx.GetSession(), ctx.GetEvent(), taskArg) }
	_, err = t.scheduler.Every(timeDur).WaitForSchedule().LimitRunsTo(1).Do(job)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx)
	}
	return discord.SuccessfulMessage(ctx, "Timer Scheduled", fmt.Sprintf("Timer scheduled for %s", timeDur.String()))
}

// Helper function to test sending a cron job message with the correct context
func sendMessageTask(s *discordgo.Session, e *discordgo.InteractionCreate, arg string) {
	ch := e.ChannelID
	msg := fmt.Sprintf("Hello from cron job: %s", arg)
	_, err := s.ChannelMessageSend(ch, msg)
	if err != nil {
		log.Println(err)
	}
}
