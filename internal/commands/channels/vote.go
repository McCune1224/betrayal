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

func (c *Channel) voteCommandGroupBuilder() ken.SubCommandGroup {
	return ken.SubCommandGroup{Name: "vote", SubHandler: []ken.CommandHandler{
		ken.SubCommandHandler{Name: "update", Run: c.updateVoteChannel},
		ken.SubCommandHandler{Name: "view", Run: c.viewVoteChannel},
	}}
}

func (c *Channel) voteCommandArgBuilder() *discordgo.ApplicationCommandOption {
	return &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
		Name:        "vote",
		Description: "add/remove a vote channel in an inventory",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "update",
				Description: "Update the vote channel",
				Options: []*discordgo.ApplicationCommandOption{
					discord.ChannelCommandArg(true),
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "view",
				Description: "View the current vote channel",
			},
		},
	}
}

func (c *Channel) updateVoteChannel(ctx ken.SubCommandContext) (err error) {
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

	q.UpsertVoteChannel(dbCtx, newChannel.ID)
	return discord.SuccessfulMessage(ctx, "Vote Channel Updated", fmt.Sprintf("Vote channel updated to %s", newChannel.Mention()))
}

func (c *(Channel)) viewVoteChannel(ctx ken.SubCommandContext) (err error) {
	if err = ctx.Defer(); err != nil {
		log.Println(err)
		return err
	}
	if !discord.IsAdminRole(ctx, discord.AdminRoles...) {
		return discord.NotAdminError(ctx)
	}
	q := models.New(c.dbPool)
	dbCtx := context.Background()

	voteChannel, err := q.GetVoteChannel(dbCtx)
	if err != nil {
		log.Println(err)
		return discord.ErrorMessage(ctx, "Failed to get vote channel", "Unable to find vote channel")
	}

	return ctx.RespondEmbed(&discordgo.MessageEmbed{
		Title:       "Current Vote Channel",
		Description: fmt.Sprintf("Vote channel is %s", discord.MentionChannel(voteChannel)),
	})
}
