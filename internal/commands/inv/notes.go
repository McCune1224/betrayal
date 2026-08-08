package inv

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/mccune1224/betrayal/internal/logger"
	"github.com/mccune1224/betrayal/internal/services/inventory"
	"github.com/mccune1224/betrayal/internal/services/playernotes"
	"github.com/zekrotja/ken"
)

func (i *Inv) notesCommandGroupBuilder() ken.SubCommandGroup {
	return ken.SubCommandGroup{Name: "notes", SubHandler: []ken.CommandHandler{
		ken.SubCommandHandler{Name: "add", Run: i.addNote},
		ken.SubCommandHandler{Name: "list", Run: i.listNote},
		ken.SubCommandHandler{Name: "update", Run: i.updateNote},
		ken.SubCommandHandler{Name: "remove", Run: i.removeNote},
	}}
}

func (i *Inv) notesCommandArgBuilder() *discordgo.ApplicationCommandOption {
	return &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
		Name:        "notes",
		Description: "add/remove a note in an inventory",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "add",
				Description: "Add a note",
				Options: []*discordgo.ApplicationCommandOption{
					discord.StringCommandArg("note", "Note to add", true),
					discord.UserCommandArg(false),
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "list",
				Description: "List notes",
				Options: []*discordgo.ApplicationCommandOption{
					discord.UserCommandArg(false),
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "update",
				Description: "Update a note",
				Options: []*discordgo.ApplicationCommandOption{
					discord.IntCommandArg("position", "Position of note to update", true),
					discord.StringCommandArg("note", "Note to update", true),
					discord.UserCommandArg(false),
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "remove",
				Description: "Remove a note",
				Options: []*discordgo.ApplicationCommandOption{
					discord.IntCommandArg("position", "Position of note to remove", true),
					discord.UserCommandArg(false),
				},
			},
		},
	}
}

func (i *Inv) addNote(ctx ken.SubCommandContext) (err error) {
	if err = ctx.Defer(); err != nil {
		logger.Get().Error().Err(err).Msg("operation failed")
		return err
	}
	if !discord.IsAdminRole(ctx, discord.AdminRoles...) {
		return discord.NotAdminError(ctx)
	}
	h, err := inventory.NewInventoryHandler(ctx, i.dbPool)
	if err != nil {
		logger.Get().Error().Err(err).Msg("operation failed")
		return discord.AlexError(ctx, "failed to init inv handler")
	}
	defer h.UpdateInventoryMessage(ctx.GetSession())
	noteArg := ctx.Options().GetByName("note").StringValue()
	_, err = playernotes.New(i.dbPool).Add(context.Background(), playernotes.DiscordAdminAuthorization(), h.GetPlayer().ID, noteArg)
	if err != nil {
		logger.Get().Error().Err(err).Msg("operation failed")
		return discord.AlexError(ctx, "Failed to add note")
	}

	return discord.SuccessfulMessage(ctx, "Note Added", fmt.Sprintf("Added note %s", noteArg))
}

func (i *Inv) listNote(ctx ken.SubCommandContext) (err error) {
	if err = ctx.Defer(); err != nil {
		logger.Get().Error().Err(err).Msg("operation failed")
		return err
	}
	if !discord.IsAdminRole(ctx, discord.AdminRoles...) {
		return discord.NotAdminError(ctx)
	}
	h, err := inventory.NewInventoryHandler(ctx, i.dbPool)
	if err != nil {
		logger.Get().Error().Err(err).Msg("operation failed")
		return discord.AlexError(ctx, "failed to init inv handler")
	}
	defer h.UpdateInventoryMessage(ctx.GetSession())
	playerNotes, err := playernotes.New(i.dbPool).List(context.Background(), playernotes.DiscordAdminAuthorization(), h.GetPlayer().ID)
	if err != nil {
		logger.Get().Error().Err(err).Msg("operation failed")
		return discord.AlexError(ctx, "Failed to get player notes")
	}

	fields := []*discordgo.MessageEmbedField{}
	for _, note := range playerNotes {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   fmt.Sprintf("%d. %s", note.Position, note.Info),
			Inline: false,
		})
	}
	return ctx.RespondEmbed(&discordgo.MessageEmbed{
		Title:       "Notes",
		Description: "Notes for player",
		Fields:      fields,
	})
}

func (i *Inv) removeNote(ctx ken.SubCommandContext) (err error) {
	if err = ctx.Defer(); err != nil {
		logger.Get().Error().Err(err).Msg("operation failed")
		return err
	}
	if !discord.IsAdminRole(ctx, discord.AdminRoles...) {
		return discord.NotAdminError(ctx)
	}
	h, err := inventory.NewInventoryHandler(ctx, i.dbPool)
	if err != nil {
		logger.Get().Error().Err(err).Msg("operation failed")
		return discord.AlexError(ctx, "failed to init inv handler")
	}
	defer h.UpdateInventoryMessage(ctx.GetSession())
	positionArg := ctx.Options().GetByName("position").IntValue()
	err = playernotes.New(i.dbPool).Delete(context.Background(), playernotes.DiscordAdminAuthorization(), h.GetPlayer().ID, int32(positionArg))
	if err != nil {
		logger.Get().Error().Err(err).Msg("operation failed")
		return discord.AlexError(ctx, "Failed to remove note")
	}

	return discord.SuccessfulMessage(ctx, "Note Removed", fmt.Sprintf("Removed note from position %d", positionArg))
}

func (i *Inv) updateNote(ctx ken.SubCommandContext) (err error) {
	if err = ctx.Defer(); err != nil {
		logger.Get().Error().Err(err).Msg("operation failed")
		return err
	}
	if !discord.IsAdminRole(ctx, discord.AdminRoles...) {
		return discord.NotAdminError(ctx)
	}
	h, err := inventory.NewInventoryHandler(ctx, i.dbPool)
	if err != nil {
		logger.Get().Error().Err(err).Msg("operation failed")
		return discord.AlexError(ctx, "failed to init inv handler")
	}
	defer h.UpdateInventoryMessage(ctx.GetSession())
	noteArg := ctx.Options().GetByName("note").StringValue()
	positionArg := ctx.Options().GetByName("position").IntValue()
	_, err = playernotes.New(i.dbPool).Update(context.Background(), playernotes.DiscordAdminAuthorization(), h.GetPlayer().ID, int(positionArg), noteArg)
	if err != nil {
		logger.Get().Error().Err(err).Msg("operation failed")
		return discord.AlexError(ctx, "Failed to update note")
	}

	return discord.SuccessfulMessage(ctx, "Note Updated", fmt.Sprintf("Updated note %s from position %d", noteArg, positionArg))
}
