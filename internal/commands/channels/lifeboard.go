package channels

import (
	"github.com/mccune1224/betrayal/internal/logger"
	"context"
	"fmt"
	"sort"
	"strings"

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
		logger.Get().Error().Err(err).Msg("operation failed")
		return discord.AlexError(ctx, "Failed to get player statuses")
	}

	oldBoard, _ := q.GetPlayerLifeboard(dbCtx)
	if oldBoard.MessageID != "" {
		ctx.GetSession().ChannelMessageDelete(oldBoard.ChannelID, oldBoard.MessageID)
		q.DeletePlayerLifeboard(dbCtx)
	}

	msg, err := UserLifeboardMessageBuilder(ctx.GetSession(), playerStatuses)
	if err != nil {
		logger.Get().Error().Err(err).Msg("operation failed")
		return discord.AlexError(ctx, "Failed to build user lifeboard message")
	}

	sentMsg, err := ctx.GetSession().ChannelMessageSendEmbed(targetChannel.ID, msg)
	if err != nil {
		logger.Get().Error().Err(err).Msg("operation failed")
		return discord.AlexError(ctx, "Failed to send status message")
	}

	err = ctx.GetSession().ChannelMessagePin(targetChannel.ID, sentMsg.ID)
	if err != nil {
		logger.Get().Error().Err(err).Msg("operation failed")
		return discord.AlexError(ctx, "Failed to pin status message")
	}

	_, err = q.CreatePlayerLifeboard(dbCtx, models.CreatePlayerLifeboardParams{
		ChannelID: targetChannel.ID,
		MessageID: sentMsg.ID,
	})
	if err != nil {
		logger.Get().Error().Err(err).Msg("operation failed")
		return discord.AlexError(ctx, "Failed to create player lifeboard")
	}

	return discord.SuccessfulMessage(ctx, "Lifeboard Channel Set", fmt.Sprintf("Lifeboard set in %s", targetChannel.Mention()))
}

func UserLifeboardMessageBuilder(sesh *discordgo.Session, playerStatuses []models.ListPlayerLifeboardRow) (*discordgo.MessageEmbed, error) {
	aliveTally := 0
	fields := []*discordgo.MessageEmbedField{}

	// temporary struct so that I can sort by alive status as well as by Nick
	type MemberAlive struct {
		Member *discordgo.Member
		Alive  bool
	}
	activePlayers := []MemberAlive{}
	for _, s := range playerStatuses {
		dgMember, _ := sesh.GuildMember(discord.BetraylGuildID, util.Itoa64(s.ID))
		activePlayers = append(activePlayers, MemberAlive{dgMember, s.Alive})
	}

	// should be sorted by alive status first, then by nick
	sort.Slice(activePlayers, func(i, j int) bool {
		if activePlayers[i].Alive == activePlayers[j].Alive {
			l := strings.ToLower(activePlayers[i].Member.DisplayName())
			r := strings.ToLower(activePlayers[j].Member.DisplayName())
			return l < r
		}
		return activePlayers[i].Alive
	})

	for i := range activePlayers {
		name := activePlayers[i].Member.DisplayName()
		if activePlayers[i].Alive {
			aliveTally++
			fields = append(fields, &discordgo.MessageEmbedField{
				Name: fmt.Sprintf("%s %s", discord.EmojiAlive, name),
			})
		} else {
			fields = append(fields, &discordgo.MessageEmbedField{
				Name: fmt.Sprintf("%s %s", discord.EmojiDead, name),
			})
		}
	}
	msg := &discordgo.MessageEmbed{
		Title:       "Player Status Board",
		Description: fmt.Sprintf("%d/%d players alive", aliveTally, len(playerStatuses)),
		Fields:      fields,
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Last updated: " + util.GetEstTimeStamp() + " (EST)",
		},
	}
	return msg, nil
}
