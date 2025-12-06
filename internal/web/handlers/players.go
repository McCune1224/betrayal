package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/mccune1224/betrayal/internal/models"
	"github.com/mccune1224/betrayal/internal/web/templates/pages"
	"github.com/mccune1224/betrayal/internal/web/templates/partials"
)

// PlayersHandler handles player-related requests
type PlayersHandler struct {
	dbPool *pgxpool.Pool
}

// NewPlayersHandler creates a new PlayersHandler
func NewPlayersHandler(pool *pgxpool.Pool) *PlayersHandler {
	return &PlayersHandler{dbPool: pool}
}

// List handles GET /players - full page
func (h *PlayersHandler) List(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()

	q := models.New(h.dbPool)

	players, err := q.ListPlayer(ctx)
	if err != nil {
		players = []models.Player{}
	}

	// Convert to view model
	playerData := make([]pages.PlayerRow, len(players))
	for i, p := range players {
		roleName := "Unknown"
		if p.RoleID.Valid {
			if role, err := q.GetRole(ctx, p.RoleID.Int32); err == nil {
				roleName = role.Name
			}
		}

		playerData[i] = pages.PlayerRow{
			ID:        p.ID,
			Name:      fmt.Sprintf("%d", p.ID), // Discord user ID as name for now
			Role:      roleName,
			Alignment: string(p.Alignment),
			Alive:     p.Alive,
			Coins:     int(p.Coins),
			Luck:      int(p.Luck),
		}
	}

	return render(c, http.StatusOK, pages.Players(playerData))
}

// Table handles GET /players/table - HTMX partial
func (h *PlayersHandler) Table(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()

	q := models.New(h.dbPool)

	players, err := q.ListPlayer(ctx)
	if err != nil {
		players = []models.Player{}
	}

	// Convert to view model
	playerData := make([]partials.PlayerRowData, len(players))
	for i, p := range players {
		roleName := "Unknown"
		if p.RoleID.Valid {
			if role, err := q.GetRole(ctx, p.RoleID.Int32); err == nil {
				roleName = role.Name
			}
		}

		playerData[i] = partials.PlayerRowData{
			ID:        p.ID,
			Name:      fmt.Sprintf("%d", p.ID), // Discord user ID as name for now
			Role:      roleName,
			Alignment: string(p.Alignment),
			Alive:     p.Alive,
			Coins:     int(p.Coins),
			Luck:      int(p.Luck),
		}
	}

	return render(c, http.StatusOK, partials.PlayerTable(playerData))
}

// Detail handles GET /players/:id
func (h *PlayersHandler) Detail(c echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid player ID")
	}

	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()

	q := models.New(h.dbPool)

	player, err := q.GetPlayer(ctx, id)
	if err != nil {
		return c.String(http.StatusNotFound, "Player not found")
	}

	// Get role
	roleName := "Unknown"
	if player.RoleID.Valid {
		if role, err := q.GetRole(ctx, player.RoleID.Int32); err == nil {
			roleName = role.Name
		}
	}

	// Get items
	items, _ := q.ListPlayerItem(ctx, player.ID)
	itemNames := make([]string, len(items))
	for i, item := range items {
		itemNames[i] = item.Name
	}

	// Get abilities
	abilities, _ := q.ListPlayerAbility(ctx, player.ID)
	abilityNames := make([]string, len(abilities))
	for i, ab := range abilities {
		abilityNames[i] = ab.Name
	}

	// Get perks
	perks, _ := q.ListPlayerPerk(ctx, player.ID)
	perkNames := make([]string, len(perks))
	for i, perk := range perks {
		perkNames[i] = perk.Name
	}

	// Get statuses
	statuses, _ := q.ListPlayerStatus(ctx, player.ID)
	statusNames := make([]string, len(statuses))
	for i, status := range statuses {
		statusNames[i] = status.Name
	}

	// Get immunities
	immunities, _ := q.ListPlayerImmunity(ctx, player.ID)
	immunityNames := make([]string, len(immunities))
	for i, imm := range immunities {
		immunityNames[i] = imm.Name
	}

	// Get notes
	notes, _ := q.ListPlayerNote(ctx, player.ID)
	noteTexts := make([]string, len(notes))
	for i, note := range notes {
		noteTexts[i] = note.Info
	}

	data := pages.PlayerDetailData{
		ID:         player.ID,
		Name:       fmt.Sprintf("%d", player.ID), // Discord user ID as name for now
		Role:       roleName,
		Alignment:  string(player.Alignment),
		Alive:      player.Alive,
		Coins:      int(player.Coins),
		Luck:       int(player.Luck),
		Items:      itemNames,
		Abilities:  abilityNames,
		Perks:      perkNames,
		Statuses:   statusNames,
		Immunities: immunityNames,
		Notes:      noteTexts,
	}

	return render(c, http.StatusOK, pages.PlayerDetail(data))
}
