package whisper

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/mccune1224/betrayal/internal/logger"
	"github.com/mccune1224/betrayal/internal/models"
	"github.com/zekrotja/ken"
)

func (w *Whisper) adminCommandGroupBuilder() ken.SubCommandGroup {
	return ken.SubCommandGroup{Name: "admin", SubHandler: []ken.CommandHandler{
		ken.SubCommandHandler{Name: "group-list", Run: w.adminGroupList},
		ken.SubCommandHandler{Name: "group-create", Run: w.adminGroupCreate},
		ken.SubCommandHandler{Name: "group-delete", Run: w.adminGroupDelete},
		ken.SubCommandHandler{Name: "member-add", Run: w.adminMemberAdd},
		ken.SubCommandHandler{Name: "member-remove", Run: w.adminMemberRemove},
		ken.SubCommandHandler{Name: "message-list", Run: w.adminMessageList},
		ken.SubCommandHandler{Name: "message-create", Run: w.adminMessageCreate},
		ken.SubCommandHandler{Name: "message-update", Run: w.adminMessageUpdate},
		ken.SubCommandHandler{Name: "message-enable", Run: w.adminMessageEnable},
		ken.SubCommandHandler{Name: "message-disable", Run: w.adminMessageDisable},
		ken.SubCommandHandler{Name: "message-delete", Run: w.adminMessageDelete},
	}}
}

func (w *Whisper) adminCommandArgBuilder() *discordgo.ApplicationCommandOption {
	return &discordgo.ApplicationCommandOption{
		Type: discordgo.ApplicationCommandOptionSubCommandGroup,
		Name: "admin", Description: "Manage whisper groups and doubt messages",
		Options: []*discordgo.ApplicationCommandOption{
			adminSubcommand("group-list", "List linked-player groups"),
			adminSubcommand("group-create", "Create a linked-player group", discord.StringCommandArg("name", "Group name", true)),
			adminSubcommand("group-delete", "Delete a linked-player group", discord.IntCommandArg("group_id", "Group ID", true)),
			adminSubcommand("member-add", "Add a player to a linked-player group", discord.IntCommandArg("group_id", "Group ID", true), discord.UserCommandArg(true)),
			adminSubcommand("member-remove", "Remove a player from a linked-player group", discord.IntCommandArg("group_id", "Group ID", true), discord.UserCommandArg(true)),
			adminSubcommand("message-list", "List doubt messages"),
			adminSubcommand("message-create", "Add a doubt message", discord.StringCommandArg("message", "Doubt message", true)),
			adminSubcommand("message-update", "Edit a doubt message", discord.IntCommandArg("message_id", "Message ID", true), discord.StringCommandArg("message", "Doubt message", true)),
			adminSubcommand("message-enable", "Enable a doubt message", discord.IntCommandArg("message_id", "Message ID", true)),
			adminSubcommand("message-disable", "Disable a doubt message", discord.IntCommandArg("message_id", "Message ID", true)),
			adminSubcommand("message-delete", "Soft-delete a doubt message", discord.IntCommandArg("message_id", "Message ID", true)),
		},
	}
}

func (w *Whisper) adminOptions() []*discordgo.ApplicationCommandOption {
	return w.adminCommandArgBuilder().Options
}

func adminSubcommand(name, description string, options ...*discordgo.ApplicationCommandOption) *discordgo.ApplicationCommandOption {
	return &discordgo.ApplicationCommandOption{Type: discordgo.ApplicationCommandOptionSubCommand, Name: name, Description: description, Options: options}
}

func (w *Whisper) adminContext(ctx ken.SubCommandContext) (context.Context, context.CancelFunc, error) {
	if err := ctx.Defer(); err != nil {
		return nil, nil, err
	}
	if !discord.IsAdminRole(ctx, discord.AdminRoles...) {
		return nil, nil, discord.NotAdminError(ctx)
	}
	dbCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	return dbCtx, cancel, nil
}

func (w *Whisper) adminGroupCreate(ctx ken.SubCommandContext) error {
	dbCtx, cancel, err := w.adminContext(ctx)
	if err != nil {
		return err
	}
	defer cancel()
	name := strings.TrimSpace(ctx.Options().GetByName("name").StringValue())
	if name == "" || len([]rune(name)) > 100 {
		return discord.ErrorMessage(ctx, "Invalid whisper group", "The group name must contain 1 to 100 characters.")
	}
	group, err := models.New(w.dbPool).CreateWhisperGroup(dbCtx, name)
	if err != nil {
		return whisperCommandDBError(ctx, "Could not create whisper group", err)
	}
	return discord.SuccessfulMessage(ctx, "Whisper group created", fmt.Sprintf("Created **%s** (group %d).", group.Name, group.ID))
}

