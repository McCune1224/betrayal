package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/mccune1224/betrayal/internal/models"
	"github.com/mccune1224/betrayal/internal/web/templates/pages"
	"github.com/mccune1224/betrayal/internal/web/templates/partials"
)

// CatalogHandler implements CRUD for the game catalog tables:
// items (/items), abilities (/abilities) and statuses (/statuses).
// This generalizes the roles CRUD pattern (list + search + detail + edit +
// create + delete) to the other catalog entities.
type CatalogHandler struct {
	dbPool *pgxpool.Pool
}

// NewCatalogHandler creates a new CatalogHandler
func NewCatalogHandler(pool *pgxpool.Pool) *CatalogHandler {
	return &CatalogHandler{dbPool: pool}
}

// ---------------------------------------------------------------------------
// Items
// ---------------------------------------------------------------------------

// Items handles GET /items — list page with search + create form
func (h *CatalogHandler) Items(c echo.Context) error {
	return render(c, http.StatusOK, pages.Items())
}

// SearchItems handles GET /items/search — HTMX partial for live search
func (h *CatalogHandler) SearchItems(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()

	q := models.New(h.dbPool)
	searchQuery := c.QueryParam("q")

	var items []models.Item
	var err error
	if searchQuery == "" {
		items, err = q.ListItem(ctx)
	} else {
		items, err = q.SearchItemByKeyword(ctx, "%"+searchQuery+"%")
	}
	if err != nil {
		items = []models.Item{}
	}

	rows := make([]partials.CatalogRow, len(items))
	for i, it := range items {
		rows[i] = partials.CatalogRow{
			ID:          it.ID,
			Name:        it.Name,
			Description: it.Description,
			Meta:        string(it.Rarity),
			Path:        "/items",
		}
	}
	return render(c, http.StatusOK, partials.CatalogSearchResults("items", rows))
}

// ItemDetail handles GET /items/:id — detail/edit page
func (h *CatalogHandler) ItemDetail(c echo.Context) error {
	id, ok := parseCatalogID(c)
	if !ok {
		return c.String(http.StatusBadRequest, "Invalid item ID")
	}

	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()

	item, err := models.New(h.dbPool).GetItem(ctx, id)
	if err != nil {
		return c.String(http.StatusNotFound, "Item not found")
	}

	data := pages.ItemDetailData{
		ID:          item.ID,
		Name:        item.Name,
		Description: item.Description,
		Rarity:      string(item.Rarity),
		Cost:        item.Cost,
	}
	return render(c, http.StatusOK, pages.ItemDetail(data))
}

// CreateItem handles POST /items — create a new item
func (h *CatalogHandler) CreateItem(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()

	rarity, ok := parseRarity(c.FormValue("rarity"))
	if !ok {
		return h.catalogToastError(c, "Invalid rarity")
	}
	cost, err := strconv.ParseInt(c.FormValue("cost"), 10, 32)
	if err != nil {
		cost = 0
	}

	item, err := models.New(h.dbPool).CreateItem(ctx, models.CreateItemParams{
		Name:        c.FormValue("name"),
		Description: c.FormValue("description"),
		Rarity:      rarity,
		Cost:        int32(cost),
	})
	if err != nil {
		return h.catalogToastError(c, "Failed to create item")
	}

	c.Response().Header().Set("HX-Trigger", `{"showToast": {"message": "Item created", "type": "success"}}`)
	return c.Redirect(http.StatusSeeOther, "/items/"+strconv.Itoa(int(item.ID)))
}

