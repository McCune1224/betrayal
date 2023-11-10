package commands

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/go-co-op/gocron"
	"github.com/mccune1224/betrayal/internal/data"
	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/mccune1224/betrayal/internal/util"
	"github.com/zekrotja/ken"
)

type ActionFunnel struct {
	models    data.Models
	scheduler *gocron.Scheduler
}

var _ ken.SlashCommand = (*ActionFunnel)(nil)

func (a *ActionFunnel) Initialize(models data.Models, scheduler *gocron.Scheduler) {
	a.models = models
	a.scheduler = scheduler
}

// Description implements ken.SlashCommand.
func (*ActionFunnel) Description() string {
	return "Send action request to admin for approval"
}

// Name implements ken.SlashCommand.
func (*ActionFunnel) Name() string {
	return discord.DebugCmd + "action"
}

// Options implements ken.SlashCommand.
func (*ActionFunnel) Options() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "req",
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
func (af *ActionFunnel) Run(ctx ken.Context) (err error) {
	return ctx.HandleSubCommands(
		ken.SubCommandHandler{Name: "req", Run: af.request},
		ken.SubCommandHandler{Name: "location", Run: af.location},
	)
}

// Version implements ken.SlashCommand.
func (*ActionFunnel) Version() string {
	return "1.0.0"
}

func (af *ActionFunnel) location(ctx ken.SubCommandContext) (err error) {
	locationArg := ctx.Options().GetByName("channel")
	if !discord.IsAdminRole(ctx, discord.AdminRoles...) {
		return discord.ErrorMessage(ctx,
			"You do not have permission to use this command",
			"You must be an admin to use this command.",
		)
	}

	gID := ctx.GetEvent().GuildID
	cID := locationArg.ChannelValue(ctx).ID
	location := data.FunnelChannel{
		ChannelID:  cID,
		GuildID:    gID,
		CurrentDay: 0,
	}
	_, err = af.models.FunnelChannels.Insert(&location)
	if err != nil {
		log.Println(err)
		return discord.ErrorMessage(ctx,
			"Error inserting funnel channel",
			"There was an error inserting the funnel channel. Let Alex know he's a bad programmer.")
	}

	_, err = af.models.FunnelChannels.Get(gID)
	if err != nil {
		log.Println(err)
		return discord.ErrorMessage(ctx,
			"Error getting funnel channel",
			"There was an error getting the funnel channel. Let Alex know he's a bad programmer.")
	}

	guildName, _ := ctx.GetSession().Guild(gID)
	if guildName == nil {
		return discord.ErrorMessage(ctx,
			"Error getting guild name",
			"There was an error getting the guild name. Let Alex know he's a bad programmer.")
	}

	channelName, _ := ctx.GetSession().Channel(cID)
	if channelName == nil {
		return discord.ErrorMessage(ctx,
			"Error getting channel name",
			"There was an error getting the channel name. Let Alex know he's a bad programmer.")
	}

	return discord.SuccessfulMessage(
		ctx,
		"Funnel Channel Set",
		fmt.Sprintf(
			"Action requests will now be sent to %s in %s",
			discord.MentionChannel(channelName.ID),
			guildName.Name,
		),
	)
}

func (af *ActionFunnel) request(ctx ken.SubCommandContext) (err error) {
	if err = ctx.Defer(); err != nil {
		log.Println(err)
		return err
	}
	event := ctx.GetEvent()

	inventory, err := af.models.Inventories.GetByPinChannel(event.ChannelID)
	if err != nil && err.Error() == "sql: no rows in result set" {
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
	// reqUser, err := ctx.GetSession().User(inventory.DiscordID)
	// if err != nil {
	// 	log.Println(err)
	// 	return discord.ErrorMessage(ctx, "Error getting user of confessional",
	// 		"There was an error getting the user. Let Alex know he's a bad programmer.")
	// }
	// actionRequest := data.Action{
	// 	RequestedAction:    reqArg,
	// 	RequesterID:        reqUser.ID,
	// 	RequestedChannelID: event.ChannelID,
	// 	RequestedMessageID: event.ID,
	// 	RequestedAt:        sqlReqTime,
	// 	// TODO: this is a hack for handling current "day" in game. Need to fix this, for now... Too bad!
	// 	RequestedDay: time.Now().Unix(),
	// }
	// _, err = af.models.Actions.Insert(&actionRequest)
	// if err != nil {
	// 	log.Println(err)
	// 	return discord.ErrorMessage(ctx, "Error inserting action request",
	// 		"There was an error inserting the action request. Let Alex know he's a bad programmer.")
	// }

	funnelChannel, err := af.models.FunnelChannels.Get(event.GuildID)
	if err != nil {
		log.Println(err)
		return discord.ErrorMessage(ctx, "Error getting funnel channel",
			"There was an error getting the funnel channel. Let Alex know he's a bad programmer.")
	}

	if funnelChannel == nil {
		log.Println(err)
		return discord.ErrorMessage(ctx, "Error getting funnel channel",
			"There was an error getting the funnel channel. Let Alex know he's a bad programmer.")
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
		"%s - %s - %s",
		guildMember.Nick,
		reqArg,
		humanReqTime,
	)

	// maybe will do something else with this but code block gives nice formatting
	// similar to that of what a logger would be...
	actionLog = discord.Code(actionLog)
	_, err = ctx.GetSession().ChannelMessageSend(funnelChannel.ChannelID, actionLog)
	if err != nil {
		log.Println(err)
		return discord.ErrorMessage(
			ctx,
			"Error sending action request",
			"Alex is a bad programmer.",
		)
	}

	return discord.SuccessfulMessage(
		ctx,
		"Action Requested",
		"Your action request has been for review.",
	)
}
