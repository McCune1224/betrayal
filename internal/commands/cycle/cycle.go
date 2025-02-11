package cycle

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/mccune1224/betrayal/internal/models"
	"github.com/mccune1224/betrayal/internal/util"
	"github.com/zekrotja/ken"
)

type Cycle struct {
	dbPool *pgxpool.Pool
}

var _ ken.SlashCommand = (*Cycle)(nil)

func (c *Cycle) Initialize(pool *pgxpool.Pool) {
	c.dbPool = pool
}

// Description implements ken.SlashCommand.
func (c *Cycle) Description() string {
	return "Message the new day/night cycle to player's confessionals, each alliance chat, and funnel channels"
}

// Name implements ken.SlashCommand.
func (c *Cycle) Name() string {
	return "cycle"
}

// Options implements ken.SlashCommand.
func (c *Cycle) Options() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Description: "Get current cycle / phase of the game",
			Name:        "current",
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Description: "move to the next phase. (i.e Day 3 -> Elimination 3, Elimination 3 -> Day 4...)",
			Name:        "next",
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Description: "Hard set the cycle of the game",
			Name:        "set",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "phase",
					Description: "choose phase day or elimination",
					Required:    true,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{Name: "Day", Value: "Day"},
						{Name: "Elimination", Value: "Elimination"},
					},
				},
				discord.IntCommandArg("number", "i.e Day [# here]", true),
			},
		},
	}
}

// Version implements ken.SlashCommand.
func (c *Cycle) Version() string {
	return "1.0.0"
}

// Run implements ken.SlashCommand.
func (c *Cycle) Run(ctx ken.Context) (err error) {
	return ctx.HandleSubCommands(
		ken.SubCommandHandler{Name: "next", Run: c.next},
		ken.SubCommandHandler{Name: "set", Run: c.set},
		ken.SubCommandHandler{Name: "current", Run: c.current})
}

func (c *Cycle) current(ctx ken.SubCommandContext) error {
	if err := ctx.Defer(); err != nil {
		log.Println(err)
		return err
	}
	q := models.New(c.dbPool)
	dbCtx := context.Background()
	currCycle, err := q.GetCycle(dbCtx)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Unable to get current game cycle")
	}

	msg := ""
	if currCycle.IsElimination {
		msg = "Elimination"
	} else {
		msg = "Day"
	}
	return ctx.RespondEmbed(&discordgo.MessageEmbed{
		Title: "Current Game Phase",
		Fields: []*discordgo.MessageEmbedField{
			{Name: msg, Value: strconv.Itoa(int(currCycle.Day))},
		},
	})
}

func (c *Cycle) set(ctx ken.SubCommandContext) error {
	if err := ctx.Defer(); err != nil {
		log.Println(err)
		return err
	}
	if !discord.IsAdminRole(ctx, discord.AdminRoles...) {
		return discord.NotAdminError(ctx)
	}
	phaseNameOpt := ctx.Options().GetByName("phase").StringValue()
	phaseNumberOpt := ctx.Options().GetByName("number").IntValue()
	channels, err := c.getCycleChannelIDs(ctx.GetSession(), ctx.GetEvent())
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Failed to get channels for cycle messages")
	}
	q := models.New(c.dbPool)
	dbCtx := context.Background()
	currCycle, err := q.GetCycle(dbCtx)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Failed to get current cycle")
	}
	isElimination := phaseNameOpt == "Elimination"
	updatedCycle, err := q.UpdateCycle(dbCtx, models.UpdateCycleParams{
		ID:            currCycle.ID,
		Day:           int32(phaseNumberOpt),
		IsElimination: isElimination,
	})

	if err != nil {
		log.Println(err)
		discord.AlexError(ctx, "Failed to set cycle")
	}

	for _, channelID := range channels {
		_, err := ctx.GetSession().ChannelMessageSend(channelID, formatCycleMessage(updatedCycle))
		if err != nil {
			log.Println(err)
			return discord.AlexError(ctx, err.Error())
		}
	}

	return discord.SuccessfulMessage(ctx, "Next Cycle messages posted", "")
}

