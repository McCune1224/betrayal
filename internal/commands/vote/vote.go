package vote

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mccune1224/betrayal/internal/discord"
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
	return ctx.HandleSubCommands(
		ken.SubCommandHandler{Name: "batch", Run: v.batch},
		ken.SubCommandHandler{Name: "player", Run: v.player},
		ken.SubCommandHandler{Name: "location", Run: v.location},
	)
}

func (v *Vote) batch(ctx ken.SubCommandContext) (err error) {
	if err := ctx.Defer(); err != nil {
		log.Println(err)
		return err
	}

	users := []*discordgo.User{ctx.Options().GetByName("user").UserValue(ctx)}

	for i := 2; i <= 5; i++ {
		user, ok := ctx.Options().GetByNameOptional(fmt.Sprintf("user%d", i))
		if ok {
			users = append(users, user.UserValue(ctx))
		}
	}

	voteLogText := fmt.Sprintf("%s voted for", ctx.User().Username)
	for _, user := range users {
		voteLogText += fmt.Sprintf(" %s", user.Username)
	}

	sesh := ctx.GetSession()
	event := ctx.GetEvent()

	votedFor := ""
	for _, user := range users {
		votedFor += fmt.Sprintf("%s ", discord.MentionUser(user.ID))
	}

	successfullMsg := discordgo.MessageEmbed{
		Title:       fmt.Sprintf("%s %s %s", discord.EmojiSuccess, "Vote Sent for Processing", discord.EmojiSuccess),
		Description: fmt.Sprintf("Voted for %s", votedFor),
		Color:       discord.ColorThemeGreen,
	}

	q := models.New(v.dbPool)
	dbCtx := context.Background()
	voteChannelID, err := q.GetVoteChannel(dbCtx)
	if err != nil {
		log.Println(err)
		return discord.ErrorMessage(ctx, "Vote location not set", "Please have admin set a vote location using /vote location")
	}

	confSuccMsg, err := sesh.ChannelMessageSendEmbed(event.ChannelID, &successfullMsg)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Failed to send vote message")
	}

	_, err = sesh.ChannelMessageSend(voteChannelID, discord.Code(voteLogText)+"\n"+discord.AbsoluteTimestamp(time.Now().Unix())+" "+discord.MessageURL(confSuccMsg.Reference()))
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Failed to send vote message")
	}

	return ctx.RespondMessage(".")
}

func (v *Vote) player(ctx ken.SubCommandContext) (err error) {
	if err := ctx.Defer(); err != nil {
		log.Println(err)
		return err
	}
	targetVoteUser := ctx.Options().GetByName("user").UserValue(ctx)
	voteContext, ok := ctx.Options().GetByNameOptional("context")

	q := models.New(v.dbPool)
	dbCtx := context.Background()

	voterID, _ := util.Atoi64(ctx.GetEvent().Member.User.ID)
	_, err = q.GetPlayer(dbCtx, voterID)
	if err != nil {
		return discord.ErrorMessage(ctx, "You are not a player", "You must be a player to vote")
	}

	voteLogMsg := ""
	if ok {
		voteContext := voteContext.StringValue()
		voteLogMsg = fmt.Sprintf("%s voted for %s with context: %s", ctx.User().Username, targetVoteUser.Username, voteContext)
	} else {
		voteLogMsg = fmt.Sprintf("%s voted for %s", ctx.User().Username, targetVoteUser.Username)
	}

	voteChannel, err := q.GetVoteChannel(dbCtx)
	if err != nil {
		log.Println(err)
		return discord.ErrorMessage(ctx, "Vote location not set", "Please have admin set a vote location using /vote location")
	}
	sesh := ctx.GetSession()
	event := ctx.GetEvent()

	successfullMsg := discordgo.MessageEmbed{
		Title:       fmt.Sprintf("%s %s %s", discord.EmojiSuccess, "Vote Sent for Processing", discord.EmojiSuccess),
		Description: fmt.Sprintf("Voted for %s", discord.MentionUser(targetVoteUser.ID)),
		Color:       discord.ColorThemeGreen,
	}

	confSuccMsg, err := sesh.ChannelMessageSendEmbed(event.ChannelID, &successfullMsg)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Failed to send vote message")
	}

	_, err = sesh.ChannelMessageSend(voteChannel, discord.Code(voteLogMsg)+"\n"+discord.AbsoluteTimestamp(time.Now().Unix())+" "+discord.MessageURL(confSuccMsg.Reference()))
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Failed to send vote message")
	}

	return ctx.RespondMessage(".")
}

func (v *Vote) location(ctx ken.SubCommandContext) (err error) {
	if err := ctx.Defer(); err != nil {
		log.Println(err)
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
		log.Println(err)
		return discord.AlexError(ctx, "Failed to set vote location")
	}
	return discord.SuccessfulMessage(ctx, "Successfully set vote location", fmt.Sprintf("Vote location set to %s", targetChannel.Mention()))
}

// Version implements ken.SlashCommand.
func (*Vote) Version() string {
	return "1.0.0"
}
