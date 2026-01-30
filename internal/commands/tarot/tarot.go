package tarot

import (
	"fmt"
	"math/rand"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/mccune1224/betrayal/internal/logger"
	"github.com/mccune1224/betrayal/internal/util"
	"github.com/zekrotja/ken"
)

type Tarot struct {
	dbPool *pgxpool.Pool
}

var _ ken.SlashCommand = (*Tarot)(nil)

func (t *Tarot) Initialize(pool *pgxpool.Pool) { t.dbPool = pool }
func (*Tarot) Name() string                    { return "tarot" }
func (*Tarot) Description() string             { return "Draw a tarot card in various modes" }
func (*Tarot) Version() string                 { return "1.0.0" }

// In-memory state for per-user assignments and guild deck
type assignment struct {
	idx      int
	reversed bool
}

type deckState struct {
	remaining []int
	dealt     map[int]bool
	orient    map[int]bool // index -> reversed?
}

var (
	assignMu        sync.RWMutex
	userAssignments = map[string]map[string]assignment{} // guildID -> userID -> assignment

	deckMu    sync.RWMutex
	guildDeck = map[string]*deckState{} // guildID -> deck
)

func (*Tarot) Options() []*discordgo.ApplicationCommandOption {
	modeChoices := []*discordgo.ApplicationCommandOptionChoice{
		{Name: "deterministic", Value: "deterministic"},
		{Name: "per_user", Value: "per_user"},
		{Name: "guild_deck", Value: "guild_deck"},
		{Name: "random", Value: "random"},
	}

	return []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "draw",
			Description: "Draw a card (default deterministic by guild+user)",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "mode",
					Description: "Draw mode",
					Required:    false,
					Choices:     modeChoices,
				},
				discord.BoolCommandArg("reversed", "Force reversed orientation", false),
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "reset",
			Description: "Reset in-memory tarot state",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "scope",
					Description: "What to reset",
					Required:    true,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{Name: "per_user", Value: "per_user"},
						{Name: "guild_deck", Value: "guild_deck"},
						{Name: "all", Value: "all"},
					},
				},
				discord.UserCommandArg(false), // required when scope=per_user for specific user reset
			},
		},
	}
}

func (t *Tarot) Run(ctx ken.Context) (err error) {
	defer logger.RecoverWithLog(*logger.Get())
	return ctx.HandleSubCommands(
		ken.SubCommandHandler{Name: "draw", Run: t.draw},
		ken.SubCommandHandler{Name: "reset", Run: t.reset},
	)
}

func (t *Tarot) draw(ctx ken.SubCommandContext) error {
	// Fast op; no defer needed
	opts := ctx.Options()
	mode := "deterministic"
	if m, ok := opts.GetByNameOptional("mode"); ok {
		mode = m.StringValue()
	}
	forceRev := false
	if r, ok := opts.GetByNameOptional("reversed"); ok {
		forceRev = r.BoolValue()
	}

	guildID := ctx.GetEvent().GuildID
	userID := ctx.GetEvent().Member.User.ID

	idx := 0
	reversed := false

	switch mode {
	case "deterministic":
		idx, reversed = DeterministicDraw(guildID, userID)
	case "random":
		idx = rand.Intn(len(TarotCards))
		reversed = rand.Intn(2) == 1
	case "per_user":
		idx, reversed = getOrAssignUser(guildID, userID)
	case "guild_deck":
		var ok bool
		idx, reversed, ok = drawFromGuildDeck(guildID)
		if !ok {
			return discord.WarningMessage(ctx, "Deck Exhausted", "All cards have been dealt for this guild. Ask an admin to run /tarot reset scope:guild_deck.")
		}
	default:
		// fallback to deterministic
		idx, reversed = DeterministicDraw(guildID, userID)
	}

	if forceRev {
		reversed = true
	}

	// Bounds check, defensive
	total := len(TarotCards)
	if idx < 0 || idx >= total {
		idx = idx % total
		if idx < 0 {
			idx = 0
		}
	}

	title := TarotCards[idx].Name
	meaning := TarotCards[idx].Upright
	if reversed {
		title += " (Reversed)"
		meaning = TarotCards[idx].Reversed
	}

	foot := fmt.Sprintf("Requested by %s • %s", discord.MentionUser(userID), util.GetEstTimeStamp())
	return discord.SuccessfulMessage(ctx, fmt.Sprintf("%s %s", discord.EmojiAbility, title), meaning, foot)
}

