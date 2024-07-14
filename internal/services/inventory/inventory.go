package inventory

import (
	"context"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/mccune1224/betrayal/internal/models"
	"github.com/mccune1224/betrayal/internal/util"
	"github.com/zekrotja/ken"
)

type PlayerInventory struct {
	models.Player
	Role       models.Role                            `json:"role"`
	Items      []models.ListPlayerItemInventoryRow    `json:"items"`
	Abilities  []models.ListPlayerAbilityInventoryRow `json:"abilities"`
	Perks      []models.PerkInfo                      `json:"perks"`
	Immunities []models.Status                        `json:"immunities"`
	Statuses   []models.ListPlayerStatusInventoryRow  `json:"statuses"`
}

type InventoryHandler struct {
	pool   *pgxpool.Pool
	player models.Player
}

// In order for this to work 1 of 2 things must happen:
// 1. This command is called within the player's confessional by an admin
// 2. This command is called within a whitelisted channel and explictly asks for the player's inventory
func NewInventoryHandler(ctx ken.SubCommandContext, db *pgxpool.Pool) (*InventoryHandler, error) {
	handler := &InventoryHandler{pool: db}
	query := models.New(db)
	playerID := int64(0)
	if playerArg, ok := ctx.Options().GetByNameOptional("user"); ok {
		playerID, _ = util.Atoi64((playerArg.UserValue(ctx).ID))
	}

	if playerID == 0 {
		channelID, _ := util.Atoi64(ctx.GetEvent().ChannelID)
		playerConfessional, err := query.GetPlayerConfessionalByChannelID(context.Background(), channelID)
		if err != nil {
			return nil, err
		}

		playerID = playerConfessional.PlayerID
	}

	player, err := query.GetPlayer(context.Background(), playerID)
	if err != nil {
		return nil, err
	}
	handler.player = player
	return handler, nil
}

// WARNING: This is a one off hack. Need to manually create this instead of using the NewInventoryHandler
// as this breaks the two checks for inventory authorization but is still *technically* correct
func Jank(player models.Player, pool *pgxpool.Pool) *InventoryHandler {
	return &InventoryHandler{pool: pool, player: player}
}

func (ih *InventoryHandler) FetchInventory() (*PlayerInventory, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	query := models.New(ih.pool)

	ih.SyncPlayer()
	abilityChan := make(chan []models.ListPlayerAbilityInventoryRow, 1)
	perksChan := make(chan []models.PerkInfo, 1)
	itemCh := make(chan []models.ListPlayerItemInventoryRow, 1)
	statusChan := make(chan []models.ListPlayerStatusInventoryRow, 1)
	immunityChan := make(chan []models.Status, 1)
	roleChan := make(chan models.Role, 1)

	go util.DbTask(ctx, roleChan, func() (models.Role, error) {
		return query.GetRole(ctx, ih.player.RoleID.Int32)
	})

	go util.DbTask(ctx, perksChan, func() ([]models.PerkInfo, error) {
		return query.ListPlayerPerk(ctx, ih.player.ID)
	})

	go util.DbTask(ctx, abilityChan, func() ([]models.ListPlayerAbilityInventoryRow, error) {
		return query.ListPlayerAbilityInventory(ctx, ih.player.ID)
	})

	go util.DbTask(ctx, itemCh, func() ([]models.ListPlayerItemInventoryRow, error) {
		return query.ListPlayerItemInventory(ctx, ih.player.ID)
	})

	go util.DbTask(ctx, statusChan, func() ([]models.ListPlayerStatusInventoryRow, error) {
		return query.ListPlayerStatusInventory(ctx, ih.player.ID)
	})

	go util.DbTask(ctx, immunityChan, func() ([]models.Status, error) {
		return query.ListPlayerImmunity(ctx, ih.player.ID)
	})

	inv := &PlayerInventory{Player: ih.player}
	inv.Role = <-roleChan
	inv.Abilities = <-abilityChan
	inv.Items = <-itemCh
	inv.Immunities = <-immunityChan
	inv.Statuses = <-statusChan
	inv.Perks = <-perksChan
	return inv, nil
}

