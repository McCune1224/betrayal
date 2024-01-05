package inventory

import (
	"errors"
	"fmt"
	"log"

	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/zekrotja/ken"
)

func (i *Inventory) addNote(ctx ken.SubCommandContext) (err error) {
	if err := ctx.Defer(); err != nil {
		log.Println(err)
		return err
	}
	handler, err := FetchHandler(ctx, i.models, true)
	if err != nil {
		if errors.Is(err, ErrNotAuthorized) {
			return discord.NotAdminError(ctx)
		}
		return discord.ErrorMessage(ctx, "Failed to find inventory.", "If not in confessional, please specify a user")
	}
	noteArg := ctx.Options().GetByName("message").StringValue()
	ctx.SetEphemeral(true)

	err = handler.AddNote(noteArg)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Failed to add note")
	}
	err = UpdateInventoryMessage(ctx.GetSession(), handler.GetInventory())
	if err != nil {
		log.Println(err)
	}

	return discord.SuccessfulMessage(ctx, "Added Note", fmt.Sprintf("Added note %s", noteArg))
}

func (i *Inventory) removeNote(ctx ken.SubCommandContext) (err error) {
	if err := ctx.Defer(); err != nil {
		log.Println(err)
		return err
	}
	handler, err := FetchHandler(ctx, i.models, true)
	if err != nil {
		if errors.Is(err, ErrNotAuthorized) {
			return discord.NotAdminError(ctx)
		}
		return discord.ErrorMessage(ctx, "Failed to find inventory.", "If not in confessional, please specify a user")
	}
	ctx.SetEphemeral(true)
	if len(handler.GetInventory().Notes) == 0 {
		return discord.ErrorMessage(ctx, "No notes to remove", "Nothing to see here officer...")
	}
	// Subtract 1 to account for 0 indexing (user input is 1 indexed)
	noteArg := int(ctx.Options().GetByName("index").IntValue()) - 1
	if noteArg < 0 || noteArg > len(handler.GetInventory().Notes)-1 {
		return discord.ErrorMessage(ctx, "Invalid note index",
			fmt.Sprintf("Please enter a number between 1 and %d", len(handler.GetInventory().Notes)))
	}
	err = handler.RemoveLimit(noteArg)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Failed to remove note")
	}
	err = UpdateInventoryMessage(ctx.GetSession(), handler.GetInventory())
	if err != nil {
		return discord.AlexError(ctx, "Failed to update inventory")
	}
	return discord.SuccessfulMessage(ctx, "Note removed", fmt.Sprintf("Removed note #%d", noteArg+1))
}