func (w *Whisper) adminGroupList(ctx ken.SubCommandContext) error {
	dbCtx, cancel, err := w.adminContext(ctx)
	if err != nil {
		return err
	}
	defer cancel()
	rows, err := models.New(w.dbPool).ListWhisperGroups(dbCtx)
	if err != nil {
		return whisperCommandDBError(ctx, "Could not list whisper groups", err)
	}
	if len(rows) == 0 {
		return discord.SuccessfulMessage(ctx, "Whisper groups", "No whisper groups are configured.")
	}
	lines := make([]string, 0, len(rows))
	seen := make(map[int64]bool)
	for _, row := range rows {
		if seen[row.ID] {
			continue
		}
		seen[row.ID] = true
		lines = append(lines, fmt.Sprintf("**%s** — group `%d`", row.Name, row.ID))
	}
	return discord.SuccessfulMessage(ctx, "Whisper groups", strings.Join(lines, "\n"))
}

func (w *Whisper) adminGroupDelete(ctx ken.SubCommandContext) error {
	dbCtx, cancel, err := w.adminContext(ctx)
	if err != nil {
		return err
	}
	defer cancel()
	id := ctx.Options().GetByName("group_id").IntValue()
	if id <= 0 {
		return discord.ErrorMessage(ctx, "Invalid whisper group", "The group ID must be positive.")
	}
	if _, err := models.New(w.dbPool).GetWhisperGroup(dbCtx, id); err != nil {
		return whisperCommandDBError(ctx, "Could not find whisper group", err)
	}
	if err := models.New(w.dbPool).DeleteWhisperGroup(dbCtx, id); err != nil {
		return whisperCommandDBError(ctx, "Could not delete whisper group", err)
	}
	return discord.SuccessfulMessage(ctx, "Whisper group deleted", fmt.Sprintf("Deleted whisper group %d.", id))
}

func (w *Whisper) adminMemberMutation(ctx ken.SubCommandContext, add bool) error {
	dbCtx, cancel, err := w.adminContext(ctx)
	if err != nil {
		return err
	}
	defer cancel()
	groupID := ctx.Options().GetByName("group_id").IntValue()
	player := ctx.Options().GetByName("user").UserValue(ctx)
	if groupID <= 0 || player == nil {
		return discord.ErrorMessage(ctx, "Invalid whisper member", "A valid Discord player is required.")
	}
	playerID, err := strconv.ParseInt(player.ID, 10, 64)
	if err != nil || playerID <= 0 {
		return discord.ErrorMessage(ctx, "Invalid whisper member", "The selected Discord player is invalid.")
	}
	q := models.New(w.dbPool)
	if _, err = q.GetWhisperGroup(dbCtx, groupID); err != nil {
		return whisperCommandDBError(ctx, "Could not find whisper group", err)
	}
	if _, err = q.GetPlayer(dbCtx, playerID); err != nil {
		return whisperCommandDBError(ctx, "Could not find player", err)
	}
	if add {
		err = q.AddWhisperGroupMember(dbCtx, models.AddWhisperGroupMemberParams{GroupID: groupID, PlayerID: playerID})
	} else {
		err = q.RemoveWhisperGroupMember(dbCtx, models.RemoveWhisperGroupMemberParams{GroupID: groupID, PlayerID: playerID})
	}
	if err != nil {
		return whisperCommandDBError(ctx, "Could not update whisper group membership", err)
	}
	action := "Removed player"
	if add {
		action = "Added player"
	}
	return discord.SuccessfulMessage(ctx, "Whisper group updated", fmt.Sprintf("%s %d %s group %d.", action, playerID, map[bool]string{true: "to", false: "from"}[add], groupID))
}
func (w *Whisper) adminMemberAdd(ctx ken.SubCommandContext) error {
	return w.adminMemberMutation(ctx, true)
}
func (w *Whisper) adminMemberRemove(ctx ken.SubCommandContext) error {
	return w.adminMemberMutation(ctx, false)
}

func (w *Whisper) adminMessageCreate(ctx ken.SubCommandContext) error {
	dbCtx, cancel, err := w.adminContext(ctx)
	if err != nil {
		return err
	}
	defer cancel()
	message := strings.TrimSpace(ctx.Options().GetByName("message").StringValue())
	if message == "" || len([]rune(message)) > maxMessageLength {
		return discord.ErrorMessage(ctx, "Invalid doubt message", "The message must contain 1 to 1000 characters.")
	}
	created, err := models.New(w.dbPool).CreateWhisperDoubtMessage(dbCtx, message)
	if err != nil {
		return whisperCommandDBError(ctx, "Could not create doubt message", err)
	}
	return discord.SuccessfulMessage(ctx, "Doubt message created", fmt.Sprintf("Created doubt message %d and enabled it.", created.ID))
}

