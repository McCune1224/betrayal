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
	"github.com/mccune1224/betrayal/internal/services/inventory"
	"github.com/mccune1224/betrayal/internal/web/templates/pages"
)

// PlayerEditHandler handles the player edit page (/players/:id/edit).
// Inventory mutations reuse internal/services/inventory so web and Discord
// edits go through the same game rules.
type PlayerEditHandler struct {
	dbPool *pgxpool.Pool
}

// NewPlayerEditHandler creates a new PlayerEditHandler
func NewPlayerEditHandler(pool *pgxpool.Pool) *PlayerEditHandler {
	return &PlayerEditHandler{dbPool: pool}
}

// Edit handles GET /players/:id/edit — full edit page
func (h *PlayerEditHandler) Edit(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()

	playerID, ok := parsePlayerID(c)
	if !ok {
		return c.String(http.StatusBadRequest, "Invalid player ID")
	}

	q := models.New(h.dbPool)
	player, err := q.GetPlayer(ctx, playerID)
	if err != nil {
		return c.String(http.StatusNotFound, "Player not found")
	}

	data, err := h.loadEditData(ctx, player)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load player inventory")
	}

	return render(c, http.StatusOK, pages.PlayerEdit(data))
}

// UpdateStats handles POST /players/:id/edit — set coins and luck
func (h *PlayerEditHandler) UpdateStats(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()

	playerID, ok := parsePlayerID(c)
	if !ok {
		return c.String(http.StatusBadRequest, "Invalid player ID")
	}

	q := models.New(h.dbPool)
	player, err := q.GetPlayer(ctx, playerID)
	if err != nil {
		return c.String(http.StatusNotFound, "Player not found")
	}

	// Reuse the inventory service for coin edits (clamps negatives to 0)
	ih := inventory.NewManualInventoryHandler(player, h.dbPool)

	if coinsStr := c.FormValue("coins"); coinsStr != "" {
		coins, err := strconv.ParseInt(coinsStr, 10, 32)
		if err != nil {
			return h.toastError(c, "Coins must be a whole number")
		}
		if err := ih.SetCoin(int32(coins)); err != nil {
			return h.toastError(c, "Failed to update coins")
		}
	}

	if luckStr := c.FormValue("luck"); luckStr != "" {
		luck, err := strconv.ParseInt(luckStr, 10, 32)
		if err != nil {
			return h.toastError(c, "Luck must be a whole number")
		}
		if _, err := q.UpdatePlayerLuck(ctx, models.UpdatePlayerLuckParams{
			ID:   playerID,
			Luck: int32(luck),
		}); err != nil {
			return h.toastError(c, "Failed to update luck")
		}
	}

	// Re-fetch and re-render the stats card
	updated, err := q.GetPlayer(ctx, playerID)
	if err != nil {
		return h.toastError(c, "Failed to reload player")
	}

	c.Response().Header().Set("HX-Trigger", `{"showToast": {"message": "Stats updated", "type": "success"}}`)
	return render(c, http.StatusOK, pages.PlayerEditStats(playerID, int(updated.Coins), int(updated.Luck)))
}

// AddItem handles POST /players/:id/items/add
func (h *PlayerEditHandler) AddItem(c echo.Context) error {
	return h.mutateInventory(c, "items", func(ih *inventory.InventoryHandler) error {
		name := c.FormValue("item_name")
		if name == "" {
			return fmt.Errorf("item name is required")
		}
		qty := parseQuantity(c.FormValue("quantity"), 1)
		_, err := ih.AddItem(name, qty)
		return err
	})
}

// RemoveItem handles POST /players/:id/items/remove (removes one at a time)
func (h *PlayerEditHandler) RemoveItem(c echo.Context) error {
	return h.mutateInventory(c, "items", func(ih *inventory.InventoryHandler) error {
		name := c.FormValue("item_name")
		if name == "" {
			return fmt.Errorf("item name is required")
		}
		_, err := ih.RemoveItem(name, 1)
		return err
	})
}

// AddAbility handles POST /players/:id/abilities/add
func (h *PlayerEditHandler) AddAbility(c echo.Context) error {
	return h.mutateInventory(c, "abilities", func(ih *inventory.InventoryHandler) error {
		name := c.FormValue("ability_name")
		if name == "" {
			return fmt.Errorf("ability name is required")
		}
		qty := parseQuantity(c.FormValue("quantity"), 0) // 0 = default charges
		_, err := ih.AddAbility(name, qty)
		return err
	})
}

