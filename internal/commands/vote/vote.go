package vote

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/mccune1224/betrayal/internal/logger"
	"github.com/mccune1224/betrayal/internal/models"
	"github.com/mccune1224/betrayal/internal/util"
	"github.com/zekrotja/ken"
)

type Vote struct {
	dbPool *pgxpool.Pool
}

// Initialize implements main.BetrayalCommand.
func (v *Vote) Initialize(pool *pgxpool.Pool) {
	v.dbPool = pool
}

var _ ken.SlashCommand = (*Vote)(nil)

// Description implements ken.SlashCommand.
func (*Vote) Description() string {
	return "Vote a player"
}

// Name implements ken.SlashCommand.
func (*Vote) Name() string {
	return discord.DebugCmd + "vote"
}

// Options implements ken.SlashCommand.
func (*Vote) Options() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "batch",
			Description: "Batch vote players up to 5 players",
			Options: []*discordgo.ApplicationCommandOption{
				discord.UserCommandArg(true),
				{
					Type:        discordgo.ApplicationCommandOptionUser,
					Name:        "user2",
					Description: "User to vote",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionUser,
					Name:        "user3",
					Description: "User to vote",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionUser,
					Name:        "user4",
					Description: "User to vote",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionUser,
					Name:        "user5",
					Description: "User to vote",
					Required:    false,
				},
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "player",
			Description: "Vote a single player",
			Options: []*discordgo.ApplicationCommandOption{
				discord.UserCommandArg(true),
				discord.StringCommandArg("context", "Additional Context/Details to provide (i.e using Gold Card)", false),
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "location",
			Description: "Set location for vote logs (Admin Only)",
			Options: []*discordgo.ApplicationCommandOption{
				discord.ChannelCommandArg(true),
			},
		},
	}
}

// Run implements ken.SlashCommand.
func (v *Vote) Run(ctx ken.Context) (err error) {
	defer logger.RecoverWithLog(*logger.Get())

	return ctx.HandleSubCommands(
		ken.SubCommandHandler{Name: "batch", Run: v.batch},
		ken.SubCommandHandler{Name: "player", Run: v.player},
		ken.SubCommandHandler{Name: "location", Run: v.location},
	)
}

func (v *Vote) batch(ctx ken.SubCommandContext) (err error) {
	if err := ctx.Defer(); err != nil {
		logger.Get().Error().Err(err).Msg("operation failed")
		return err
	}

	sesh := ctx.GetSession()
	event := ctx.GetEvent()

	q := models.New(v.dbPool)
	dbCtx := context.Background()

	// Validate voter is a player
	voterID, _ := util.Atoi64(ctx.GetEvent().Member.User.ID)
	_, err = q.GetPlayer(dbCtx, voterID)
	if err != nil {
		return discord.ErrorMessage(ctx, "You are not a player", "You must be a player to vote")
	}

	// Get current game cycle
	cycle, err := q.GetCycle(dbCtx)
	if err != nil {
		logger.Get().Error().Err(err).Msg("failed to get game cycle")
		return discord.ErrorMessage(ctx, "Game cycle not found", "Please have admin set up the game cycle")
	}

	firstTarget, _ := sesh.GuildMember(discord.BetraylGuildID, ctx.Options().GetByName("user").UserValue(ctx).ID)
	votedMembers := []*discordgo.Member{firstTarget}

	for i := 2; i <= 5; i++ {
		user, ok := ctx.Options().GetByNameOptional(fmt.Sprintf("user%d", i))
		if ok {
			nextMember, _ := sesh.GuildMember(discord.BetraylGuildID, user.UserValue(ctx).ID)
			votedMembers = append(votedMembers, nextMember)
		}
	}

	// Store each vote in database
	for _, member := range votedMembers {
		targetID, _ := util.Atoi64(member.User.ID)

		// Validate target is a player
		_, err = q.GetPlayer(dbCtx, targetID)
		if err != nil {
			logger.Get().Warn().Str("target", member.DisplayName()).Msg("batch vote target is not a player, skipping")
			continue
		}

		_, err = q.UpsertVote(dbCtx, models.UpsertVoteParams{
			VoterID:       voterID,
			TargetID:      targetID,
			CycleDay:      cycle.Day,
			IsElimination: cycle.IsElimination,
			Weight:        1,
			Context:       pgtype.Text{Valid: false},
		})
		if err != nil {
			logger.Get().Error().Err(err).Str("target", member.DisplayName()).Msg("failed to save batch vote")
		}
	}

	voteLogText := fmt.Sprintf("%s voted for", ctx.User().Username)
	for _, member := range votedMembers {
		voteLogText += fmt.Sprintf(" %s", member.DisplayName())
	}

	votedFor := ""
	for _, member := range votedMembers {
		votedFor += fmt.Sprintf("%s ", discord.MentionUser(member.User.ID))
	}

	successfullMsg := discordgo.MessageEmbed{
		Title:       "Vote Sent for Processing",
		Description: fmt.Sprintf("Voted for %s", votedFor),
		Color:       discord.ColorThemeYellow,
	}

	voteChannelID, err := q.GetVoteChannel(dbCtx)
	if err != nil {
		logger.Get().Error().Err(err).Msg("operation failed")
		return discord.ErrorMessage(ctx, "Vote location not set", "Please have admin set a vote location using /vote location")
	}

	_, err = sesh.ChannelMessageSendEmbed(event.ChannelID, &successfullMsg)
	if err != nil {
		logger.Get().Error().Err(err).Msg("operation failed")
		return discord.AlexError(ctx, "Failed to send vote message")
	}

	_, err = sesh.ChannelMessageSend(voteChannelID, discord.Code(voteLogText))
	if err != nil {
		logger.Get().Error().Err(err).Msg("operation failed")
		return discord.AlexError(ctx, "Failed to send vote message")
	}

	return ctx.RespondMessage(".")
}

