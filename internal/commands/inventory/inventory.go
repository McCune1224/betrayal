package inventory

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/mccune1224/betrayal/internal/data"
	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/zekrotja/ken"
)

const (
	defaultCoins      = 200
	defaultItemsLimit = 4
)

var optional = discordgo.ApplicationCommandOption{
	Type:        discordgo.ApplicationCommandOptionBoolean,
	Name:        "hidden",
	Description: "Make view hidden or public (default hidden)",
	Required:    false,
}

type Inventory struct {
	models data.Models
}

var (
	_ ken.SlashCommand = (*Inventory)(nil)
)

func (i *Inventory) Type() discordgo.ApplicationCommandType {
	return discordgo.ChatApplicationCommand
}

func (i *Inventory) SetModels(models data.Models) {
	i.models = models
}

// Description implements ken.SlashCommand.
func (*Inventory) Description() string {
	return "Command for managing inventory"
}

// Name implements ken.SlashCommand.
func (*Inventory) Name() string {
	return discord.DebugCmd + "inventory"
}

// Options implements ken.SlashCommand.
func (*Inventory) Options() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "get",
			Description: "get player's inventory",
			Options: []*discordgo.ApplicationCommandOption{
				discord.UserCommandArg(true),
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "create",
			Description: "create a new player",
			Options: []*discordgo.ApplicationCommandOption{
				discord.UserCommandArg(true),
				discord.StringCommandArg("role", "Role to assign to player", true),
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "delete",
			Description: "delete inventory",
			Options: []*discordgo.ApplicationCommandOption{
				discord.UserCommandArg(true),
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
			Name:        "whitelist",
			Description: "whitelist channel for inventory",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "add",
					Description: "add whitelist channel",
					Options: []*discordgo.ApplicationCommandOption{
						discord.ChannelCommandArg(true),
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "remove",
					Description: "remove whitelist channel",
					Options: []*discordgo.ApplicationCommandOption{
						discord.ChannelCommandArg(true),
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "list",
					Description: "list whitelist channels",
				},
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
			Name:        "add",
			Description: "add to player's inventory",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "ability",
					Description: "add an ability",
					Options: []*discordgo.ApplicationCommandOption{
						discord.StringCommandArg("name", "Name of the ability", true),
						discord.IntCommandArg("charges", "Number of charges", false),
						discord.UserCommandArg(false),
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "perk",
					Description: "add a perk",
					Options: []*discordgo.ApplicationCommandOption{
						discord.StringCommandArg("name", "Name of the perk", true),
						discord.UserCommandArg(false),
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "item",
					Description: "add an item",
					Options: []*discordgo.ApplicationCommandOption{
						discord.StringCommandArg("name", "Name of the item", true),
						discord.UserCommandArg(false),
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "status",
					Description: "add a status",
					Options: []*discordgo.ApplicationCommandOption{
						discord.StringCommandArg("name", "Name of the status", true),
						discord.UserCommandArg(false),
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "immunity",
					Description: "add an immunity",
					Options: []*discordgo.ApplicationCommandOption{
						discord.StringCommandArg("name", "Name of the immunity", true),
						discord.UserCommandArg(false),
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "coins",
					Description: "add coins",
					Options: []*discordgo.ApplicationCommandOption{
						discord.IntCommandArg("amount", "Amount of coins to add", true),
						discord.UserCommandArg(false),
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "note",
					Description: "add note",
					Options: []*discordgo.ApplicationCommandOption{
						discord.StringCommandArg("data", "Note to add", true),
						discord.UserCommandArg(false),
					},
				},
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
			Name:        "remove",
			Description: "remove to player's inventory",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "ability",
					Description: "remove an ability",
					Options: []*discordgo.ApplicationCommandOption{
						discord.UserCommandArg(true),
						discord.StringCommandArg("name", "Name of the ability", true),
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "perk",
					Description: "remove a perk",
					Options: []*discordgo.ApplicationCommandOption{
						discord.UserCommandArg(true),
						discord.StringCommandArg("name", "Name of the perk", true),
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "item",
					Description: "remove an item",
					Options: []*discordgo.ApplicationCommandOption{
						discord.UserCommandArg(true),
						discord.StringCommandArg("name", "Name of the item", true),
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "status",
					Description: "remove a status",
					Options: []*discordgo.ApplicationCommandOption{
						discord.UserCommandArg(true),
						discord.StringCommandArg("name", "Name of the status", true),
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "immunity",
					Description: "remove an immunity",
					Options: []*discordgo.ApplicationCommandOption{
						discord.UserCommandArg(true),
						discord.StringCommandArg("name", "Name of the immunity", true),
					},
				},
			},
		},
	}
}

// Run implements ken.SlashCommand.
func (i *Inventory) Run(ctx ken.Context) (err error) {
	return ctx.HandleSubCommands(
		ken.SubCommandHandler{Name: "get", Run: i.get},
		ken.SubCommandHandler{Name: "create", Run: i.create},
		ken.SubCommandHandler{Name: "delete", Run: i.delete},
		ken.SubCommandGroup{Name: "whitelist", SubHandler: []ken.CommandHandler{
			ken.SubCommandHandler{Name: "add", Run: i.addWhitelist},
			ken.SubCommandHandler{Name: "remove", Run: i.removeWhitelist},
			ken.SubCommandHandler{Name: "list", Run: i.listWhitelist},
		}},
		ken.SubCommandGroup{Name: "add", SubHandler: []ken.CommandHandler{
			ken.SubCommandHandler{Name: "ability", Run: i.addAnyAbility},
			ken.SubCommandHandler{Name: "perk", Run: i.addPerk},
			ken.SubCommandHandler{Name: "item", Run: i.addItem},
			ken.SubCommandHandler{Name: "status", Run: i.addStatus},
			ken.SubCommandHandler{Name: "immunity", Run: i.addImmunity},
			ken.SubCommandHandler{Name: "coins", Run: i.addCoins},
			ken.SubCommandHandler{Name: "note", Run: i.addNote},
		}},
		ken.SubCommandGroup{Name: "remove", SubHandler: []ken.CommandHandler{
			ken.SubCommandHandler{Name: "ability", Run: i.removeAbility},
			ken.SubCommandHandler{Name: "perk", Run: i.removePerk},
			ken.SubCommandHandler{Name: "item", Run: i.removeItem},
			ken.SubCommandHandler{Name: "status", Run: i.removeStatus},
			ken.SubCommandHandler{Name: "immunity", Run: i.addImmunity},
			// ken.SubCommandHandler{Name: "coins", Run: i.removeCoins},
			// ken.SubCommandHandler{Name: "note", Run: i.removeNote},
		}},
	)
}

func (i *Inventory) get(ctx ken.SubCommandContext) (err error) {
	ctx.SetEphemeral(true)

	player := ctx.Options().GetByName("user").UserValue(ctx)
	inv, err := i.models.Inventories.GetByDiscordID(player.ID)
	if err != nil {
		discord.SendSilentError(
			ctx,
			"Failed to Find Inventory",
			fmt.Sprintf("Failed to find inventory for %s", player.Username),
		)
		return err
	}

	allowed := i.authorized(ctx, inv)

	if !allowed {
		ctx.SetEphemeral(true)
		err = discord.SendSilentError(ctx, "Unauthorized",
			"You are not authorized to use this command.")
		ctx.SetEphemeral(false)
		return err
	}

	host := discord.IsAdminRole(ctx, discord.AdminRoles...)
	embd := InventoryEmbedMessage(ctx.GetEvent(), inv, host)
	err = ctx.RespondEmbed(embd)
	if err != nil {
		return err
	}
	return nil
}

func (i *Inventory) create(ctx ken.SubCommandContext) (err error) {

	if !discord.IsAdminRole(ctx, discord.AdminRoles...) {
		err = discord.SendSilentError(
			ctx,
			"Unauthorized",
			"You are not authorized to use this command.",
		)
		return err
	}

	playerArg := ctx.Options().GetByName("user").UserValue(ctx)
	roleArg := ctx.Options().GetByName("role").StringValue()
	channelID := ctx.GetEvent().ChannelID

	// Make sure role exists before creating inventory
	role, err := i.models.Roles.GetByName(roleArg)
	if err != nil {
		discord.SendSilentError(ctx, "Failed to get Role", err.Error())
		return err
	}

	// Check if inventory already exists
	existingInv, _ := i.models.Inventories.GetByDiscordID(playerArg.ID)
	if existingInv != nil {
		discord.SendSilentError(
			ctx,
			"Inventory Already Exists",
			fmt.Sprintf("Inventory already exists for %s", playerArg.Username),
		)
		return err
	}

	inventoryCreateMsg := discordgo.MessageEmbed{
		Title:       "Creating Inventory...",
		Description: fmt.Sprintf("Creating inventory for %s", playerArg.Username),
	}
	pinMsg, err := ctx.GetSession().ChannelMessageSendEmbed(channelID, &inventoryCreateMsg)
	if err != nil {
		discord.SendSilentError(ctx, "Failed to send message", err.Error())
		return err
	}
	roleAbilities, err := i.models.Roles.GetAbilities(role.ID)
	if err != nil {
		discord.SendSilentError(ctx, "Failed to get Role Abilities", err.Error())
		return err
	}
	rolePerks, err := i.models.Roles.GetPerks(role.ID)
	if err != nil {
		discord.SendSilentError(ctx, "Failed to get Role Perks", err.Error())
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

	newInv := &data.Inventory{
		DiscordID:      playerArg.ID,
		UserPinChannel: channelID,
		UserPinMessage: pinMsg.ChannelID,
		Alignment:      role.Alignment,
		RoleName:       roleArg,
		Abilities:      abilityNames,
		Perks:          perkNames,
		Coins:          defaultCoins,
		ItemsLimit:     defaultItemsLimit,
	}

	_, err = i.models.Inventories.Insert(newInv)
	if err != nil {
		log.Println(err)
		discord.SendSilentError(ctx, "Alex is a bad programmer", "Failed to insert inventory")
		return err
	}
	embd := InventoryEmbedMessage(ctx.GetEvent(), newInv, false)
	msg, err := ctx.GetSession().ChannelMessageEditEmbed(channelID, pinMsg.ID, embd)
	if err != nil {
		log.Println(err)
		discord.SendSilentError(ctx, "Alex is a bad programmer", "Failed to edit message")
		return err
	}
	newInv.UserPinChannel = msg.ChannelID
	newInv.UserPinMessage = msg.ID
	err = i.models.Inventories.Update(newInv)
	if err != nil {
		log.Println(err)
		discord.SendSilentError(ctx, "Alex is a bad programmer", "Failed to set Pinned Message")
		return err
	}
	err = ctx.GetSession().ChannelMessagePin(channelID, pinMsg.ID)
	if err != nil {
		discord.SendSilentError(ctx, "Discord is at fault for once", err.Error())
		return err
	}

	return err
}

func (i *Inventory) delete(ctx ken.SubCommandContext) (err error) {
	authed := discord.IsAdminRole(ctx, discord.AdminRoles...)
	if !authed {
		err = discord.SendSilentError(
			ctx,
			"Unauthorized",
			"You are not authorized to use this command.",
		)
		return err
	}

	userArg := ctx.Options().GetByName("user").UserValue(ctx)
	ctx.SetEphemeral(true)
	inv, err := i.models.Inventories.GetByDiscordID(userArg.ID)
	if err != nil {
		discord.SendSilentError(ctx, "Failed to Find Inventory",
			fmt.Sprintf("Failed to find inventory for %s", userArg.Username))
		return err
	}
	sesh := ctx.GetSession()
	err = sesh.ChannelMessageDelete(inv.UserPinChannel, inv.UserPinMessage)
	if err != nil {
		channel, _ := sesh.Channel(inv.UserPinChannel)
		discord.SendSilentError(ctx, "Failed to Delete Message",
			fmt.Sprintf("Failed to delete message for %s, could not find message in channel %s",
				userArg.Username, channel.Name))
	}
	err = i.models.Inventories.Delete(userArg.ID)
	if err != nil {
		discord.SendSilentError(ctx, "Failed to Delete Inventory",
			fmt.Sprintf("Failed to delete inventory for %s", userArg.Username))
	}
	ctx.RespondMessage(fmt.Sprintf("Inventory for %s deleted", userArg.Username))
	return err
}

// Version implements ken.SlashCommand.
func (*Inventory) Version() string {
	return "1.0.0"
}

// Will check if user has role "Host, "Co-Host", or "Bot Developer"
// If so append notes to inventory view and make hidden
func InventoryEmbedMessage(
	event *discordgo.InteractionCreate,
	inv *data.Inventory,
	host bool,
) *discordgo.MessageEmbed {

	roleField := &discordgo.MessageEmbedField{
		Name:   "Role",
		Value:  inv.RoleName,
		Inline: true,
	}
	alignmentEmoji := ""
	switch inv.Alignment {
	case "GOOD":
		alignmentEmoji += "👼"
	case "EVIL":
		alignmentEmoji += "👿"
	case "NEUTRAL":
		alignmentEmoji += "😐"
	}
	alignmentField := &discordgo.MessageEmbedField{
		Name:   "Alignment " + alignmentEmoji,
		Value:  inv.Alignment,
		Inline: true,
	}

	coinStr := fmt.Sprintf("%d", inv.Coins) + " [" + fmt.Sprintf("%d", inv.Coin_Bonus) + "%]"
	coinField := &discordgo.MessageEmbedField{
		Name:   "Coins 💰",
		Value:  coinStr,
		Inline: false,
	}
	abilitiesField := &discordgo.MessageEmbedField{
		Name:   "Base Abilities 💪",
		Value:  strings.Join(inv.Abilities, "\n"),
		Inline: true,
	}
	perksField := &discordgo.MessageEmbedField{
		Name:   "Perks ➕",
		Value:  strings.Join(inv.Perks, "\n"),
		Inline: true,
	}
	anyAbilitiesField := &discordgo.MessageEmbedField{
		Name:   "Any Abilities 🃏",
		Value:  strings.Join(inv.AnyAbilities, "\n"),
		Inline: true,
	}
	itemsField := &discordgo.MessageEmbedField{
		Name:   fmt.Sprintf("Items (%d/%d) ⚔️", len(inv.Items), inv.ItemsLimit),
		Value:  strings.Join(inv.Items, "\n"),
		Inline: false,
	}
	statusesField := &discordgo.MessageEmbedField{
		Name:   "Statuses 🔥",
		Value:  strings.Join(inv.Statuses, "\n"),
		Inline: true,
	}

	immunitiesField := &discordgo.MessageEmbedField{
		Name:   "Immunities 🛡️",
		Value:  strings.Join(inv.Immunities, "\n"),
		Inline: true,
	}
	effectsField := &discordgo.MessageEmbedField{
		// firework emoji: 🎆
		Name:   "Effects 🎆",
		Value:  strings.Join(inv.Effects, "\n"),
		Inline: true,
	}

	embd := &discordgo.MessageEmbed{
		Title: "Inventory 🎒",
		Fields: []*discordgo.MessageEmbedField{
			roleField,
			alignmentField,
			coinField,
			abilitiesField,
			perksField,
			anyAbilitiesField,
			itemsField,
			statusesField,
			immunitiesField,
			effectsField,
		},
	}

	if host {
		embd.Fields = append(embd.Fields, &discordgo.MessageEmbedField{
			Name:   "Notes 📝",
			Value:  strings.Join(inv.Notes, "\n"),
			Inline: false,
		})
		embd.Color = 0x00ff00
	}

	return embd
}

// In order to use the inventory channel you must meet one of the following criteria:
// 1. Call inventory command in confessional channel
// 2. Have the role "Host", "Co-Host", or "Bot Developer" AND
//   - Be in the same channel as the pinned inventory message
//   - Be within a whiteilsted channel (admin only channel...etc)
func (i *Inventory) authorized(ctx ken.SubCommandContext, inv *data.Inventory) bool {
	event := ctx.GetEvent()
	invokeChannelID := event.ChannelID
	invoker := event.Member

	// Base case of user is in confessional channel and is the owner of the inventory
	if inv.DiscordID == invoker.User.ID && inv.UserPinChannel == invokeChannelID {
		return true
	}

	// If not in confessional channel, check if in whitelist
	whitelistChannels, _ := i.models.Whitelists.GetAll()
	if invokeChannelID != inv.UserPinChannel {
		for _, whitelist := range whitelistChannels {
			if whitelist.ChannelID == invokeChannelID {
				return true
			}
		}
		return false
	}

	// --- We know from this point on that the user is in the confessional channel ---
	guildID := event.GuildID
	guildRoles, err := ctx.GetSession().GuildRoles(guildID)
	if err != nil {
		return false
	}
	// Go through and make sure user has one of the following roles:
	if inv.DiscordID != invoker.User.ID {
		for _, roleID := range invoker.Roles {
			for _, guildRole := range guildRoles {
				if roleID == guildRole.ID {
					if guildRole.Name == "Host" || guildRole.Name == "Co-Host" ||
						guildRole.Name == "Bot Developer" {
						return true
					}
				}
			}
		}
		return false
	}
	return true
}

func (i *Inventory) listWhitelist(ctx ken.SubCommandContext) (err error) {
	wishlists, err := i.models.Whitelists.GetAll()
	if len(wishlists) == 0 {
		err = discord.SendSilentError(ctx, "No whitelisted channels", "Nothing here...")
		return err
	}

	fields := []*discordgo.MessageEmbedField{}
	for _, v := range wishlists {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   v.ChannelName,
			Inline: false,
		})
	}
	err = ctx.RespondEmbed(&discordgo.MessageEmbed{
		Title:       "Whitelisted Channels",
		Description: "Whitelisted channels for inventory",
		Fields:      fields,
	})
	return err
}

// Set ephemeral to true to hide response, check if user is calling in confessional channel,
// attempt to fetch inventory, if inventory is found check if user is authorized to view inventory
// if not authorized send error message and return
func (i *Inventory) imLazyMiddleware(ctx ken.SubCommandContext) (inv *data.Inventory, err error) {
	ctx.SetEphemeral(true)
	userArg, ok := ctx.Options().GetByNameOptional("user")
	channelID := ctx.GetEvent().ChannelID
	if !ok {
		inv, err = i.models.Inventories.GetByPinChannel(channelID)
		if err != nil {
			discord.SendSilentError(
				ctx,
				"Cannot get inventory",
				"It appears you're not using this in a confessional channel, please specify a user.",
			)
			return nil, err
		}
	}
	if inv == nil {
		inv, err = i.models.Inventories.GetByDiscordID(userArg.UserValue(ctx).ID)
		if err != nil {
			log.Println(err)
			discord.SendSilentError(
				ctx,
				fmt.Sprint("Cannot find Inventory for user: ", userArg.UserValue(ctx).Username),
				fmt.Sprintf("Verify if %s has an inventory", userArg.UserValue(ctx).Username),
			)
			return nil, err
		}
	}
	if !i.authorized(ctx, inv) {
		return nil, discord.SendSilentError(
			ctx,
			"Unauthorized",
			"You are not authorized to use this command.",
		)
	}
	if inv == nil {
		return nil, errors.New("Somehow inventory is nil in middleware...")
	}
	return inv, nil
}

// Helper to handle getting the pinned message for inventory and updating it
func (i *Inventory) updateInventoryMessage(
	ctx ken.SubCommandContext,
	inventory *data.Inventory,
) (err error) {
	sesh := ctx.GetSession()
	_, err = sesh.ChannelMessageEditEmbed(
		inventory.UserPinChannel,
		inventory.UserPinMessage,
		InventoryEmbedMessage(ctx.GetEvent(), inventory, false),
	)
	if err != nil {
		log.Println(err)
		return discord.SendSilentError(
			ctx,
			"Failed to update inventory message",
			"Alex is a bad programmer, and this is his fault.",
		)
	}
	return nil
}
