package channels

import (
	"context"
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/mccune1224/betrayal/internal/models"
	"github.com/zekrotja/ken"
)

func (c *Channel) actionCommandGroupBuilder() ken.SubCommandGroup {
	return ken.SubCommandGroup{Name: "action", SubHandler: []ken.CommandHandler{
		ken.SubCommandHandler{Name: "update", Run: c.updateActionChannel},
		ken.SubCommandHandler{Name: "view", Run: c.viewActionChannel},
	}}
}

func (c *Channel) actionCommandArgBuilder() *discordgo.ApplicationCommandOption {
	return &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
		Name:        "action",
		Description: "update and view current action funnel channel",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "update",
				Description: "Update the current action funnel channel",
				Options: []*discordgo.ApplicationCommandOption{
					discord.ChannelCommandArg(true),
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "view",
				Description: "view the current action funnel channel",
				Options:     []*discordgo.ApplicationCommandOption{},
			},
		},
	}
}

func (c *Channel) viewActionChannel(ctx ken.SubCommandContext) (err error) {
	if err = ctx.Defer(); err != nil {
		log.Println(err)
		return err
	}
	if !discord.IsAdminRole(ctx, discord.AdminRoles...) {
		return discord.NotAdminError(ctx)
	}

	q := models.New(c.dbPool)
	dbCtx := context.Background()

	actionChannel, err := q.GetActionChannel(dbCtx)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Unable to find Action Channel")
	}

	return ctx.RespondEmbed(&discordgo.MessageEmbed{
		Title:       "Current Action Channel",
		Description: discord.MentionChannel(actionChannel),
	})
}

func (c *Channel) updateActionChannel(ctx ken.SubCommandContext) (err error) {
	if err = ctx.Defer(); err != nil {
		log.Println(err)
		return err
	}
	if !discord.IsAdminRole(ctx, discord.AdminRoles...) {
		return discord.NotAdminError(ctx)
	}
	newChannel := ctx.Options().GetByName("channel").ChannelValue(ctx)
	q := models.New(c.dbPool)
	dbCtx := context.Background()
	q.WipeActionChannel(dbCtx)

	err = q.UpsertActionChannel(dbCtx, newChannel.ID)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "")
	}
	return discord.SuccessfulMessage(ctx, "Action Channel Updated", fmt.Sprintf("Admin channel updated to %s", newChannel.Mention()))
}
