package inv

import (
	"context"
	"fmt"
	"github.com/mccune1224/betrayal/internal/logger"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/mccune1224/betrayal/internal/models"
	"github.com/mccune1224/betrayal/internal/services/inventory"
	"github.com/mccune1224/betrayal/internal/util"
	"github.com/zekrotja/ken"
)

// Configurable defaults for `/inv create`, backed by the game_config table
// (seeded by migration 000029). These constants are the fallback when a row is
// missing or unparseable.
const (
	cfgDefaultCoins      = 200
	cfgDefaultItemsLimit = 4
	cfgDefaultLuck       = 0

	configKeyDefaultCoins      = "default_coins"
	configKeyDefaultItemsLimit = "default_items_limit"
	configKeyDefaultLuck       = "default_luck"
)

func (i *Inv) create(ctx ken.SubCommandContext) (err error) {
	if err = ctx.Defer(); err != nil {
		logger.Get().Error().Err(err).Msg("operation failed")
		return err
	}
	if !discord.IsAdminRole(ctx, discord.AdminRoles...) {
		err = discord.ErrorMessage(ctx, "Unauthorized", "You are not authorized to use this command.")
		return err
	}

	playerArg := ctx.Options().GetByName("user").UserValue(ctx)
	roleArg := ctx.Options().GetByName("role").StringValue()
	channelID := ctx.GetEvent().ChannelID
	query := models.New(i.dbPool)

	// use channels to spawn goroutines for fetching role, role abilities, and role perks

	bgCtx := context.Background()

	// make generic struct to handle a channel of type T, and has an error property
	type channel[T any] struct {
		data T
		err  error
	}

	// roleCh := make(chan models.Role, 1)
	roleCh := make(chan channel[models.Role], 1)
	go func() {
		role, err := query.GetRoleByFuzzy(bgCtx, roleArg)
		roleCh <- channel[models.Role]{data: role, err: err}
	}()

	roleResult := <-roleCh
	if roleResult.err != nil {
		return discord.ErrorMessage(ctx, "Failed to get Role", fmt.Sprintf("Cannot find role %s", roleArg))
	}

	role := roleResult.data

	abilitiesCh := make(chan channel[[]models.AbilityInfo], 1)
	perksCh := make(chan channel[[]models.PerkInfo], 1)

	go func() {
		abilities, err := query.ListRoleAbilityForRole(bgCtx, role.ID)
		abilitiesCh <- channel[[]models.AbilityInfo]{data: abilities, err: err}
	}()

	go func() {
		perks, err := query.ListRolePerkForRole(bgCtx, role.ID)
		perksCh <- channel[[]models.PerkInfo]{data: perks, err: err}
	}()

	abilitiesResult := <-abilitiesCh
	if abilitiesResult.err != nil {
		return discord.ErrorMessage(ctx, "Failed to get Role Abilities", abilitiesResult.err.Error())
	}

	perksResult := <-perksCh
	if perksResult.err != nil {
		return discord.ErrorMessage(ctx, "Failed to get Role Perks", perksResult.err.Error())
	}

	abilityNames := make([]string, len(abilitiesResult.data))
	for i, ability := range abilitiesResult.data {
		chargeNumber := ""
		if ability.DefaultCharges == 999999 {
			chargeNumber = "∞"
		} else {
			chargeNumber = fmt.Sprintf("%d", ability.DefaultCharges)
		}

		abilityNames[i] = fmt.Sprintf("%s [%s]", ability.Name, chargeNumber)
	}
	perkNames := make([]string, len(perksResult.data))
	for i, perk := range perksResult.data {
		perkNames[i] = perk.Name
	}

	//1. Create the player
	discordID, err := util.Atoi64(playerArg.ID)
	if err != nil {
		logger.Get().Error().Err(err).Msg("operation failed")
		return discord.ErrorMessage(ctx, "Failed to create player", "Unable to create player in database")
	}

	num, err := util.Numeric(0.0)
	if err != nil {
		logger.Get().Error().Err(err).Msg("operation failed")
		return discord.AlexError(ctx, "Failed to create player")
	}

	player, err := query.CreatePlayer(bgCtx,
		models.CreatePlayerParams{
			ID:        int64(discordID),
			RoleID:    pgtype.Int4{Int32: roleResult.data.ID, Valid: true},
			Alive:     true,
			Coins:     gameConfigInt(bgCtx, query, configKeyDefaultCoins, cfgDefaultCoins),
			CoinBonus: num,
			Luck:      gameConfigInt(bgCtx, query, configKeyDefaultLuck, cfgDefaultLuck),
			ItemLimit: gameConfigInt(bgCtx, query, configKeyDefaultItemsLimit, cfgDefaultItemsLimit),
			Alignment: roleResult.data.Alignment,
		},
	)
	if err != nil {
		logger.Get().Error().Err(err).Msg("operation failed")
		return discord.ErrorMessage(ctx, "Failed to create player", "Unable to create player in database")
	}

	// return ctx.RespondMessage(fmt.Sprintf("%v", player))

	//2. Create the player_ability
	for _, ability := range abilitiesResult.data {
		_, err := query.CreatePlayerAbilityJoin(bgCtx, models.CreatePlayerAbilityJoinParams{
			PlayerID:  player.ID,
			AbilityID: ability.ID,
			Quantity:  ability.DefaultCharges,
		})
		if err != nil {
			logger.Get().Error().Err(err).Msg("operation failed")
			query.DeletePlayer(bgCtx, player.ID)
			return discord.ErrorMessage(ctx, "Failed to create player ability", "Unable to create player ability in database")
		}
	}

	//3. Create the player_perk

	for _, perk := range perksResult.data {
		_, err := query.CreatePlayerPerkJoin(bgCtx, models.CreatePlayerPerkJoinParams{
			PlayerID: player.ID,
			PerkID:   perk.ID,
		})
		if err != nil {
			logger.Get().Error().Err(err).Msg("operation failed")
			query.DeletePlayer(bgCtx, player.ID)
			return discord.ErrorMessage(ctx, "Failed to create player perk", "Unable to create player perk in database")
		}
	}

	// Apply per-role post-creation adjustments (immunities, statuses, item
	// limit) driven by role perks. Kept as data in roleOpsByRole so a role's
	// setup is a one-line change instead of a switch arm.
	statuses, err := query.ListStatus(bgCtx)
	if err != nil {
		logger.Get().Error().Err(err).Msg("operation failed")
		query.DeletePlayer(bgCtx, player.ID)
		return discord.ErrorMessage(ctx, "Failed to get statuses", "Unable to fetch statuses in database")
	}
	statusMap := make(map[string]int32, len(statuses))
	for _, status := range statuses {
		statusMap[status.Name] = status.ID
	}

	if ops, ok := roleOpsByRole[strings.ToLower(roleResult.data.Name)]; ok {
		if err := applyRoleOps(bgCtx, query, player, ops, statusMap); err != nil {
			logger.Get().Error().Err(err).Msg("operation failed")
			query.DeletePlayer(bgCtx, player.ID)
			return discord.ErrorMessage(ctx, "Failed to apply role setup", "Unable to apply role immunities/statuses in database")
		}
	}

	//4. Create the player_confessional
	// embd := InventoryEmbedBuilder(defaultInv, false)
	embd := &discordgo.MessageEmbed{
		Title: fmt.Sprintf("Idk finished inventory lol %s", playerArg.Username),
	}

	pinMsg, err := ctx.GetSession().ChannelMessageSendEmbed(channelID, &discordgo.MessageEmbed{
		Title: fmt.Sprintf("🏗 %s Inventory in creation 🏗 ", playerArg.Username),
	})
	if err != nil {
		query.DeletePlayer(bgCtx, player.ID)
		return discord.ErrorMessage(ctx, "Failed to send message", err.Error())
	}
	_, err = ctx.GetSession().ChannelMessageEditEmbed(channelID, pinMsg.ID, embd)
	if err != nil {
		query.DeletePlayer(bgCtx, player.ID)
		logger.Get().Error().Err(err).Msg("operation failed")
		ctx.GetSession().ChannelMessageDelete(channelID, pinMsg.ID)
		return discord.ErrorMessage(ctx, "Failed to edit inventory message", fmt.Sprintf("Could not send to channel %s", discord.MentionChannel(channelID)))
	}

	iChannelID, _ := util.Atoi64(channelID)
	iPinMessageID, _ := util.Atoi64(pinMsg.ID)
	_, err = query.CreatePlayerConfessional(bgCtx, models.CreatePlayerConfessionalParams{
		PlayerID:     player.ID,
		ChannelID:    iChannelID,
		PinMessageID: iPinMessageID,
	})
	if err != nil {
		logger.Get().Error().Err(err).Msg("operation failed")
		query.DeletePlayer(bgCtx, player.ID)
		ctx.GetSession().ChannelMessageDelete(channelID, pinMsg.ID)
		return discord.ErrorMessage(ctx, "Failed to update inventory", fmt.Sprintf("Unable to update inventory for %s", playerArg.Username))
	}
	err = ctx.GetSession().ChannelMessagePin(channelID, pinMsg.ID)
	if err != nil {
		logger.Get().Error().Err(err).Msg("operation failed")
		query.DeletePlayer(bgCtx, player.ID)
		return discord.ErrorMessage(ctx, "Failed to pin inventory message", fmt.Sprintf("Unable to pin inventory message for %s", playerArg.Username))
	}

	h, err := inventory.NewInventoryHandler(ctx, i.dbPool)
	if err != nil {
		logger.Get().Error().Err(err).Msg("operation failed")
		return discord.AlexError(ctx, "failed to init inv handler")
	}
	defer h.UpdateInventoryMessage(ctx.GetSession())

	return discord.SuccessfulMessage(ctx, "Inventory Created", fmt.Sprintf("Created and pinined inventory for %s", playerArg.Username))
}

