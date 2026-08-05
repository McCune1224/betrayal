package channels

import (
	"context"
	"errors"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v5"
	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/mccune1224/betrayal/internal/logger"
	"github.com/mccune1224/betrayal/internal/models"
	"github.com/zekrotja/ken"
)

func (c *Channel) logCommandGroupBuilder() ken.SubCommandGroup {
	return ken.SubCommandGroup{Name: "log", SubHandler: []ken.CommandHandler{
		ken.SubCommandHandler{Name: "update", Run: c.updateLogChannel},
		ken.SubCommandHandler{Name: "view", Run: c.viewLogChannel},
		ken.SubCommandHandler{Name: "remove", Run: c.removeLogChannel},
	}}
}

func (c *Channel) logCommandArgBuilder() *discordgo.ApplicationCommandOption {
	return &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
		Name:        "log",
		Description: "Configure the channel that slash command usage is logged to",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "update",
				Description: "Set the command log channel",
				Options: []*discordgo.ApplicationCommandOption{
					discord.ChannelCommandArg(true),
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "view",
				Description: "View the current command log channel",
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "remove",
				Description: "Stop logging slash command usage",
			},
		},
	}
}

func (c *Channel) updateLogChannel(ctx ken.SubCommandContext) (err error) {
	if err = ctx.Defer(); err != nil {
		logger.Get().Error().Err(err).Msg("operation failed")
		return err
	}
	if !discord.IsAdminRole(ctx, discord.AdminRoles...) {
		return discord.NotAdminError(ctx)
	}
	newChannel := ctx.Options().GetByName("channel").ChannelValue(ctx)
	q := models.New(c.dbPool)
	dbCtx := context.Background()

	channelID, err := q.SetCommandLogChannel(dbCtx, newChannel.ID)
	if err != nil {
		logger.Get().Error().Err(err).Msg("operation failed")
		return discord.ErrorMessage(ctx, "Failed to update log channel", err.Error())
	}
	return discord.SuccessfulMessage(ctx, "Log Channel Updated", fmt.Sprintf("Command logging channel updated to %s", discord.MentionChannel(channelID)))
}

func (c *Channel) viewLogChannel(ctx ken.SubCommandContext) (err error) {
	if err = ctx.Defer(); err != nil {
		logger.Get().Error().Err(err).Msg("operation failed")
		return err
	}
	if !discord.IsAdminRole(ctx, discord.AdminRoles...) {
		return discord.NotAdminError(ctx)
	}
	q := models.New(c.dbPool)
	dbCtx := context.Background()

	channelID, err := q.GetCommandLogChannel(dbCtx)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return discord.ErrorMessage(ctx, "No Log Channel", "Command logging is not configured. Use `/channel log update` to set one.")
		}
		logger.Get().Error().Err(err).Msg("operation failed")
		return discord.ErrorMessage(ctx, "Failed to get log channel", "Unable to find log channel")
	}

	return ctx.RespondEmbed(&discordgo.MessageEmbed{
		Title:       "Current Log Channel",
		Description: fmt.Sprintf("Command logging channel is %s", discord.MentionChannel(channelID)),
	})
}

func (c *Channel) removeLogChannel(ctx ken.SubCommandContext) (err error) {
	if err = ctx.Defer(); err != nil {
		logger.Get().Error().Err(err).Msg("operation failed")
		return err
	}
	if !discord.IsAdminRole(ctx, discord.AdminRoles...) {
		return discord.NotAdminError(ctx)
	}
	q := models.New(c.dbPool)
	dbCtx := context.Background()

	if err := q.DeleteCommandLogChannel(dbCtx); err != nil {
		logger.Get().Error().Err(err).Msg("operation failed")
		return discord.ErrorMessage(ctx, "Failed to remove log channel", err.Error())
	}
	return discord.SuccessfulMessage(ctx, "Log Channel Removed", "Command logging is now disabled.")
}
