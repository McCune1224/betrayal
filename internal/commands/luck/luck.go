package roll

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/mccune1224/betrayal/internal/commands/inventory"
	"github.com/mccune1224/betrayal/internal/cron"
	"github.com/mccune1224/betrayal/internal/data"
	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/zekrotja/ken"
)

type Roll struct {
	models    data.Models
	scheduler *cron.BetrayalScheduler
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

	options := []*discordgo.ApplicationCommandOptionChoice{}
	for _, t := range targetTypes {
		options = append(options, &discordgo.ApplicationCommandOptionChoice{
			Name:  t,
			Value: t,
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
			Description: "Manual roll for item or ability. Optional argument to add to an inventory",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "target",
					Description: "target type",
					Required:    true,
					Choices:     options,
				},
				discord.IntCommandArg("level", "level to roll for", true),
				discord.UserCommandArg(false),
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "debug",
			Description: "simulate roll and show chances",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "target",
					Description: "target type",
					Required:    true,
					Choices:     options,
				},
				discord.IntCommandArg("level", "level to roll for", true),
				discord.UserCommandArg(false),
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "table",
			Description: "Table view of luck calculations",
			Options: []*discordgo.ApplicationCommandOption{
				discord.IntCommandArg("low", "low range for luck", false),
				discord.IntCommandArg("high", "high range for luck", false),
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "wheel",
			Description: "Spin the wheel for a game event",
		},
	}
}

// Run implements ken.SlashCommand.
func (r *Roll) Run(ctx ken.Context) (err error) {
	return ctx.HandleSubCommands(
		ken.SubCommandHandler{Name: "manual", Run: r.luckManual},
		ken.SubCommandHandler{Name: "debug", Run: r.luckDebug},
		ken.SubCommandHandler{Name: "care_package", Run: r.luckCarePackage},
		ken.SubCommandHandler{Name: "table", Run: r.luckTable},
		ken.SubCommandHandler{Name: "item_rain", Run: r.luckItemRain},
		ken.SubCommandHandler{Name: "power_drop", Run: r.luckPowerDrop},
		ken.SubCommandHandler{Name: "wheel", Run: r.wheel},
	)
}

func (r *Roll) luckManual(ctx ken.SubCommandContext) (err error) {
	var inv *data.Inventory
	opts := ctx.Options()
	target := opts.GetByName("target").StringValue()
	level := opts.GetByName("level").IntValue()
	user, ok := opts.GetByNameOptional("user")
	if ok {
		inv, err = inventory.Fetch(ctx, r.models, true)
		if err != nil {
			if errors.Is(err, inventory.ErrNotAuthorized) {
				return discord.NotAuthorizedError(ctx)
			}
		}
	}

	rng := rand.Float64()
	luckType := RollLuck(float64(level), rng)

	if target == "item" {
		isAdding := false
		item, err := r.getRandomItem(luckType)
		if err != nil {
			isAdding = true
			log.Println(err)
			return discord.AlexError(ctx)
		}
		if inv != nil {
			inv.Items = append(inv.Items, item.Name)
			err = r.models.Inventories.UpdateItems(inv)
			if err != nil {
				log.Println(err)
				return discord.AlexError(ctx)
			}
			err = inventory.UpdateInventoryMessage(ctx, inv)
			if err != nil {
				log.Println(err)
				return discord.AlexError(ctx)
			}
		}
		foot := &discordgo.MessageEmbedFooter{}
		if isAdding {
			foot.Text = fmt.Sprintf("Added to %s's inventory", user.UserValue(ctx).Username)
		}
		return ctx.RespondEmbed(&discordgo.MessageEmbed{
			Title:       fmt.Sprintf("Got Item %s (%s)", item.Name, luckType),
			Description: item.Description,
			Footer:      foot,
		})
	}

	if target == "ability" {
		isAdding := false
		ability, err := r.getRandomAnyAbility(inv.RoleName, luckType)
		if err != nil {
			isAdding = true
			log.Println(err)
			return discord.AlexError(ctx)
		}
		if inv != nil {
			if ability.RoleSpecific == inv.RoleName {
				ab, err := r.models.Abilities.GetByName(ability.RoleSpecific)
				if err != nil {
					log.Println(err)
					return discord.AlexError(ctx)
				}
				inventory.UpsertAbility(inv, ab)
				err = r.models.Inventories.UpdateAbilities(inv)
				if err != nil {
					log.Println(err)
					return discord.AlexError(ctx)
				}

			} else {
				inventory.UpsertAA(inv, ability)
				err = r.models.Inventories.UpdateAnyAbilities(inv)
				if err != nil {
					log.Println(err)
					return discord.AlexError(ctx)
				}
			}
			err = inventory.UpdateInventoryMessage(ctx, inv)
			if err != nil {
				log.Println(err)
				return discord.AlexError(ctx)
			}
		}
		foot := &discordgo.MessageEmbedFooter{}
		if isAdding {
			foot.Text = fmt.Sprintf("Added to %s's inventory", user.UserValue(ctx).Username)
		}
		return ctx.RespondEmbed(&discordgo.MessageEmbed{
			Title:       fmt.Sprintf("Got Ability %s (%s)", ability.Name, luckType),
			Description: ability.Description,
			Footer:      foot,
		})
	}

	return discord.ErrorMessage(ctx, "Failed to get category", "Alex is a bad programmer")
}

