package inv

import (
	"context"
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/mccune1224/betrayal/internal/models"
	"github.com/mccune1224/betrayal/internal/services/inventory"
	"github.com/zekrotja/ken"
)

func (i *Inv) alignmentCommandGroupBuilder() ken.SubCommandGroup {
	return ken.SubCommandGroup{Name: "alignment", SubHandler: []ken.CommandHandler{
		ken.SubCommandHandler{Name: "set", Run: i.setAlignment},
	}}
}
func (i *Inv) alignmentCommandArgBuilder() *discordgo.ApplicationCommandOption {
	return &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
		Name:        "alignment",
		Description: "set the alignment of the player",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "set",
				Description: "set the alignment of the player",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "alignment",
						Description: "set the alignment of the player",
						Required:    true,
						Choices: []*discordgo.ApplicationCommandOptionChoice{
							{Name: "Good", Value: string(models.AlignmentGOOD)},
							{Name: "Neutral", Value: string(models.AlignmentNEUTRAL)},
							{Name: "Evil", Value: string(models.AlignmentEVIL)},
						},
					},
					discord.UserCommandArg(false),
				},
			},
		},
	}
}

func (i *Inv) setAlignment(ctx ken.SubCommandContext) (err error) {
	if err = ctx.Defer(); err != nil {
		log.Println(err)
		return err
	}
	if !discord.IsAdminRole(ctx, discord.AdminRoles...) {
		return discord.NotAdminError(ctx)
	}
	h, err := inventory.NewInventoryHandler(ctx, i.dbPool)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "failed to init inv handler")
	}
	defer h.UpdateInventoryMessage(ctx.GetSession())

	alignment := ctx.Options().GetByName("alignment").StringValue()
	q := models.New(i.dbPool)
	_, err = q.UpdatePlayerAlignment(context.Background(), models.UpdatePlayerAlignmentParams{
		ID:        h.GetPlayer().ID,
		Alignment: models.Alignment(alignment),
	})
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "")
	}
	return discord.SuccessfulMessage(ctx, "Alignment Set", fmt.Sprintf("Set alignment to %s", alignment))
}
