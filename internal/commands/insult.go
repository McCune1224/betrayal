package commands

import (
	"fmt"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/mccune1224/betrayal/internal/cron"
	"github.com/mccune1224/betrayal/internal/data"
	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/mccune1224/betrayal/internal/util"
	"github.com/zekrotja/ken"
)

type Insult struct {
	models    data.Models
	scheduler *cron.BetrayalScheduler
}

func (i *Insult) Initialize(models data.Models, scheduler *cron.BetrayalScheduler) {
	i.models = models
	i.scheduler = scheduler
}

var _ ken.SlashCommand = (*Insult)(nil)

// Description implements ken.SlashCommand.
func (*Insult) Description() string {
	return "Get and add insults for Alex to read"
}

// Name implements ken.SlashCommand.
func (*Insult) Name() string {
	return discord.DebugCmd + "insult"
}

// Options implements ken.SlashCommand.
func (*Insult) Options() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "add",
			Description: "Add an insult",
			Options: []*discordgo.ApplicationCommandOption{
				discord.StringCommandArg("insult", "The insult to add", true),
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "get",
			Description: "Get an insult",
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "delayed",
			Description: "Get an insult after a delay...for a suprise :)",
			Options: []*discordgo.ApplicationCommandOption{
				discord.StringCommandArg("duration", "The duration to wait before sending the insult", true),
			},
		},
	}
}

// Run implements ken.SlashCommand.
func (i *Insult) Run(ctx ken.Context) (err error) {
	err = ctx.HandleSubCommands(
		ken.SubCommandHandler{Name: "add", Run: i.add},
		ken.SubCommandHandler{Name: "get", Run: i.get},
		ken.SubCommandHandler{Name: "delayed", Run: i.getDelayed},
	)
	return err
}

func (i *Insult) add(ctx ken.SubCommandContext) (err error) {
	args := ctx.Options()
	insultArg := args.GetByName("insult")
	var insult data.Insult
	insult.Insult = insultArg.StringValue()
	insult.AuthorID = ctx.GetEvent().Member.User.ID
	err = i.models.Insults.Insert(&insult)
	if err != nil {
		discord.ErrorMessage(
			ctx,
			fmt.Sprintf("Error adding insult: %s", err.Error()),
			"Alex is a bag programmer and didn't handle this error",
		)
	}
	err = ctx.RespondMessage(
		fmt.Sprintf("Hey %s, %s", discord.MentionUser(discord.McKusaID), insult.Insult),
	)
	return err
}

func (i *Insult) get(ctx ken.SubCommandContext) (err error) {
	insult, err := i.models.Insults.GetRandom()
	if err != nil {
		ctx.SetEphemeral(true)
		return err
	}
	err = ctx.RespondMessage(
		fmt.Sprintf("Hey %s, %s", discord.MentionUser(discord.McKusaID), insult.Insult),
	)
	return err
}

// Version implements ken.SlashCommand.
func (*Insult) Version() string {
	return "1.0.0"
}

func (i *Insult) getDelayed(ctx ken.SubCommandContext) (err error) {
	startTime := util.GetEstTimeStamp()
	duration := ctx.Options().GetByName("duration").StringValue()
	timeDuration, err := time.ParseDuration(duration)
	if err != nil {
		log.Println(err)
		return discord.ErrorMessage(ctx, "Failed to parse duration string", `A duration string is a possibly signed sequence of decimal numbers, each with optional fraction and a unit suffix, such as "300ms", "-1.5h" or "2h45m". Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".`)
	}

	job := func() {
		invokeTime := util.GetEstTimeStamp()
		randInsult, err := i.models.Insults.GetRandom()
		if err != nil {
			log.Println(err)
			return
		}
		msg := &discordgo.MessageEmbed{
			Title:       "Insult",
			Description: fmt.Sprintf("Hey %s, %s", discord.MentionUser(discord.McKusaID), randInsult.Insult),
			Color:       discord.ColorThemeOrange,
			Footer: &discordgo.MessageEmbedFooter{
				Text: fmt.Sprintf("Insult invoked by %s. Queued at %s, resolevd at %s", ctx.User().Username, startTime, invokeTime),
			},
		}
		// we will see if this works with the ken.Context interface...(please work)
		_, err = ctx.GetSession().ChannelMessageSendEmbed(ctx.GetEvent().ChannelID, msg)
	}
	// one time job
	// random id
	foo := time.Now().String()
	err = i.scheduler.UpsertJob(foo, timeDuration, job)
	if err != nil {
		log.Println(err)
		discord.ErrorMessage(ctx, "Failed to queue insult", "Alex is a bad programmer and didn't handle this error")
	}
	return discord.SuccessfulMessage(ctx, "Insult Scheduled", "Your insult has been scheduled... :)")
}
