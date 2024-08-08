package channels

import (
	"context"
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/mccune1224/betrayal/internal/models"
	"github.com/mccune1224/betrayal/internal/util"
	"github.com/zekrotja/ken"
)

func (c *Channel) lifeboardCommandGroupBuilder() ken.SubCommandGroup {
	return ken.SubCommandGroup{Name: "lifeboard", SubHandler: []ken.CommandHandler{
		ken.SubCommandHandler{Name: "set", Run: c.setLifeboardChannel},
	}}
}

func (c *Channel) lifeboardCommandArgBuilder() *discordgo.ApplicationCommandOption {
	return &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
		Name:        "lifeboard",
		Description: "update and view current lifeboard channel",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "set",
				Description: "Update the current action funnel channel",
				Options: []*discordgo.ApplicationCommandOption{
					discord.ChannelCommandArg(true),
				},
			},
		},
	}
}

func (c *Channel) setLifeboardChannel(ctx ken.SubCommandContext) (err error) {
	targetChannel := ctx.Options().GetByName("channel").ChannelValue(ctx)

	q := models.New(c.dbPool)
	dbCtx := context.Background()
	playerStatuses, err := q.ListPlayerLifeboard(dbCtx)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Failed to get player statuses")
	}

	oldBoard, _ := q.GetPlayerLifeboard(dbCtx)
	if oldBoard.MessageID != "" {
		ctx.GetSession().ChannelMessageDelete(oldBoard.ChannelID, oldBoard.MessageID)
		q.DeletePlayerLifeboard(dbCtx)
	}

	aliveTally := 0
	fields := []*discordgo.MessageEmbedField{}
	for i := range playerStatuses {
		if playerStatuses[i].Alive {
			aliveTally++
			fields = append(fields, &discordgo.MessageEmbedField{
				Name: fmt.Sprintf("%s - %s", discord.MentionUser(util.Itoa64(playerStatuses[i].ID)), discord.EmojiAlive),
			})
		} else {
			fields = append(fields, &discordgo.MessageEmbedField{
				Name: fmt.Sprintf("%s - %s", discord.MentionUser(util.Itoa64(playerStatuses[i].ID)), discord.EmojiDead),
			})
		}
	}

	msg := &discordgo.MessageEmbed{
		Title:       "Current Player Status Board",
		Description: fmt.Sprintf("%d/%d players alive", aliveTally, len(playerStatuses)),
		Fields:      fields,
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Last updated: " + util.GetEstTimeStamp(),
		},
	}

	sentMsg, err := ctx.GetSession().ChannelMessageSendEmbed(targetChannel.ID, msg)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Failed to send status message")
	}

	err = ctx.GetSession().ChannelMessagePin(targetChannel.ID, sentMsg.ID)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Failed to pin status message")
	}

	_, err = q.CreatePlayerLifeboard(dbCtx, models.CreatePlayerLifeboardParams{
		ChannelID: targetChannel.ID,
		MessageID: sentMsg.ID,
	})
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Failed to create player lifeboard")
	}

	return discord.SuccessfulMessage(ctx, "Lifeboard Channel Set", fmt.Sprintf("Lifeboard set in %s", targetChannel.Mention()))
}