type confessionalChannelDetails struct {
	player  *discordgo.Member
	channel *discordgo.Channel
}

func (c *Cycle) next(ctx ken.SubCommandContext) error {
	if err := ctx.Defer(); err != nil {
		log.Println(err)
		return err
	}
	if !discord.IsAdminRole(ctx, discord.AdminRoles...) {
		return discord.NotAdminError(ctx)
	}

	sesh := ctx.GetSession()
	channelIDSendList, err := c.getCycleChannelIDs(sesh, ctx.GetEvent())
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Failed to get channels for cycle update messages")
	}

	updatedCycle, err := c.incrementCycle()
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Failed to update game cycle")
	}

	msg := formatCycleMessage(updatedCycle)

	for _, channelID := range channelIDSendList {
		_, err := sesh.ChannelMessageSend(channelID, msg)
		if err != nil {
			log.Println(err)
			return discord.AlexError(ctx, err.Error())
		}
	}

	return discord.SuccessfulMessage(ctx, "Next Cycle messages posted", "")
}

// All channels that need to be notified of the next phase of the game.
// This includes confessionals, funnel, and alliance channels
func (c *Cycle) getCycleChannelIDs(sesh *discordgo.Session, event *discordgo.InteractionCreate) ([]string, error) {
	channels := []string{}

	q := models.New(c.dbPool)
	dbCtx := context.Background()
	playerConfessionals, err := q.ListPlayerConfessional(dbCtx)
	if err != nil {
		return channels, err
	}
	playerChannelIds := []string{}
	for _, conf := range playerConfessionals {
		channel, err := sesh.Channel(util.Itoa64(conf.ChannelID))
		if err != nil {
			return channels, fmt.Errorf("Unable to get confessional channel %s", util.Itoa64(conf.ChannelID))
		}
		playerChannelIds = append(playerChannelIds, channel.ID)
	}

	actionChannelID, err := q.GetActionChannel(dbCtx)
	if err != nil {
		return channels, err
	}
	voteChannelID, err := q.GetVoteChannel(dbCtx)
	if err != nil {
		return channels, err
	}
	allianceChannels, err := discord.GetChannelsWithinCategory(sesh, event, "alliances")
	allianceChannelIDs := []string{}
	for _, ch := range *allianceChannels {
		allianceChannelIDs = append(allianceChannelIDs, ch.ID)
	}
	channels = []string{voteChannelID, actionChannelID}
	channels = append(channels, playerChannelIds...)
	channels = append(channels, allianceChannelIDs...)
	return channels, nil
}

// Wrapper for models.UpdateCycle
func (c *Cycle) incrementCycle() (models.GameCycle, error) {
	q := models.New(c.dbPool)
	dbCtx := context.Background()

	currCycle, err := q.GetCycle(dbCtx)
	if err != nil {
		return currCycle, err
	}

	if currCycle.Day == 0 {
		return q.UpdateCycle(dbCtx, models.UpdateCycleParams{
			IsElimination: false,
			Day:           1,
			ID:            currCycle.ID,
		})
	}

	if currCycle.IsElimination {
		return q.UpdateCycle(dbCtx, models.UpdateCycleParams{
			IsElimination: false,
			Day:           currCycle.Day + 1,
			ID:            currCycle.ID,
		})
	}

	return q.UpdateCycle(dbCtx, models.UpdateCycleParams{
		IsElimination: true,
		Day:           currCycle.Day,
		ID:            currCycle.ID,
	})
}

func formatCycleMessage(updatedCycle models.GameCycle) string {
	// Handle elimination cycle
	if updatedCycle.IsElimination {
		return fmt.Sprintf("`=== END OF DAY %d, START OF ELIMINATION %d ===`",
			updatedCycle.Day, updatedCycle.Day)
	}

	// Handle day 0 transition
	if updatedCycle.Day-1 == 0 {
		return fmt.Sprintf("`=== END OF DAY 0, START OF DAY %d ===`",
			updatedCycle.Day)
	}

	// Handle regular elimination to day transition
	return fmt.Sprintf("`=== END OF ELIMINATION %d, START OF DAY %d ===`",
		updatedCycle.Day-1, updatedCycle.Day)
}
