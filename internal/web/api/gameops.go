package api

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/mccune1224/betrayal/internal/models"
	"github.com/mccune1224/betrayal/internal/util"
)

// CycleDTO is the public representation of the current game phase.
type CycleDTO struct {
	Day           int    `json:"day"`
	Phase         string `json:"phase"`
	IsElimination bool   `json:"is_elimination"`
}

// CycleHandler exposes read-only cycle state.
type CycleHandler struct{ pool *pgxpool.Pool }

func NewCycleHandler(pool *pgxpool.Pool) *CycleHandler { return &CycleHandler{pool: pool} }

func (h *CycleHandler) Get(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()
	cycle, err := models.New(h.pool).GetCycle(ctx)
	if err != nil {
		WriteError(c.Response(), http.StatusInternalServerError, "cycle_unavailable", "could not load cycle", nil)
		return nil
	}
	WriteJSON(c.Response(), http.StatusOK, cycleDTO(cycle))
	return nil
}

func cycleDTO(cycle models.GameCycle) CycleDTO {
	phase := "Day"
	if cycle.IsElimination {
		phase = "Elimination"
	}
	return CycleDTO{Day: int(cycle.Day), Phase: phase, IsElimination: cycle.IsElimination}
}

type ChannelEntryDTO struct {
	Name      string `json:"name"`
	Kind      string `json:"kind"`
	ChannelID string `json:"channel_id,omitempty"`
	Status    string `json:"status"`
	Note      string `json:"note,omitempty"`
}
type ChannelSummaryDTO struct {
	Total      int `json:"total"`
	Configured int `json:"configured"`
	Missing    int `json:"missing"`
	Orphaned   int `json:"orphaned"`
	Unverified int `json:"unverified"`
}
type ChannelsDTO struct {
	DiscordConnected bool              `json:"discord_connected"`
	Entries          []ChannelEntryDTO `json:"entries"`
	Summary          ChannelSummaryDTO `json:"summary"`
}

type ChannelsHandler struct {
	pool    *pgxpool.Pool
	discord *discordgo.Session
}

func NewChannelsHandler(pool *pgxpool.Pool, discord *discordgo.Session) *ChannelsHandler {
	return &ChannelsHandler{pool: pool, discord: discord}
}

func (h *ChannelsHandler) Get(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()
	q := models.New(h.pool)
	entries := make([]ChannelEntryDTO, 0, 8)
	addSingleton := func(name, kind string, load func(context.Context) (string, error)) {
		id, err := load(ctx)
		if err != nil {
			entries = append(entries, ChannelEntryDTO{Name: name, Kind: kind, Status: "missing", Note: "Not configured"})
			return
		}
		entries = append(entries, h.channelEntry(name, kind, id))
	}
	addSingleton("Vote Channel", "vote", q.GetVoteChannel)
	addSingleton("Action Channel", "action", q.GetActionChannel)
	addSingleton("Command Log Channel", "log", q.GetCommandLogChannel)
	addSingleton("Lifeboard Channel", "lifeboard", func(ctx context.Context) (string, error) {
		lb, err := q.GetPlayerLifeboard(ctx)
		return lb.ChannelID, err
	})
	admins, err := q.ListAdminChannel(ctx)
	if err == nil && len(admins) > 0 {
		for _, ch := range admins {
			entries = append(entries, h.channelEntry("Admin Channel", "admin", ch))
		}
	} else {
		entries = append(entries, ChannelEntryDTO{Name: "Admin Channels", Kind: "admin", Status: "missing", Note: "No admin channels configured"})
	}
	if confessionals, err := q.ListPlayerConfessional(ctx); err == nil {
		for _, conf := range confessionals {
			entries = append(entries, h.channelEntry("Confessional", "confessional", util.Itoa64(conf.ChannelID)))
		}
	}
	summary := ChannelSummaryDTO{Total: len(entries)}
	for _, entry := range entries {
		switch entry.Status {
		case "configured":
			summary.Configured++
		case "missing":
			summary.Missing++
		case "orphaned":
			summary.Orphaned++
		case "unverified":
			summary.Unverified++
		}
	}
	WriteJSON(c.Response(), http.StatusOK, ChannelsDTO{DiscordConnected: h.discord != nil, Entries: entries, Summary: summary})
	return nil
}

func (h *ChannelsHandler) channelEntry(name, kind, id string) ChannelEntryDTO {
	entry := ChannelEntryDTO{Name: name, Kind: kind, ChannelID: id}
	if h.discord == nil {
		entry.Status, entry.Note = "unverified", "Discord is disabled"
		return entry
	}
	if _, err := h.discord.Channel(id); err != nil {
		entry.Status, entry.Note = "orphaned", "Discord channel no longer exists"
		return entry
	}
	entry.Status = "configured"
	return entry
}