func (w *Whisper) adminMessageList(ctx ken.SubCommandContext) error {
	dbCtx, cancel, err := w.adminContext(ctx)
	if err != nil {
		return err
	}
	defer cancel()
	rows, err := models.New(w.dbPool).ListWhisperDoubtMessages(dbCtx)
	if err != nil {
		return whisperCommandDBError(ctx, "Could not list doubt messages", err)
	}
	if len(rows) == 0 {
		return discord.SuccessfulMessage(ctx, "Doubt messages", "No doubt messages are configured.")
	}
	lines := make([]string, 0, len(rows))
	for _, row := range rows {
		lines = append(lines, fmt.Sprintf("**%d** — %s (%s)", row.ID, row.Message, enabledLabel(row.Enabled)))
	}
	return discord.SuccessfulMessage(ctx, "Doubt messages", strings.Join(lines, "\n"))
}

func (w *Whisper) adminMessageUpdate(ctx ken.SubCommandContext) error {
	dbCtx, cancel, err := w.adminContext(ctx)
	if err != nil {
		return err
	}
	defer cancel()
	id := ctx.Options().GetByName("message_id").IntValue()
	message := strings.TrimSpace(ctx.Options().GetByName("message").StringValue())
	if id <= 0 || message == "" || len([]rune(message)) > maxMessageLength {
		return discord.ErrorMessage(ctx, "Invalid doubt message", "The message ID must be positive and the text must contain 1 to 1000 characters.")
	}
	current, err := models.New(w.dbPool).GetWhisperDoubtMessage(dbCtx, id)
	if err != nil {
		return whisperCommandDBError(ctx, "Could not find doubt message", err)
	}
	updated, err := models.New(w.dbPool).UpdateWhisperDoubtMessage(dbCtx, models.UpdateWhisperDoubtMessageParams{ID: id, Message: message, Enabled: current.Enabled})
	if err != nil {
		return whisperCommandDBError(ctx, "Could not update doubt message", err)
	}
	return discord.SuccessfulMessage(ctx, "Doubt message updated", fmt.Sprintf("Updated doubt message %d (%s).", updated.ID, enabledLabel(updated.Enabled)))
}
func (w *Whisper) adminMessageEnable(ctx ken.SubCommandContext) error {
	return w.adminMessageSetEnabled(ctx, true)
}
func (w *Whisper) adminMessageDisable(ctx ken.SubCommandContext) error {
	return w.adminMessageSetEnabled(ctx, false)
}
func (w *Whisper) adminMessageSetEnabled(ctx ken.SubCommandContext, enabled bool) error {
	dbCtx, cancel, err := w.adminContext(ctx)
	if err != nil {
		return err
	}
	defer cancel()
	id := ctx.Options().GetByName("message_id").IntValue()
	if id <= 0 {
		return discord.ErrorMessage(ctx, "Invalid doubt message", "The message ID must be positive.")
	}
	current, err := models.New(w.dbPool).GetWhisperDoubtMessage(dbCtx, id)
	if err != nil {
		return whisperCommandDBError(ctx, "Could not find doubt message", err)
	}
	updated, err := models.New(w.dbPool).UpdateWhisperDoubtMessage(dbCtx, models.UpdateWhisperDoubtMessageParams{ID: id, Message: current.Message, Enabled: enabled})
	if err != nil {
		return whisperCommandDBError(ctx, "Could not update doubt message", err)
	}
	return discord.SuccessfulMessage(ctx, "Doubt message updated", fmt.Sprintf("Doubt message %d is now %s.", updated.ID, enabledLabel(updated.Enabled)))
}
func (w *Whisper) adminMessageDelete(ctx ken.SubCommandContext) error {
	dbCtx, cancel, err := w.adminContext(ctx)
	if err != nil {
		return err
	}
	defer cancel()
	id := ctx.Options().GetByName("message_id").IntValue()
	if id <= 0 {
		return discord.ErrorMessage(ctx, "Invalid doubt message", "The message ID must be positive.")
	}
	if err := models.New(w.dbPool).DeleteWhisperDoubtMessage(dbCtx, id); err != nil {
		return whisperCommandDBError(ctx, "Could not delete doubt message", err)
	}
	return discord.SuccessfulMessage(ctx, "Doubt message removed", fmt.Sprintf("Soft-deleted doubt message %d.", id))
}

func enabledLabel(enabled bool) string {
	if enabled {
		return "enabled"
	}
	return "disabled"
}
func whisperCommandDBError(ctx ken.SubCommandContext, title string, err error) error {
	logger.Get().Error().Err(err).Msg("whisper admin command failed")
	return discord.ErrorMessage(ctx, title, "The whisper setting could not be changed.")
}
