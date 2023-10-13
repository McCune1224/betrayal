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
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// emoji constants
// TODO: Maybe make these configurable?

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

// Components implements main.BetrayalCommand.
func (*Inventory) Components() []*discordgo.MessageComponent {
	return nil
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
	return discord.DebugCmd + "inv"
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
					Description: "add a base ability",
					Options: []*discordgo.ApplicationCommandOption{
						discord.StringCommandArg("name", "Name of the ability", true),
						discord.IntCommandArg("charges", "Number of charges", false),
						discord.UserCommandArg(false),
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "aa",
					Description: "add an any ability",
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
					Name:        "effect",
					Description: "add an effect",
					Options: []*discordgo.ApplicationCommandOption{
						discord.StringCommandArg("name", "Name of the effect", true),
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
					Name:        "bonus",
					Description: "add coin bonus",
					Options: []*discordgo.ApplicationCommandOption{
						// Discord is fucking stupid and doesn't allow decimals
						discord.StringCommandArg("amount", "Amount of coin bonus to add", true),
						discord.UserCommandArg(false),
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "note",
					Description: "add a note",
					Options: []*discordgo.ApplicationCommandOption{
						discord.StringCommandArg("message", "Note to add", true),
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
					Description: "remove a base ability",
					Options: []*discordgo.ApplicationCommandOption{
						discord.StringCommandArg("name", "Name of the ability", true),
						discord.UserCommandArg(false),
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "aa",
					Description: "remove an any ability",
					Options: []*discordgo.ApplicationCommandOption{
						discord.StringCommandArg("name", "Name of the ability", true),
						discord.UserCommandArg(false),
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "perk",
					Description: "remove a perk",
					Options: []*discordgo.ApplicationCommandOption{
						discord.StringCommandArg("name", "Name of the perk", true),
						discord.UserCommandArg(false),
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "item",
					Description: "remove an item",
					Options: []*discordgo.ApplicationCommandOption{
						discord.StringCommandArg("name", "Name of the item", true),
						discord.UserCommandArg(false),
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "status",
					Description: "remove a status",
					Options: []*discordgo.ApplicationCommandOption{
						discord.StringCommandArg("name", "Name of the status", true),
						discord.UserCommandArg(false),
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "immunity",
					Description: "remove an immunity",
					Options: []*discordgo.ApplicationCommandOption{
						discord.StringCommandArg("name", "Name of the immunity", true),
						discord.UserCommandArg(false),
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "effect",
					Description: "remove an effect",
					Options: []*discordgo.ApplicationCommandOption{
						discord.StringCommandArg("name", "Name of the effect", true),
						discord.UserCommandArg(false),
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "coins",
					Description: "remove coins",
					Options: []*discordgo.ApplicationCommandOption{
						discord.IntCommandArg("amount", "Amount of coins to remove", true),
						discord.UserCommandArg(false),
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "bonus",
					Description: "remove coin bonus",
					Options: []*discordgo.ApplicationCommandOption{
						// Discord is fucking stupid and doesn't take decimals...need to use string arg
						discord.StringCommandArg("amount", "Amount of coin bonus to remove", true),
						discord.UserCommandArg(false),
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "note",
					Description: "remove a note by index number",
					Options: []*discordgo.ApplicationCommandOption{
						discord.IntCommandArg("index", "Index # to remove", true),
						discord.UserCommandArg(false),
					},
				},
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
			Name:        "set",
			Description: "set to player's inventory",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "ability",
					Description: "set a base ability",
					Options: []*discordgo.ApplicationCommandOption{
						discord.StringCommandArg("name", "Name of the ability", true),
						discord.IntCommandArg("charges", "Number of charges", true),
						discord.UserCommandArg(false),
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "aa",
					Description: "set an any ability",
					Options: []*discordgo.ApplicationCommandOption{
						discord.StringCommandArg("name", "Name of the ability", true),
						discord.IntCommandArg("charges", "Number of charges", true),
						discord.UserCommandArg(false),
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "coins",
					Description: "set coins",
					Options: []*discordgo.ApplicationCommandOption{
						discord.IntCommandArg("amount", "Amount of coins to set", true),
						discord.UserCommandArg(false),
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "bonus",
					Description: "set coin bonus",
					Options: []*discordgo.ApplicationCommandOption{
						// Discord is fucking stupid and doesn't take decimals...need to use string arg
						discord.StringCommandArg("amount", "Amount of coin bonus to set", true),
						discord.UserCommandArg(false),
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "limit",
					Description: "Set item limit",
					Options: []*discordgo.ApplicationCommandOption{
						discord.IntCommandArg("size", "New size to set", true),
						discord.UserCommandArg(false),
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
			ken.SubCommandHandler{Name: "ability", Run: i.addAbility},
			ken.SubCommandHandler{Name: "aa", Run: i.addAnyAbility},
			ken.SubCommandHandler{Name: "perk", Run: i.addPerk},
			ken.SubCommandHandler{Name: "item", Run: i.addItem},
			ken.SubCommandHandler{Name: "status", Run: i.addStatus},
			ken.SubCommandHandler{Name: "immunity", Run: i.addImmunity},
			ken.SubCommandHandler{Name: "effect", Run: i.addEffect},
			ken.SubCommandHandler{Name: "coins", Run: i.addCoins},
			ken.SubCommandHandler{Name: "bonus", Run: i.addCoinBonus},
			ken.SubCommandHandler{Name: "note", Run: i.addNote},
		}},
		ken.SubCommandGroup{Name: "remove", SubHandler: []ken.CommandHandler{
			ken.SubCommandHandler{Name: "ability", Run: i.removeAbility},
			ken.SubCommandHandler{Name: "aa", Run: i.removeAnyAbility},
			ken.SubCommandHandler{Name: "perk", Run: i.removePerk},
			ken.SubCommandHandler{Name: "item", Run: i.removeItem},
			ken.SubCommandHandler{Name: "status", Run: i.removeStatus},
			ken.SubCommandHandler{Name: "immunity", Run: i.removeImmunity},
			ken.SubCommandHandler{Name: "effect", Run: i.removeEffect},
			ken.SubCommandHandler{Name: "coins", Run: i.removeCoins},
			ken.SubCommandHandler{Name: "bonus", Run: i.removeCoinBonus},
			ken.SubCommandHandler{Name: "note", Run: i.removeNote},
		}},
		ken.SubCommandGroup{Name: "set", SubHandler: []ken.CommandHandler{
			ken.SubCommandHandler{Name: "ability", Run: i.setAbility},
			ken.SubCommandHandler{Name: "aa", Run: i.setAnyAbility},
			ken.SubCommandHandler{Name: "coins", Run: i.setCoins},
			ken.SubCommandHandler{Name: "bonus", Run: i.setCoinBonus},
			ken.SubCommandHandler{Name: "limit", Run: i.setItemsLimit},
		}},
	)
}

func (i *Inventory) get(ctx ken.SubCommandContext) (err error) {
	ctx.SetEphemeral(true)

	player := ctx.Options().GetByName("user").UserValue(ctx)
	inv, err := i.models.Inventories.GetByDiscordID(player.ID)
	if err != nil {
		discord.ErrorMessage(
			ctx,
			"Failed to Find Inventory",
			fmt.Sprintf("Are you sure there's an inventory for %s?", player.Username),
		)
		return err
	}

	allowed := i.inventoryAuthorized(ctx, inv)

	if !allowed {
		ctx.SetEphemeral(true)
		err = discord.ErrorMessage(ctx, "Unauthorized",
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
	role, err := i.models.Roles.GetByName(roleArg)
	if err != nil {
		discord.ErrorMessage(ctx, "Failed to get Role", err.Error())
		return err
	}

	// Check if inventory already exists
	existingInv, _ := i.models.Inventories.GetByDiscordID(playerArg.ID)
	if existingInv != nil {
		discord.ErrorMessage(
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
		discord.ErrorMessage(ctx, "Failed to send message", err.Error())
		return err
	}
	roleAbilities, err := i.models.Roles.GetAbilities(role.ID)
	if err != nil {
		discord.ErrorMessage(ctx, "Failed to get Role Abilities", err.Error())
		return err
	}
	rolePerks, err := i.models.Roles.GetPerks(role.ID)
	if err != nil {
		discord.ErrorMessage(ctx, "Failed to get Role Perks", err.Error())
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
		ItemLimit:      defaultItemsLimit,
	}

	_, err = i.models.Inventories.Insert(newInv)
	if err != nil {
		log.Println(err)
		discord.ErrorMessage(ctx, "Alex is a bad programmer", "Failed to insert inventory")
		return err
	}
	embd := InventoryEmbedMessage(ctx.GetEvent(), newInv, false)
	msg, err := ctx.GetSession().ChannelMessageEditEmbed(channelID, pinMsg.ID, embd)
	if err != nil {
		log.Println(err)
		discord.ErrorMessage(ctx, "Alex is a bad programmer", "Failed to edit message")
		return err
	}
	newInv.UserPinChannel = msg.ChannelID
	newInv.UserPinMessage = msg.ID
	err = i.models.Inventories.Update(newInv)
	if err != nil {
		log.Println(err)
		discord.ErrorMessage(ctx, "Alex is a bad programmer", "Failed to set Pinned Message")
		return err
	}
	err = ctx.GetSession().ChannelMessagePin(channelID, pinMsg.ID)
	if err != nil {
		discord.ErrorMessage(ctx, "Discord is at fault for once", err.Error())
		return err
	}

	return err
}

func (i *Inventory) delete(ctx ken.SubCommandContext) (err error) {
	authed := discord.IsAdminRole(ctx, discord.AdminRoles...)
	if !authed {
		err = discord.ErrorMessage(
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
		discord.ErrorMessage(ctx, "Failed to Find Inventory",
			fmt.Sprintf("Failed to find inventory for %s", userArg.Username))
		return err
	}
	sesh := ctx.GetSession()
	err = sesh.ChannelMessageDelete(inv.UserPinChannel, inv.UserPinMessage)
	if err != nil {
		channel, _ := sesh.Channel(inv.UserPinChannel)
		discord.ErrorMessage(ctx, "Failed to Delete Message",
			fmt.Sprintf("Failed to delete message for %s, could not find message in channel %s",
				userArg.Username, channel.Name))
	}
	err = i.models.Inventories.Delete(userArg.ID)
	if err != nil {
		discord.ErrorMessage(ctx, "Failed to Delete Inventory",
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
	alignmentEmoji := discord.EmojiAlignment
	alignmentField := &discordgo.MessageEmbedField{
		Name:   fmt.Sprintf("Alignment %s", alignmentEmoji),
		Value:  inv.Alignment,
		Inline: true,
	}

	//show coin bonus x100
	cb := inv.CoinBonus * 100
	coinStr := fmt.Sprintf("%d", inv.Coins) + " [" + fmt.Sprintf("%.2f", cb) + "%]"
	coinField := &discordgo.MessageEmbedField{
		Name:   fmt.Sprintf("Coins %s", discord.EmojiCoins),
		Value:  coinStr,
		Inline: false,
	}
	abilitiesField := &discordgo.MessageEmbedField{
		Name:   fmt.Sprintf("Abilities %s", discord.EmojiAbility),
		Value:  strings.Join(inv.Abilities, "\n"),
		Inline: true,
	}
	perksField := &discordgo.MessageEmbedField{
		Name:   fmt.Sprintf("Perks %s", discord.EmojiPerk),
		Value:  strings.Join(inv.Perks, "\n"),
		Inline: true,
	}
	anyAbilitiesField := &discordgo.MessageEmbedField{
		Name:   fmt.Sprintf("Any Abilities %s", discord.EmojiAnyAbility),
		Value:  strings.Join(inv.AnyAbilities, "\n"),
		Inline: true,
	}
	itemsField := &discordgo.MessageEmbedField{
		Name:   fmt.Sprintf("Items (%d/%d) %s", len(inv.Items), inv.ItemLimit, discord.EmojiItem),
		Value:  strings.Join(inv.Items, "\n"),
		Inline: false,
	}
	statusesField := &discordgo.MessageEmbedField{
		Name:   fmt.Sprintf("Statuses %s", discord.EmojiStatus),
		Value:  strings.Join(inv.Statuses, "\n"),
		Inline: true,
	}

	immunitiesField := &discordgo.MessageEmbedField{
		Name:   fmt.Sprintf("Immunities %s", discord.EmojiImmunity),
		Value:  strings.Join(inv.Immunities, "\n"),
		Inline: true,
	}
	effectsField := &discordgo.MessageEmbedField{
		Name:   fmt.Sprintf("Effects %s", discord.EmojiEffect),
		Value:  strings.Join(inv.Effects, "\n"),
		Inline: true,
	}

	embd := &discordgo.MessageEmbed{
		Title: fmt.Sprintf("Inventory %s", discord.EmojiInventory),
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

		noteListString := ""
		for i, note := range inv.Notes {
			noteListString += fmt.Sprintf("%d. %s\n", i+1, note)
		}

		embd.Fields = append(embd.Fields, &discordgo.MessageEmbedField{
			Name:   fmt.Sprintf("Notes %s", discord.EmojiNote),
			Value:  noteListString,
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
func (i *Inventory) inventoryAuthorized(ctx ken.SubCommandContext, inv *data.Inventory) bool {
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

	// Go through and make sure user has one of the allowed roles:
	for _, role := range invoker.Roles {
		for _, allowedRole := range discord.AdminRoles {
			if role == allowedRole {
				return true
			}
		}
	}
	return true
}

func (i *Inventory) listWhitelist(ctx ken.SubCommandContext) (err error) {
	wishlists, err := i.models.Whitelists.GetAll()
	if len(wishlists) == 0 {
		err = discord.ErrorMessage(ctx, "No whitelisted channels", "Nothing here...")
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
			discord.ErrorMessage(
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
			discord.ErrorMessage(
				ctx,
				fmt.Sprint("Cannot find Inventory for user: ", userArg.UserValue(ctx).Username),
				fmt.Sprintf("Verify if %s has an inventory", userArg.UserValue(ctx).Username),
			)
			return nil, err
		}
	}
	if !i.inventoryAuthorized(ctx, inv) {
		discord.ErrorMessage(
			ctx,
			"Unauthorized",
			"You are not authorized to use this command.",
		)
		return nil, errors.New("Unauthorized to use command")
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
		return discord.ErrorMessage(
			ctx,
			"Failed to update inventory message",
			"Alex is a bad programmer, and this is his fault.",
		)
	}
	return nil
}

func UpdateInventoryMessage(ctx ken.Context, i *data.Inventory) (err error) {
	sesh := ctx.GetSession()
	_, err = sesh.ChannelMessageEditEmbed(
		i.UserPinChannel,
		i.UserPinMessage,
		InventoryEmbedMessage(ctx.GetEvent(), i, false),
	)
	if err != nil {
		log.Println(err)
		return discord.ErrorMessage(
			ctx,
			"Failed to update inventory message",
			"Alex is a bad programmer, and this is his fault.",
		)
	}
	return nil
}