type VoteDTO struct {
	ID        int32      `json:"id"`
	VoterID   int64      `json:"voter_id"`
	TargetID  int64      `json:"target_id"`
	Weight    int32      `json:"weight"`
	Context   string     `json:"context,omitempty"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}
type VoteTallyDTO struct {
	TargetID   int64 `json:"target_id"`
	TotalVotes int32 `json:"total_votes"`
	VoteCount  int64 `json:"vote_count"`
}
type VoteCycleDTO struct {
	Day           int    `json:"day"`
	Phase         string `json:"phase"`
	IsElimination bool   `json:"is_elimination"`
	IsCurrent     bool   `json:"is_current"`
}
type VoteCycleOptionDTO struct {
	Day           int    `json:"day"`
	IsElimination bool   `json:"is_elimination"`
	Label         string `json:"label"`
}
type VotePlayerStatDTO struct {
	PlayerID int64 `json:"player_id"`
	Count    int   `json:"count"`
}
type VoteStatsDTO struct {
	MostVotedPlayers  []VotePlayerStatDTO `json:"most_voted_players"`
	MostActiveVoters  []VotePlayerStatDTO `json:"most_active_voters"`
	LeastVotedPlayers []VotePlayerStatDTO `json:"least_voted_players"`
}
type VotesDTO struct {
	Cycle      VoteCycleDTO         `json:"cycle"`
	Votes      []VoteDTO            `json:"votes"`
	Tallies    []VoteTallyDTO       `json:"tallies"`
	Cycles     []VoteCycleOptionDTO `json:"cycles"`
	Stats      VoteStatsDTO         `json:"stats"`
	TotalVotes int                  `json:"total_votes"`
}

type VotesHandler struct{ pool *pgxpool.Pool }

func NewVotesHandler(pool *pgxpool.Pool) *VotesHandler { return &VotesHandler{pool: pool} }
func (h *VotesHandler) Get(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()
	q := models.New(h.pool)
	current, err := q.GetCycle(ctx)
	if err != nil {
		WriteError(c.Response(), 500, "votes_unavailable", "could not load cycle", nil)
		return nil
	}
	day, elim := int(current.Day), current.IsElimination
	if value := c.QueryParam("day"); value != "" {
		if parsed, parseErr := strconv.Atoi(value); parseErr == nil {
			day = parsed
		}
	}
	if value := c.QueryParam("elimination"); value != "" {
		elim = value == "true"
	}
	params := models.ListVotesByCycleParams{CycleDay: int32(day), IsElimination: elim}
	votes, err := q.ListVotesByCycle(ctx, params)
	if err != nil {
		WriteError(c.Response(), 500, "votes_unavailable", "could not load votes", nil)
		return nil
	}
	tallies, err := q.GetVoteTalliesByCycle(ctx, models.GetVoteTalliesByCycleParams{CycleDay: int32(day), IsElimination: elim})
	if err != nil {
		WriteError(c.Response(), 500, "votes_unavailable", "could not load vote tallies", nil)
		return nil
	}
	cycles, err := q.GetDistinctCyclesWithVotes(ctx)
	if err != nil {
		WriteError(c.Response(), 500, "votes_unavailable", "could not load vote history", nil)
		return nil
	}
	stats, err := q.GetVoteStatsByPlayer(ctx)
	if err != nil {
		WriteError(c.Response(), 500, "votes_unavailable", "could not load vote stats", nil)
		return nil
	}
	participation, err := q.GetVoterParticipation(ctx)
	if err != nil {
		WriteError(c.Response(), 500, "votes_unavailable", "could not load voter participation", nil)
		return nil
	}
	voteDTOs := make([]VoteDTO, len(votes))
	for i, vote := range votes {
		voteDTOs[i] = VoteDTO{ID: vote.ID, VoterID: vote.VoterID, TargetID: vote.TargetID, Weight: vote.Weight, Context: nullableText(vote.Context), UpdatedAt: nullableTime(vote.UpdatedAt)}
	}
	tallyDTOs := make([]VoteTallyDTO, len(tallies))
	for i, tally := range tallies {
		tallyDTOs[i] = VoteTallyDTO{TargetID: tally.TargetID, TotalVotes: tally.TotalVotes, VoteCount: tally.VoteCount}
	}
	cycleDTOs := make([]VoteCycleOptionDTO, len(cycles))
	for i, item := range cycles {
		phase := "Day"
		if item.IsElimination {
			phase = "Elimination"
		}
		cycleDTOs[i] = VoteCycleOptionDTO{Day: int(item.CycleDay), IsElimination: item.IsElimination, Label: phase + " " + strconv.Itoa(int(item.CycleDay))}
	}
	statsDTO := VoteStatsDTO{MostVotedPlayers: playerVoteStats(stats), MostActiveVoters: voterStats(participation)}
	if len(stats) > 0 {
		for i := len(stats) - 1; i >= 0 && len(statsDTO.LeastVotedPlayers) < 5; i-- {
			statsDTO.LeastVotedPlayers = append(statsDTO.LeastVotedPlayers, VotePlayerStatDTO{PlayerID: stats[i].TargetID, Count: int(stats[i].TotalVotesReceived)})
		}
	}
	phase := "Day"
	if elim {
		phase = "Elimination"
	}
	WriteJSON(c.Response(), 200, VotesDTO{Cycle: VoteCycleDTO{Day: day, Phase: phase, IsElimination: elim, IsCurrent: day == int(current.Day) && elim == current.IsElimination}, Votes: voteDTOs, Tallies: tallyDTOs, Cycles: cycleDTOs, Stats: statsDTO, TotalVotes: len(voteDTOs)})
	return nil
}
func nullableText(value pgtype.Text) string {
	if value.Valid {
		return value.String
	}
	return ""
}
func nullableTime(value pgtype.Timestamp) *time.Time {
	if value.Valid {
		result := value.Time
		return &result
	}
	return nil
}
func playerVoteStats(rows []models.GetVoteStatsByPlayerRow) []VotePlayerStatDTO {
	result := make([]VotePlayerStatDTO, len(rows))
	for i, row := range rows {
		result[i] = VotePlayerStatDTO{PlayerID: row.TargetID, Count: int(row.TotalVotesReceived)}
	}
	return result
}
func voterStats(rows []models.GetVoterParticipationRow) []VotePlayerStatDTO {
	result := make([]VotePlayerStatDTO, len(rows))
	for i, row := range rows {
		result[i] = VotePlayerStatDTO{PlayerID: row.VoterID, Count: int(row.VotesCast)}
	}
	return result
}

type ReadinessDTO struct {
	Ready            bool `json:"ready"`
	DiscordConnected bool `json:"discord_connected"`
	Channels         struct {
		Admin          int    `json:"admin"`
		AdminReady     bool   `json:"admin_ready"`
		Vote           string `json:"vote"`
		VoteReady      bool   `json:"vote_ready"`
		Action         string `json:"action"`
		ActionReady    bool   `json:"action_ready"`
		LifeboardReady bool   `json:"lifeboard_ready"`
	} `json:"channels"`
	Players struct {
		Total int `json:"total"`
		Alive int `json:"alive"`
		Dead  int `json:"dead"`
	} `json:"players"`
	Cycle struct {
		Ready bool   `json:"ready"`
		Day   int    `json:"day"`
		Phase string `json:"phase"`
	} `json:"cycle"`
}
type ReadinessHandler struct {
	pool    *pgxpool.Pool
	discord *discordgo.Session
}

func NewReadinessHandler(pool *pgxpool.Pool, discord *discordgo.Session) *ReadinessHandler {
	return &ReadinessHandler{pool: pool, discord: discord}
}
func (h *ReadinessHandler) Get(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()
	q := models.New(h.pool)
	data := ReadinessDTO{DiscordConnected: h.discord != nil}
	admins, adminErr := q.ListAdminChannel(ctx)
	data.Channels.Admin = len(admins)
	data.Channels.AdminReady = adminErr == nil && len(admins) > 0
	data.Channels.Vote, _ = q.GetVoteChannel(ctx)
	data.Channels.VoteReady = data.Channels.Vote != ""
	data.Channels.Action, _ = q.GetActionChannel(ctx)
	data.Channels.ActionReady = data.Channels.Action != ""
	if lifeboard, err := q.GetPlayerLifeboard(ctx); err == nil {
		data.Channels.LifeboardReady = lifeboard.ChannelID != ""
	}
	players, _ := q.ListPlayer(ctx)
	data.Players.Total = len(players)
	for _, player := range players {
		if player.Alive {
			data.Players.Alive++
		} else {
			data.Players.Dead++
		}
	}
	if cycle, err := q.GetCycle(ctx); err == nil {
		data.Cycle.Ready = true
		data.Cycle.Day = int(cycle.Day)
		data.Cycle.Phase = cycleDTO(cycle).Phase
	}
	data.Ready = data.Channels.AdminReady && data.Channels.VoteReady && data.Channels.ActionReady && data.Players.Total > 0
	WriteJSON(c.Response(), 200, data)
	return nil
}
