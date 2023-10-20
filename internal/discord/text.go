package discord

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/zekrotja/ken"
)

type Emoji string

// String implements fmt.Stringer.
func (e Emoji) String() string {
	return string(e)
}

// status emojis
const (
	EmojiSuccess = Emoji("✅")
	EmojiError   = Emoji("❌")
	EmojiWarning = Emoji("⚠️")
	EmojiInfo    = Emoji("ℹ️")
)

// inventory emojis
const (
	EmojiInventory  = Emoji("🎒")
	EmojiAlignment  = Emoji("⚖️")
	EmojiAbility    = Emoji("💪")
	EmojiPerk       = Emoji("➕")
	EmojiItem       = Emoji("📦")
	EmojiStatus     = Emoji("🔵")
	EmojiImmunity   = Emoji("🛡️")
	EmojiEffect     = Emoji("🌟")
	EmojiCoins      = Emoji("💰")
	EmojiCoinBonus  = Emoji("🔥")
	EmojiNote       = Emoji("📝")
	EmojiAnyAbility = Emoji("🔮")
	EmojiLimit      = Emoji("📏")
	EmojiLuck       = Emoji("🍀")
)

// Hex colors / color themes
const (
	ColorThemeRed      = 0xff0000
	ColorThemeGreen    = 0x00ff00
	ColorThemeBlue     = 0x0000ff
	ColorThemeYellow   = 0xffff00
	ColorThemePurple   = 0xff00ff
	ColorThemeOrange   = 0xffa500
	ColorThemePink     = 0xffc0cb
	ColorThemeBlack    = 0x000000
	ColorThemeWhite    = 0xffffff
	ColorThemeGrey     = 0x808080
	ColorThemeBrown    = 0x8b4513
	ColorThemeGold     = 0xffd700
	ColorThemeSilver   = 0xc0c0c0
	ColorThemeBronze   = 0xcd7f32
	ColorThemeCopper   = 0xb87333
	ColorThemePlatinum = 0xe5e4e2
	ColorThemeDiamond  = 0x00ffff
	ColorThemeEmerald  = 0x50c878
	ColorThemeRuby     = 0xe0115f
	ColorThemeSapphire = 0x082567
	ColorThemeAmethyst = 0x9966cc
	ColorThemeTopaz    = 0xffc87c
	ColorThemePearl    = 0xfdeef4
	ColorThemeOpal     = 0x9fd8cb

	ColorItemCommon    = 0x00ff00
	ColorItemUncommon  = 0x0000ff
	ColorItemRare      = 0x00008b
	ColorItemEpic      = 0xff00ff
	ColorItemLegendary = 0xffc0cb
	ColorItemMythical  = 0x800080
	ColorItemUnique    = 0xffa500
)

// ID of bot owner
const McKusaID = "206268866714796032"

func MentionUser(userID string) string {
	return "<@" + userID + ">"
}

func MentionChannel(channelID string) string {
	return "<#" + channelID + ">"
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

func Strike(s string) string {
	return "~~" + s + "~~"
}

func Code(s string) string {
	return "`" + s + "`"
}

// Temporary prefix for debugging commands.
const DebugCmd = "z_"

// Send Pre-Formatted Error Message after slash command
func ErrorMessage(ctx ken.Context, title string, message string) (err error) {
	// default to ephemeral, but sometimes we want to show the error to everyone
	resp := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       fmt.Sprintf("%s %s %s", EmojiError, title, EmojiError),
					Description: message,
					Color:       ColorThemeRed,
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

// Send Pre-Formatted Successful Message after slash command
func SuccessfulMessage(ctx ken.Context,
	title string,
	message string,
) (err error) {
	resp := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags: 0,
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       fmt.Sprintf("%s %s %s", EmojiSuccess, title, EmojiSuccess),
					Description: message,
					Color:       ColorThemeGreen,
				},
			},
		},
	}
	err = ctx.Respond(resp)
	return err
}

func SilentSuccessfulMessage(ctx ken.Context,
	title string,
	message string,
) (err error) {
	resp := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags: discordgo.MessageFlagsEphemeral,
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       fmt.Sprintf("%s %s %s", EmojiSuccess, title, EmojiSuccess),
					Description: message,
					Color:       ColorThemeGreen,
				},
			},
		},
	}
	err = ctx.Respond(resp)
	return err
}

func WarningMessage(ctx ken.Context,
	title string,
	message string,
) (err error) {
	resp := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags: 0,
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       fmt.Sprintf("%s %s %s", EmojiWarning, title, EmojiWarning),
					Description: message,
					Color:       ColorThemeYellow,
				},
			},
		},
	}

	return ctx.Respond(resp)
}

func SilentWarningMessage(ctx ken.Context,
	title string,
	message string,
) (err error) {
	resp := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags: discordgo.MessageFlagsEphemeral,
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       fmt.Sprintf("%s %s %s", EmojiWarning, title, EmojiWarning),
					Description: message,
					Color:       ColorThemeYellow,
				},
			},
		},
	}

	return ctx.Respond(resp)
}
