package commands

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/mccune1224/betrayal/internal/data"
	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/mccune1224/betrayal/internal/util"
	"github.com/zekrotja/ken"
)

type Alliance struct {
	models data.Models
}

func (a *Alliance) SetModels(models data.Models) {
	a.models = models
}

var _ ken.SlashCommand = (*Alliance)(nil)

// Description implements ken.SlashCommand.
func (*Alliance) Description() string {
	return "Create and join alliances."
}

// Name implements ken.SlashCommand.
func (*Alliance) Name() string {
	return discord.DebugCmd + "alliance"
}

// Options implements ken.SlashCommand.
func (*Alliance) Options() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "request",
			Description: "Request to create an alliance.",
			Options: []*discordgo.ApplicationCommandOption{
				discord.StringCommandArg("name", "The name of the alliance.", true),
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "invite",
			Description: "Invite a player to join your alliance",
			Options: []*discordgo.ApplicationCommandOption{
				discord.UserCommandArg(true),
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "accept",
			Description: "Accept an invitation to join an alliance.",
			Options: []*discordgo.ApplicationCommandOption{
				discord.StringCommandArg("name", "The name of the alliance.", true),
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
			Name:        "admin",
			Description: "Admin commands for alliances.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "approve",
					Description: "Approve a request to create an alliance.",
					Options: []*discordgo.ApplicationCommandOption{
						discord.StringCommandArg("name", "The name of the alliance.", true),
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "decline",
					Description: "Decline a request to create an alliance.",
					Options: []*discordgo.ApplicationCommandOption{
						discord.StringCommandArg("name", "The name of the alliance.", true),
					},
				},
			},
		},
	}
}

// Run implements ken.SlashCommand.
func (a *Alliance) Run(ctx ken.Context) (err error) {
	return ctx.HandleSubCommands(
		ken.SubCommandHandler{Name: "request", Run: a.request},
		ken.SubCommandHandler{Name: "invite", Run: a.invite},
		ken.SubCommandHandler{Name: "accept", Run: a.accept},
		ken.SubCommandGroup{
			Name: "admin", SubHandler: []ken.CommandHandler{
				ken.SubCommandHandler{Name: "approve", Run: a.adminApprove},
				ken.SubCommandHandler{Name: "decline", Run: a.adminDecline},
			},
		},
	)
}

func (a *Alliance) request(ctx ken.SubCommandContext) (err error) {
	aName := ctx.Options().GetByName("name").StringValue()
	s := ctx.GetSession()
	e := ctx.GetEvent()
	requester := e.Member.User

	// Check to make sure they're not already the owner of an alliance.
	currReqs, err := a.models.Alliances.GetRequestByOwnerID(requester.ID)
	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
		return discord.AlexError(ctx)
	}
	if currReqs != nil {
		return discord.ErrorMessage(ctx, "Already Within Alliance",
			fmt.Sprintf("You already have a pending alliance creation request (%s).", currReqs.Name))
	}

	// Check to make sure they're not already a member of an alliance.
	currAlliances, err := a.models.Alliances.GetByMemberID(requester.ID)
	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
		return discord.AlexError(ctx)
	}
	if currAlliances != nil {
		return discord.ErrorMessage(ctx, "Already Within Alliance",
			fmt.Sprintf("You are already a member of an alliance (%s).", currAlliances.Name))
	}
	// Check to make sure the alliance name is not already taken.
	currAlliances, err = a.models.Alliances.GetByName(aName)
	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
		return discord.AlexError(ctx)
	}
	if currAlliances != nil {
		return discord.ErrorMessage(ctx, "Alliance Name Taken",
			fmt.Sprintf("The alliance name (%s) is already taken.", aName))
	}

	// Create the request.
	req := &data.AllianceRequest{
		RequesterID: requester.ID,
		Name:        aName,
	}
	err = a.models.Alliances.InsertRequest(req)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx)
	}
	// Send the request to the action channel
	actionChannel, err := a.models.FunnelChannels.Get(e.GuildID)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx)
	}
	reqMsg := fmt.Sprintf("%s - alliance create request: %s - %s", requester.Username, aName, util.GetEstTimeStamp())
	_, err = s.ChannelMessageSend(actionChannel.ChannelID, reqMsg)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx)
	}

	return discord.SuccessfulMessage(ctx, "Alliance Requested", fmt.Sprintf("Your alliance request (%s) has been sent for review.", aName))
}

func (a *Alliance) invite(ctx ken.SubCommandContext) (err error) {
	// This will more than likely take more than 3 seconds to complete.
	if err = ctx.Defer(); err != nil {
		log.Println(err)
		return err
	}
	target := ctx.Options().GetByName("user").UserValue(ctx)
	s := ctx.GetSession()
	e := ctx.GetEvent()
	alliance, err := a.models.Alliances.GetRequestByOwnerID(e.Member.User.ID)
	if err != nil {
		log.Println(err)
		return discord.ErrorMessage(ctx, "Unable to find alliance", "Are you sure you're the owner of an alliance?")
	}
	targetInv, err := a.models.Inventories.GetByDiscordID(target.ID)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx)
	}
	inviteMsg := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("%s You have been invited to join %s %s", discord.EmojiMail, alliance.Name, discord.EmojiMail),
		Description: fmt.Sprintf("To accept, type `/alliance accept %s`", alliance.Name),
	}
	_, err = s.ChannelMessageSendEmbed(targetInv.UserPinChannel, inviteMsg)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx)
	}
	return discord.SuccessfulMessage(ctx, "Invite Sent", fmt.Sprintf("Invite sent to %s", target.Username))
}

func (a *Alliance) accept(ctx ken.SubCommandContext) (err error) {
	return discord.AlexError(ctx)
}

func (a *Alliance) adminApprove(ctx ken.SubCommandContext) (err error) {
	return discord.AlexError(ctx)
}

func (a *Alliance) adminDecline(ctx ken.SubCommandContext) (err error) {
	return discord.AlexError(ctx)
}

// Version implements ken.SlashCommand.
func (*Alliance) Version() string {
	return "1.0.0"
}
