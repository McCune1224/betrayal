package roll

import (
	"context"
	"fmt"
	"log"
	"math/rand"

	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/mccune1224/betrayal/internal/models"
	"github.com/zekrotja/ken"
)

type Roll struct {
	dbPool *pgxpool.Pool
}

func (r *Roll) Initialize(pool *pgxpool.Pool) {
	r.dbPool = pool
}

var _ ken.SlashCommand = (*Roll)(nil)

// Description implements ken.SlashCommand.
func (*Roll) Description() string {
	return "Determine luck for a given level"
}

// Name implements ken.SlashCommand.
func (*Roll) Name() string {
	return discord.DebugCmd + "roll"
}

// Options implements ken.SlashCommand.
func (*Roll) Options() []*discordgo.ApplicationCommandOption {
	targetTypes := []string{"item", "aa"}
	targetOpts := []*discordgo.ApplicationCommandOptionChoice{}
	for _, t := range targetTypes {
		targetOpts = append(targetOpts, &discordgo.ApplicationCommandOptionChoice{
			Name:  t,
			Value: t,
		})
	}
	minRarityOpts := []*discordgo.ApplicationCommandOptionChoice{}
	for _, r := range rarityPriorities {
		minRarityOpts = append(minRarityOpts, &discordgo.ApplicationCommandOptionChoice{
			Name:  string(r),
			Value: string(r),
		})
	}

	return []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "care_package",
			Description: "Give a random item and ability",
			Options: []*discordgo.ApplicationCommandOption{
				discord.UserCommandArg(false),
				discord.IntCommandArg("luck", "optional override of luck level", false),
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "item_rain",
			Description: "Make it rain!!",
			Options: []*discordgo.ApplicationCommandOption{
				discord.UserCommandArg(false),
				discord.IntCommandArg("luck", "optional override of luck level", false),
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "power_drop",
			Description: "give a random AA",
			Options: []*discordgo.ApplicationCommandOption{
				discord.UserCommandArg(false),
				discord.IntCommandArg("luck", "optional override of luck level", false),
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "manual",
			Description: "Manual roll for item or ability.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "target",
					Description: "target type",
					Required:    true,
					Choices:     targetOpts,
				},
				discord.IntCommandArg("luck", "level to roll for", true),
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "rarity",
			Description: "Roll for an item/any ability at a minimum rarity",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "target",
					Description: "target type",
					Required:    true,
					Choices:     targetOpts,
				},
				discord.IntCommandArg("luck", "level to roll for", true),
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "min_rarity",
					Description: "minimum rarity to roll for",
					Required:    true,
					Choices:     minRarityOpts,
				},
				discord.UserCommandArg(true),
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "player",
			Description: "Pick a random player",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionUser,
					Name:        "target_a",
					Description: "First player to choose from",
				},
				{
					Type:        discordgo.ApplicationCommandOptionUser,
					Name:        "target_b",
					Description: "Second player to choose from",
				},
			},
		},
	}
}

// Run implements ken.SlashCommand.
func (r *Roll) Run(ctx ken.Context) (err error) {
	return ctx.HandleSubCommands(
		ken.SubCommandHandler{Name: "manual", Run: r.luckManual},
		// ken.SubCommandHandler{Name: "rarity", Run: r.rollByMinimumRarity},
		// ken.SubCommandHandler{Name: "table", Run: r.luckTable},
		ken.SubCommandHandler{Name: "care_package", Run: r.luckCarePackage},
		ken.SubCommandHandler{Name: "item_rain", Run: r.luckItemRain},
		ken.SubCommandHandler{Name: "power_drop", Run: r.luckPowerDrop},
		ken.SubCommandHandler{Name: "player", Run: r.player},
	)
}

func (r *Roll) luckManual(ctx ken.SubCommandContext) (err error) {
	if err := ctx.Defer(); err != nil {
		log.Println(err)
		return err
	}

	if !discord.IsAdminRole(ctx, discord.AdminRoles...) {
		return discord.NotAdminError(ctx)
	}

	opts := ctx.Options()
	target := opts.GetByName("target").StringValue()
	level := opts.GetByName("luck").IntValue()

	rng := rand.Float64()
	rarity := RollRarityLevel(float64(level), rng)

	q := models.New(r.dbPool)
	dbCtx := context.Background()
	if target == "item" {
		item, err := q.GetRandomItemByRarity(dbCtx, rarity)
		if err != nil {
			log.Println(err)
			return discord.AlexError(ctx, "Failed to get random item")
		}

		return ctx.RespondEmbed(&discordgo.MessageEmbed{
			Title:       fmt.Sprintf("Got Item %s", item.Name),
			Description: item.Description,
			Footer: &discordgo.MessageEmbedFooter{
				Text: string(item.Rarity),
			},
		})
	} else {
		aa, err := q.GetRandomAnyAbilityByRarity(dbCtx, rarity)
		if err != nil {
			log.Println(err)
			return discord.AlexError(ctx, "")
		}
		return ctx.RespondEmbed(&discordgo.MessageEmbed{
			Title:       fmt.Sprintf("Got Ability %s", aa.Name),
			Description: aa.Description,
			Footer: &discordgo.MessageEmbedFooter{
				Text: string(aa.Rarity),
			},
		})
	}
}