func (ih *InventoryHandler) InventoryEmbedBuilder(
	inv *PlayerInventory,
	host bool,
) *discordgo.MessageEmbed {
	roleField := &discordgo.MessageEmbedField{
		Name:   "Role",
		Value:  inv.Role.Name,
		Inline: true,
	}
	alignmentEmoji := discord.EmojiAlignment
	alignmentField := &discordgo.MessageEmbedField{
		Name:   fmt.Sprintf("%s Alignment", alignmentEmoji),
		Value:  string(inv.Player.Alignment),
		Inline: true,
	}

	coinBonusStr, _ := util.NumericToString(inv.CoinBonus)
	coinStr := fmt.Sprintf("%d [%s%% bonus]", inv.Coins, coinBonusStr[:len(coinBonusStr)-1])
	coinField := &discordgo.MessageEmbedField{
		Name:   fmt.Sprintf("%s Coins", discord.EmojiCoins),
		Value:  coinStr,
		Inline: true,
	}
	abSts := []string{}
	for _, ab := range inv.Abilities {
		str := fmt.Sprintf("[%d] %s", ab.Quantity, ab.Name)
		if ab.AnyAbility {
			abSts = append(abSts, str+" (AA)")
		} else {
			abSts = append(abSts, str)
		}

	}
	abilitiesField := &discordgo.MessageEmbedField{
		Name:   fmt.Sprintf("%s Abilities", discord.EmojiAbility),
		Value:  strings.Join(abSts, "\n"),
		Inline: true,
	}

	perksSts := []string{}
	for _, perk := range inv.Perks {
		perksSts = append(perksSts, perk.Name)
	}
	perksField := &discordgo.MessageEmbedField{
		Name:   fmt.Sprintf("%s Perks", discord.EmojiPerk),
		Value:  strings.Join(perksSts, "\n"),
		Inline: true,
	}

	itemsSts := []string{}
	itemQuantity := int32(0)
	for _, item := range inv.Items {
		itemQuantity += item.Quantity
		itemsSts = append(itemsSts, fmt.Sprintf("%s [%d]", item.Name, item.Quantity))
	}

	itemsField := &discordgo.MessageEmbedField{
		Name:   fmt.Sprintf("%s Items (%d/%d)", discord.EmojiItem, itemQuantity, inv.ItemLimit),
		Value:  strings.Join(itemsSts, "\n"),
		Inline: true,
	}

	statusStrs := []string{}
	for _, status := range inv.Statuses {
		statusStrs = append(statusStrs, fmt.Sprintf("%s [%d]", status.Name, status.Quantity))
	}
	statusesField := &discordgo.MessageEmbedField{
		Name:   fmt.Sprintf("%s Statuses", discord.EmojiStatus),
		Value:  strings.Join(statusStrs, "\n"),
		Inline: true,
	}

	immusSts := []string{}
	for _, immu := range inv.Immunities {
		immusSts = append(immusSts, immu.Name)
	}
	immunitiesField := &discordgo.MessageEmbedField{
		Name:   fmt.Sprintf("%s Immunities", discord.EmojiImmunity),
		Value:  strings.Join(immusSts, "\n"),
		Inline: true,
	}

	isAlive := ""
	if inv.Player.Alive {
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
			// anyAbilitiesField,
			perksField,
			itemsField,
			statusesField,
			immunitiesField,
			// effectsField,
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

		// noteListString := ""
		// for i, note := range inv.Notes {
		// 	noteListString += fmt.Sprintf("%d. %s\n", i+1, note)
		// }

		// embd.Fields = append(embd.Fields, &discordgo.MessageEmbedField{
		// 	Name:   fmt.Sprintf("%s Notes", discord.EmojiNote),
		// 	Value:  noteListString,
		// 	Inline: false,
		// })

		embd.Color = discord.ColorThemeAmethyst

	}

	return embd
}

func (ih *InventoryHandler) UpdateInventoryMessage(sesh *discordgo.Session) (err error) {
	query := models.New(ih.pool)
	playerConf, err := query.GetPlayerConfessional(context.Background(), ih.player.ID)
	if err != nil {
		return err
	}

	inv, err := ih.FetchInventory()
	if err != nil {
		return err
	}
	_, err = sesh.ChannelMessageEditEmbed(
		util.Itoa64(playerConf.ChannelID),
		util.Itoa64(playerConf.PinMessageID),
		ih.InventoryEmbedBuilder(inv, false),
	)
	if err != nil {
		return err
	}
	return nil
}

// Will check to see if the caller is authorized based off the following:
//  1. If the caller is in a player's confessional channel and is either:
//     1a. The owner of the inventory
//     1b. Has a whitelisted admin role AND is in the owners's confessional channel
//  2. If the caller is an admin and is in a whitelisted channel
func (ih *InventoryHandler) InventoryAuthorized(ctx ken.SubCommandContext) (bool, error) {
	event := ctx.GetEvent()
	invokeChannelID := event.ChannelID
	invoker := event.Member
	query := models.New(ih.pool)
	playerConf, err := query.GetPlayerConfessional(context.Background(), ih.player.ID)
	if err != nil {
		return false, err
	}

	// Base case of user is in confessional channel and is the owner of the inventory
	if util.Itoa64(ih.player.ID) == invoker.User.ID && util.Itoa64(playerConf.ChannelID) == invokeChannelID {
		return true, nil
	}

	// If not in confessional channel, check if in whitelist
	whitelistChannels, _ := query.ListAdminChannel(context.Background())
	if invokeChannelID != util.Itoa64(playerConf.ChannelID) {
		for _, whitelistChannelID := range whitelistChannels {
			if whitelistChannelID == invokeChannelID {
				return true, nil
			}
		}
		return false, nil
	}

	// Go through and make sure user has one of the allowed roles:
	for _, role := range invoker.Roles {
		for _, allowedRole := range discord.AdminRoles {
			if role == allowedRole {
				return true, nil
			}
		}
	}
	return true, nil
}

// Get player but will fetch from DB first to ensure accurate data
func (ih *InventoryHandler) SyncPlayer() models.Player {
	query := models.New(ih.pool)
	newPlayer, _ := query.GetPlayer(context.Background(), ih.player.ID)
	ih.player = newPlayer
	return newPlayer
}

// Pull local / current cached player
func (ih *InventoryHandler) GetPlayer() models.Player {
	return ih.player
}
