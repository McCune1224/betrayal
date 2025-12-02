package echo

import (
	"fmt"
	"github.com/mccune1224/betrayal/internal/logger"

	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/zekrotja/ken"
)

type Echo struct {
	dbPool *pgxpool.Pool
}

var _ ken.SlashCommand = (*Echo)(nil)

func (e *Echo) Initialize(pool *pgxpool.Pool) {
	e.dbPool = pool
}

// Description implements ken.SlashCommand.
func (e *Echo) Description() string {
	return "Echo text via another channel"
}

// Name implements ken.SlashCommand.
func (e *Echo) Name() string {
	return "echo"
}

// Options implements ken.SlashCommand.
func (e *Echo) Options() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "message",
			Description: "Echo a message via another channel",
			Options: []*discordgo.ApplicationCommandOption{
				discord.ChannelCommandArg(true),
				discord.StringCommandArg("text", "Text to echo", true),
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "embed",
			Description: "Echo an embeded message via another channel",
			Options: []*discordgo.ApplicationCommandOption{
				discord.ChannelCommandArg(true),
				discord.StringCommandArg("title", "Title of the embed", true),
				discord.StringCommandArg("body", "Body of the embed", true),
			},
		},
	}
}

// Run implements ken.SlashCommand.
func (e *Echo) Run(ctx ken.Context) (err error) {
	defer logger.RecoverWithLog(*logger.Get())

	return ctx.HandleSubCommands(
		ken.SubCommandHandler{Name: "message", Run: e.message},
		ken.SubCommandHandler{Name: "embed", Run: e.embed},
	)
}

// Version implements ken.SlashCommand.
func (e *Echo) Version() string {
	return "1.0.0"
}

func (e *Echo) message(ctx ken.SubCommandContext) (err error) {
	if err = ctx.Defer(); err != nil {
		logger.Get().Error().Err(err).Msg("operation failed")
		return err
	}
	targetChannel := ctx.Options().GetByName("channel").ChannelValue(ctx)
	text := ctx.Options().GetByName("text").StringValue()
	_, err = ctx.GetSession().ChannelMessageSend(targetChannel.ID, text)
	if err != nil {
		logger.Get().Error().Err(err).Msg("operation failed")
		return err
	}
	return discord.SuccessfulMessage(ctx, "Message Sent", fmt.Sprintf("Sent message to %s", targetChannel.Mention()))
}

func (e *Echo) embed(ctx ken.SubCommandContext) (err error) {
	if err = ctx.Defer(); err != nil {
		logger.Get().Error().Err(err).Msg("operation failed")
		return err
	}

	targetChannel := ctx.Options().GetByName("channel").ChannelValue(ctx)
	msg := discordgo.MessageEmbed{}
	msg.Title = ctx.Options().GetByName("title").StringValue()
	msg.Fields = append(msg.Fields, &discordgo.MessageEmbedField{
		Value: ctx.Options().GetByName("body").StringValue(),
	})
	_, err = ctx.GetSession().ChannelMessageSendEmbed(targetChannel.ID, &msg)
	if err != nil {
		logger.Get().Error().Err(err).Msg("operation failed")
		return discord.AlexError(ctx, "Failed to send message")
	}
	return discord.SuccessfulMessage(ctx, "Message Sent", fmt.Sprintf("Sent message to %s", targetChannel.Mention()))
}