func (i Inv) delete(ctx ken.SubCommandContext) (err error) {
	if err = ctx.Defer(); err != nil {
		logger.Get().Error().Err(err).Msg("operation failed")
		return err
	}
	if !discord.IsAdminRole(ctx, discord.AdminRoles...) {
		err = discord.ErrorMessage(ctx, "Unauthorized", "You are not authorized to use this command.")
		return err
	}

	playerArg := ctx.Options().GetByName("user").UserValue(ctx)
	query := models.New(i.dbPool)
	bgCtx := context.Background()

	pId, _ := util.Atoi64(playerArg.ID)
	player, err := query.GetPlayer(bgCtx, int64(pId))
	if err != nil {
		logger.Get().Error().Err(err).Msg("operation failed")
		return discord.ErrorMessage(ctx, "Failed to get player", fmt.Sprintf("Unable to get player %s", playerArg.Username))
	}

	playerConf, err := query.GetPlayerConfessional(bgCtx, player.ID)
	if err != nil {
		logger.Get().Error().Err(err).Msg("operation failed")
		return discord.ErrorMessage(ctx, "Failed to get player confessional", fmt.Sprintf("Unable to get player confessional for %s", playerArg.Username))
	}

	err = ctx.GetSession().ChannelMessageDelete(util.Itoa64(playerConf.ChannelID), strconv.Itoa(int(playerConf.PinMessageID)))
	if err != nil {
		logger.Get().Error().Err(err).Msg("operation failed")
		return discord.ErrorMessage(ctx, "Failed to delete message", fmt.Sprintf("Unable to delete message for %s", playerArg.Username))
	}

	err = query.DeletePlayer(bgCtx, player.ID)
	if err != nil {
		logger.Get().Error().Err(err).Msg("operation failed")
		return discord.ErrorMessage(ctx, "Failed to delete player", fmt.Sprintf("Unable to delete player %s", playerArg.Username))
	}

	return discord.SuccessfulMessage(ctx, "Deleted Player Inventory", fmt.Sprintf("Deleted inventory for %s", playerArg.Username))
}

