package inventory

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/zekrotja/ken"
)

func (i *Inventory) addStatus(ctx ken.SubCommandContext) (err error) {
	ctx.SetEphemeral(false)
	inventory, err := Fetch(ctx, i.models, true)
	if err != nil {
		if errors.Is(err, ErrNotAuthorized) {
			return discord.NotAdminError(ctx)
		}
		return discord.ErrorMessage(ctx, "Failed to find inventory.", "If not in confessional, please specify a user")
	}
	statusNameArg := ctx.Options().GetByName("name").StringValue()
	status, err := i.models.Statuses.GetByName(statusNameArg)
	if err != nil {
		return discord.ErrorMessage(
			ctx,
			fmt.Sprint("Cannot find Status: ", statusNameArg),
			"Verify if the status exists.",
		)
	}

	inventory.Statuses = append(inventory.Statuses, status.Name)
	err = i.models.Inventories.UpdateStatuses(inventory)
	if err != nil {
		log.Println(err)
		return discord.ErrorMessage(
			ctx,
			"Failed to add status",
			"Alex is a bad programmer, and this is his fault.",
		)
	}
	err = i.updateInventoryMessage(ctx, inventory)
	if err != nil {
		log.Println(err)
	}

	err = discord.SuccessfulMessage(
		ctx,
		"Added Status",
		fmt.Sprintf("Status %s added", statusNameArg),
	)
	return err
}

func (i *Inventory) removeStatus(ctx ken.SubCommandContext) (err error) {
	inventory, err := Fetch(ctx, i.models, true)
	if err != nil {
		if errors.Is(err, ErrNotAuthorized) {
			return discord.NotAdminError(ctx)
		}
		return discord.ErrorMessage(ctx, "Failed to find inventory.", "If not in confessional, please specify a user")
	}
	ctx.SetEphemeral(false)

	statusArg := ctx.Options().GetByName("name").StringValue()

	for k, v := range inventory.Statuses {
		if strings.EqualFold(v, statusArg) {
			inventory.Statuses = append(inventory.Statuses[:k], inventory.Statuses[k+1:]...)
			err = i.models.Inventories.UpdateStatuses(inventory)
			if err != nil {
				log.Println(err)
				return discord.ErrorMessage(
					ctx,
					"Failed to remove status",
					"Alex is a bad programmer, and this is his fault.",
				)
			}
			err = i.updateInventoryMessage(ctx, inventory)
			if err != nil {
				return err
			}
			return discord.SuccessfulMessage(
				ctx,
				"Status removed from inventory",
				fmt.Sprintf("Removed %s from inventory.", statusArg),
			)
		}
	}

	discord.ErrorMessage(
		ctx,
		"Failed to find Status",
		fmt.Sprintf("Status %s not found in inventory.", statusArg),
	)
	return err
}
