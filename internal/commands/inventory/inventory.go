package inventory

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/mccune1224/betrayal/internal/scheduler"
	"github.com/mccune1224/betrayal/internal/services/inventory"
	"github.com/mccune1224/betrayal/internal/util"
	"github.com/mccune1224/betrayal/pkg/data"
	"github.com/zekrotja/ken"
)

// errors that can occur
var (
	ErrNotAuthorized = errors.New("you are not an admin role")
)

// TODO: Maybe make these configurable?
const (
	defaultCoins      = 0
	defaultItemsLimit = 4
	defaultLuck       = 0
)

var optional = discordgo.ApplicationCommandOption{
	Type:        discordgo.ApplicationCommandOptionBoolean,
	Name:        "hidden",
	Description: "hide inventory message (admin only)",
	Required:    false,
}

type Inventory struct {
	models    data.Models
	scheduler *scheduler.BetrayalScheduler
}

// Components implements main.BetrayalCommand.
func (*Inventory) Components() []*discordgo.MessageComponent {
	return nil
}

var _ ken.SlashCommand = (*Inventory)(nil)

func (i *Inventory) Type() discordgo.ApplicationCommandType {
	return discordgo.ChatApplicationCommand
}

func (i *Inventory) Initialize(m data.Models, s *scheduler.BetrayalScheduler) {
	i.models = m
	i.scheduler = s
}

// Description implements ken.SlashCommand.
func (*Inventory) Description() string {
	return "Command for managing inventory"
}

// Name implements ken.SlashCommand.
func (*Inventory) Name() string {
	return discord.DebugCmd + "inv"
}

func (i *Inventory) get(ctx ken.SubCommandContext) (err error) {
	if err := ctx.Defer(); err != nil {
		log.Println(err)
		return err
	}

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
	// embd := InventoryEmbedBuilder(inv, host)
	embd := &discordgo.MessageEmbed{}

	// Edge case where if user is in their own confessional, make it public (helpful for admins)
	e := ctx.GetEvent()
	if e.ChannelID == inv.UserPinChannel {
		if host {
			showArg, ok := ctx.Options().GetByNameOptional("show")
			hide := true
			if ok {
				hide = !showArg.BoolValue()
			}

			if hide {
				ctx.SetEphemeral(true)
				embd = InventoryEmbedBuilder(inv, true)
				return ctx.RespondEmbed(embd)
			} else {
				ctx.SetEphemeral(false)
				embd = InventoryEmbedBuilder(inv, false)
				return ctx.RespondEmbed(embd)
			}
		}

		ctx.SetEphemeral(false)
		embd = InventoryEmbedBuilder(inv, false)
	} else {
		embd = InventoryEmbedBuilder(inv, host)
	}
	return ctx.RespondEmbed(embd)
}

func (i *Inventory) me(ctx ken.SubCommandContext) (err error) {
	if err := ctx.Defer(); err != nil {
		log.Println(err)
		return err
	}

	ctx.SetEphemeral(true)
	player := ctx.GetEvent().Member.User
	inv, err := i.models.Inventories.GetByDiscordID(player.ID)
	if err != nil {
		discord.ErrorMessage(ctx, "Failed to find your inventory.", "Are you sure you're an active player or in your confessional?")
		return err
	}
	allowed := i.inventoryAuthorized(ctx, inv)

	if !allowed {
		ctx.SetEphemeral(true)
		err = discord.ErrorMessage(ctx, "Unauthorized",
			"Please only use this command in your confessional.")
		return err
	}

	host := discord.IsAdminRole(ctx, discord.AdminRoles...)
	// embd := InventoryEmbedBuilder(inv, host)
	embd := &discordgo.MessageEmbed{}

	e := ctx.GetEvent()
	if e.ChannelID == inv.UserPinChannel {
		ctx.SetEphemeral(false)
		embd = InventoryEmbedBuilder(inv, false)
	} else {
		embd = InventoryEmbedBuilder(inv, host)
	}
	return ctx.RespondEmbed(embd)
}

func (i *Inventory) delete(ctx ken.SubCommandContext) (err error) {
	if err := ctx.Defer(); err != nil {
		log.Println(err)
		return err
	}
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
	inv, err := i.models.Inventories.GetByDiscordID(userArg.ID)
	if err != nil {
		log.Println(err)
		return discord.ErrorMessage(ctx, "Failed to Find Inventory",
			fmt.Sprintf("Failed to find inventory for %s", userArg.Username))
	}
	sesh := ctx.GetSession()
	err = sesh.ChannelMessageDelete(inv.UserPinChannel, inv.UserPinMessage)
	if err != nil {
		channel, _ := sesh.Channel(inv.UserPinChannel)
		return discord.ErrorMessage(ctx, "Failed to Delete Message",
			fmt.Sprintf("Failed to delete message for %s, could not find message in channel %s",
				userArg.Username, channel.Name))
	}
	err = i.models.Inventories.Delete(userArg.ID)
	log.Println(err)
	if err != nil {
		return discord.ErrorMessage(ctx, "Failed to Delete Inventory",
			fmt.Sprintf("Failed to delete inventory for %s", userArg.Username))
	}
	return discord.SuccessfulMessage(ctx, "Inventory Deleted", fmt.Sprintf("Removed inventory for channel %s", discord.MentionChannel(inv.UserPinChannel)))
}

