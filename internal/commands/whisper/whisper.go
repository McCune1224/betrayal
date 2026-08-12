package whisper

import (
	"context"
	cryptorand "crypto/rand"
	"errors"
	"math/big"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/mccune1224/betrayal/internal/logger"
	"github.com/mccune1224/betrayal/internal/models"
	whispersvc "github.com/mccune1224/betrayal/internal/services/whisper"
	"github.com/mccune1224/betrayal/internal/util"
	"github.com/zekrotja/ken"
)

const maxMessageLength = 1000

type Whisper struct{ dbPool *pgxpool.Pool }

var _ ken.SlashCommand = (*Whisper)(nil)

func (*Whisper) Name() string                    { return "whisper" }
func (*Whisper) Description() string             { return "Send a private bot-delivered message to a player" }
func (*Whisper) Version() string                 { return "1.0.0" }
func (w *Whisper) Initialize(pool *pgxpool.Pool) { w.dbPool = pool }

func (*Whisper) Options() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{{
		Type:        discordgo.ApplicationCommandOptionSubCommand,
		Name:        "send",
		Description: "Send a message to a player's complete linked group",
		Options: []*discordgo.ApplicationCommandOption{
			discord.UserCommandArg(true),
			discord.StringCommandArg("message", "Message to send", true),
		},
	}}
}

func (w *Whisper) Run(ctx ken.Context) error {
	defer logger.RecoverWithLog(*logger.Get())
	return ctx.HandleSubCommands(ken.SubCommandHandler{Name: "send", Run: w.send})
}

func (w *Whisper) send(ctx ken.SubCommandContext) error {
	if err := ctx.Defer(); err != nil {
		return err
	}
	event := ctx.GetEvent()
	if event == nil || event.Member == nil || event.Member.User == nil {
		return discord.ErrorMessage(ctx, "Whisper unavailable", "This command could not identify the sending player.")
	}
	senderID, err := util.Atoi64(event.Member.User.ID)
	if err != nil {
		return discord.ErrorMessage(ctx, "Whisper unavailable", "This command could not identify the sending player.")
	}
	targetOption, ok := ctx.Options().GetByNameOptional("user")
	if !ok || targetOption == nil {
		return discord.ErrorMessage(ctx, "Whisper unavailable", "A target player is required.")
	}
	target := targetOption.UserValue(ctx)
	if target == nil {
		return discord.ErrorMessage(ctx, "Whisper unavailable", "The target player could not be resolved.")
	}
	targetID, err := util.Atoi64(target.ID)
	if err != nil {
		return discord.ErrorMessage(ctx, "Whisper unavailable", "The target player could not be resolved.")
	}
	message := strings.TrimSpace(ctx.Options().GetByName("message").StringValue())
	if message == "" || len([]rune(message)) > maxMessageLength {
		return discord.ErrorMessage(ctx, "Whisper unavailable", "Your message must contain between 1 and 1000 characters.")
	}

	q := models.New(w.dbPool)
	dbCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if _, err := q.GetPlayer(dbCtx, targetID); err != nil {
		return discord.ErrorMessage(ctx, "Whisper unavailable", "That player is not available for whisper delivery.")
	}
	senderConf, err := q.GetPlayerConfessional(dbCtx, senderID)
	if err != nil {
		return discord.ErrorMessage(ctx, "Whisper unavailable", "Your confessional is not available for message receipts.")
	}

	rows, err := q.ListWhisperGroupMembers(dbCtx)
	if err != nil {
		return discord.ErrorMessage(ctx, "Whisper unavailable", "Whisper delivery is temporarily unavailable.")
	}
	confessionals := make([]whispersvc.PlayerConfessional, 0, len(rows)+1)
	for _, row := range rows {
		channelID := ""
		if row.ChannelID.Valid {
			channelID = util.Itoa64(row.ChannelID.Int64)
		}
		confessionals = append(confessionals, whispersvc.PlayerConfessional{PlayerID: row.PlayerID, ChannelID: channelID, GroupID: util.Itoa64(row.GroupID)})
	}
	if _, err := q.GetPlayerConfessional(dbCtx, targetID); err == nil {
		targetConf, _ := q.GetPlayerConfessional(dbCtx, targetID)
		found := false
		for _, conf := range confessionals {
			if conf.PlayerID == targetID {
				found = true
				break
			}
		}
		if !found {
			confessionals = append(confessionals, whispersvc.PlayerConfessional{PlayerID: targetID, ChannelID: util.Itoa64(targetConf.ChannelID)})
		}
	}
	recipients, err := whispersvc.ResolveRecipients(senderID, targetID, confessionals)
	if err != nil {
		if errors.Is(err, whispersvc.ErrSelfTarget) {
			return discord.ErrorMessage(ctx, "Whisper unavailable", "You cannot whisper to yourself.")
		}
		return discord.ErrorMessage(ctx, "Whisper unavailable", "That player is not available for complete whisper delivery.")
	}
	warnings, err := q.ListEnabledWhisperDoubtMessages(dbCtx)
	if err != nil {
		return discord.ErrorMessage(ctx, "Whisper unavailable", "Whisper delivery is temporarily unavailable.")
	}
	warningPool := make([]string, 0, len(warnings))
	for _, warning := range warnings {
		warningPool = append(warningPool, warning.Message)
	}

	sender := sessionSender{session: ctx.GetSession()}
	_, err = whispersvc.Deliver(whispersvc.DeliveryRequest{
		SenderChannelID:     util.Itoa64(senderConf.ChannelID),
		RecipientChannelIDs: recipients,
		Message:             message,
		Timestamp:           discord.AbsoluteTimestamp(time.Now().Unix()),
	}, sender, secureRoller{}, warningPool)
	if err != nil {
		return discord.ErrorMessage(ctx, "Whisper delivery failed", "The complete whisper could not be delivered. Please try again later.")
	}
	ctx.SetEphemeral(true)
	return ctx.RespondEmbed(&discordgo.MessageEmbed{Title: "Whisper sent", Description: "Your message was delivered by the bot.", Color: discord.ColorThemeGreen})
}

type sessionSender struct{ session *discordgo.Session }

func (s sessionSender) Send(channelID, content string) error {
	_, err := s.session.ChannelMessageSend(channelID, content)
	return err
}

type secureRoller struct{}

func (secureRoller) Hit(chance float64) bool {
	return secureRoller{}.Intn(10000) < int(chance*10000)
}
func (secureRoller) Intn(n int) int {
	if n <= 1 {
		return 0
	}
	number, err := cryptorand.Int(cryptorand.Reader, big.NewInt(int64(n)))
	if err != nil {
		return 0
	}
	return int(number.Int64())
}
