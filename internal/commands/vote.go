package commands

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/mccune1224/betrayal/internal/data"
	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/mccune1224/betrayal/internal/scheduler"
	"github.com/zekrotja/ken"
)

type Vote struct {
	models    data.Models
	scheduler *scheduler.BetrayalScheduler
}

// Initialize implements main.BetrayalCommand.
func (v *Vote) Initialize(models data.Models, scheduler *scheduler.BetrayalScheduler) {
	v.models = models
	v.scheduler = scheduler
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
	return discord.SuccessfulMessage(ctx, "Alex needs to make this", "TODO: stop being lazy")
}

func (v *Vote) player(ctx ken.SubCommandContext) (err error) {
	voteUser := ctx.Options().GetByName("user").UserValue(ctx)
	voteContext, ok := ctx.Options().GetByNameOptional("context")

	voteMsg := ""

	if ok {
		voteContext := voteContext.StringValue()
		voteMsg = fmt.Sprintf("%s voted for %s with context: %s", ctx.User().Username, voteUser.Username, voteContext)
	} else {
		voteMsg = fmt.Sprintf("%s voted for %s", ctx.User().Username, voteUser.Username)
	}

	voteChannel, err := v.models.Votes.Get()
	if err != nil {
		log.Println(err)
		return discord.ErrorMessage(ctx, "Vote location not set", "Please have admin set a vote location using /vote location")
	}
	sesh := ctx.GetSession()
	_, err = sesh.ChannelMessageSend(voteChannel.ChannelID, discord.Code(voteMsg))
	if err != nil {
		return discord.AlexError(ctx)
	}

	return discord.SuccessfulMessage(ctx, "Vote Sent for Processing.", fmt.Sprintf("Voted for %s", voteUser.Username))
}

func (v *Vote) location(ctx ken.SubCommandContext) (err error) {
	if !discord.IsAdminRole(ctx, discord.AdminRoles...) {
		return discord.NotAuthorizedError(ctx)
	}

	targetChannel := ctx.Options().GetByName("channel").ChannelValue(ctx)

	vote := data.Vote{
		ChannelID: targetChannel.ID,
	}
	err = v.models.Votes.Insert(vote)
	if err != nil {
		return discord.AlexError(ctx)
	}

	return discord.SuccessfulMessage(ctx, "Successfully set vote location", fmt.Sprintf("Vote location set to %s", targetChannel.Mention()))
}

// Version implements ken.SlashCommand.
func (*Vote) Version() string {
	return "1.0.0"
}