// RemoveAbility handles POST /players/:id/abilities/remove
func (h *PlayerEditHandler) RemoveAbility(c echo.Context) error {
	return h.mutateInventory(c, "abilities", func(ih *inventory.InventoryHandler) error {
		name := c.FormValue("ability_name")
		if name == "" {
			return fmt.Errorf("ability name is required")
		}
		_, err := ih.RemoveAbility(name)
		return err
	})
}

// AddStatus handles POST /players/:id/statuses/add
func (h *PlayerEditHandler) AddStatus(c echo.Context) error {
	return h.mutateInventory(c, "statuses", func(ih *inventory.InventoryHandler) error {
		name := c.FormValue("status_name")
		if name == "" {
			return fmt.Errorf("status name is required")
		}
		qty := parseQuantity(c.FormValue("quantity"), 1)
		_, err := ih.AddStatus(name, qty)
		return err
	})
}

// RemoveStatus handles POST /players/:id/statuses/remove (removes one at a time)
func (h *PlayerEditHandler) RemoveStatus(c echo.Context) error {
	return h.mutateInventory(c, "statuses", func(ih *inventory.InventoryHandler) error {
		name := c.FormValue("status_name")
		if name == "" {
			return fmt.Errorf("status name is required")
		}
		_, err := ih.RemoveStatus(name, 1)
		return err
	})
}

// AddPerk handles POST /players/:id/perks/add.
// The inventory service has no perk helpers (perks are a plain join), so this
// uses the models queries directly.
func (h *PlayerEditHandler) AddPerk(c echo.Context) error {
	return h.mutateInventory(c, "perks", func(ih *inventory.InventoryHandler) error {
		name := c.FormValue("perk_name")
		if name == "" {
			return fmt.Errorf("perk name is required")
		}
		q := models.New(h.dbPool)
		ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
		defer cancel()
		perk, err := q.GetPerkInfoByFuzzy(ctx, name)
		if err != nil {
			return fmt.Errorf("perk not found: %s", name)
		}
		_, err = q.CreatePlayerPerkJoin(ctx, models.CreatePlayerPerkJoinParams{
			PlayerID: ih.GetPlayer().ID,
			PerkID:   perk.ID,
		})
		return err
	})
}

// RemovePerk handles POST /players/:id/perks/remove
func (h *PlayerEditHandler) RemovePerk(c echo.Context) error {
	return h.mutateInventory(c, "perks", func(ih *inventory.InventoryHandler) error {
		name := c.FormValue("perk_name")
		if name == "" {
			return fmt.Errorf("perk name is required")
		}
		q := models.New(h.dbPool)
		ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
		defer cancel()
		perk, err := q.GetPerkInfoByFuzzy(ctx, name)
		if err != nil {
			return fmt.Errorf("perk not found: %s", name)
		}
		return q.DeletePlayerPerk(ctx, models.DeletePlayerPerkParams{
			PlayerID: ih.GetPlayer().ID,
			PerkID:   perk.ID,
		})
	})
}

// mutateInventory loads the player, applies a mutation, then re-renders the
// affected inventory section so HTMX can swap it in place.
func (h *PlayerEditHandler) mutateInventory(c echo.Context, section string, mutate func(*inventory.InventoryHandler) error) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()

	playerID, ok := parsePlayerID(c)
	if !ok {
		return c.String(http.StatusBadRequest, "Invalid player ID")
	}

	q := models.New(h.dbPool)
	player, err := q.GetPlayer(ctx, playerID)
	if err != nil {
		return c.String(http.StatusNotFound, "Player not found")
	}

	ih := inventory.NewManualInventoryHandler(player, h.dbPool)
	if err := mutate(ih); err != nil {
		c.Response().Header().Set("HX-Trigger", `{"showToast": {"message": "`+err.Error()+`", "type": "error"}}`)
		return c.String(http.StatusBadRequest, err.Error())
	}

	c.Response().Header().Set("HX-Trigger", `{"showToast": {"message": "Inventory updated", "type": "success"}}`)
	return h.renderSection(c, ctx, section, playerID)
}

