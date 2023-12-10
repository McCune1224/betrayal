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
	"github.com/mccune1224/betrayal/internal/util"
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
			Name:        "create",
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
					Name:        "create",
					Description: "Approve a request to create an alliance.",
					Options: []*discordgo.ApplicationCommandOption{
						discord.StringCommandArg("name", "The name of the alliance. (admin only)", true),
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "decline",
					Description: "Decline a request to create an alliance. (admin only)",
					Options: []*discordgo.ApplicationCommandOption{
						discord.StringCommandArg("name", "The name of the alliance.", true),
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "invite",
					Description: "Allow a member into an alliance. (admin only)",
					Options: []*discordgo.ApplicationCommandOption{
						discord.UserCommandArg(true),
						discord.ChannelCommandArg(true),
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "pending",
					Description: "View all pending alliance requests and invites. (admin only)",
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "wipe",
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
		ken.SubCommandHandler{Name: "create", Run: a.createRequest},
		ken.SubCommandHandler{Name: "invite", Run: a.invite},
		ken.SubCommandHandler{Name: "accept", Run: a.acceptRequest},
		ken.SubCommandHandler{Name: "pending", Run: a.pending},
		ken.SubCommandHandler{Name: "leave", Run: a.leave},
		ken.SubCommandGroup{
			Name: "admin", SubHandler: []ken.CommandHandler{
				ken.SubCommandHandler{Name: "create", Run: a.adminApproveCreate},
				ken.SubCommandHandler{Name: "decline", Run: a.adminDeclineCreate},
				ken.SubCommandHandler{Name: "invite", Run: a.adminApproveInvite},
				ken.SubCommandHandler{Name: "pending", Run: a.adminPending},
				ken.SubCommandHandler{Name: "wipe", Run: a.adminWipe},
			},
		},
	)
}

// Version implements ken.SlashCommand.
func (*Alliance) Version() string {
	return "1.0.0"
}

func (a *Alliance) createRequest(ctx ken.SubCommandContext) (err error) {
	if err = ctx.Defer(); err != nil {
		log.Println(err)
		return err
	}

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

	s := ctx.GetSession()
	actionChannel, err := discord.GetChannelByName(s, e, "action-funnel")
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Unable to get action funnel channel")
	}

	logMsg := discord.Code(fmt.Sprintf("%s - alliance create request '%s' - %s", requester.Username, aName, util.GetEstTimeStamp()))
	_, err = s.ChannelMessageSend(actionChannel.ID, logMsg)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Unable to log to action funnel channel")
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

	inviteeInventory, err := a.models.Inventories.GetByDiscordID(target.ID)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx,
			fmt.Sprintf("Unable to find existing player %s.", discord.MentionUser(target.ID)))
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

	_, err = s.ChannelMessageSendEmbed(inviteeInventory.UserPinChannel, &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("%s Alliance Invite", discord.EmojiInfo),
		Description: fmt.Sprintf("You have been invited to join %s. Type `/alliance accept %s` to request admin approval to join alliance.", currentAlliance.Name, currentAlliance.Name),
	})
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, fmt.Sprintf("Failed to send invite message to %s", discord.MentionUser(target.ID)))
	}

	actionChannel, err := discord.GetChannelByName(s, e, "action-funnel")
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Unable to get action funnel channel")
	}
	logMsg := discord.Code(fmt.Sprintf("%s - alliance invite for %s to %s - %s", e.Member.User.Username, target.Username, currentAlliance.Name, util.GetEstTimeStamp()))
	_, err = s.ChannelMessageSend(actionChannel.ID, logMsg)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Unable to log to action funnel channel")
	}

	return discord.SuccessfulMessage(ctx, "Alliance invite Sent", fmt.Sprintf("Invite sent to %s's confessional.", discord.MentionUser(target.ID)))
}

func (a *Alliance) acceptRequest(ctx ken.SubCommandContext) (err error) {
	if err = ctx.Defer(); err != nil {
		log.Println(err)
		return err
	}

	allianceName := ctx.Options().GetByName("name").StringValue()
	e := ctx.GetEvent()

	handler := alliance.InitAllianceHandler(a.models)
	playerInv, err := a.models.Inventories.GetByDiscordID(e.Member.User.ID)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, fmt.Sprintf("Unable to get/find %s's confessional channel", e.Member.User.Username))
	}

	if playerInv.UserPinChannel != e.ChannelID {
		return discord.ErrorMessage(ctx, "You must be in your confessional channel to accept an invite.", "Unable to accept invite")
	}

	err = handler.AcceptInvite(ctx.GetSession(), e.Member.User.ID, allianceName)
	if err != nil {
		if errors.Is(err, alliance.ErrAlreadyAllianceMember) {
			return discord.ErrorMessage(ctx, "Already Alliance Member", "You are already a member of an alliance.")
		} else if errors.Is(err, alliance.ErrInviteNotFound) {
			return discord.ErrorMessage(ctx, "Invite Not Found", "You do not have a pending invite for that alliance. see `/alliance pending`")
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

	s := ctx.GetSession()
	targetChannel, err := discord.GetChannelByName(s, e, "action-funnel")
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Unable to get action funnel channel")
	}

	logMsg := discord.Code(fmt.Sprintf("%s - alliance invite accepted for '%s' - %s", e.Member.User.Username, allianceName, util.GetEstTimeStamp()))
	_, err = s.ChannelMessageSend(targetChannel.ID, logMsg)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Unable to log to action funnel channel")
	}

	return discord.SuccessfulMessage(ctx,
		fmt.Sprintf("Request to accept joining %s sent.", allianceName),
		fmt.Sprintf("Request to join %s has been sent. An admin will need to approve your request.", allianceName))
}

