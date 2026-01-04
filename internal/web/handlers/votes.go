package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/mccune1224/betrayal/internal/models"
	"github.com/mccune1224/betrayal/internal/web/templates/pages"
)

// VotesHandler handles vote-related web requests
type VotesHandler struct {
	dbPool *pgxpool.Pool
}

// NewVotesHandler creates a new VotesHandler
func NewVotesHandler(pool *pgxpool.Pool) *VotesHandler {
	return &VotesHandler{dbPool: pool}
}

// Votes handles GET /votes - main votes page
func (h *VotesHandler) Votes(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()

	q := models.New(h.dbPool)

	// Get current cycle
	cycle, err := q.GetCycle(ctx)
	if err != nil {
		log.Printf("votes: failed to get cycle: %v", err)
		cycle = models.GameCycle{Day: 1, IsElimination: false}
	}

	// Get query params for which cycle to view (default to current)
	viewDay := int(cycle.Day)
	viewElim := cycle.IsElimination

	if dayStr := c.QueryParam("day"); dayStr != "" {
		if d, err := strconv.Atoi(dayStr); err == nil {
			viewDay = d
		}
	}
	if elimStr := c.QueryParam("elimination"); elimStr != "" {
		viewElim = elimStr == "true"
	}

	// Get votes for the selected cycle
	votes, err := q.ListVotesByCycle(ctx, models.ListVotesByCycleParams{
		CycleDay:      int32(viewDay),
		IsElimination: viewElim,
	})
	if err != nil {
		log.Printf("votes: failed to get votes: %v", err)
		votes = []models.Vote{}
	}

	// Get vote tallies for the selected cycle
	tallies, err := q.GetVoteTalliesByCycle(ctx, models.GetVoteTalliesByCycleParams{
		CycleDay:      int32(viewDay),
		IsElimination: viewElim,
	})
	if err != nil {
		log.Printf("votes: failed to get tallies: %v", err)
		tallies = []models.GetVoteTalliesByCycleRow{}
	}

	// Get all cycles that have votes (for history dropdown)
	cyclesWithVotes, err := q.GetDistinctCyclesWithVotes(ctx)
	if err != nil {
		log.Printf("votes: failed to get cycles with votes: %v", err)
		cyclesWithVotes = []models.GetDistinctCyclesWithVotesRow{}
	}

	// Get overall stats
	voteStats, err := q.GetVoteStatsByPlayer(ctx)
	if err != nil {
		log.Printf("votes: failed to get vote stats: %v", err)
		voteStats = []models.GetVoteStatsByPlayerRow{}
	}

	voterParticipation, err := q.GetVoterParticipation(ctx)
	if err != nil {
		log.Printf("votes: failed to get voter participation: %v", err)
		voterParticipation = []models.GetVoterParticipationRow{}
	}

	// Build view models
	voteRows := h.buildVoteRows(ctx, q, votes)
	tallyRows := h.buildTallyRows(ctx, q, tallies)
	cycleOptions := h.buildCycleOptions(cyclesWithVotes, viewDay, viewElim)
	statsData := h.buildStatsData(ctx, q, voteStats, voterParticipation)

	// Determine phase name
	phaseName := "Day"
	if viewElim {
		phaseName = "Elimination"
	}

	data := pages.VotesData{
		CurrentPhase:    phaseName,
		CurrentDay:      viewDay,
		IsCurrentCycle:  viewDay == int(cycle.Day) && viewElim == cycle.IsElimination,
		Votes:           voteRows,
		Tallies:         tallyRows,
		CycleOptions:    cycleOptions,
		Stats:           statsData,
		TotalVotesCount: len(votes),
	}

	return render(c, http.StatusOK, pages.Votes(data))
}

// VoteTally handles GET /votes/tally - HTMX partial for vote tally
func (h *VotesHandler) VoteTally(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()

	q := models.New(h.dbPool)

	// Get cycle from query params
	dayStr := c.QueryParam("day")
	elimStr := c.QueryParam("elimination")

	day := 1
	if d, err := strconv.Atoi(dayStr); err == nil {
		day = d
	}
	isElim := elimStr == "true"

	tallies, err := q.GetVoteTalliesByCycle(ctx, models.GetVoteTalliesByCycleParams{
		CycleDay:      int32(day),
		IsElimination: isElim,
	})
	if err != nil {
		tallies = []models.GetVoteTalliesByCycleRow{}
	}

	tallyRows := h.buildTallyRows(ctx, q, tallies)
	return render(c, http.StatusOK, pages.VoteTallyPartial(tallyRows))
}

