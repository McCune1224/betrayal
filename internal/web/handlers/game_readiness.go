package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/mccune1224/betrayal/internal/models"
	"github.com/mccune1224/betrayal/internal/web/templates/pages"
)

// GameReadinessHandler renders the web equivalent of the Discord
// /healthcheck command. It intentionally uses game language rather than
// database terminology so hosts can use it during setup.
type GameReadinessHandler struct {
	pool           *pgxpool.Pool
	discordSession *discordgo.Session
}

func NewGameReadinessHandler(pool *pgxpool.Pool, discordSession *discordgo.Session) *GameReadinessHandler {
	return &GameReadinessHandler{pool: pool, discordSession: discordSession}
}

func (h *GameReadinessHandler) Page(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()
	q := models.New(h.pool)

	data := pages.GameReadinessData{DiscordConnected: h.discordSession != nil}
	admin, adminErr := q.ListAdminChannel(ctx)
	data.AdminChannels = len(admin)
	data.AdminReady = adminErr == nil && len(admin) > 0

	vote, voteErr := q.GetVoteChannel(ctx)
	data.VoteChannel = vote
	data.VoteReady = voteErr == nil && vote != ""

	action, actionErr := q.GetActionChannel(ctx)
	data.ActionChannel = action
	data.ActionReady = actionErr == nil && action != ""

	lifeboard, lifeboardErr := q.GetPlayerLifeboard(ctx)
	data.LifeboardReady = lifeboardErr == nil && lifeboard.ChannelID != ""

	players, playersErr := q.ListPlayer(ctx)
	if playersErr == nil {
		data.PlayerCount = len(players)
		for _, player := range players {
			if player.Alive {
				data.Alive++
			} else {
				data.Dead++
			}
		}
	}

	cycle, cycleErr := q.GetCycle(ctx)
	if cycleErr == nil {
		data.CyclePhase = "Day"
		if cycle.IsElimination {
			data.CyclePhase = "Elimination"
		}
		data.CycleDay = int(cycle.Day)
		data.CycleReady = true
	}

	data.Ready = data.AdminReady && data.VoteReady && data.ActionReady && data.PlayerCount > 0
	return render(c, http.StatusOK, pages.GameReadiness(data))
}
