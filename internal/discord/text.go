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

func SendSilentError(ctx ken.Context, title string, message string) error {
	ctx.SetEphemeral(true)
	err := ctx.RespondError(message, title)
	ctx.SetEphemeral(false)
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
