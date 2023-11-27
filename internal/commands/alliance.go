package commands

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/mccune1224/betrayal/internal/data"
	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/mccune1224/betrayal/internal/scheduler"
	"github.com/mccune1224/betrayal/internal/services/alliance"
	"github.com/mccune1224/betrayal/internal/services/betrayal"
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
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "pending",
			Description: "view your pending alliance requests.",
			Options:     []*discordgo.ApplicationCommandOption{},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "leave",
			Description: "Leave an alliance",
			Options: []*discordgo.ApplicationCommandOption{
				discord.StringCommandArg("name", "name of the alliance", true),
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
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "delete",
					Description: "Delete an alliance and associated channel. (admin only)",
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
		ken.SubCommandHandler{Name: "pending", Run: a.pending},
		ken.SubCommandGroup{
			Name: "admin", SubHandler: []ken.CommandHandler{
				ken.SubCommandHandler{Name: "approve", Run: a.adminApprove},
				ken.SubCommandHandler{Name: "decline", Run: a.adminDecline},
				ken.SubCommandHandler{Name: "delete", Run: a.adminDelete},
			},
		},
	)
}

func (a *Alliance) request(ctx ken.SubCommandContext) (err error) {
	aName := ctx.Options().GetByName("name").StringValue()
	e := ctx.GetEvent()
	requester := e.Member.User

	requesterInventory, err := a.models.Inventories.GetByDiscordID(requester.ID)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, fmt.Sprintf("Unable to get/find %s's confessional channel", requester.Username))
	}

	if ctx.GetEvent().ChannelID != requesterInventory.UserPinChannel {
		return discord.ErrorMessage(ctx, "Invalid Channel", fmt.Sprintf("You must be in your %s channel to request an alliance.", discord.MentionChannel(requesterInventory.UserPinChannel)))
	}

	handler := alliance.InitAllianceHandler(a.models)
	err = handler.CreateAllinaceRequest(aName, requester.ID)
	if err != nil {
		if errors.Is(err, alliance.ErrCreateRequestAlreadyExists) {
			return discord.ErrorMessage(ctx, "Alliance Already Exists", fmt.Sprintf("An alliance with the name %s already exists.", aName))
		} else if errors.Is(err, alliance.ErrMemberAlreadyExists) {
			return discord.ErrorMessage(ctx, "Already Alliance Member", "You are already a member of an alliance.")
		} else if errors.Is(err, alliance.ErrAllianceAlreadyExists) {
			return discord.ErrorMessage(ctx, "Alliance Already Exists", fmt.Sprintf("An alliance with the name %s already exists.", aName))
		}
		log.Println(err)
		return discord.AlexError(ctx, "Unable to create alliance request")
	}

	return discord.SuccessfulMessage(ctx, "Alliance Successfully Requested.", fmt.Sprintf("Your alliance request '%s' has been sent for review.", aName))
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

	if target.ID == e.Member.User.ID {
		return discord.ErrorMessage(ctx, "Invalid Target", "You cannot invite yourself to an alliance.")
	}

	inviterInventory, err := a.models.Inventories.GetByDiscordID(e.Member.User.ID)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, fmt.Sprintf("Unable to get/find %s's confessional channel", e.Member.User.Username))
	}

	if inviterInventory.UserPinChannel != e.ChannelID {
		return discord.ErrorMessage(ctx, "Invalid Channel",
			fmt.Sprintf("You must be in your confessional %s channel to invite a player.", discord.MentionChannel(inviterInventory.UserPinChannel)))
	}

	handler := alliance.InitAllianceHandler(a.models)
	currentAlliance, err := a.models.Alliances.GetByMemberID(e.Member.User.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return discord.ErrorMessage(ctx, "Failed to find a valid alliance",
				"You must be the within an alliance to invite a player. Use `/alliance request` to create an alliance.")
		}
		log.Println(err)
	}
	err = handler.InvitePlayer(e.Member.User.ID, target.ID, currentAlliance.Name)
	if err != nil {
		if errors.Is(err, alliance.ErrAllianceNotFound) {
			return discord.ErrorMessage(ctx, "Failed to find a valid alliance",
				"You must be within an alliance to invite a player. Use `/alliance request` to create an alliance.")
		}
		log.Println(err)
		return discord.AlexError(ctx, fmt.Sprintf("Failed to invite create %s to alliance", discord.MentionUser(target.ID)))
	}
	inviteeInventory, err := a.models.Inventories.GetByDiscordID(target.ID)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx,
			fmt.Sprintf("Unable to get/find %s's confessional channel", discord.MentionUser(target.ID)))
	}

	_, err = s.ChannelMessageSendEmbed(inviteeInventory.UserPinChannel, &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("%s Alliance Invite", discord.EmojiInfo),
		Description: fmt.Sprintf("You have been invited to join %s. Type `/alliance accept %s` to accept the invite.", currentAlliance.Name, currentAlliance.Name),
	})
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, fmt.Sprintf("Failed to send invite message to %s", discord.MentionUser(target.ID)))
	}

	return discord.SuccessfulMessage(ctx, "Invite Sent", fmt.Sprintf("Invite sent to %s", discord.MentionUser(target.ID)))
}

