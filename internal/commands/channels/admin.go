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

func (c *Channel) adminCommandGroupBuilder() ken.SubCommandGroup {
	return ken.SubCommandGroup{Name: "admin", SubHandler: []ken.CommandHandler{
		ken.SubCommandHandler{Name: "add", Run: c.addAdminChannel},
		ken.SubCommandHandler{Name: "list", Run: c.listAdminChannel},
		ken.SubCommandHandler{Name: "delete", Run: c.deleteAdminChannel},
	}}
}

func (c *Channel) adminCommandArgBuilder() *discordgo.ApplicationCommandOption {
	return &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
		Name:        "admin",
		Description: "add/remove a admin channel in an inventory",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "add",
				Description: "Add an admin channel",
				Options: []*discordgo.ApplicationCommandOption{
					discord.ChannelCommandArg(true),
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "list",
				Description: "list the current admin channel",
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "delete",
				Description: "Delete an admin channel",
				Options: []*discordgo.ApplicationCommandOption{
					discord.ChannelCommandArg(true),
				},
			},
		},
	}
}

func (c *Channel) addAdminChannel(ctx ken.SubCommandContext) (err error) {
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

	q.CreateAdminChannel(dbCtx, newChannel.ID)
	return discord.SuccessfulMessage(ctx, "Admin Channel Updated", fmt.Sprintf("Admin channel updated to %s", newChannel.Mention()))
}

func (c *Channel) deleteAdminChannel(ctx ken.SubCommandContext) (err error) {
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

	q.DeleteAdminChannel(dbCtx, newChannel.ID)
	return discord.SuccessfulMessage(ctx, "Admin Channel Deleted", fmt.Sprintf("Admin channel deleted from %s", newChannel.Mention()))
}

func (c *Channel) listAdminChannel(ctx ken.SubCommandContext) (err error) {
	if err = ctx.Defer(); err != nil {
		log.Println(err)
		return err
	}
	if !discord.IsAdminRole(ctx, discord.AdminRoles...) {
		return discord.NotAdminError(ctx)
	}
	q := models.New(c.dbPool)
	dbCtx := context.Background()

	adminChannel, err := q.ListAdminChannel(dbCtx)
	if err != nil {
		log.Println(err)
		return discord.ErrorMessage(ctx, "Failed to get admin channel", "Unable to find admin channel")
	}

	msg := ""
	if len(adminChannel) == 0 {
		msg = "No Current Admin Channels"
	} else {
		for _, k := range adminChannel {
			msg += fmt.Sprintf("%s\n", discord.MentionChannel(k))
		}
	}
	return ctx.RespondEmbed(&discordgo.MessageEmbed{
		Title:       "Current Admin Channel",
		Description: fmt.Sprintf("%s", msg),
	})
}