func (r *Roll) luckTable(ctx ken.SubCommandContext) (err error) {
	if !discord.IsAdminRole(ctx, discord.AdminRoles...) {
		return discord.NotAdminError(ctx)
	}
	// Setting to eph for now to avoid flooding channels with bulky messages
	ctx.SetEphemeral(true)
	low := 0
	high := 100
	lArg, lOk := ctx.Options().GetByNameOptional("low")
	hArg, hOk := ctx.Options().GetByNameOptional("high")
	if lOk {
		low = int(lArg.IntValue())
		if !hOk {
			high = int(lArg.IntValue()) + 10
		}
	}

	if hOk {
		high = int(hArg.IntValue())
	}

	if low < 0 || high < 0 {
		return discord.ErrorMessage(ctx, "Invalid Range", "Please provide a non-negative number")
	}

	tMsg := ""

	for level := float64(low); level < float64(high); level++ {
		currChances := []float64{
			commonLuckChance(level) * 100,
			uncommonLuckChance(level) * 100,
			rareLuckChance(level) * 100,
			epicLuckChance(level) * 100,
			legendaryLuckChance(level) * 100,
			mythicalLuckChance(level) * 100,
		}

		tMsg += fmt.Sprintf("%d - ,", int(level))
		for i := range currChances {
			tMsg += fmt.Sprintf("%.2f%%\t", currChances[i])
		}
		tMsg += "\n"
	}

	return ctx.RespondMessage(discord.Code(tMsg))
}

// Version implements ken.SlashCommand.
func (*Roll) Version() string {
	return "1.0.0"
}

// func (r *Roll) rollByMinimumRarity(ctx ken.SubCommandContext) (err error) {
// 	if err := ctx.Defer(); err != nil {
// 		log.Println(err)
// 		return err
// 	}
//
// 	level := ctx.Options().GetByName("luck").IntValue()
// 	minimumRarity := models.Rarity(ctx.Options().GetByName("min_rarity").StringValue())
// 	target := ctx.Options().GetByName("target").StringValue()
//
// 	// q := models.New(r.dbPool)
// 	// dbCtx := context.Background()
// 	// userInv, err := inventory.NewInventoryHandler(ctx, r.dbPool)
// 	// if err != nil {
// 	// 	return discord.ErrorMessage(ctx, "Failed to get user inventory", err.Error())
// 	// }
//
// 	start := slices.Index(rarityPriorities, minimumRarity)
// 	rarityOptions := rarityPriorities[start:]
// 	rarity := rollAtRarity(float64(level), rarityOptions)
//
// 	if target == "item" {
// 		item, err := r.getRandomItem(rarity)
// 		if err != nil {
// 			log.Println(err)
// 			return discord.AlexError(ctx, "Failed to get random item")
// 		}
// 		return ctx.RespondEmbed(&discordgo.MessageEmbed{
// 			Title:       fmt.Sprintf("Got Item %s (%s)", item.Name, rarity),
// 			Description: item.Description,
// 			Footer: &discordgo.MessageEmbedFooter{
// 				Text: fmt.Sprintf("%s Note, this will not auto add to an inventory.", discord.EmojiWarning),
// 			},
// 		})
// 	} else {
// 		// FIXME: This is broken
// 		ability, err := q.
// 		if err != nil {
// 			log.Println(err)
// 			return discord.AlexError(ctx, "Failed to get random ability")
// 		}
// 		return ctx.RespondEmbed(&discordgo.MessageEmbed{
// 			Title:       fmt.Sprintf("Got Ability %s (%s)", ability.Name, rarity),
// 			Description: ability.Description,
//
// 			Footer: &discordgo.MessageEmbedFooter{
// 				Text: fmt.Sprintf("%s Note, this will not auto add to an inventory.", discord.EmojiWarning),
// 			},
// 		})
// 	}
// 	return discord.ErrorMessage(ctx, "Failed to get category", "Alex is a bad programmer")
// }

func (r *Roll) player(ctx ken.SubCommandContext) (err error) {
	playerA := ctx.Options().GetByName("target_a").UserValue(ctx)
	playerB := ctx.Options().GetByName("target_b").UserValue(ctx)

	roll := rand.Intn(2)
	if roll == 0 {
		return ctx.RespondMessage(fmt.Sprintf("%s %s was chosen", discord.EmojiRoll, playerA.Mention()))
	} else {
		return ctx.RespondMessage(fmt.Sprintf("%s %s was chosen", discord.EmojiRoll, playerB.Mention()))
	}
}