// gameConfigInt reads an integer game config value, falling back to the given
// default when the row is missing or not a valid integer.
func gameConfigInt(ctx context.Context, q *models.Queries, key string, fallback int32) int32 {
	raw, err := q.GetGameConfig(ctx, key)
	if err != nil {
		return fallback
	}
	n, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		logger.Get().Warn().Str("key", key).Str("value", raw).Msg("game config value is not an integer; using fallback")
		return fallback
	}
	return int32(n)
}

// roleOps describes the post-creation adjustments applied to a player based on
// their role's perks. A nil itemLimit leaves the player's item limit untouched.
type roleOps struct {
	immunities []string
	statuses   []string
	itemLimit  *int32
}

func int32Ptr(v int32) *int32 { return &v }

// roleOpsByRole maps a lowercased role name to its creation-time adjustments
// (driven by the role's perks). Previously this was a giant switch statement;
// it is data now so a role's setup is a one-line change.
//
// NOTE: three roles were fixed while converting to this map:
//   - magician's "Lucky" status was accidentally mapped as an immunity (the
//     call went to mapImmunities instead of mapStatuses); it now gets the Lucky
//     status like entertainer (same perk, Top-Hat Tip).
//   - succubus referenced "Blackmail" and cultist referenced "Curse", neither
//     of which exists in the status table (migration 000008 seeds "Blackmailed"
//     and "Cursed"). The typo'd names made `/inv create` fail with a foreign
//     key violation on those roles.
var roleOpsByRole = map[string]roleOps{
	// Good roles
	"cerberus":  {immunities: []string{"Frozen", "Burned"}},                                                                                                                                // Hades' Hound
	"detective": {immunities: []string{"Blackmailed", "Disabled", "Despaired"}},                                                                                                            // Clever
	"fisherman": {itemLimit: int32Ptr(8)},                                                                                                                                                  // Barrels
	"hero":      {immunities: []string{"Madness"}},                                                                                                                                         // Compos Mentis
	"nurse":     {immunities: []string{"Death Cursed", "Frozen", "Paralyzed", "Burned", "Empowered", "Drunk", "Restrained", "Disabled", "Blackmailed", "Despaired", "Madness", "Unlucky"}}, // Powerful Immunity
	"terminal":  {immunities: []string{"Death Cursed", "Frozen", "Paralyzed", "Burned", "Empowered", "Drunk", "Restrained", "Disabled", "Blackmailed", "Despaired", "Madness", "Unlucky"}}, // Heartbeats
	"wizard":    {immunities: []string{"Frozen", "Paralyzed", "Burned", "Cursed"}},                                                                                                         // Magic Barrier
	"yeti":      {immunities: []string{"Frozen"}},                                                                                                                                          // Winter Coat
	// Neutral roles
	"cyborg":      {immunities: []string{"Paralyzed", "Frozen", "Burned", "Despaired", "Blackmailed", "Drunk"}},
	"entertainer": {immunities: []string{"Unlucky"}, statuses: []string{"Lucky"}}, // Top-Hat Tip
	"magician":    {immunities: []string{"Unlucky"}, statuses: []string{"Lucky"}}, // Top-Hat Tip
	"masochist":   {immunities: []string{"Lucky"}},                                // One Track Mind
	"succubus":    {immunities: []string{"Blackmailed"}},                          // Dominatrix
	// Evil roles
	"arsonist":   {immunities: []string{"Burned"}}, // Ashes to Ashes / Flamed
	"cultist":    {immunities: []string{"Cursed"}},
	"director":   {immunities: []string{"Despaired", "Blackmailed", "Drunk"}},
	"gatekeeper": {immunities: []string{"Restrained", "Paralyzed", "Frozen"}},
	"hacker":     {immunities: []string{"Disabled", "Blackmailed"}},
	"highwayman": {immunities: []string{"Madness"}},
	"imp":        {immunities: []string{"Despaired", "Paralyzed"}},
	"threatener": {itemLimit: int32Ptr(6)},
}

