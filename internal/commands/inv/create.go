package inv

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/mccune1224/betrayal/internal/models"
	"github.com/mccune1224/betrayal/internal/util"
	"github.com/zekrotja/ken"
)

// TODO: Maybe make these configurable?
const (
	defaultCoins      = 200
	defaultItemsLimit = 4
	defaultLuck       = 0
)

func (i *Inv) create(ctx ken.SubCommandContext) (err error) {
	if err = ctx.Defer(); err != nil {
		log.Println(err)
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
		discord.ErrorMessage(ctx, "Failed to get Role", fmt.Sprintf("Cannot find role %s", roleArg))
		return roleResult.err
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
		discord.ErrorMessage(ctx, "Failed to get Role Abilities", abilitiesResult.err.Error())
		return abilitiesResult.err
	}

	perksResult := <-perksCh
	if perksResult.err != nil {
		discord.ErrorMessage(ctx, "Failed to get Role Perks", perksResult.err.Error())
		return perksResult.err
	}

	abilityNames := make([]string, len(abilitiesResult.data))
	for i, ability := range abilitiesResult.data {
		chargeNumber := ""
		if ability.DefaultCharges == -1 {
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
		log.Println(err)
		return discord.ErrorMessage(ctx, "Failed to create player", "Unable to create player in database")
	}
	player, err := query.CreatePlayer(bgCtx,
		models.CreatePlayerParams{
			ID:        int64(discordID),
			RoleID:    pgtype.Int4{Int32: roleResult.data.ID, Valid: true},
			Alive:     true,
			Coins:     defaultCoins,
			Luck:      defaultLuck,
			Alignment: roleResult.data.Alignment,
		},
	)
	if err != nil {
		log.Println(err)
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
			log.Println(err)
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
			log.Println(err)
			query.DeletePlayer(bgCtx, player.ID)
			return discord.ErrorMessage(ctx, "Failed to create player perk", "Unable to create player perk in database")
		}
	}

	// FIXME: Lord please forgive for the unholy amount of switch statements I am about to unleash
	// Will need to make some sort of Website or UI to allow for custom roles to be created instead of me hardcoding them
	roleName := strings.ToLower(roleResult.data.Name)
	statuses, _ := query.ListStatus(bgCtx)
	statusMap := make(map[string]int32, len(statuses))
	for _, status := range statuses {
		statusMap[status.Name] = status.ID
	}
	switch roleName {
	// --- GOOD ROLES ---
	case "cerberus":
		// Due to perk Hades' Hound
		immunities := []string{"Frozen", "Burned"}
		err := mapImmunities(query, player, immunities, statusMap)
		if err != nil {
			log.Println(err)
			query.DeletePlayer(bgCtx, player.ID)
			return discord.ErrorMessage(ctx, "Failed to create immunity", "Unable to create immunity in database")
		}

	case "detective":
		// Due to perk Clever
		immunities := []string{"Blackmailed", "Disabled", "Despaired"}
		err := mapImmunities(query, player, immunities, statusMap)
		if err != nil {
			log.Println(err)
			query.DeletePlayer(bgCtx, player.ID)
			return discord.ErrorMessage(ctx, "Failed to create immunity", "Unable to create immunity in database")
		}
	case "fisherman":
		// Due to perk Barrels
		_, err := query.UpdatePlayerItemLimit(bgCtx, models.UpdatePlayerItemLimitParams{
			ID:        player.ID,
			ItemLimit: 8,
		})
		if err != nil {
			log.Println(err)
			query.DeletePlayer(bgCtx, player.ID)
			return discord.ErrorMessage(ctx, "Failed to update item limit", "Unable to update item limit in database")
		}
	case "hero":
		// Due to perk Compos Mentis
		immunities := []string{"Madness"}
		err := mapImmunities(query, player, immunities, statusMap)
		if err != nil {
			log.Println(err)
			query.DeletePlayer(bgCtx, player.ID)
			return discord.ErrorMessage(ctx, "Failed to create immunity", "Unable to create immunity in database")
		}
	case "nurse":
		// Due to perk Powerful Immunity
		immunities := []string{"Death Cursed", "Frozen", "Paralyzed", "Burned", "Empowered", "Drunk", "Restrained", "Disabled", "Blackmailed", "Despaired", "Madness", "Unlucky"}
		err := mapImmunities(query, player, immunities, statusMap)
		if err != nil {
			log.Println(err)
			query.DeletePlayer(bgCtx, player.ID)
			return discord.ErrorMessage(ctx, "Failed to create immunity", "Unable to create immunity in database")
		}
	case "terminal":
		// Due to perk Heartbeats
		immunities := []string{"Death Cursed", "Frozen", "Paralyzed", "Burned", "Empowered", "Drunk", "Restrained", "Disabled", "Blackmailed", "Despaired", "Madness", "Unlucky"}
		err := mapImmunities(query, player, immunities, statusMap)
		if err != nil {
			log.Println(err)
			query.DeletePlayer(bgCtx, player.ID)
			return discord.ErrorMessage(ctx, "Failed to create immunity", "Unable to create immunity in database")
		}
	case "wizard":
		// due to perk Magic Barrier
		immunities := []string{"Frozen", "Paralyzed", "Burned", "Cursed"}
		err := mapImmunities(query, player, immunities, statusMap)
		if err != nil {
			log.Println(err)
			query.DeletePlayer(bgCtx, player.ID)
			return discord.ErrorMessage(ctx, "Failed to create immunity", "Unable to create immunity in database")
		}
	case "yeti":
		// Due to perk Winter Coat
		immunities := []string{"Frozen"}
		err := mapImmunities(query, player, immunities, statusMap)
		if err != nil {
			log.Println(err)
			query.DeletePlayer(bgCtx, player.ID)
			return discord.ErrorMessage(ctx, "Failed to create immunity", "Unable to create immunity in database")
		}

		// Neutral Roles
	case "cyborg":
		immunities := []string{"Paralyzed", "Frozen", "Burned", "Despaired", "Blackmailed", "Drunk"}
		err := mapImmunities(query, player, immunities, statusMap)
		if err != nil {
			log.Println(err)
			query.DeletePlayer(bgCtx, player.ID)
			return discord.ErrorMessage(ctx, "Failed to create immunity", "Unable to create immunity in database")
		}
	case "entertainer":
		// Due to perk Top-Hat Tip
		immunities := []string{"Unlucky"}
		err := mapImmunities(query, player, immunities, statusMap)
		if err != nil {
			log.Println(err)
			query.DeletePlayer(bgCtx, player.ID)
			return discord.ErrorMessage(ctx, "Failed to create immunity", "Unable to create immunity in database")
		}
		statuses := []string{"Lucky"}
		err = mapStatuses(query, player, statuses, statusMap)
		if err != nil {
			log.Println(err)
			query.DeletePlayer(bgCtx, player.ID)
			return discord.ErrorMessage(ctx, "Failed to create immunity", "Unable to create immunity in database")
		}
	case "magician":
		// Due to perk Top-Hat Tip
		statuses := []string{"Lucky"}
		err := mapImmunities(query, player, statuses, statusMap)
		if err != nil {
			log.Println(err)
			query.DeletePlayer(bgCtx, player.ID)
			return discord.ErrorMessage(ctx, "Failed to create immunity", "Unable to create immunity in database")
		}
		immunities := []string{"Unlucky"}
		err = mapImmunities(query, player, immunities, statusMap)
		if err != nil {
			log.Println(err)
			query.DeletePlayer(bgCtx, player.ID)
			return discord.ErrorMessage(ctx, "Failed to create immunity", "Unable to create immunity in database")
		}
	case "masochist":
		// Due to perk One Track Mind
		immunities := []string{"Lucky"}
		err := mapImmunities(query, player, immunities, statusMap)
		if err != nil {
			log.Println(err)
			query.DeletePlayer(bgCtx, player.ID)
			return discord.ErrorMessage(ctx, "Failed to create immunity", "Unable to create immunity in database")
		}
	case "succubus":
		// Due to perk Dominatrix
		immunities := []string{"Blackmail"}
		err := mapImmunities(query, player, immunities, statusMap)
		if err != nil {
			log.Println(err)
			query.DeletePlayer(bgCtx, player.ID)
			return discord.ErrorMessage(ctx, "Failed to create immunity", "Unable to create immunity in database")
		}
	//
	// Evil Roles
	case "arsonist":
		// Due to perk Ashes to Ashes / Flamed
		immunities := []string{"Burned"}
		err := mapImmunities(query, player, immunities, statusMap)
		if err != nil {
			log.Println(err)
			query.DeletePlayer(bgCtx, player.ID)
			return discord.ErrorMessage(ctx, "Failed to create immunity", "Unable to create immunity in database")
		}
	case "cultist":
		immunities := []string{"Curse"}
		err := mapImmunities(query, player, immunities, statusMap)
		if err != nil {
			log.Println(err)
			query.DeletePlayer(bgCtx, player.ID)
			return discord.ErrorMessage(ctx, "Failed to create immunity", "Unable to create immunity in database")
		}
	case "director":
		immunities := []string{"Despaired", "Blackmailed", "Drunk"}
		err := mapImmunities(query, player, immunities, statusMap)
		if err != nil {
			log.Println(err)
			query.DeletePlayer(bgCtx, player.ID)
			return discord.ErrorMessage(ctx, "Failed to create immunity", "Unable to create immunity in database")
		}
	case "gatekeeper":
		immunities := []string{"Restrained", "Paralyzed", "Frozen"}
		err := mapImmunities(query, player, immunities, statusMap)
		if err != nil {
			log.Println(err)
			query.DeletePlayer(bgCtx, player.ID)
			return discord.ErrorMessage(ctx, "Failed to create immunity", "Unable to create immunity in database")
		}
	case "hacker":
		immunities := []string{"Disabled", "Blackmailed"}
		err := mapImmunities(query, player, immunities, statusMap)
		if err != nil {
			log.Println(err)
			query.DeletePlayer(bgCtx, player.ID)
			return discord.ErrorMessage(ctx, "Failed to create immunity", "Unable to create immunity in database")
		}
	case "highwayman":
		immunities := []string{"Madness"}
		err := mapImmunities(query, player, immunities, statusMap)
		if err != nil {
			log.Println(err)
			query.DeletePlayer(bgCtx, player.ID)
			return discord.ErrorMessage(ctx, "Failed to create immunity", "Unable to create immunity in database")
		}
	case "imp":
		immunities := []string{"Despaired", "Paralyzed"}
		err := mapImmunities(query, player, immunities, statusMap)
		if err != nil {
			log.Println(err)
			query.DeletePlayer(bgCtx, player.ID)
			return discord.ErrorMessage(ctx, "Failed to create immunity", "Unable to create immunity in database")
		}
	case "threatener":
		_, err := query.UpdatePlayerItemLimit(bgCtx, models.UpdatePlayerItemLimitParams{
			ID:        player.ID,
			ItemLimit: 6,
		})
		if err != nil {
			log.Println(err)
			query.DeletePlayer(bgCtx, player.ID)
			return discord.ErrorMessage(ctx, "Failed to update item limit", "Unable to update item limit in database")
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
		discord.ErrorMessage(ctx, "Failed to send message", err.Error())
		ctx.GetSession().ChannelMessageDelete(channelID, pinMsg.ID)
		return err
	}
	_, err = ctx.GetSession().ChannelMessageEditEmbed(channelID, pinMsg.ID, embd)
	if err != nil {
		query.DeletePlayer(bgCtx, player.ID)
		log.Println(err)
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
		log.Println(err)
		query.DeletePlayer(bgCtx, player.ID)
		ctx.GetSession().ChannelMessageDelete(channelID, pinMsg.ID)
		return discord.ErrorMessage(ctx, "Failed to update inventory", fmt.Sprintf("Unable to update inventory for %s", playerArg.Username))
	}
	err = ctx.GetSession().ChannelMessagePin(channelID, pinMsg.ID)
	if err != nil {
		log.Println(err)
		query.DeletePlayer(bgCtx, player.ID)
		return discord.ErrorMessage(ctx, "Failed to pin inventory message", fmt.Sprintf("Unable to pin inventory message for %s", playerArg.Username))
	}

	return discord.SuccessfulMessage(ctx, "Inventory Created", fmt.Sprintf("Created and pinined inventory for %s", playerArg.Username))
}

func (i Inv) delete(ctx ken.SubCommandContext) (err error) {
	if err = ctx.Defer(); err != nil {
		log.Println(err)
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
		log.Println(err)
		discord.ErrorMessage(ctx, "Failed to get player", fmt.Sprintf("Unable to get player %s", playerArg.Username))
		return err
	}

	playerConf, err := query.GetPlayerConfessional(bgCtx, player.ID)
	if err != nil {
		log.Println(err)
		discord.ErrorMessage(ctx, "Failed to get player confessional", fmt.Sprintf("Unable to get player confessional for %s", playerArg.Username))
		return err
	}

	err = ctx.GetSession().ChannelMessageDelete(util.Itoa64(playerConf.ChannelID), strconv.Itoa(int(playerConf.PinMessageID)))
	if err != nil {
		log.Println(err)
		discord.ErrorMessage(ctx, "Failed to delete message", fmt.Sprintf("Unable to delete message for %s", playerArg.Username))
		return err
	}

	err = query.DeletePlayer(bgCtx, player.ID)
	if err != nil {
		log.Println(err)
		discord.ErrorMessage(ctx, "Failed to delete player", fmt.Sprintf("Unable to delete player %s", playerArg.Username))
		return err
	}

	return discord.SuccessfulMessage(ctx, "Deleted Player Inventory", fmt.Sprintf("Deleted inventory for %s", playerArg.Username))
}

func mapImmunities(query *models.Queries, player models.Player, immunities []string, statusMap map[string]int32) (err error) {
	for _, immunity := range immunities {
		_, err := query.CreatePlayerImmunityJoin(context.Background(), models.CreatePlayerImmunityJoinParams{
			PlayerID: player.ID,
			StatusID: statusMap[immunity],
		})
		if err != nil {
			log.Println(err)
			return err
		}
	}
	return nil
}

func mapStatuses(query *models.Queries, player models.Player, statuses []string, statusMap map[string]int32) (err error) {
	for _, status := range statuses {
		_, err := query.CreatePlayerStatusJoin(context.Background(), models.CreatePlayerStatusJoinParams{
			PlayerID: player.ID,
			StatusID: statusMap[status],
		})
		if err != nil {
			log.Println(err)
			return err
		}
	}
	return nil
}