// UpdateItem handles POST /items/:id — update item fields
func (h *CatalogHandler) UpdateItem(c echo.Context) error {
	id, ok := parseCatalogID(c)
	if !ok {
		return c.String(http.StatusBadRequest, "Invalid item ID")
	}

	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()

	rarity, ok := parseRarity(c.FormValue("rarity"))
	if !ok {
		return h.catalogToastError(c, "Invalid rarity")
	}
	cost, err := strconv.ParseInt(c.FormValue("cost"), 10, 32)
	if err != nil {
		cost = 0
	}

	_, err = models.New(h.dbPool).UpdateItem(ctx, models.UpdateItemParams{
		ID:          id,
		Name:        c.FormValue("name"),
		Description: c.FormValue("description"),
		Rarity:      rarity,
		Cost:        int32(cost),
	})
	if err != nil {
		return h.catalogToastError(c, "Failed to update item")
	}

	c.Response().Header().Set("HX-Trigger", `{"showToast": {"message": "Item updated", "type": "success"}}`)
	return c.String(http.StatusOK, "")
}

// DeleteItem handles POST /items/:id/delete — delete an item
func (h *CatalogHandler) DeleteItem(c echo.Context) error {
	id, ok := parseCatalogID(c)
	if !ok {
		return c.String(http.StatusBadRequest, "Invalid item ID")
	}

	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()

	if err := models.New(h.dbPool).DeleteItem(ctx, id); err != nil {
		return h.catalogToastError(c, "Failed to delete item")
	}

	c.Response().Header().Set("HX-Trigger", `{"showToast": {"message": "Item deleted", "type": "success"}}`)
	return c.Redirect(http.StatusSeeOther, "/items")
}

// ---------------------------------------------------------------------------
// Abilities
// ---------------------------------------------------------------------------

// Abilities handles GET /abilities — list page with search + create form
func (h *CatalogHandler) Abilities(c echo.Context) error {
	return render(c, http.StatusOK, pages.Abilities())
}

// SearchAbilities handles GET /abilities/search — HTMX partial
func (h *CatalogHandler) SearchAbilities(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()

	q := models.New(h.dbPool)
	searchQuery := c.QueryParam("q")

	var abilities []models.AbilityInfo
	var err error
	if searchQuery == "" {
		abilities, err = q.ListAbilityInfo(ctx)
	} else {
		abilities, err = q.SearchAbilityByKeyword(ctx, "%"+searchQuery+"%")
	}
	if err != nil {
		abilities = []models.AbilityInfo{}
	}

	rows := make([]partials.CatalogRow, len(abilities))
	for i, ab := range abilities {
		meta := string(ab.Rarity)
		if ab.AnyAbility {
			meta += " · AA"
		}
		rows[i] = partials.CatalogRow{
			ID:          ab.ID,
			Name:        ab.Name,
			Description: ab.Description,
			Meta:        meta,
			Path:        "/abilities",
		}
	}
	return render(c, http.StatusOK, partials.CatalogSearchResults("abilities", rows))
}

// AbilityDetail handles GET /abilities/:id
func (h *CatalogHandler) AbilityDetail(c echo.Context) error {
	id, ok := parseCatalogID(c)
	if !ok {
		return c.String(http.StatusBadRequest, "Invalid ability ID")
	}

	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()

	ability, err := models.New(h.dbPool).GetAbilityInfo(ctx, id)
	if err != nil {
		return c.String(http.StatusNotFound, "Ability not found")
	}

	data := pages.AbilityDetailData{
		ID:             ability.ID,
		Name:           ability.Name,
		Description:    ability.Description,
		DefaultCharges: ability.DefaultCharges,
		AnyAbility:     ability.AnyAbility,
		Rarity:         string(ability.Rarity),
	}
	return render(c, http.StatusOK, pages.AbilityDetail(data))
}

// CreateAbility handles POST /abilities
func (h *CatalogHandler) CreateAbility(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()

	rarity, ok := parseRarity(c.FormValue("rarity"))
	if !ok {
		return h.catalogToastError(c, "Invalid rarity")
	}
	charges, _ := strconv.ParseInt(c.FormValue("default_charges"), 10, 32)
	anyAbility := c.FormValue("any_ability") == "on" || c.FormValue("any_ability") == "true"

	ability, err := models.New(h.dbPool).CreateAbilityInfo(ctx, models.CreateAbilityInfoParams{
		Name:           c.FormValue("name"),
		Description:    c.FormValue("description"),
		DefaultCharges: int32(charges),
		AnyAbility:     anyAbility,
		Rarity:         rarity,
	})
	if err != nil {
		return h.catalogToastError(c, "Failed to create ability")
	}

	c.Response().Header().Set("HX-Trigger", `{"showToast": {"message": "Ability created", "type": "success"}}`)
	return c.Redirect(http.StatusSeeOther, "/abilities/"+strconv.Itoa(int(ability.ID)))
}