func (v *Vote) player(ctx ken.SubCommandContext) (err error) {
	if err := ctx.Defer(); err != nil {
		logger.Get().Error().Err(err).Msg("operation failed")
		return err
	}

	sesh := ctx.GetSession()
	event := ctx.GetEvent()

	targetVoteUser, _ := sesh.GuildMember(discord.BetraylGuildID, ctx.Options().GetByName("user").UserValue(ctx).ID)
	voteContext, ok := ctx.Options().GetByNameOptional("context")

	q := models.New(v.dbPool)
	dbCtx := context.Background()

	voterID, _ := util.Atoi64(ctx.GetEvent().Member.User.ID)
	_, err = q.GetPlayer(dbCtx, voterID)
	if err != nil {
		return discord.ErrorMessage(ctx, "You are not a player", "You must be a player to vote")
	}

	// Get target player ID
	targetID, _ := util.Atoi64(targetVoteUser.User.ID)
	_, err = q.GetPlayer(dbCtx, targetID)
	if err != nil {
		return discord.ErrorMessage(ctx, "Target is not a player", "You can only vote for registered players")
	}

	// Get current game cycle
	cycle, err := q.GetCycle(dbCtx)
	if err != nil {
		logger.Get().Error().Err(err).Msg("failed to get game cycle")
		return discord.ErrorMessage(ctx, "Game cycle not found", "Please have admin set up the game cycle")
	}

	voteLogMsg := ""
	var voteContextText pgtype.Text
	if ok {
		voteContextStr := voteContext.StringValue()
		voteContextText = pgtype.Text{String: voteContextStr, Valid: true}
		voteLogMsg = fmt.Sprintf("%s voted for %s with context: %s", event.Member.DisplayName(), targetVoteUser.DisplayName(), voteContextStr)
	} else {
		voteContextText = pgtype.Text{Valid: false}
		voteLogMsg = fmt.Sprintf("%s voted for %s", event.Member.DisplayName(), targetVoteUser.DisplayName())
	}

	// Store vote in database (upsert - will update if player already voted this cycle)
	_, err = q.UpsertVote(dbCtx, models.UpsertVoteParams{
		VoterID:       voterID,
		TargetID:      targetID,
		CycleDay:      cycle.Day,
		IsElimination: cycle.IsElimination,
		Weight:        1, // Default weight, can be modified for special abilities
		Context:       voteContextText,
	})
	if err != nil {
		logger.Get().Error().Err(err).Msg("failed to save vote to database")
		return discord.AlexError(ctx, "Failed to save vote")
	}

	voteChannel, err := q.GetVoteChannel(dbCtx)
	if err != nil {
		logger.Get().Error().Err(err).Msg("operation failed")
		return discord.ErrorMessage(ctx, "Vote location not set", "Please have admin set a vote location using /vote location")
	}

	successfullMsg := discordgo.MessageEmbed{
		Title:       "Vote Sent for Processing",
		Description: fmt.Sprintf("Voted for %s", discord.MentionUser(targetVoteUser.User.ID)),
		Color:       discord.ColorThemeYellow,
	}

	_, err = sesh.ChannelMessageSendEmbed(event.ChannelID, &successfullMsg)
	if err != nil {
		logger.Get().Error().Err(err).Msg("operation failed")
		return discord.AlexError(ctx, "Failed to send vote message")
	}

	_, err = sesh.ChannelMessageSend(voteChannel, discord.Code(voteLogMsg))
	if err != nil {
		logger.Get().Error().Err(err).Msg("operation failed")
		return discord.AlexError(ctx, "Failed to send vote message")
	}

	return ctx.RespondMessage(".")
}

func (v *Vote) location(ctx ken.SubCommandContext) (err error) {
	if err := ctx.Defer(); err != nil {
		logger.Get().Error().Err(err).Msg("operation failed")
		return err
	}
	if !discord.IsAdminRole(ctx, discord.AdminRoles...) {
		return discord.NotAdminError(ctx)
	}

	q := models.New(v.dbPool)
	dbCtx := context.Background()
	targetChannel := ctx.Options().GetByName("channel").ChannelValue(ctx)

	err = q.UpsertVoteChannel(dbCtx, targetChannel.ID)
	if err != nil {
		logger.Get().Error().Err(err).Msg("operation failed")
		return discord.AlexError(ctx, "Failed to set vote location")
	}
	return discord.SuccessfulMessage(ctx, "Successfully set vote location", fmt.Sprintf("Vote location set to %s", targetChannel.Mention()))
}

// Version implements ken.SlashCommand.
func (*Vote) Version() string {
	return "1.0.0"
}