func (r *Roll) luckDebug(ctx ken.SubCommandContext) (err error) {
	if !discord.IsAdminRole(ctx, discord.AdminRoles...) {
		return discord.ErrorMessage(
			ctx,
			"You do not have permission to use this command",
			fmt.Sprintf(
				"You must have one of the following roles: %v",
				strings.Join(discord.AdminRoles, ","),
			),
		)
	}

	rng := rand.Float64()
	level := ctx.Options().GetByName("level").IntValue()
	view := tableView(float64(level), rng)

	return ctx.RespondMessage(view)
}

func (r *Roll) luckTable(ctx ken.SubCommandContext) (err error) {
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

func (r *Roll) Initialize(m data.Models, s *cron.BetrayalScheduler) {
	r.models = m
	r.scheduler = s
}

// Helper to get a random ability. If it is not an any ability, need to check to
// make sure that the ability is the same as the user's current class roll
func (r *Roll) getRandomAnyAbility(role string, rarity string, rec ...int) (*data.AnyAbility, error) {
	// lil saftey net to prevent infinite recursion (hopefully)
	rec = append(rec, 1)
	if len(rec) > 0 && rec[0] > 10 {
		return nil, errors.New("too many attempts to get ability")
	}
	ab, err := r.models.Abilities.GetRandomAnyAbilityByRarity(rarity)
	if err != nil {
		return nil, err
	}

	if ab.RoleSpecific != "" {
		if !strings.EqualFold(ab.RoleSpecific, role) {
			// FIXME: Every time a recursive call is made an angel loses its wings
			return r.getRandomAnyAbility(role, rarity, rec[0]+1)
		}
	}
	return ab, nil
}

// Will roll for random item excluding uniques
func (r *Roll) getRandomItem(rarity string) (*data.Item, error) {
	item, err := r.models.Items.GetRandomByRarity(rarity)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	if item.Rarity == "Unique" {
		return r.getRandomItem(rarity)
	}

	return item, nil
}

func (r *Roll) wheel(ctx ken.SubCommandContext) (err error) {
	// Def need to defer as this will 100% take longer than 3 seconds to respond
	if err = ctx.Defer(); err != nil {
		log.Println(err)
		return err
	}

	s := ctx.GetSession()
	e := ctx.GetEvent()
	events := []string{
		"Sunder",
		"Lawful",
		"Sunder Lawful",
		"Everyone Gets a Doggo",
		"Random 6pb",
		"Duels",
		"Everyone gets 3k",
		"Random polymorph",
		"Votes are public 24hr",
		"Actions are public 24hr",
		"Random Zingy",
		"Random revival",
		"Random role swap",
		"Dimensional shatter",
		"Random Russian revolver present",
		"Care package present",
		"Double event to next roll",
		"RPS event",
		"Coin bonuses randomized",
		"Remove negative statuses from everyone",
		"Everyone is drunk",
		"Jury vote determines game winner",
		"Game winner is determined by the wheel",
		"Host quiz",
		"Everyone can only use gifs/emojis for 6 hours",
		"Everyone is made mad as a random role",
		"Host choice",
		"Random mythical item for all",
		"Random legendary AA for all",
		"Someone explodes",
		"Graveyard and living switch places",
		"Two people revive",
		"oops all villagers",
		"All good roles get elim immunity",
		"All neut roles get elim immunity",
		"All evil roles get elim immunity",
		"Shotgun present",
		"Two players explode",
		"Three players randomly bent",
		"Everyone can pick one AA to get",
	}
	// Send Placeholder message
	base := fmt.Sprintf("%s Spinning the wheel...", discord.EmojiRoll)
	tempMsg, err := s.ChannelMessageSend(e.ChannelID, base)
	if err != nil {
		log.Println(err)
	}

	// precalculate rolls to reduce delay
	rolls := []int{}
	for i := 0; i < 7; i++ {
		rolls = append(rolls, rand.Intn(len(events)))
	}

	for i := 0; i < 7; i++ {
		time.Sleep(450 * time.Millisecond)
		// increase delay by 100 each iteration
		event := events[rolls[i]]
		_, err = s.ChannelMessageEdit(e.ChannelID, tempMsg.ID, fmt.Sprintf("%s %s", base, event))
		if err != nil {
			log.Println(err)
			return discord.AlexError(ctx)
		}
	}
	final := events[rand.Intn(len(events))]
	// delet tempMsg
	err = s.ChannelMessageDelete(e.ChannelID, tempMsg.ID)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx)
	}
	return discord.SuccessfulMessage(ctx, fmt.Sprintf("%s Event Rolled!", final), "(use /list events to see all events)")
}