// UpdateAbility handles POST /abilities/:id
func (h *CatalogHandler) UpdateAbility(c echo.Context) error {
	id, ok := parseCatalogID(c)
	if !ok {
		return c.String(http.StatusBadRequest, "Invalid ability ID")
	}

	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()

	rarity, ok := parseRarity(c.FormValue("rarity"))
	if !ok {
		return h.catalogToastError(c, "Invalid rarity")
	}
	charges, _ := strconv.ParseInt(c.FormValue("default_charges"), 10, 32)
	anyAbility := c.FormValue("any_ability") == "on" || c.FormValue("any_ability") == "true"

	_, err := models.New(h.dbPool).UpdateAbilityInfo(ctx, models.UpdateAbilityInfoParams{
		ID:             id,
		Name:           c.FormValue("name"),
		Description:    c.FormValue("description"),
		DefaultCharges: int32(charges),
		AnyAbility:     anyAbility,
		Rarity:         rarity,
	})
	if err != nil {
		return h.catalogToastError(c, "Failed to update ability")
	}

	c.Response().Header().Set("HX-Trigger", `{"showToast": {"message": "Ability updated", "type": "success"}}`)
	return c.String(http.StatusOK, "")
}

// DeleteAbility handles POST /abilities/:id/delete
func (h *CatalogHandler) DeleteAbility(c echo.Context) error {
	id, ok := parseCatalogID(c)
	if !ok {
		return c.String(http.StatusBadRequest, "Invalid ability ID")
	}

	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()

	if err := models.New(h.dbPool).DeleteAbilityInfo(ctx, id); err != nil {
		return h.catalogToastError(c, "Failed to delete ability")
	}

	c.Response().Header().Set("HX-Trigger", `{"showToast": {"message": "Ability deleted", "type": "success"}}`)
	return c.Redirect(http.StatusSeeOther, "/abilities")
}

// ---------------------------------------------------------------------------
// Statuses
// ---------------------------------------------------------------------------

// Statuses handles GET /statuses — list page with search + create form
func (h *CatalogHandler) Statuses(c echo.Context) error {
	return render(c, http.StatusOK, pages.Statuses())
}

// SearchStatuses handles GET /statuses/search — HTMX partial
func (h *CatalogHandler) SearchStatuses(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()

	q := models.New(h.dbPool)
	searchQuery := c.QueryParam("q")

	var statuses []models.Status
	var err error
	if searchQuery == "" {
		statuses, err = q.ListStatus(ctx)
	} else {
		statuses, err = q.SearchStatusByKeyword(ctx, "%"+searchQuery+"%")
	}
	if err != nil {
		statuses = []models.Status{}
	}

	rows := make([]partials.CatalogRow, len(statuses))
	for i, st := range statuses {
		meta := ""
		if st.HourDuration > 0 {
			meta = strconv.Itoa(int(st.HourDuration)) + "h"
		}
		rows[i] = partials.CatalogRow{
			ID:          st.ID,
			Name:        st.Name,
			Description: st.Description,
			Meta:        meta,
			Path:        "/statuses",
		}
	}
	return render(c, http.StatusOK, partials.CatalogSearchResults("statuses", rows))
}