func (a *Alliance) accept(ctx ken.SubCommandContext) (err error) {
	allianceName := ctx.Options().GetByName("name").StringValue()
	e := ctx.GetEvent()

	handler := alliance.InitAllianceHandler(a.models)
	existingAlliance, err := a.models.Alliances.GetByName(allianceName)
	if err != nil {
		log.Println(err)
		return discord.ErrorMessage(ctx, "Failed to find a valid alliance",
			"You must be a member of an alliance to accept an invite.")
	}
	playerInv, err := a.models.Inventories.GetByDiscordID(e.Member.User.ID)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, fmt.Sprintf("Unable to get/find %s's confessional channel", e.Member.User.Username))
	}

	bypassMemberLimit := false
	if len(existingAlliance.MemberIDs) == 4 {
		bypassMemberLimit = betrayal.AllianceMemberLimitBypass(playerInv.RoleName)
	}

	bypassAllianceLimit := false

	err = handler.AcceptInvite(e.Member.User.ID, allianceName, bypassMemberLimit, bypassAllianceLimit)
	if err != nil {
		if errors.Is(err, alliance.ErrAlreadyAllianceMember) {
			return discord.ErrorMessage(ctx, "Already Alliance Member", "You are already a member of an alliance.")
		} else if errors.Is(err, alliance.ErrInviteNotFound) {
			return discord.ErrorMessage(ctx, "Invite Not Found", "You do not have a pending invite for that alliance. see `/alliance pending`")
		} else if errors.Is(err, alliance.ErrAllianceMemberLimitExceeded) {
			return discord.ErrorMessage(ctx, "Alliance Member Limit Exceeded", "The alliance you are trying to join is full.")
		} else if errors.Is(err, alliance.ErrAllianceNotFound) {
			return discord.ErrorMessage(ctx, "Alliance Not Found", "The alliance you are trying to join does not exist.")
		} else if errors.Is(err, alliance.ErrMemberAlreadyExists) {
			return discord.ErrorMessage(ctx, "Already Alliance Member", "You are already a member of an alliance.")
		} else if errors.Is(err, alliance.ErrAllianceAlreadyExists) {
			return discord.ErrorMessage(ctx, "Alliance Already Exists", fmt.Sprintf("An alliance with the name %s already exists.", allianceName))
		}
		log.Println(err)
		return discord.AlexError(ctx, "Unable to accept invite")
	}
	return discord.SuccessfulMessage(ctx, "Alliance Joined", fmt.Sprintf("You have joined %s.", allianceName))
}

func (a *Alliance) adminApprove(ctx ken.SubCommandContext) (err error) {
	if !discord.IsAdminRole(ctx, discord.AdminRoles...) {
		return discord.NotAdminError(ctx)
	}

	allianceName := ctx.Options().GetByName("name").StringValue()
	handler := alliance.InitAllianceHandler(a.models)
	newAlliance, err := handler.ApproveCreateRequest(allianceName, ctx.GetSession(), ctx.GetEvent())
	if err != nil {
		if errors.Is(err, alliance.ErrCreateRequestAlreadyExists) {
			return discord.ErrorMessage(ctx, "Alliance Already Exists", fmt.Sprintf("An alliance with the name %s already exists.", allianceName))
		}
		log.Println(err)
		return discord.AlexError(ctx, "Unable to complete alliance request")
	}
	playerID := newAlliance.MemberIDs[0]
	playerInventory, err := a.models.Inventories.GetByDiscordID(playerID)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, fmt.Sprintf("Unable to get player %s inventory", discord.MentionUser(playerID)))
	}

	s := ctx.GetSession()
	_, err = s.ChannelMessageSendEmbed(playerInventory.UserPinChannel, &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("%s Alliance Created, %s", discord.EmojiInfo, discord.EmojiInfo),
		Description: fmt.Sprintf("Your alliance %s has been created. Check it out in %s. Start inviting people with `/alliance invite`", newAlliance.Name, discord.MentionChannel(newAlliance.ChannelID)),
	})
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Unable to send message to owner")
	}

	return discord.SuccessfulMessage(ctx, fmt.Sprintf("Successfully created alliance %s", newAlliance.Name),
		fmt.Sprintf("Alliance %s has been created. Check it out in %s", newAlliance.Name, discord.MentionChannel(newAlliance.ChannelID)))
}

