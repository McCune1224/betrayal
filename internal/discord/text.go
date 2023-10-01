package discord

import (
	"github.com/bwmarrin/discordgo"
	"github.com/zekrotja/ken"
)

// ID of bot owner (me)
const McKusaID = "206268866714796032"

func Mention(userID string) string {
	return "<@" + userID + ">"
}

func Underline(s string) string {
	return "__" + s + "__"
}

func Bold(s string) string {
	return "**" + s + "**"
}

func Italic(s string) string {
	return "*" + s + "*"
}

// Temporary prefix for debugging commands.
const DebugCmd = "z_"

func SendSilentError(ctx ken.Context, title string, message string) (err error) {
	resp := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags: discordgo.MessageFlagsEphemeral,
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       title,
					Description: message,
					Color:       0xff0000,
				},
			},
		},
	}
	err = ctx.Respond(resp)
	return err
}

func UpdatePinnedMessage(
	ctx ken.Context,
	channelID string,
	messageID string,
	content string,
) (*discordgo.Message, error) {
	return ctx.GetSession().ChannelMessageEdit(channelID, messageID, content)
}
