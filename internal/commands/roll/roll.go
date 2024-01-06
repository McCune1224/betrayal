package roll

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"slices"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/mccune1224/betrayal/internal/commands"
	"github.com/mccune1224/betrayal/internal/commands/inventory"
	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/mccune1224/betrayal/internal/scheduler"
	"github.com/mccune1224/betrayal/pkg/data"
	"github.com/zekrotja/ken"
)

type Roll struct {
	models    data.Models
	scheduler *scheduler.BetrayalScheduler
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
			Name:  r,
			Value: r,
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
				discord.UserCommandArg(true),
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
			Name:        "wheel",
			Description: "Spin the wheel for a game event",
		},
	}
}

// Run implements ken.SlashCommand.
func (r *Roll) Run(ctx ken.Context) (err error) {
	return ctx.HandleSubCommands(
		ken.SubCommandHandler{Name: "manual", Run: r.luckManual},
		ken.SubCommandHandler{Name: "rarity", Run: r.rollByMinimumRarity},
		ken.SubCommandHandler{Name: "care_package", Run: r.luckCarePackage},
		// ken.SubCommandHandler{Name: "table", Run: r.luckTable},
		ken.SubCommandHandler{Name: "item_rain", Run: r.luckItemRain},
		ken.SubCommandHandler{Name: "power_drop", Run: r.luckPowerDrop},
		ken.SubCommandHandler{Name: "wheel", Run: r.wheel},
	)
}

func (r *Roll) luckManual(ctx ken.SubCommandContext) (err error) {
	if err := ctx.Defer(); err != nil {
		log.Println(err)
		return err
	}

	handler, err := inventory.FetchHandler(ctx, r.models, true)
	if err != nil {
		if errors.Is(err, inventory.ErrNotAuthorized) {
			return discord.NotAdminError(ctx)
		}
		return discord.ErrorMessage(ctx, "Failed to find inventory.", "If not in confessional, please specify a user")
	}

	opts := ctx.Options()
	target := opts.GetByName("target").StringValue()
	level := opts.GetByName("luck").IntValue()

	rng := rand.Float64()
	luckType := RollLuck(float64(level), rng)

	if target == "item" {
		item, err := r.getRandomItem(luckType)
		if err != nil {
			log.Println(err)
			return discord.AlexError(ctx, "Failed to get random item")
		}

		return ctx.RespondEmbed(&discordgo.MessageEmbed{
			Title:       fmt.Sprintf("Got Item %s (%s)", item.Name, luckType),
			Description: item.Description,
		})
	}

	if target == "aa" {
		ability, err := r.getRandomAnyAbility(handler.GetInventory().RoleName, luckType)
		if err != nil {
			log.Println(err)
			return discord.AlexError(ctx, "Failed to get random ability")
		}

		return ctx.RespondEmbed(&discordgo.MessageEmbed{
			Title:       fmt.Sprintf("Got Ability %s (%s)", ability.Name, luckType),
			Description: ability.Description,
		})
	}

	return discord.ErrorMessage(ctx, "Failed to get category", "Alex is a bad programmer")
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

func (r *Roll) Initialize(m data.Models, s *scheduler.BetrayalScheduler) {
	r.models = m
	r.scheduler = s
}

func (r *Roll) getRandomAnyAbility(role string, rarity string) (*data.AnyAbility, error) {
	// lil saftey net to prevent infinite recursion (hopefully)
	ab, err := r.models.Abilities.GetRandomAnyAbilityByRarity(rarity)
	if err != nil {
		return nil, err
	}
	if ab.RoleSpecific != "" && !strings.EqualFold(ab.RoleSpecific, role) {
		return r.getRandomAnyAbility(role, rarity)
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
	if err = ctx.Defer(); err != nil {
		log.Println(err)
		return err
	}

	if !discord.IsAdminRole(ctx, discord.AdminRoles...) {
		return discord.NotAdminError(ctx)
	}

	s := ctx.GetSession()
	e := ctx.GetEvent()
	// Send Placeholder message
	base := fmt.Sprintf("%s Spinning the wheel...", discord.EmojiRoll)
	tempMsg, err := s.ChannelMessageSend(e.ChannelID, base)
	if err != nil {
		log.Println(err)
	}

	// precalculate rolls to reduce delay
	rolls := []int{}
	for i := 0; i < 7; i++ {
		rolls = append(rolls, rand.Intn(len(commands.WheelEvents)))
	}

	for i := 0; i < 7; i++ {
		time.Sleep(450 * time.Millisecond)
		// increase delay by 100 each iteration
		event := commands.WheelEvents[rolls[i]]
		_, err = s.ChannelMessageEdit(e.ChannelID, tempMsg.ID, fmt.Sprintf("%s %s", base, event))
		if err != nil {
			log.Println(err)
			return discord.AlexError(ctx, "Failed to edit message")
		}
	}
	final := commands.WheelEvents[rand.Intn(len(commands.WheelEvents))]
	// delet tempMsg
	err = s.ChannelMessageDelete(e.ChannelID, tempMsg.ID)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Failed to delete message")
	}
	return discord.SuccessfulMessage(ctx, fmt.Sprintf("%s Event Rolled!", final), "(use `/list wheel_events` to see all possible random events.)")
}

func (r *Roll) rollByMinimumRarity(ctx ken.SubCommandContext) (err error) {
	if err := ctx.Defer(); err != nil {
		log.Println(err)
		return err
	}

	level := ctx.Options().GetByName("luck").IntValue()
	minimumRarity := ctx.Options().GetByName("min_rarity").StringValue()
	target := ctx.Options().GetByName("target").StringValue()

	userInv, err := inventory.FetchHandler(ctx, r.models, true)
	if err != nil {
		return discord.ErrorMessage(ctx, "Failed to get user inventory", err.Error())
	}

	start := slices.Index(rarityPriorities, minimumRarity)

	rarityOptions := rarityPriorities[start:]

	rarity := rollAtRarity(float64(level), rarityOptions)

	if target == "item" {
		item, err := r.getRandomItem(rarity)
		if err != nil {
			log.Println(err)
			return discord.AlexError(ctx, "Failed to get random item")
		}
		return ctx.RespondEmbed(&discordgo.MessageEmbed{
			Title:       fmt.Sprintf("Got Item %s (%s)", item.Name, rarity),
			Description: item.Description,
			Footer: &discordgo.MessageEmbedFooter{
				Text: fmt.Sprintf("%s Note, this will not auto add to an inventory.", discord.EmojiWarning),
			},
		})
	} else {
		ability, err := r.getRandomAnyAbility(userInv.GetInventory().RoleName, rarity)
		if err != nil {
			log.Println(err)
			return discord.AlexError(ctx, "Failed to get random ability")
		}
		return ctx.RespondEmbed(&discordgo.MessageEmbed{
			Title:       fmt.Sprintf("Got Ability %s (%s)", ability.Name, rarity),
			Description: ability.Description,

			Footer: &discordgo.MessageEmbedFooter{
				Text: fmt.Sprintf("%s Note, this will not auto add to an inventory.", discord.EmojiWarning),
			},
		})
	}
}