func (a *Alliance) pending(ctx ken.SubCommandContext) (err error) {
	if err = ctx.Defer(); err != nil {
		log.Println(err)
		return err
	}
	pendingInvites, err := a.models.Alliances.GetAllInvitesForUser(ctx.GetEvent().Member.User.ID)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		log.Println(err)
		return discord.AlexError(ctx, "Failed to fetch all pending allinace invites")
	}

	msg := &discordgo.MessageEmbed{
		Title:       "Pending Alliance Requests",
		Description: fmt.Sprintf("You have %d pending invites. To join one, do `/alliance accept [name]`", len(pendingInvites)),
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
	if err = ctx.Defer(); err != nil {
		log.Println(err)
		return err
	}

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

	aHandler := alliance.InitAllianceHandler(a.models)
	currAlliance, err := a.models.Alliances.GetByMemberID(ctx.GetEvent().Member.User.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return discord.ErrorMessage(ctx, "Not in Alliance", "You are not in an alliance.")
		}
	}
	err = aHandler.LeaveAlliance(currAlliance, ctx.GetEvent().Member.User.ID, ctx.GetSession())
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

func (a *Alliance) adminApproveCreate(ctx ken.SubCommandContext) (err error) {
	if err = ctx.Defer(); err != nil {
		log.Println(err)
		return err
	}

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
		Title:       fmt.Sprintf("%s Alliance Created %s", discord.EmojiInfo, discord.EmojiInfo),
		Description: fmt.Sprintf("Your alliance %s has been created. Check it out in %s. Start inviting people with `/alliance invite`", newAlliance.Name, discord.MentionChannel(newAlliance.ChannelID)),
	})
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Unable to send message to owner")
	}

	return discord.SuccessfulMessage(ctx, fmt.Sprintf("Successfully created alliance %s", newAlliance.Name), "")
}

func (a *Alliance) adminDeclineCreate(ctx ken.SubCommandContext) (err error) {
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

func (a *Alliance) adminWipe(ctx ken.SubCommandContext) (err error) {
	if !discord.IsAdminRole(ctx, discord.AdminRoles...) {
		return discord.NotAdminError(ctx)
	}

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

func (a *Alliance) adminApproveInvite(ctx ken.SubCommandContext) (err error) {
	if !discord.IsAdminRole(ctx, discord.AdminRoles...) {
		return discord.NotAdminError(ctx)
	}

	invitee := ctx.Options().GetByName("user").UserValue(ctx)
	allianceChannel := ctx.Options().GetByName("channel").ChannelValue(ctx)

	inviteeInv, err := a.models.Inventories.GetByDiscordID(invitee.ID)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Failed to get invitee confessional channel")
	}

	handler := alliance.InitAllianceHandler(a.models)
	err = handler.AdminApproveInvite(ctx.GetSession(), invitee.ID, allianceChannel.ID)
	if err != nil {
		if errors.Is(err, alliance.ErrInviteNotFound) {
			return discord.ErrorMessage(ctx, "Invite Not Found", "You do not have a pending invite for that alliance. see `/alliance pending`")
		}
		if errors.Is(err, alliance.ErrAllianceNotFound) {
			return discord.ErrorMessage(ctx, "Alliance Not Found", "The alliance you are trying to join does not exist.")
		}

		log.Println(err)
		return discord.AlexError(ctx, "Unable to process admin invite.")
	}

	_, err = ctx.GetSession().ChannelMessageSendEmbed(inviteeInv.UserPinChannel, &discordgo.MessageEmbed{})
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Unable to send message to invitee")
	}

	return discord.SuccessfulMessage(ctx, "Invite Accepted", fmt.Sprintf("%s has been invited to %s.", discord.MentionUser(invitee.ID), allianceChannel.Name))
}

func (a *Alliance) adminPending(ctx ken.SubCommandContext) (err error) {
	if !discord.IsAdminRole(ctx, discord.AdminRoles...) {
		return discord.NotAdminError(ctx)
	}

	allCreateRequests, err := a.models.Alliances.GetAllRequests()
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Println(err)
		return discord.AlexError(ctx, "Unable to fetch all alliance requests")
	}
	allInvites, err := a.models.Alliances.GetAllInvites()
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Println(err)
		return discord.AlexError(ctx, "Unable to fetch all alliance invites")
	}

	createFields := []*discordgo.MessageEmbedField{}
	for _, v := range allCreateRequests {
		createFields = append(createFields, &discordgo.MessageEmbedField{
			Name:   v.Name,
			Value:  fmt.Sprintf("Requested by %s", discord.MentionUser(v.RequesterID)),
			Inline: false,
		})
	}

	inviteFields := []*discordgo.MessageEmbedField{}
	for _, v := range allInvites {
		inviteFields = append(inviteFields, &discordgo.MessageEmbedField{
			Name:   v.AllianceName,
			Value:  fmt.Sprintf("%s Invited by %s", discord.MentionUser(v.InviteeID), discord.MentionUser(v.InviterID)),
			Inline: false,
		})
	}

	msg := &discordgo.MessageEmbed{
		Title:  "Pending Alliance Creates and Player Invite Requests",
		Fields: append(createFields, inviteFields...),
	}

	return ctx.RespondEmbed(msg)
}
