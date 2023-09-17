package commands

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/mccune1224/betrayal/internal/data"
	"github.com/zekrotja/ken"
)

type Insult struct {
	models data.Models
}

func (i *Insult) SetModels(models data.Models) {
	i.models = models
}

var _ ken.SlashCommand = (*Insult)(nil)

// Description implements ken.SlashCommand.
func (*Insult) Description() string {
	return "Get and add insults for Alex to read"
}

// Name implements ken.SlashCommand.
func (*Insult) Name() string {
	return "insult"
}

// Options implements ken.SlashCommand.
func (*Insult) Options() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "add",
			Description: "Add an insult",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "insult",
					Description: "The insult to add",
					Required:    true,
				},
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
	insult.AuthorID = ctx.GetEvent().User.ID
	err = i.models.Insults.Insert(&insult)
	if err != nil {
		ctx.RespondError(err.Error(), "Error adding insult")
		return err
	}
	return err
}

func (i *Insult) get(ctx ken.SubCommandContext) (err error) {
	insult, err := i.models.Insults.GetRandom()
	if err != nil {
		ctx.SetEphemeral(true)
		return err
	}
	err = ctx.RespondMessage(
		fmt.Sprintf("Hey %s, %s", Mention(mckusaID), insult.Insult),
	)
	return err
}

// Version implements ken.SlashCommand.
func (*Insult) Version() string {
	return "1.0.0"
}