// Version implements ken.SlashCommand.
func (*Inventory) Version() string {
	return "1.0.0"
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
	if err := ctx.Defer(); err != nil {
		log.Println(err)
		return err
	}
	wishlists, _ := i.models.Whitelists.GetAll()
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

// Helper to handle getting the pinned message for inventory and updating it
func (i *Inventory) updateInventoryMessage(
	ctx ken.SubCommandContext,
	inventory *data.Inventory,
) (err error) {
	sesh := ctx.GetSession()
	_, err = sesh.ChannelMessageEditEmbed(
		inventory.UserPinChannel,
		inventory.UserPinMessage,
		InventoryEmbedBuilder(inventory, false),
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

func UpdateInventoryMessage(sesh *discordgo.Session, i *data.Inventory) (err error) {
	_, err = sesh.ChannelMessageEditEmbed(
		i.UserPinChannel,
		i.UserPinMessage,
		InventoryEmbedBuilder(i, false),
	)
	if err != nil {
		return err
	}
	return nil
}

// Helper to determine if user is authorized to use inventory command based on:
// 1. In their confessional (and owner of inventory)
// 2. In a whitelisted channel (and an admin)
func InventoryAuthorized(
	ctx ken.SubCommandContext,
	inv *data.Inventory,
	wl []*data.Whitelist,
) bool {
	event := ctx.GetEvent()
	invokeChannelID := event.ChannelID
	invoker := event.Member

	// Base case of user is in confessional channel and is the owner of the inventory
	if inv.DiscordID == invoker.User.ID && inv.UserPinChannel == invokeChannelID {
		return true
	}

	// If not in confessional channel, check if in whitelist
	if invokeChannelID != inv.UserPinChannel {
		for _, whitelist := range wl {
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

// Helper to attempt to fetch given user's inventory from user command option
func Fetch(ctx ken.SubCommandContext, m data.Models, adminOnly bool) (inv *data.Inventory, err error) {
	if adminOnly && !discord.IsAdminRole(ctx, discord.AdminRoles...) {
		return nil, ErrNotAuthorized
	}
	userArg, ok := ctx.Options().GetByNameOptional("user")
	event := ctx.GetEvent()
	channelID := event.ChannelID
	if !ok {
		inv, err = m.Inventories.GetByPinChannel(channelID)
		if err != nil {
			log.Println(err)
			return nil, err
		}
	}
	if inv == nil {
		inv, err = m.Inventories.GetByDiscordID(userArg.UserValue(ctx).ID)
		if err != nil {
			log.Println(err)
			return nil, err
		}
	}
	wl, err := m.Whitelists.GetAll()
	if err != nil {
		log.Println(err)
		return nil, err
	}
	if !InventoryAuthorized(ctx, inv, wl) {
		return nil, ErrNotAuthorized
	}
	if inv == nil {
		return nil, errors.New("somehow inventory is nil in middleware")
	}
	return inv, nil
}

// Middleware for inventory commands to fetch inventory and ensure user is authorized
func FetchHandler(ctx ken.SubCommandContext, m data.Models, adminOnly bool) (handler *inventory.InventoryHandler, err error) {
	inv := &data.Inventory{}
	if adminOnly && !discord.IsAdminRole(ctx, discord.AdminRoles...) {
		return nil, ErrNotAuthorized
	}
	userArg, ok := ctx.Options().GetByNameOptional("user")
	event := ctx.GetEvent()
	channelID := event.ChannelID
	if !ok {
		inv, err = m.Inventories.GetByPinChannel(channelID)
		if err != nil {
			log.Println(err)
			return nil, err
		}
	}
	if inv == nil {
		inv, err = m.Inventories.GetByDiscordID(userArg.UserValue(ctx).ID)
		if err != nil {
			log.Println(err)
			return nil, err
		}
	}
	wl, err := m.Whitelists.GetAll()
	if err != nil {
		log.Println(err)
		return nil, err
	}
	if !InventoryAuthorized(ctx, inv, wl) {
		return nil, ErrNotAuthorized
	}
	if inv == nil {
		return nil, errors.New("somehow inventory is nil in middleware")
	}
	handler = inventory.InitInventoryHandler(m, inv)
	return handler, nil
}

// Helper to build inventory embed message based off if user is host or not
func InventoryEmbedBuilder(
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
		Name:   fmt.Sprintf("%s Alignment", alignmentEmoji),
		Value:  inv.Alignment,
		Inline: true,
	}

	coinStr := fmt.Sprintf("%d", inv.Coins) + " [" + fmt.Sprintf("%.2f", inv.CoinBonus) + "%]"
	coinField := &discordgo.MessageEmbedField{
		Name:   fmt.Sprintf("%s Coins", discord.EmojiCoins),
		Value:  coinStr,
		Inline: true,
	}
	abilitiesField := &discordgo.MessageEmbedField{
		Name:   fmt.Sprintf("%s Abilities", discord.EmojiAbility),
		Value:  strings.Join(inv.Abilities, "\n"),
		Inline: true,
	}
	perksField := &discordgo.MessageEmbedField{
		Name:   fmt.Sprintf("%s Perks", discord.EmojiPerk),
		Value:  strings.Join(inv.Perks, "\n"),
		Inline: true,
	}
	anyAbilitiesField := &discordgo.MessageEmbedField{
		Name:   fmt.Sprintf("%s Any Abilities", discord.EmojiAnyAbility),
		Value:  strings.Join(inv.AnyAbilities, "\n"),
		Inline: true,
	}
	itemsField := &discordgo.MessageEmbedField{
		Name:   fmt.Sprintf("%s Items (%d/%d)", discord.EmojiItem, len(inv.Items), inv.ItemLimit),
		Value:  strings.Join(inv.Items, "\n"),
		Inline: true,
	}
	statusesField := &discordgo.MessageEmbedField{
		Name:   fmt.Sprintf("%s Statuses", discord.EmojiStatus),
		Value:  strings.Join(inv.Statuses, "\n"),
		Inline: true,
	}

	immunitiesField := &discordgo.MessageEmbedField{
		Name:   fmt.Sprintf("%s Immunities", discord.EmojiImmunity),
		Value:  strings.Join(inv.Immunities, "\n"),
		Inline: true,
	}
	effectsField := &discordgo.MessageEmbedField{
		Name:   fmt.Sprintf("%s Effects", discord.EmojiEffect),
		Value:  strings.Join(inv.Effects, "\n"),
		Inline: true,
	}
	isAlive := ""
	if inv.IsAlive {
		isAlive = fmt.Sprintf("%s Alive", discord.EmojiAlive)
	} else {
		isAlive = fmt.Sprintf("%s Dead", discord.EmojiDead)
	}

	deadField := &discordgo.MessageEmbedField{
		Name:   isAlive,
		Inline: true,
	}

	embd := &discordgo.MessageEmbed{
		Title: fmt.Sprintf("Inventory %s", discord.EmojiInventory),
		Fields: []*discordgo.MessageEmbedField{
			roleField,
			alignmentField,
			coinField,
			abilitiesField,
			anyAbilitiesField,
			perksField,
			itemsField,
			statusesField,
			immunitiesField,
			effectsField,
			deadField,
		},
		Color: discord.ColorThemeDiamond,
	}

	humanReqTime := util.GetEstTimeStamp()
	embd.Footer = &discordgo.MessageEmbedFooter{
		Text: fmt.Sprintf("Last updated: %s", humanReqTime),
	}

	if host {

		embd.Fields = append(embd.Fields, &discordgo.MessageEmbedField{
			Name:   fmt.Sprintf("%s Luck", discord.EmojiLuck),
			Value:  fmt.Sprintf("%d", inv.Luck),
			Inline: true,
		})

		noteListString := ""
		for i, note := range inv.Notes {
			noteListString += fmt.Sprintf("%d. %s\n", i+1, note)
		}

		embd.Fields = append(embd.Fields, &discordgo.MessageEmbedField{
			Name:   fmt.Sprintf("%s Notes", discord.EmojiNote),
			Value:  noteListString,
			Inline: false,
		})

		embd.Color = discord.ColorThemeAmethyst

	}

	return embd
}

// Ability strings follow the format of 'Name [#]'
func ParseAbilityString(raw string) (name string, charges int, err error) {
	// Check if there's a charge amount
	charges = 1
	split := strings.Split(raw, " ")
	if len(split) > 1 {
		charges, err = strconv.Atoi(split[len(split)-1])
		if err != nil {
			return "", 0, err
		}
	}
	name = strings.Join(split[:len(split)-1], " ")
	return name, charges, nil
}

// Will attempt to upate the given any ability in the inventory and if not present will add it
func UpsertAA(inv *data.Inventory, aa *data.AnyAbility, charges ...int) {
	defaultCharge := 1
	if len(charges) > 0 {
		defaultCharge = charges[0]
	}

	for i, a := range inv.AnyAbilities {
		invName, invCharge, _ := ParseAbilityString(a)
		if strings.EqualFold(invName, aa.Name) {
			inv.AnyAbilities[i] = fmt.Sprintf("%s [%d]", invName, invCharge+defaultCharge)
			return
		}
	}
	inv.AnyAbilities = append(inv.AnyAbilities, fmt.Sprintf("%s [%d]", aa.Name, defaultCharge))
}

// Will attempt to upate the given ability in the inventory and if not present will add it
func UpsertAbility(inv *data.Inventory, aa *data.Ability, charges ...int) {
	defaultCharges := 1
	if len(charges) > 0 {
		defaultCharges = charges[0]
	}
	for i, a := range inv.Abilities {
		invName, invCharge, _ := ParseAbilityString(a)
		if strings.EqualFold(invName, aa.Name) {
			inv.Abilities[i] = fmt.Sprintf("%s [%d]", invName, invCharge+defaultCharges)
			return
		}
	}
	inv.Abilities = append(inv.Abilities, fmt.Sprintf("%s [%d]", aa.Name, defaultCharges))
}
