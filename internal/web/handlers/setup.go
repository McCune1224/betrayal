package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/mccune1224/betrayal/internal/services/roledraft"
	"github.com/mccune1224/betrayal/internal/web/templates/pages"
)

type SetupHandler struct {
	dbPool         *pgxpool.Pool
	discordSession *discordgo.Session
}

func NewSetupHandler(pool *pgxpool.Pool, session *discordgo.Session) *SetupHandler {
	return &SetupHandler{dbPool: pool, discordSession: session}
}

func (h *SetupHandler) Page(c echo.Context) error {
	return render(c, http.StatusOK, pages.SetupPage(pages.SetupData{DiscordConnected: h.discordSession != nil, DeceptionistDefault: h.deceptionistCount()}))
}
func (h *SetupHandler) Generate(c echo.Context) error {
	players, err := strconv.Atoi(c.FormValue("player_count"))
	if err != nil {
		return c.String(http.StatusBadRequest, "player count must be a whole number")
	}
	deceps, err := strconv.Atoi(c.FormValue("deceptionist_count"))
	if err != nil {
		return c.String(http.StatusBadRequest, "deceptionist count must be a whole number")
	}
	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()
	roles, err := roledraft.LoadRoles(ctx, h.dbPool)
	if err != nil {
		return c.String(http.StatusInternalServerError, "failed to load active roles")
	}
	pool, err := roledraft.Generate(roles, players, deceps)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	return render(c, http.StatusOK, pages.SetupResult(pages.SetupResultData{Pool: pool, PlayerCount: players, DeceptionistCount: deceps}))
}
func (h *SetupHandler) deceptionistCount() int {
	if h.discordSession == nil {
		return 0
	}
	members, err := h.discordSession.GuildMembers(discord.BetraylGuildID, "", 1000)
	if err != nil {
		return 0
	}
	roles, err := h.discordSession.GuildRoles(discord.BetraylGuildID)
	if err != nil {
		return 0
	}
	var id string
	for _, r := range roles {
		if r.Name == "Deceptionist" {
			id = r.ID
			break
		}
	}
	n := 0
	for _, m := range members {
		for _, rid := range m.Roles {
			if rid == id {
				n++
				break
			}
		}
	}
	return n
}
