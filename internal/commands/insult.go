package commands

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/mccune1224/betrayal/internal/data"
	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/mccune1224/betrayal/internal/scheduler"
	"github.com/zekrotja/ken"
)

type Insult struct {
	models    data.Models
	scheduler *scheduler.BetrayalScheduler
}

func (i *Insult) Initialize(models data.Models, scheduler *scheduler.BetrayalScheduler) {
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
	}
}

// Run implements ken.SlashCommand.
func (i *Insult) Run(ctx ken.Context) (err error) {
	err = ctx.HandleSubCommands(
		ken.SubCommandHandler{Name: "add", Run: i.add},
		ken.SubCommandHandler{Name: "get", Run: i.get},
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
