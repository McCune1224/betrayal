package inventory

import (
	"errors"
	"fmt"
	"log"

	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/zekrotja/ken"
)

func (i *Inventory) SetRoleName(ctx ken.SubCommandContext) (err error) {
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
	roleArg := ctx.Options().GetByName("name").StringValue()
	role, err := i.models.Roles.GetByFuzzy(roleArg)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Failed to update role")
	}
	_, err = handler.SetRole(role.Name)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Failed to update role")
	}
	err = handler.SetAlignment(role.Alignment)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Failed to update alignment")
	}
	ctx.SetEphemeral(false)
	err = UpdateInventoryMessage(ctx.GetSession(), handler.GetInventory())
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Failed to update inventory message")
	}

	return discord.SuccessfulMessage(ctx, fmt.Sprintf("Updated role to %s", role.Name), fmt.Sprintf("Updated for %s", discord.MentionUser(handler.GetInventory().DiscordID)))
}