// applyRoleOps applies the creation-time adjustments for a role to a player.
func applyRoleOps(ctx context.Context, query *models.Queries, player models.Player, ops roleOps, statusMap map[string]int32) error {
	if ops.itemLimit != nil {
		if _, err := query.UpdatePlayerItemLimit(ctx, models.UpdatePlayerItemLimitParams{
			ID:        player.ID,
			ItemLimit: *ops.itemLimit,
		}); err != nil {
			return err
		}
	}
	if err := mapStatuses(ctx, query, player, ops.statuses, statusMap); err != nil {
		return err
	}
	return mapImmunities(ctx, query, player, ops.immunities, statusMap)
}

// mapImmunities creates player immunity joins for each named status. Unknown
// status names are skipped with a warning instead of aborting the player
// creation (previously a typo'd name inserted status id 0 and failed the
// foreign key, deleting the freshly created player).
func mapImmunities(ctx context.Context, query *models.Queries, player models.Player, immunities []string, statusMap map[string]int32) (err error) {
	for _, immunity := range immunities {
		statusID, ok := statusMap[immunity]
		if !ok {
			logger.Get().Warn().Str("status", immunity).Msg("unknown status name in role ops; skipping")
			continue
		}
		_, err := query.CreatePlayerImmunityJoin(ctx, models.CreatePlayerImmunityJoinParams{
			PlayerID: player.ID,
			StatusID: statusID,
		})
		if err != nil {
			logger.Get().Error().Err(err).Msg("operation failed")
			return err
		}
	}
	return nil
}

func mapStatuses(ctx context.Context, query *models.Queries, player models.Player, statuses []string, statusMap map[string]int32) (err error) {
	for _, status := range statuses {
		statusID, ok := statusMap[status]
		if !ok {
			logger.Get().Warn().Str("status", status).Msg("unknown status name in role ops; skipping")
			continue
		}
		_, err := query.CreatePlayerStatusJoin(ctx, models.CreatePlayerStatusJoinParams{
			PlayerID: player.ID,
			StatusID: statusID,
		})
		if err != nil {
			logger.Get().Error().Err(err).Msg("operation failed")
			return err
		}
	}
	return nil
}