func (h *VotesHandler) buildVoteRows(ctx context.Context, q *models.Queries, votes []models.Vote) []pages.VoteRow {
	rows := make([]pages.VoteRow, len(votes))
	for i, v := range votes {
		voterName := formatPlayerID(v.VoterID)
		targetName := formatPlayerID(v.TargetID)

		// Try to get role names for more context
		if player, err := q.GetPlayer(ctx, v.VoterID); err == nil && player.RoleID.Valid {
			if role, err := q.GetRole(ctx, player.RoleID.Int32); err == nil {
				voterName = fmt.Sprintf("%s (%s)", voterName, role.Name)
			}
		}
		if player, err := q.GetPlayer(ctx, v.TargetID); err == nil && player.RoleID.Valid {
			if role, err := q.GetRole(ctx, player.RoleID.Int32); err == nil {
				targetName = fmt.Sprintf("%s (%s)", targetName, role.Name)
			}
		}

		context := ""
		if v.Context.Valid {
			context = v.Context.String
		}

		updatedAt := time.Time{}
		if v.UpdatedAt.Valid {
			updatedAt = v.UpdatedAt.Time
		}

		rows[i] = pages.VoteRow{
			ID:         int(v.ID),
			VoterID:    v.VoterID,
			VoterName:  voterName,
			TargetID:   v.TargetID,
			TargetName: targetName,
			Weight:     int(v.Weight),
			Context:    context,
			UpdatedAt:  updatedAt,
		}
	}
	return rows
}

func (h *VotesHandler) buildTallyRows(ctx context.Context, q *models.Queries, tallies []models.GetVoteTalliesByCycleRow) []pages.TallyRow {
	rows := make([]pages.TallyRow, len(tallies))
	for i, t := range tallies {
		targetName := formatPlayerID(t.TargetID)
		roleName := ""
		alive := true

		if player, err := q.GetPlayer(ctx, t.TargetID); err == nil {
			alive = player.Alive
			if player.RoleID.Valid {
				if role, err := q.GetRole(ctx, player.RoleID.Int32); err == nil {
					roleName = role.Name
				}
			}
		}

		rows[i] = pages.TallyRow{
			TargetID:   t.TargetID,
			TargetName: targetName,
			RoleName:   roleName,
			TotalVotes: int(t.TotalVotes),
			VoteCount:  int(t.VoteCount),
			IsAlive:    alive,
			Rank:       i + 1,
		}
	}
	return rows
}

func (h *VotesHandler) buildCycleOptions(cycles []models.GetDistinctCyclesWithVotesRow, currentDay int, currentElim bool) []pages.CycleOption {
	options := make([]pages.CycleOption, len(cycles))
	for i, c := range cycles {
		phaseName := "Day"
		if c.IsElimination {
			phaseName = "Elimination"
		}
		options[i] = pages.CycleOption{
			Day:           int(c.CycleDay),
			IsElimination: c.IsElimination,
			Label:         fmt.Sprintf("%s %d", phaseName, c.CycleDay),
			IsSelected:    int(c.CycleDay) == currentDay && c.IsElimination == currentElim,
		}
	}
	return options
}

func (h *VotesHandler) buildStatsData(ctx context.Context, q *models.Queries, voteStats []models.GetVoteStatsByPlayerRow, participation []models.GetVoterParticipationRow) *pages.VoteStatsData {
	if len(voteStats) == 0 && len(participation) == 0 {
		return nil
	}

	stats := &pages.VoteStatsData{
		MostVotedPlayers:  make([]pages.PlayerVoteStat, 0),
		MostActiveVoters:  make([]pages.PlayerVoteStat, 0),
		LeastVotedPlayers: make([]pages.PlayerVoteStat, 0),
	}

	// Most voted (top 5)
	for i, vs := range voteStats {
		if i >= 5 {
			break
		}
		name := formatPlayerID(vs.TargetID)
		if player, err := q.GetPlayer(ctx, vs.TargetID); err == nil && player.RoleID.Valid {
			if role, err := q.GetRole(ctx, player.RoleID.Int32); err == nil {
				name = fmt.Sprintf("%s (%s)", name, role.Name)
			}
		}
		stats.MostVotedPlayers = append(stats.MostVotedPlayers, pages.PlayerVoteStat{
			PlayerID:   vs.TargetID,
			PlayerName: name,
			Count:      int(vs.TotalVotesReceived),
		})
	}

	// Least voted (bottom 5, but only if they have votes)
	if len(voteStats) > 0 {
		startIdx := len(voteStats) - 5
		if startIdx < 0 {
			startIdx = 0
		}
		for i := len(voteStats) - 1; i >= startIdx; i-- {
			vs := voteStats[i]
			name := formatPlayerID(vs.TargetID)
			if player, err := q.GetPlayer(ctx, vs.TargetID); err == nil && player.RoleID.Valid {
				if role, err := q.GetRole(ctx, player.RoleID.Int32); err == nil {
					name = fmt.Sprintf("%s (%s)", name, role.Name)
				}
			}
			stats.LeastVotedPlayers = append(stats.LeastVotedPlayers, pages.PlayerVoteStat{
				PlayerID:   vs.TargetID,
				PlayerName: name,
				Count:      int(vs.TotalVotesReceived),
			})
		}
	}

	// Most active voters (top 5)
	for i, vp := range participation {
		if i >= 5 {
			break
		}
		name := formatPlayerID(vp.VoterID)
		if player, err := q.GetPlayer(ctx, vp.VoterID); err == nil && player.RoleID.Valid {
			if role, err := q.GetRole(ctx, player.RoleID.Int32); err == nil {
				name = fmt.Sprintf("%s (%s)", name, role.Name)
			}
		}
		stats.MostActiveVoters = append(stats.MostActiveVoters, pages.PlayerVoteStat{
			PlayerID:   vp.VoterID,
			PlayerName: name,
			Count:      int(vp.VotesCast),
		})
	}

	return stats
}
