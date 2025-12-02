package inv

import (
	"github.com/mccune1224/betrayal/internal/logger"
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/mccune1224/betrayal/internal/models"
	"github.com/mccune1224/betrayal/internal/services/inventory"
	"github.com/zekrotja/ken"
)

func (i *Inv) immunityCommandGroupBuilder() ken.SubCommandGroup {
	return ken.SubCommandGroup{Name: "immunity", SubHandler: []ken.CommandHandler{
		ken.SubCommandHandler{Name: "add", Run: i.addImmunity},
		ken.SubCommandHandler{Name: "remove", Run: i.removeImmunity},
	}}
}

func (i *Inv) immunityCommandArgBuilder() *discordgo.ApplicationCommandOption {
	return &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
		Name:        "immunity",
		Description: "add/delete an immunity in an inventory",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "add",
				Description: "Add immunity",
				Options: []*discordgo.ApplicationCommandOption{
					discord.StatusCommandArg("immunity", "Immunity to add", true),
					discord.UserCommandArg(false),
					discord.BoolCommandArg("one_time", "One Time Immunity", false),
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "remove",
				Description: "Remove immunity",
				Options: []*discordgo.ApplicationCommandOption{
					discord.StatusCommandArg("immunity", "Immunity to remove", true),
					discord.UserCommandArg(false),
				},
			},
		},
	}
}

func (i *Inv) addImmunity(ctx ken.SubCommandContext) (err error) {
	if err = ctx.Defer(); err != nil {
		logger.Get().Error().Err(err).Msg("operation failed")
		return err
	}
	if !discord.IsAdminRole(ctx, discord.AdminRoles...) {
		return discord.NotAdminError(ctx)
	}
	h, err := inventory.NewInventoryHandler(ctx, i.dbPool)
	if err != nil {
		logger.Get().Error().Err(err).Msg("operation failed")
		return discord.AlexError(ctx, "failed to init inv handler")
	}
	defer h.UpdateInventoryMessage(ctx.GetSession())
	immunityArg := ctx.Options().GetByName("immunity").StringValue()
	isOneTime := false
	if oneTimeArg, ok := ctx.Options().GetByNameOptional("one_time"); ok {
		isOneTime = oneTimeArg.BoolValue()
	}

	q := models.New(i.dbPool)
	existingImmunities, _ := q.ListPlayerImmunity(context.Background(), h.SyncPlayer().ID)
	if len(existingImmunities) > 0 {
		for _, immunity := range existingImmunities {
			if immunity.Name == immunityArg {
				return discord.ErrorMessage(ctx, "Immunity already exists", fmt.Sprintf("Immunity %s already exists", immunityArg))
			}
		}
	}
	immunity, err := q.GetStatusByFuzzy(context.Background(), immunityArg)
	if err != nil {
		logger.Get().Error().Err(err).Msg("operation failed")
		return discord.AlexError(ctx, "failed to add immunity")
	}
	_, err = q.CreateOneTimePlayerImmunityJoin(context.Background(), models.CreateOneTimePlayerImmunityJoinParams{
		PlayerID: h.SyncPlayer().ID,
		StatusID: immunity.ID,
		OneTime:  isOneTime,
	})
	if err != nil {
		logger.Get().Error().Err(err).Msg("operation failed")
		return discord.AlexError(ctx, "failed to add immunity")
	}

	return discord.SuccessfulMessage(ctx, "Immunity Added", fmt.Sprintf("Added immunity %s", immunity.Name))
}

func (i *Inv) removeImmunity(ctx ken.SubCommandContext) (err error) {
	if err = ctx.Defer(); err != nil {
		logger.Get().Error().Err(err).Msg("operation failed")
		return err
	}
	if !discord.IsAdminRole(ctx, discord.AdminRoles...) {
		return discord.NotAdminError(ctx)
	}
	h, err := inventory.NewInventoryHandler(ctx, i.dbPool)
	if err != nil {
		logger.Get().Error().Err(err).Msg("operation failed")
		return discord.AlexError(ctx, "failed to init inv handler")
	}
	defer h.UpdateInventoryMessage(ctx.GetSession())
	q := models.New(i.dbPool)
	immunityArg := ctx.Options().GetByName("immunity").StringValue()
	targetImmunity, err := q.GetStatusByFuzzy(context.Background(), immunityArg)
	if err != nil {
		logger.Get().Error().Err(err).Msg("operation failed")
		return discord.AlexError(ctx, "failed to find immunity")
	}
	existingImmunities, _ := q.ListPlayerImmunity(context.Background(), h.SyncPlayer().ID)
	target := &models.ListPlayerImmunityRow{}
	if len(existingImmunities) > 0 {
		for _, immunity := range existingImmunities {
			if targetImmunity.ID == immunity.ID {
				target = &immunity
				break
			}
		}
	}
	if target == nil {
		return discord.ErrorMessage(ctx, "Immunity not found", fmt.Sprintf("Unable to find immunity %s", immunityArg))
	}
	err = q.DeletePlayerImmunity(context.Background(), models.DeletePlayerImmunityParams{
		PlayerID: h.SyncPlayer().ID,
		StatusID: target.ID,
	})
	if err != nil {
		return discord.AlexError(ctx, "failed to remove immunity")
	}
	return discord.SuccessfulMessage(ctx, "Immunity Removed", fmt.Sprintf("Removed immunity %s", target.Name))

}
