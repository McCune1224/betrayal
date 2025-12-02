package healthcheck

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/mccune1224/betrayal/internal/logger"
	"github.com/mccune1224/betrayal/internal/models"
	"github.com/zekrotja/ken"
)

type Healthcheck struct {
	dbPool *pgxpool.Pool
}

func (h *Healthcheck) Initialize(pool *pgxpool.Pool) {
	h.dbPool = pool
}

var _ ken.SlashCommand = (*Healthcheck)(nil)

// Name implements ken.SlashCommand.
func (*Healthcheck) Name() string {
	return "healthcheck"
}

// Description implements ken.SlashCommand.
func (*Healthcheck) Description() string {
	return "Admin only: Verify server configuration and channel setup"
}

// Version implements ken.SlashCommand.
func (*Healthcheck) Version() string {
	return "1.0.0"
}

// Options implements ken.SlashCommand.
func (*Healthcheck) Options() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{}
}

// Run implements ken.SlashCommand.
func (h *Healthcheck) Run(ctx ken.Context) (err error) {
	defer logger.RecoverWithLog(*logger.Get())

	if err = ctx.Defer(); err != nil {
		logger.Get().Error().Err(err).Msg("operation failed")
		return err
	}

	// Check admin role
	if !discord.IsAdminRole(ctx, discord.AdminRoles...) {
		return discord.NotAdminError(ctx)
	}

	// Run healthcheck
	return h.runHealthcheck(ctx)
}

func (h *Healthcheck) runHealthcheck(ctx ken.Context) error {
	q := models.New(h.dbPool)
	appCtx := context.Background()

	// Collect all status info
	var fields []*discordgo.MessageEmbedField

	// Check Admin Channels
	adminChannels, err := q.ListAdminChannel(appCtx)
	if err != nil {
		logger.Get().Error().Err(err).Msg("failed to get admin channels")
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:  "❌ Admin Channels",
			Value: fmt.Sprintf("Error fetching: %v", err),
		})
	} else if len(adminChannels) == 0 {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:  "⚠️ Admin Channels",
			Value: "No admin channels configured. Run `/channel admin add` to set up.",
		})
	} else {
		channelList := ""
		for _, chID := range adminChannels {
			channelList += fmt.Sprintf("%s\n", discord.MentionChannel(chID))
		}
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:  "✅ Admin Channels",
			Value: channelList,
		})
	}

	// Check Vote Channel
	voteChannel, err := q.GetVoteChannel(appCtx)
	if err != nil || voteChannel == "" {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:  "⚠️ Vote Channel",
			Value: "No vote channel configured. Run `/channel vote add` to set up.",
		})
	} else {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:  "✅ Vote Channel",
			Value: discord.MentionChannel(voteChannel),
		})
	}

	// Check Action Channel
	actionChannel, err := q.GetActionChannel(appCtx)
	if err != nil || actionChannel == "" {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:  "⚠️ Action Channel",
			Value: "No action channel configured. Run `/channel action add` to set up.",
		})
	} else {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:  "✅ Action Channel",
			Value: discord.MentionChannel(actionChannel),
		})
	}

	// Check Lifeboard
	lifeboard, err := q.GetPlayerLifeboard(appCtx)
	if err != nil || lifeboard.ChannelID == "" {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:  "⚠️ Lifeboard",
			Value: "No lifeboard configured. Run `/channel lifeboard set` to set up.",
		})
	} else {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:  "✅ Lifeboard",
			Value: discord.MentionChannel(lifeboard.ChannelID),
		})
	}

	// Check Player Confessionals
	confessionals, err := q.ListPlayerConfessional(appCtx)
	if err != nil {
		logger.Get().Error().Err(err).Msg("failed to get confessionals")
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:  "❌ Player Confessionals",
			Value: fmt.Sprintf("Error fetching: %v", err),
		})
	} else if len(confessionals) == 0 {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:  "ℹ️ Player Confessionals",
			Value: "No player confessionals created yet. Players will get these as they join the game.",
		})
	} else {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:  "✅ Player Confessionals",
			Value: fmt.Sprintf("%d confessional channel(s) active", len(confessionals)),
		})
	}

	// Check Players
	players, err := q.ListPlayer(appCtx)
	if err != nil {
		logger.Get().Error().Err(err).Msg("failed to get players")
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:  "❌ Players",
			Value: fmt.Sprintf("Error fetching: %v", err),
		})
	} else if len(players) == 0 {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:  "ℹ️ Players",
			Value: "No players registered in the game yet.",
		})
	} else {
		aliveCount := 0
		deadCount := 0
		for _, p := range players {
			if p.Alive {
				aliveCount++
			} else {
				deadCount++
			}
		}
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:  "✅ Players",
			Value: fmt.Sprintf("Total: %d | Alive: %d | Dead: %d", len(players), aliveCount, deadCount),
		})
	}

	// Check Current Cycle
	cycle, err := q.GetCycle(appCtx)
	if err != nil {
		logger.Get().Error().Err(err).Msg("failed to get game cycle")
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:  "ℹ️ Game Cycle",
			Value: "No game cycle initialized yet.",
		})
	} else {
		cycleType := "Day"
		if cycle.IsElimination {
			cycleType = "Elimination"
		}
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:  "✅ Game Cycle",
			Value: fmt.Sprintf("%s %d", cycleType, cycle.Day),
		})
	}

	// Determine overall health
	color := discord.ColorThemeGreen
	status := "✅ All systems operational"

	if len(adminChannels) == 0 || voteChannel == "" || actionChannel == "" {
		color = discord.ColorThemeOrange
		status = "⚠️ Missing critical channel configuration"
	}

	embed := &discordgo.MessageEmbed{
		Title:       "Server Healthcheck",
		Description: status,
		Fields:      fields,
		Color:       color,
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Run this before starting a new game to verify everything is configured correctly.",
		},
	}

	return ctx.RespondEmbed(embed)
}
