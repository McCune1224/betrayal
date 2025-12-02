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

func (i *Inv) statusCommandGroupBuilder() ken.SubCommandGroup {
	return ken.SubCommandGroup{Name: "status", SubHandler: []ken.CommandHandler{
		ken.SubCommandHandler{Name: "add", Run: i.addStatus},
		ken.SubCommandHandler{Name: "remove", Run: i.removeStatus},
	}}
}
func (i *Inv) statusCommandArgBuilder() *discordgo.ApplicationCommandOption {
	return &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
		Name:        "status",
		Description: "add/remove a status in an inventory",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "add",
				Description: "Add a status",
				Options: []*discordgo.ApplicationCommandOption{
					discord.StatusCommandArg("status", "Status to add", true),
					discord.IntCommandArg("quantity", "amount of the status to add (default 1)", false),
					discord.UserCommandArg(false),
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "remove",
				Description: "Remove a status",
				Options: []*discordgo.ApplicationCommandOption{
					discord.StatusCommandArg("status", "Status to remove", true),
					discord.IntCommandArg("quantity", "amount of the status to add (default 1)", false),
					discord.UserCommandArg(false),
				},
			},
		},
	}
}

func (i *Inv) addStatus(ctx ken.SubCommandContext) (err error) {
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

	statusNameArg := ctx.Options().GetByName("status").StringValue()
	quantity := 1
	if quantityArg, ok := ctx.Options().GetByNameOptional("quantity"); ok {
		quantity = int(quantityArg.IntValue())
	}

	q := models.New(i.dbPool)
	warningMsg := ""
	currentStatuses, _ := q.ListPlayerStatus(context.Background(), h.GetPlayer().ID)
	for _, status := range currentStatuses {
		if status.Name == "Lucky" {
			warningMsg += fmt.Sprintf("%s the player has Lucky status, consider adding now that a new status has been added. %s\n", discord.EmojiWarning, discord.EmojiWarning)
			break
		}
		if status.Name == "Unlucky" {
			warningMsg += fmt.Sprintf("%s the player has Unlucky status, consider adding now that a new status has been added. %s\n", discord.EmojiWarning, discord.EmojiWarning)
			break
		}
	}

	status, err := h.AddStatus(statusNameArg, int32(quantity))
	if err != nil {
		logger.Get().Error().Err(err).Msg("operation failed")
		return discord.AlexError(ctx, "")
	}

	immunities, err := q.ListPlayerImmunity(context.Background(), h.GetPlayer().ID)
	for _, immunity := range immunities {
		if immunity.Name == status.Name {
			if immunity.OneTime {
				warningMsg = fmt.Sprintf("%s The player has one time immunity for %s. Consider removing the immunity.%s\n", discord.EmojiWarning, status.Name, discord.EmojiWarning)
			} else {
				warningMsg = fmt.Sprintf("%s The player is immune to %s. If this is okay, consider removing immunity and if not the status. %s\n", discord.EmojiWarning, status.Name, discord.EmojiWarning)
			}
		}
	}
	return discord.SuccessfulMessage(ctx, "Status Added", fmt.Sprintf("Added status %s", status.Name), warningMsg)
}
func (i *Inv) removeStatus(ctx ken.SubCommandContext) (err error) {
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
	statusNameArg := ctx.Options().GetByName("status").StringValue()
	quantity := 1
	if quantityArg, ok := ctx.Options().GetByNameOptional("quantity"); ok {
		quantity = int(quantityArg.IntValue())
	}
	h.RemoveStatus(statusNameArg, int32(quantity))
	if err != nil {
		logger.Get().Error().Err(err).Msg("operation failed")
		return discord.AlexError(ctx, "Failed to remove status")
	}

	return discord.SuccessfulMessage(ctx, "Status Removed", fmt.Sprintf("Removed %s Status", statusNameArg))
}
