package inventory

import (
	"errors"
	"fmt"
	"log"

	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/mccune1224/betrayal/internal/services/inventory"
	"github.com/zekrotja/ken"
)

func (i *Inventory) addItem(ctx ken.SubCommandContext) (err error) {
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

	itemNameArg := ctx.Options().GetByName("name").StringValue()
	item, err := handler.AddItem(itemNameArg)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Failed to add item")
	}
	err = i.updateInventoryMessage(ctx, handler.GetInventory())
	if err != nil {
		log.Println(err)
	}

	err = discord.SuccessfulMessage(ctx, fmt.Sprintf("Added Item %s", item),
		fmt.Sprintf("Added item for %s", discord.MentionUser(handler.GetInventory().DiscordID)))
	return err
}

func (i *Inventory) removeItem(ctx ken.SubCommandContext) (err error) {
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
	ctx.SetEphemeral(false)
	itemArg := ctx.Options().GetByName("name").StringValue()
	item, err := handler.RemoveItem(itemArg)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Failed to remove item")
	}
	err = UpdateInventoryMessage(ctx.GetSession(), handler.GetInventory())
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Failed to update inventory message")
	}
	return discord.SuccessfulMessage(ctx, fmt.Sprintf("Removed item %s", item),
		fmt.Sprintf("Removed item %s to %s", item, discord.MentionUser(handler.GetInventory().DiscordID)))
}

func (i *Inventory) addItemLimit(ctx ken.SubCommandContext) (err error) {
  if err := ctx.Defer(); err != nil {
    log.Println(err)
    return err
  }
	handler, err := FetchHandler(ctx, i.models, true)
	if err != nil {
		log.Println(err)
		return discord.ErrorMessage(ctx, "Failed to find inventory.", "If not in confessional, please specify a user")
	}
	ctx.SetEphemeral(false)
	itemLimitArg := ctx.Options().GetByName("amount").IntValue()

	err = handler.AddLimit(int(itemLimitArg))
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Failed to add item limit")
	}
	err = i.updateInventoryMessage(ctx, handler.GetInventory())
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Failed to update inventory message")
	}
	return discord.SuccessfulMessage(ctx, "Item Limit Updated", fmt.Sprintf("Item limit set to %d", handler.GetInventory().ItemLimit))
}

func (i *Inventory) removeItemLimit(ctx ken.SubCommandContext) (err error) {
  if err := ctx.Defer(); err != nil {
    log.Println(err)
    return err
  }
	inv, err := Fetch(ctx, i.models, true)
	if err != nil {
		log.Println(err)
		return discord.ErrorMessage(ctx, "Failed to find inventory.", "If not in confessional, please specify a user")
	}
	ctx.SetEphemeral(false)
	itemLimitArg := ctx.Options().GetByName("amount").IntValue()
	ih := inventory.InitInventoryHandler(i.models, inv)
	err = ih.RemoveLimit(int(itemLimitArg))
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Failed to remove item limit")
	}
	err = i.updateInventoryMessage(ctx, inv)
	if err != nil {
		log.Println(err)
		return err
	}
	return discord.SuccessfulMessage(ctx, "Item Limit Updated", fmt.Sprintf("Item limit set to %d", inv.ItemLimit))
}


func (i *Inventory) setItemsLimit(ctx ken.SubCommandContext) (err error) {
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
	ctx.SetEphemeral(false)
	itemsLimitArg := ctx.Options().GetByName("size").IntValue()

	err = handler.SetLimit(int(itemsLimitArg))
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Failed to set item limit")
	}
	err = UpdateInventoryMessage(ctx.GetSession(), handler.GetInventory())
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Failed to update inventory message")
	}

	return discord.SuccessfulMessage(
		ctx, "Items Limit updated", fmt.Sprintf("Items Limit set to %d", handler.GetInventory().ItemLimit),
	)
}
