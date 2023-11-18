package inventory

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/mccune1224/betrayal/internal/data"
	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/mccune1224/betrayal/internal/services/inventory"
	"github.com/zekrotja/ken"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func (i *Inventory) create(ctx ken.SubCommandContext) (err error) {
	// Defer the command to make sure it is always responded to
	if err = ctx.Defer(); err != nil {
		log.Println(err)
		return err
	}
	if !discord.IsAdminRole(ctx, discord.AdminRoles...) {
		err = discord.ErrorMessage(
			ctx,
			"Unauthorized",
			"You are not authorized to use this command.",
		)
		return err
	}

	playerArg := ctx.Options().GetByName("user").UserValue(ctx)
	roleArg := ctx.Options().GetByName("role").StringValue()
	channelID := ctx.GetEvent().ChannelID

	caser := cases.Title(language.AmericanEnglish)
	roleArg = caser.String(roleArg)
	// Make sure role exists before creating inventory
	role, err := i.models.Roles.GetByFuzzy(roleArg)
	if err != nil {
		discord.ErrorMessage(ctx, "Failed to get Role", fmt.Sprintf("Cannot find role %s", roleArg))
		return err
	}
	inventoryCreateMsg := discordgo.MessageEmbed{
		Title:       "Creating Inventory...",
		Description: fmt.Sprintf("Creating inventory for %s", playerArg.Username),
	}

	pinMsg, err := ctx.GetSession().ChannelMessageSendEmbed(channelID, &inventoryCreateMsg)
	if err != nil {
		discord.ErrorMessage(ctx, "Failed to send message", err.Error())
		// delete message
		ctx.GetSession().ChannelMessageDelete(channelID, pinMsg.ID)
		return err
	}
	roleAbilities, err := i.models.Roles.GetAbilities(role.ID)
	if err != nil {
		discord.ErrorMessage(ctx, "Failed to get Role Abilities", err.Error())
		ctx.GetSession().ChannelMessageDelete(channelID, pinMsg.ID)
		return err
	}
	rolePerks, err := i.models.Roles.GetPerks(role.ID)
	if err != nil {
		discord.ErrorMessage(ctx, "Failed to get Role Perks", err.Error())
		ctx.GetSession().ChannelMessageDelete(channelID, pinMsg.ID)
		return err
	}
	abilityNames := make([]string, len(roleAbilities))
	for i, ability := range roleAbilities {
		chargeNumber := ""
		if ability.Charges == -1 {
			chargeNumber = "∞"
		} else {
			chargeNumber = fmt.Sprintf("%d", ability.Charges)
		}

		abilityNames[i] = fmt.Sprintf("%s [%s]", ability.Name, chargeNumber)
	}
	perkNames := make([]string, len(rolePerks))
	for i, perk := range rolePerks {
		perkNames[i] = perk.Name
	}

	defaultInv := &data.Inventory{
		DiscordID:      playerArg.ID,
		UserPinChannel: channelID,
		UserPinMessage: pinMsg.ChannelID,
		Alignment:      role.Alignment,
		RoleName:       role.Name,
		Abilities:      abilityNames,
		Perks:          perkNames,
		Coins:          defaultCoins,
		Luck:           defaultLuck,
		IsAlive:        true,
		ItemLimit:      defaultItemsLimit,
	}

	handler := inventory.InitInventoryHandler(i.models)
	err = handler.CreateInventory(defaultInv)
	if err != nil {
		log.Println(err)
		discord.ErrorMessage(ctx, "Failed to create inventory", "Unable to create inventory in database")
		return err
	}
	embd := InventoryEmbedBuilder(defaultInv, false)
	msg, err := ctx.GetSession().ChannelMessageEditEmbed(channelID, pinMsg.ID, embd)
	if err != nil {
		log.Println(err)
		discord.ErrorMessage(ctx, "Failed to edit message", fmt.Sprintf("Could not send to channel %s", discord.MentionChannel(channelID)))
		ctx.GetSession().ChannelMessageDelete(channelID, pinMsg.ID)
		return err
	}
	defaultInv.UserPinChannel = msg.ChannelID
	defaultInv.UserPinMessage = msg.ID
	err = i.models.Inventories.Update(defaultInv)
	if err != nil {
		log.Println(err)
		discord.ErrorMessage(ctx, "Failed to update inventory", fmt.Sprintf("Unable to update inventory for %s", playerArg.Username))
		ctx.GetSession().ChannelMessageDelete(channelID, pinMsg.ID)
		return err
	}
	err = ctx.GetSession().ChannelMessagePin(channelID, pinMsg.ID)
	if err != nil {
		log.Println(err)
		discord.ErrorMessage(ctx, "Failed to pin inventory message", fmt.Sprintf("Unable to pin inventory message for %s", playerArg.Username))
		return err
	}
	return discord.SuccessfulMessage(ctx, "Inventory Created", fmt.Sprintf("Created inventory for %s", playerArg.Username))
}