// renderSection renders the HTMX partial for one inventory section
func (h *PlayerEditHandler) renderSection(c echo.Context, ctx context.Context, section string, playerID int64) error {
	q := models.New(h.dbPool)

	switch section {
	case "items":
		items, err := q.ListPlayerItemInventory(ctx, playerID)
		if err != nil {
			return c.String(http.StatusInternalServerError, "Failed to load items")
		}
		rows := make([]pages.PlayerEditItemRow, len(items))
		for i, it := range items {
			rows[i] = pages.PlayerEditItemRow{ID: it.ID, Name: it.Name, Quantity: it.Quantity}
		}
		return render(c, http.StatusOK, pages.PlayerEditItems(playerID, rows))
	case "abilities":
		abilities, err := q.ListPlayerAbilityInventory(ctx, playerID)
		if err != nil {
			return c.String(http.StatusInternalServerError, "Failed to load abilities")
		}
		rows := make([]pages.PlayerEditAbilityRow, len(abilities))
		for i, ab := range abilities {
			rows[i] = pages.PlayerEditAbilityRow{ID: ab.ID, Name: ab.Name, Quantity: ab.Quantity}
		}
		return render(c, http.StatusOK, pages.PlayerEditAbilities(playerID, rows))
	case "statuses":
		statuses, err := q.ListPlayerStatusInventory(ctx, playerID)
		if err != nil {
			return c.String(http.StatusInternalServerError, "Failed to load statuses")
		}
		rows := make([]pages.PlayerEditStatusRow, len(statuses))
		for i, st := range statuses {
			rows[i] = pages.PlayerEditStatusRow{ID: st.ID, Name: st.Name, Quantity: st.Quantity}
		}
		return render(c, http.StatusOK, pages.PlayerEditStatuses(playerID, rows))
	case "perks":
		perks, err := q.ListPlayerPerk(ctx, playerID)
		if err != nil {
			return c.String(http.StatusInternalServerError, "Failed to load perks")
		}
		rows := make([]pages.PlayerEditPerkRow, len(perks))
		for i, pk := range perks {
			rows[i] = pages.PlayerEditPerkRow{ID: pk.ID, Name: pk.Name}
		}
		return render(c, http.StatusOK, pages.PlayerEditPerks(playerID, rows))
	default:
		return c.String(http.StatusInternalServerError, "Unknown section")
	}
}

// loadEditData gathers everything the full edit page needs
func (h *PlayerEditHandler) loadEditData(ctx context.Context, player models.Player) (pages.PlayerEditData, error) {
	q := models.New(h.dbPool)

	roleName := "Unknown"
	if player.RoleID.Valid {
		if role, err := q.GetRole(ctx, player.RoleID.Int32); err == nil {
			roleName = role.Name
		}
	}

	items, err := q.ListPlayerItemInventory(ctx, player.ID)
	if err != nil {
		return pages.PlayerEditData{}, err
	}
	itemRows := make([]pages.PlayerEditItemRow, len(items))
	for i, it := range items {
		itemRows[i] = pages.PlayerEditItemRow{ID: it.ID, Name: it.Name, Quantity: it.Quantity}
	}

	abilities, err := q.ListPlayerAbilityInventory(ctx, player.ID)
	if err != nil {
		return pages.PlayerEditData{}, err
	}
	abilityRows := make([]pages.PlayerEditAbilityRow, len(abilities))
	for i, ab := range abilities {
		abilityRows[i] = pages.PlayerEditAbilityRow{ID: ab.ID, Name: ab.Name, Quantity: ab.Quantity}
	}

	statuses, err := q.ListPlayerStatusInventory(ctx, player.ID)
	if err != nil {
		return pages.PlayerEditData{}, err
	}
	statusRows := make([]pages.PlayerEditStatusRow, len(statuses))
	for i, st := range statuses {
		statusRows[i] = pages.PlayerEditStatusRow{ID: st.ID, Name: st.Name, Quantity: st.Quantity}
	}

	perks, err := q.ListPlayerPerk(ctx, player.ID)
	if err != nil {
		return pages.PlayerEditData{}, err
	}
	perkRows := make([]pages.PlayerEditPerkRow, len(perks))
	for i, pk := range perks {
		perkRows[i] = pages.PlayerEditPerkRow{ID: pk.ID, Name: pk.Name}
	}

	return pages.PlayerEditData{
		ID:        player.ID,
		Role:      roleName,
		Alive:     player.Alive,
		Coins:     int(player.Coins),
		Luck:      int(player.Luck),
		Items:     itemRows,
		Abilities: abilityRows,
		Statuses:  statusRows,
		Perks:     perkRows,
	}, nil
}

func parsePlayerID(c echo.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return 0, false
	}
	return id, true
}

func parseQuantity(raw string, def int32) int32 {
	if raw == "" {
		return def
	}
	q, err := strconv.ParseInt(raw, 10, 32)
	if err != nil || q < 1 {
		return def
	}
	return int32(q)
}

func (h *PlayerEditHandler) toastError(c echo.Context, msg string) error {
	c.Response().Header().Set("HX-Trigger", `{"showToast": {"message": "`+msg+`", "type": "error"}}`)
	return c.String(http.StatusBadRequest, msg)
}
