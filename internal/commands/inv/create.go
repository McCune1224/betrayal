package inv

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/mccune1224/betrayal/internal/models"
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
	discordID, _ := strconv.Atoi(playerArg.ID)
	player, err := query.CreatePlayer(bgCtx,
		models.CreatePlayerParams{
			ID:        int32(discordID),
			RoleID:    pgtype.Int4{Int32: roleResult.data.ID, Valid: true},
			Alive:     true,
			Coins:     defaultCoins,
			Luck:      defaultLuck,
			Alignment: roleResult.data.Alignment,
		},
	)
	if err != nil {
		log.Println(err)
		discord.ErrorMessage(ctx, "Failed to create player", "Unable to create player in database")
	}
	log.Println("hit part 1")

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
			discord.ErrorMessage(ctx, "Failed to create player ability", "Unable to create player ability in database")
		}
	}
	log.Println("hit part 2")

	//3. Create the player_perk

	for _, perk := range perksResult.data {
		_, err := query.CreatePlayerPerkJoin(bgCtx, models.CreatePlayerPerkJoinParams{
			PlayerID: player.ID,
			PerkID:   perk.ID,
		})
		if err != nil {
			log.Println(err)
			query.DeletePlayer(bgCtx, player.ID)
			discord.ErrorMessage(ctx, "Failed to create player perk", "Unable to create player perk in database")
		}
	}
	log.Println("hit part 3")

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
		discord.ErrorMessage(ctx, "Failed to edit inventory message", fmt.Sprintf("Could not send to channel %s", discord.MentionChannel(channelID)))
		ctx.GetSession().ChannelMessageDelete(channelID, pinMsg.ID)
		return err
	}

	iChannelID, _ := strconv.Atoi(channelID)
	_, err = query.CreatePlayerConfessional(bgCtx, models.CreatePlayerConfessionalParams{
		PlayerID:     player.ID,
		ChannelID:    int32(iChannelID),
		PinMessageID: 0,
	})
	if err != nil {
		log.Println(err)
		query.DeletePlayer(bgCtx, player.ID)
		discord.ErrorMessage(ctx, "Failed to update inventory", fmt.Sprintf("Unable to update inventory for %s", playerArg.Username))
		ctx.GetSession().ChannelMessageDelete(channelID, pinMsg.ID)
		return err
	}
	err = ctx.GetSession().ChannelMessagePin(channelID, pinMsg.ID)
	if err != nil {
		log.Println(err)
		query.DeletePlayer(bgCtx, player.ID)
		discord.ErrorMessage(ctx, "Failed to pin inventory message", fmt.Sprintf("Unable to pin inventory message for %s", playerArg.Username))
		return err
	}

	return discord.SuccessfulMessage(ctx, "Inventory Created", fmt.Sprintf("Created and pinined inventory for %s", playerArg.Username))
}