func (a *Alliance) adminDecline(ctx ken.SubCommandContext) (err error) {
	if !discord.IsAdminRole(ctx, discord.AdminRoles...) {
		return discord.NotAdminError(ctx)
	}
	allianceName := ctx.Options().GetByName("name").StringValue()
	handler := alliance.InitAllianceHandler(a.models)

	pendingRequest, err := a.models.Alliances.GetRequestByName(allianceName)
	if err != nil {
		log.Println(err)
		return discord.ErrorMessage(ctx, "Alliance Not Found", fmt.Sprintf("An alliance with the name '%s' was not found.", allianceName))
	}

	err = handler.DeclineRequest(allianceName)
	if err != nil {
		if errors.Is(err, alliance.ErrCreateRequestAlreadyExists) {
			return discord.ErrorMessage(ctx, "Alliance Already Exists", fmt.Sprintf("An alliance with the name %s already exists.", allianceName))
		}
		log.Println(err)
		return discord.AlexError(ctx, "Unable to complete alliance request")
	}

	requesterInventory, err := a.models.Inventories.GetByDiscordID(pendingRequest.RequesterID)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Unable to get requester inventory")
	}

	s := ctx.GetSession()
	_, err = s.ChannelMessageSendEmbed(requesterInventory.UserPinChannel, &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("%s Alliance Request Declined", discord.EmojiInfo),
		Description: fmt.Sprintf("Your alliance create request for %s has been declined by admin.", allianceName),
	})
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Unable to send message to requester confessional")
	}

	return discord.SuccessfulMessage(ctx, "Alliance Request Declined", fmt.Sprintf("Alliance request %s has been declined.", allianceName))
}

// Version implements ken.SlashCommand.
func (*Alliance) Version() string {
	return "1.0.0"
}

func (a *Alliance) adminDelete(ctx ken.SubCommandContext) (err error) {
	allainceArgName := ctx.Options().GetByName("name").StringValue()
	handler := alliance.InitAllianceHandler(a.models)
	targetAlliance, err := a.models.Alliances.GetByName(allainceArgName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return discord.ErrorMessage(ctx, "Alliance Not Found",
				fmt.Sprintf("An alliance with the name %s was not found.", allainceArgName))
		}
		log.Println(err)
		return discord.AlexError(ctx, "Unable to get alliance")
	}

	withinAlliance := false
	for _, memberID := range targetAlliance.MemberIDs {
		if memberID == ctx.GetEvent().Member.User.ID {
			withinAlliance = true
			break
		}
	}

	if !discord.IsAdminRole(ctx, discord.AdminRoles...) || !withinAlliance {
		return discord.NotAdminError(ctx)
	}

	err = handler.DeleteAlliance(targetAlliance.Name, ctx.GetSession())
	if err != nil {
		if errors.Is(err, alliance.ErrAllianceNotFound) {
			return discord.ErrorMessage(ctx, "Alliance Not Found",
				fmt.Sprintf("An alliance with the name %s was not found.", allainceArgName))
		}
		log.Println(err)
		return discord.AlexError(ctx, "Unable to delete alliance or channel")
	}
	return discord.SuccessfulMessage(ctx, "Alliance Deleted", fmt.Sprintf("Alliance %s has been deleted.", targetAlliance.Name))
}

func (a *Alliance) pending(ctx ken.SubCommandContext) (err error) {
	if err = ctx.Defer(); err != nil {
		log.Println(err)
		return err
	}
	pendingInvites, err := a.models.Alliances.GetAllInvitesForUser(ctx.GetEvent().Member.User.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return discord.ErrorMessage(ctx, "No Pending Requests", "You have no pending alliance requests.")
		}
		log.Println(err)
	}

	msg := &discordgo.MessageEmbed{
		Title:       "Pending Alliance Requests",
		Description: fmt.Sprintf("You have %d pending invites", len(pendingInvites)),
	}
	fields := []*discordgo.MessageEmbedField{}
	for _, invite := range pendingInvites {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   invite.AllianceName,
			Value:  fmt.Sprintf("Invited by %s", discord.MentionUser(invite.InviterID)),
			Inline: false,
		})
	}

	msg.Fields = fields
	return ctx.RespondEmbed(msg)
}

func (a *Alliance) leave(ctx ken.SubCommandContext) (err error) {
	allianceName := ctx.Options().GetByName("name").StringValue()
	_, err = a.models.Alliances.GetByName(allianceName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return discord.ErrorMessage(ctx, "Alliance Not Found",
				fmt.Sprintf("An alliance with the name %s was not found.", allianceName))
		}
		log.Println(err)
		return discord.AlexError(ctx, "Unable to find alliance")
	}

	handler := alliance.InitAllianceHandler(a.models)
	err = handler.LeaveAlliance(ctx.GetEvent().Member.User.ID, allianceName, ctx.GetSession())
	if err != nil {
		if errors.Is(err, alliance.ErrAllianceNotFound) {
			return discord.ErrorMessage(ctx, "Alliance Not Found",
				fmt.Sprintf("An alliance with the name %s was not found.", allianceName))
		}
		log.Println(err)
		return discord.AlexError(ctx, "Unable to leave allinace")
	}
	return discord.SuccessfulMessage(ctx, fmt.Sprintf("Left alliance %s", allianceName), fmt.Sprintf("You have left alliance %s.", allianceName))
}
