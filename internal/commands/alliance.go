package commands

import (
	"errors"
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/mccune1224/betrayal/internal/data"
	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/mccune1224/betrayal/internal/scheduler"
	"github.com/mccune1224/betrayal/internal/services/alliance"
	"github.com/zekrotja/ken"
)

type Alliance struct {
	models    data.Models
	scheduler *scheduler.BetrayalScheduler
}

func (a *Alliance) Initialize(models data.Models, scheduler *scheduler.BetrayalScheduler) {
	a.models = models
	a.scheduler = scheduler
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

	handler := alliance.InitAllianceHandler(a.models)

	err = handler.CreateAllinaceRequest(aName, requester.ID)
	if err != nil {
		if errors.Is(err, alliance.ErrAlreadyExists) {
			return discord.ErrorMessage(ctx, "Alliance Already Exists", fmt.Sprintf("An alliance with the name %s already exists.", aName))
		}
		log.Println(err)
		return discord.AlexError(ctx, "Unable to create alliance request")
	}

	sentReq, err := a.models.Alliances.GetRequestByName(aName)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Oopsie Woopsie\t"+err.Error())
	}

	reqMsg := fmt.Sprintf("MAKE THIS LOOK NICE OR SOMETHING IUNNO \n%v", sentReq)

	_, err = s.ChannelMessageSend(e.ChannelID, discord.Code(reqMsg))
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Failed to send request message")
	}
	return discord.SuccessfulMessage(ctx, "Alliance Successfully Requested.", fmt.Sprintf("Your alliance request (%s) has been sent for review.", aName))
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
	alliance, err := a.models.Alliances.GetRequestByRequesterID(e.Member.User.ID)
	if err != nil {
		log.Println(err)
		return discord.ErrorMessage(ctx, "Unable to find alliance", "Are you sure you're the owner of an alliance?")
	}
	targetInv, err := a.models.Inventories.GetByDiscordID(target.ID)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Unable to get target inventory")
	}
	inviteMsg := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("%s You have been invited to join %s %s", discord.EmojiMail, alliance.Name, discord.EmojiMail),
		Description: fmt.Sprintf("To accept, type `/alliance accept %s`", alliance.Name),
	}
	_, err = s.ChannelMessageSendEmbed(targetInv.UserPinChannel, inviteMsg)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Unable to send invite")
	}
	return discord.SuccessfulMessage(ctx, "Invite Sent", fmt.Sprintf("Invite sent to %s", target.Username))
}

func (a *Alliance) accept(ctx ken.SubCommandContext) (err error) {
	return discord.AlexError(ctx, "Not Implemented")
}

func (a *Alliance) adminApprove(ctx ken.SubCommandContext) (err error) {
	if !discord.IsAdminRole(ctx, discord.AdminRoles...) {
		return discord.NotAdminError(ctx)
	}

	allianceName := ctx.Options().GetByName("name").StringValue()
	handler := alliance.InitAllianceHandler(a.models)
	newAlliance, err := handler.ApproveCreateRequest(allianceName, ctx.GetSession(), ctx.GetEvent())
	if err != nil {
		if errors.Is(err, alliance.ErrAlreadyExists) {
			return discord.ErrorMessage(ctx, "Alliance Already Exists", fmt.Sprintf("An alliance with the name %s already exists.", allianceName))
		}
		log.Println(err)
		return discord.AlexError(ctx, "Unable to complete alliance request")
	}
	return discord.SuccessfulMessage(ctx, fmt.Sprintf("Successfully created alliance %s", newAlliance.Name),
		fmt.Sprintf("Alliance %s has been created. Check it out in %s", newAlliance.Name, discord.MentionChannel(newAlliance.ChannelID)))
}

func (a *Alliance) adminDecline(ctx ken.SubCommandContext) (err error) {
	return discord.AlexError(ctx, "Not Implemented")
}

// Version implements ken.SlashCommand.
func (*Alliance) Version() string {
	return "1.0.0"
}
