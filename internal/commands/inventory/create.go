package inventory

import (
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/lib/pq"
	"github.com/mccune1224/betrayal/internal/data"
	"github.com/mccune1224/betrayal/internal/discord"
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

	defaultInv = roleInventoryBuilder(defaultInv)

	_, err = i.models.Inventories.Insert(defaultInv)
	if err != nil {
		log.Println(err)
		discord.ErrorMessage(ctx, "Alex is a bad programmer", "Failed to insert inventory")
		return err
	}
	embd := InventoryEmbedBuilder(defaultInv, false)
	msg, err := ctx.GetSession().ChannelMessageEditEmbed(channelID, pinMsg.ID, embd)
	if err != nil {
		log.Println(err)
		discord.ErrorMessage(ctx, "Alex is a bad programmer", "Failed to edit message")
		ctx.GetSession().ChannelMessageDelete(channelID, pinMsg.ID)
		return err
	}
	defaultInv.UserPinChannel = msg.ChannelID
	defaultInv.UserPinMessage = msg.ID
	err = i.models.Inventories.Update(defaultInv)
	if err != nil {
		log.Println(err)
		discord.ErrorMessage(ctx, "Alex is a bad programmer", "Failed to set Pinned Message")
		ctx.GetSession().ChannelMessageDelete(channelID, pinMsg.ID)
		return err
	}
	err = ctx.GetSession().ChannelMessagePin(channelID, pinMsg.ID)
	if err != nil {
		discord.ErrorMessage(ctx, "Failed to Pin Inventoory Message", err.Error())
		return err
	}
	return discord.SuccessfulMessage(ctx, "Inventory Created", fmt.Sprintf("Created inventory for %s", playerArg.Username))
}

// Handle edge cases for special roles with non-default setups (immunities, item limit...)
func roleInventoryBuilder(initInv *data.Inventory) *data.Inventory {
	inv := initInv

	// FIXME: Lord please forgive for the unholy amount of switch statements I am about to unleash
	// Will need to make some sort of Website or UI to allow for custom roles to be created instead of me hardcoding them
	roleName := strings.ToLower(inv.RoleName)
	switch roleName {
	// --- GOOD ROLES ---
	case "cerberus":
		// Due to perk Hades' Hound
		inv.Immunities = pq.StringArray{"Frozen", "Burned"}
	case "detective":
		// Due to perk Clever
		inv.Immunities = pq.StringArray{"Blackmailed", "Disabled", "Despaired"}
	case "fisherman":
		// Due to perk Barrels
		inv.ItemLimit = 8
	case "hero":
		// Due to perk Compos Mentis
		inv.Immunities = pq.StringArray{"Madness"}
	case "nurse":
		// Due to perk Powerful Immunity
		inv.Immunities = pq.StringArray{"Death Cursed", "Frozen", "Paralyzed", "Burned", "Empowered", "Drunk", "Restrained", "Disabled", "Blackmailed", "Despaired", "Madness", "Unlucky"}
	case "terminal":
		// Due to perk Heartbeats
		inv.Immunities = pq.StringArray{"Death Cursed", "Frozen", "Paralyzed", "Burned", "Empowered", "Drunk", "Restrained", "Disabled", "Blackmailed", "Despaired", "Madness", "Unlucky"}
	case "wizard":
		// due to perk Magic Barrier
		inv.Immunities = pq.StringArray{"Frozen", "Paralyzed", "Burned", "Cursed"}
	case "yeti":
		// Due to perk Winter Coat
		inv.Immunities = pq.StringArray{"Frozen"}

		// Neutral Roles
	case "cyborg":
		inv.Immunities = pq.StringArray{"Paralyzed", "Frozen", "Burned", "Despaired", "Blackmailed", "Drunk"}
	case "entertainer":
		// Due to perk Top-Hat Tip
		inv.Immunities = pq.StringArray{"Unlucky"}
		inv.Statuses = pq.StringArray{"Lucky"}
	case "magician":
		// Due to perk Top-Hat Tip
		inv.Statuses = pq.StringArray{"Lucky"}
		inv.Immunities = pq.StringArray{"Unlucky"}
	case "masochist":
		// Due to perk One Track Mind
		inv.Immunities = pq.StringArray{"Lucky"}
	case "succubus":
		// Due to perk Dominatrix
		inv.Immunities = pq.StringArray{"Blackmail"}
	//
	// Evil Roles
	case "arsonist":
		// Due to perk Ashes to Ashes / Flamed
		inv.Immunities = pq.StringArray{"Burned"}
	case "cultist":
		inv.Immunities = pq.StringArray{"Curse"}
	case "director":
		inv.Immunities = pq.StringArray{"Despaired", "Blackmailed", "Drunk"}
	case "gatekeeper":
		inv.Immunities = pq.StringArray{"Restrained", "Paralyzed", "Frozen"}
	case "hacker":
		inv.Immunities = pq.StringArray{"Disabled", "Blackmailed"}
	case "highwayman":
		inv.Immunities = pq.StringArray{"Madness"}
	case "imp":
		inv.Immunities = pq.StringArray{"Despaired", "Paralyzed"}
	case "threatener":
		inv.ItemLimit = 6
	default: // Do nothing
		return inv
	}

	return inv
}
