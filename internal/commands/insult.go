package commands

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/mccune1224/betrayal/internal/data"
	"github.com/zekrotja/ken"
)

type InsultAdd struct {
	models data.Models
}

type InsultGet struct {
	models data.Models
}

func (igm *InsultGet) SetModels(models data.Models) {
	igm.models = models
}

func (iac *InsultAdd) SetModels(models data.Models) {
	iac.models = models
}

var _ ken.SlashCommand = (*InsultAdd)(nil)
var _ ken.SlashCommand = (*InsultGet)(nil)

// Description implements ken.SlashCommand.
func (*InsultAdd) Description() string {
	return "Insult to provide to Alex"
}

// Name implements ken.SlashCommand.
func (*InsultAdd) Name() string {
	return "insult_add"
}

// Options implements ken.SlashCommand.
func (*InsultAdd) Options() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "insult",
			Description: "Insult to provide to Alex",
			Required:    true,
		},
	}
}

// Run implements ken.SlashCommand.
func (iac *InsultAdd) Run(ctx ken.Context) (err error) {

	var insult data.Insult
	data := ctx.Options().GetByName("insult").StringValue()

	insult.Insult = data
	insult.AuthorID = ctx.User().ID

	err = iac.models.Insults.Insert(&insult)
	if err != nil {
		ctx.RespondError("Failed to add insult", err.Error())
		return err
	}

	err = ctx.RespondMessage(
		fmt.Sprintf(
			"Hey %s, %s",
			Mention(mckusaID),
			insult.Insult,
		),
	)

	return err
}

// Version implements ken.SlashCommand.
func (*InsultAdd) Version() string {
	return "1.0.0"
}

// Description implements ken.SlashCommand.
func (*InsultGet) Description() string {
	return "Get a random insult to throw at Alex"
}

// Name implements ken.SlashCommand.
func (*InsultGet) Name() string {
	return "insult_get"
}

// Options implements ken.SlashCommand.
func (*InsultGet) Options() []*discordgo.ApplicationCommandOption {
	return nil
}

// Run implements ken.SlashCommand.
func (igc *InsultGet) Run(ctx ken.Context) (err error) {
	if err = ctx.Defer(); err != nil {
		return
	}

	b := ctx.FollowUpEmbed(&discordgo.MessageEmbed{
		Description: "Press the button below to get a random insult for Alex to read",
	})

	b.AddComponents(func(cb *ken.ComponentBuilder) {
		cb.AddActionsRow(func(b ken.ComponentAssembler) {
			b.Add(discordgo.Button{
				CustomID: "b_insult_get",
				Label:    "Get Insult",
			}, func(ctx ken.ComponentContext) bool {
				insult, err := igc.models.Insults.GetRandom()
				if err != nil {
					ctx.RespondError("Failed to get insult", err.Error())
					return false
				}
				ctx.RespondMessage(
					fmt.Sprintf(
						"Hey %s, %s",
						Mention(mckusaID),
						insult.Insult,
					),
				)
				return true
			})
		})
	})
	fum := b.Send()
	return fum.Error
}

// Version implements ken.SlashCommand.
func (*InsultGet) Version() string {
	return "1.0.0"
}