// StatusDetail handles GET /statuses/:id
func (h *CatalogHandler) StatusDetail(c echo.Context) error {
	id, ok := parseCatalogID(c)
	if !ok {
		return c.String(http.StatusBadRequest, "Invalid status ID")
	}

	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()

	// There is no GetStatusByID query, so list all and match by id.
	var found *models.Status
	all, err := models.New(h.dbPool).ListStatus(ctx)
	if err == nil {
		for i := range all {
			if all[i].ID == id {
				found = &all[i]
				break
			}
		}
	}
	if found == nil {
		return c.String(http.StatusNotFound, "Status not found")
	}

	data := pages.StatusDetailData{
		ID:           found.ID,
		Name:         found.Name,
		Description:  found.Description,
		HourDuration: found.HourDuration,
	}
	return render(c, http.StatusOK, pages.StatusDetail(data))
}

// CreateStatus handles POST /statuses
func (h *CatalogHandler) CreateStatus(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()

	hours, _ := strconv.ParseInt(c.FormValue("hour_duration"), 10, 32)

	status, err := models.New(h.dbPool).CreateStatus(ctx, models.CreateStatusParams{
		Name:        c.FormValue("name"),
		Description: c.FormValue("description"),
	})
	if err != nil {
		return h.catalogToastError(c, "Failed to create status")
	}

	if hours > 0 {
		// UpdateStatus doubles as the setter for hour_duration
		_, err = models.New(h.dbPool).UpdateStatus(ctx, models.UpdateStatusParams{
			ID:           status.ID,
			Name:         status.Name,
			Description:  status.Description,
			HourDuration: int32(hours),
		})
		if err != nil {
			return h.catalogToastError(c, "Failed to set status duration")
		}
	}

	c.Response().Header().Set("HX-Trigger", `{"showToast": {"message": "Status created", "type": "success"}}`)
	return c.Redirect(http.StatusSeeOther, "/statuses/"+strconv.Itoa(int(status.ID)))
}

// UpdateStatus handles POST /statuses/:id
func (h *CatalogHandler) UpdateStatus(c echo.Context) error {
	id, ok := parseCatalogID(c)
	if !ok {
		return c.String(http.StatusBadRequest, "Invalid status ID")
	}

	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()

	hours, _ := strconv.ParseInt(c.FormValue("hour_duration"), 10, 32)

	_, err := models.New(h.dbPool).UpdateStatus(ctx, models.UpdateStatusParams{
		ID:           id,
		Name:         c.FormValue("name"),
		Description:  c.FormValue("description"),
		HourDuration: int32(hours),
	})
	if err != nil {
		return h.catalogToastError(c, "Failed to update status")
	}

	c.Response().Header().Set("HX-Trigger", `{"showToast": {"message": "Status updated", "type": "success"}}`)
	return c.String(http.StatusOK, "")
}

// DeleteStatus handles POST /statuses/:id/delete
func (h *CatalogHandler) DeleteStatus(c echo.Context) error {
	id, ok := parseCatalogID(c)
	if !ok {
		return c.String(http.StatusBadRequest, "Invalid status ID")
	}

	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()

	if err := models.New(h.dbPool).DeleteStatus(ctx, id); err != nil {
		return h.catalogToastError(c, "Failed to delete status")
	}

	c.Response().Header().Set("HX-Trigger", `{"showToast": {"message": "Status deleted", "type": "success"}}`)
	return c.Redirect(http.StatusSeeOther, "/statuses")
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func parseCatalogID(c echo.Context) (int32, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		return 0, false
	}
	return int32(id), true
}

func parseRarity(raw string) (models.Rarity, bool) {
	switch models.Rarity(raw) {
	case models.RarityCOMMON,
		models.RarityUNCOMMON,
		models.RarityRARE,
		models.RarityEPIC,
		models.RarityLEGENDARY,
		models.RarityMYTHICAL,
		models.RarityROLESPECIFIC,
		models.RarityUNIQUE:
		return models.Rarity(raw), true
	default:
		return "", false
	}
}

func (h *CatalogHandler) catalogToastError(c echo.Context, msg string) error {
	c.Response().Header().Set("HX-Trigger", `{"showToast": {"message": "`+msg+`", "type": "error"}}`)
	return c.String(http.StatusBadRequest, msg)
}
