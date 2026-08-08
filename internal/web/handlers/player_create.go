package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/mccune1224/betrayal/internal/models"
	"github.com/mccune1224/betrayal/internal/util"
	"github.com/mccune1224/betrayal/internal/web/templates/pages"
)

type PlayerCreateHandler struct{ dbPool *pgxpool.Pool }

func NewPlayerCreateHandler(pool *pgxpool.Pool) *PlayerCreateHandler {
	return &PlayerCreateHandler{dbPool: pool}
}
func (h *PlayerCreateHandler) Page(c echo.Context) error {
	return render(c, http.StatusOK, pages.PlayerCreate())
}
func (h *PlayerCreateHandler) Create(c echo.Context) error {
	id, err := strconv.ParseInt(c.FormValue("player_id"), 10, 64)
	if err != nil || id <= 0 {
		return c.String(http.StatusBadRequest, "Player ID must be a positive integer")
	}
	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()
	q := models.New(h.dbPool)
	role, err := q.GetRoleByFuzzy(ctx, c.FormValue("role_name"))
	if err != nil {
		return c.String(http.StatusBadRequest, "Role not found")
	}
	numeric, err := util.Numeric(0)
	if err != nil {
		return c.String(http.StatusInternalServerError, "failed to initialize player")
	}
	coins, luck, limit := int32(200), int32(0), int32(4)
	if value, e := q.GetGameConfig(ctx, "default_coins"); e == nil {
		if n, e := strconv.ParseInt(value, 10, 32); e == nil {
			coins = int32(n)
		}
	}
	if value, e := q.GetGameConfig(ctx, "default_luck"); e == nil {
		if n, e := strconv.ParseInt(value, 10, 32); e == nil {
			luck = int32(n)
		}
	}
	if value, e := q.GetGameConfig(ctx, "default_item_limit"); e == nil {
		if n, e := strconv.ParseInt(value, 10, 32); e == nil {
			limit = int32(n)
		}
	}
	_, err = q.CreatePlayer(ctx, models.CreatePlayerParams{ID: id, RoleID: pgtype.Int4{Int32: role.ID, Valid: true}, Alive: true, Coins: coins, CoinBonus: numeric, Luck: luck, ItemLimit: limit, Alignment: role.Alignment})
	if err != nil {
		return c.String(http.StatusBadRequest, "Could not create player (the ID may already exist)")
	}
	return c.Redirect(http.StatusSeeOther, "/players/"+strconv.FormatInt(id, 10)+"/edit")
}
