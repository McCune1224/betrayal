package action

import (
	"context"
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/mccune1224/betrayal/internal/models"
	"github.com/mccune1224/betrayal/internal/services/inventory"
	"github.com/mccune1224/betrayal/internal/util"
	"github.com/zekrotja/ken"
)

type Action struct {
	dbPool *pgxpool.Pool
}

var _ ken.SlashCommand = (*Action)(nil)

// Description implements ken.SlashCommand.
func (a *Action) Description() string {
	return "ALl things actions. actions, items, abilities, and more."
}

// Name implements ken.SlashCommand.
func (a *Action) Name() string {
	return "action"
}

func (a *Action) Version() string {
	return "1.0.0"
}

// Initialize implements main.BetrayalCommand.
func (a *Action) Initialize(pool *pgxpool.Pool) {
	a.dbPool = pool
}

// Options implements ken.SlashCommand.
func (*Action) Options() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "request",
			Description: "Request an action to be performed",
			Options: []*discordgo.ApplicationCommandOption{
				discord.StringCommandArg("action", "Action to be preformed", true),
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "location",
			Description: "Select a channel to funnel actions request to",
			Options: []*discordgo.ApplicationCommandOption{
				discord.ChannelCommandArg(true),
			},
		},
	}
}

// Run implements ken.SlashCommand.
func (a *Action) Run(ctx ken.Context) (err error) {
	return ctx.HandleSubCommands(
		ken.SubCommandHandler{Name: "request", Run: a.request},
		ken.SubCommandHandler{Name: "location", Run: a.location},
	)
}

// Version implements ken.SlashCommand.

func (a *Action) location(ctx ken.SubCommandContext) (err error) {
	if err := ctx.Defer(); err != nil {
		log.Println(err)
		return err
	}

	locationArg := ctx.Options().GetByName("channel")
	if !discord.IsAdminRole(ctx, discord.AdminRoles...) {
		return discord.ErrorMessage(ctx,
			"You do not have permission to use this command",
			"You must be an admin to use this command.",
		)
	}

	cID := locationArg.ChannelValue(ctx).ID
	q := models.New(a.dbPool)
	dbCtx := context.Background()

	err = q.WipeActionChannel(dbCtx)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Failed to wipe action channel")
	}

	err = q.UpsertActionChannel(dbCtx, cID)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Failed to set action channel")
	}

	return discord.SuccessfulMessage(
		ctx,
		"Action Channel Set",
		fmt.Sprintf(
			"Action requests will now be sent to in %s",
			discord.MentionChannel(cID),
		),
	)
}

func (a *Action) request(ctx ken.SubCommandContext) (err error) {
	if err = ctx.Defer(); err != nil {
		log.Println(err)
		return err
	}
	event := ctx.GetEvent()

	messageId := event.Message.ID
	channelId := event.ChannelID
	guildId := event.GuildID

	url := fmt.Sprintf("https://discord.com/channels/%s/%s/%s", guildId, channelId, messageId)

	inventory, err := inventory.NewInventoryHandler(ctx, a.dbPool)
	if err != nil && err.Error() == "no rows in result set" {
		log.Println(err)
		return discord.ErrorMessage(
			ctx,
			"Failed to send action request",
			"Actions can only be requested in your confessional. Is this your confessional channel?",
		)
	}
	if err != nil {
		log.Println(err)
		return discord.ErrorMessage(ctx,
			"Unknown error processing action request",
			"Let Alex know he's a bad programmer.")
	}
	if inventory == nil {
		return discord.ErrorMessage(
			ctx,
			"Failed to send action request",
			"Actions can only be requested in your confessional. Is this your confessional channel?",
		)
	}

	reqArg := ctx.Options().GetByName("action").StringValue()
	// East coast time babyyy
	humanReqTime := util.GetEstTimeStamp()

	q := models.New(a.dbPool)
	dbCtx := context.Background()

	actionChannel, err := q.GetActionChannel(dbCtx)
	if err != nil {
		log.Println(err)
		return discord.ErrorMessage(ctx, "Error getting action channel",
			"There was an error getting the action channel. Let Alex know he's a bad programmer.")
	}
	// get the user's name in guild
	guildMember, err := ctx.GetSession().
		GuildMember(event.GuildID, event.Member.User.ID)
	if err != nil {
		log.Println(err)
		return discord.ErrorMessage(ctx, "Error getting guild member",
			"There was an error getting the guild member. Let Alex know he's a bad programmer.")
	}

	if guildMember.Nick == "" {
		guildMember.Nick = guildMember.User.Username
	}

	actionLog := fmt.Sprintf(
		"%s || %s || %s",
		guildMember.Nick,
		reqArg,
		humanReqTime,
	)

	// maybe will do something else with this but code block gives nice formatting
	// similar to that of what a logger would be...

	actionLog = fmt.Sprintf("%s %s", url, discord.Code(actionLog))
	_, err = ctx.GetSession().ChannelMessageSend(actionChannel, actionLog)
	if err != nil {
		log.Println(err)
		return discord.ErrorMessage(
			ctx,
			"Error sending action request",
			"Alex is a bad programmer.",
		)
	}

	return discord.SuccessfulMessage(ctx, "Action Requested", fmt.Sprintf("Request '%s' sent for processing", reqArg))
}