func (t *Tarot) reset(ctx ken.SubCommandContext) error {
	if !discord.IsAdminRole(ctx, discord.AdminRoles...) {
		return discord.NotAdminError(ctx)
	}
	scope := ctx.Options().GetByName("scope").StringValue()
	guildID := ctx.GetEvent().GuildID

	switch scope {
	case "per_user":
		// Optional specific user; if omitted, clear all user assignments in guild
		if uOpt, ok := ctx.Options().GetByNameOptional("user"); ok {
			user := uOpt.UserValue(ctx)
			assignMu.Lock()
			if ua, exists := userAssignments[guildID]; exists {
				delete(ua, user.ID)
				if len(ua) == 0 {
					delete(userAssignments, guildID)
				}
			}
			assignMu.Unlock()
			return discord.SuccessfulMessage(ctx, "Tarot Reset", fmt.Sprintf("Cleared assignment for %s", user.Mention()))
		}
		// Clear all in guild
		assignMu.Lock()
		delete(userAssignments, guildID)
		assignMu.Unlock()
		return discord.SuccessfulMessage(ctx, "Tarot Reset", "Cleared all per-user assignments for this guild")
	case "guild_deck":
		deckMu.Lock()
		delete(guildDeck, guildID)
		deckMu.Unlock()
		return discord.SuccessfulMessage(ctx, "Tarot Reset", "Shuffled and cleared the guild deck; next draw will start fresh")
	case "all":
		assignMu.Lock()
		userAssignments = map[string]map[string]assignment{}
		assignMu.Unlock()
		deckMu.Lock()
		guildDeck = map[string]*deckState{}
		deckMu.Unlock()
		return discord.SuccessfulMessage(ctx, "Tarot Reset", "Cleared all tarot state (per-user and guild deck)")
	default:
		return discord.ErrorMessage(ctx, "Invalid Scope", "Use one of: per_user, guild_deck, all")
	}
}

func getOrAssignUser(guildID, userID string) (idx int, reversed bool) {
	assignMu.RLock()
	if ua, ok := userAssignments[guildID]; ok {
		if a, ok := ua[userID]; ok {
			assignMu.RUnlock()
			return a.idx, a.reversed
		}
	}
	assignMu.RUnlock()

	// Not found; assign
	idx = rand.Intn(len(TarotCards))
	reversed = rand.Intn(2) == 1

	assignMu.Lock()
	if _, ok := userAssignments[guildID]; !ok {
		userAssignments[guildID] = map[string]assignment{}
	}
	userAssignments[guildID][userID] = assignment{idx: idx, reversed: reversed}
	assignMu.Unlock()
	return idx, reversed
}

func drawFromGuildDeck(guildID string) (idx int, reversed bool, ok bool) {
	deckMu.Lock()
	defer deckMu.Unlock()

	ds, exists := guildDeck[guildID]
	if !exists {
		// Initialize fresh deck with all indices
		total := len(TarotCards)
		remaining := make([]int, total)
		for i := 0; i < total; i++ {
			remaining[i] = i
		}
		ds = &deckState{remaining: remaining, dealt: map[int]bool{}, orient: map[int]bool{}}
		guildDeck[guildID] = ds
	}

	if len(ds.remaining) == 0 {
		return 0, false, false
	}

	// Pick random from remaining
	r := rand.Intn(len(ds.remaining))
	idx = ds.remaining[r]
	// Remove from remaining (swap with last)
	last := len(ds.remaining) - 1
	ds.remaining[r], ds.remaining[last] = ds.remaining[last], ds.remaining[r]
	ds.remaining = ds.remaining[:last]
	ds.dealt[idx] = true
	reversed = rand.Intn(2) == 1
	ds.orient[idx] = reversed
	return idx, reversed, true
}
