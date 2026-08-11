package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/mccune1224/betrayal/internal/models"
)

// CatalogHandler exposes catalog records as versioned JSON DTOs. sqlc models
// deliberately never cross this boundary so the frontend contract stays stable.
type CatalogHandler struct{ pool *pgxpool.Pool }

func NewCatalogHandler(pool *pgxpool.Pool) *CatalogHandler { return &CatalogHandler{pool: pool} }

type catalogAbilityDTO struct {
	ID             int32  `json:"id"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	DefaultCharges int32  `json:"default_charges"`
	AnyAbility     bool   `json:"any_ability"`
	Rarity         string `json:"rarity"`
}
type catalogPerkDTO struct {
	ID          int32  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}
type catalogRoleDTO struct {
	ID          int32               `json:"id"`
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Alignment   string              `json:"alignment"`
	Abilities   []catalogAbilityDTO `json:"abilities"`
	Perks       []catalogPerkDTO    `json:"perks"`
}
type catalogItemDTO struct {
	ID          int32  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Rarity      string `json:"rarity"`
	Cost        int32  `json:"cost"`
}
type catalogStatusDTO struct {
	ID           int32  `json:"id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	HourDuration int32  `json:"hour_duration"`
}

type catalogRoleInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Alignment   string `json:"alignment"`
}
type catalogItemInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Rarity      string `json:"rarity"`
	Cost        int32  `json:"cost"`
}
type catalogAbilityInput struct {
	Name           string `json:"name"`
	Description    string `json:"description"`
	DefaultCharges int32  `json:"default_charges"`
	AnyAbility     bool   `json:"any_ability"`
	Rarity         string `json:"rarity"`
}
type catalogStatusInput struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	HourDuration int32  `json:"hour_duration"`
}

func catalogContext(c echo.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(c.Request().Context(), 10*time.Second)
}
func decodeCatalog(c echo.Context, dst any) error {
	if err := json.NewDecoder(c.Request().Body).Decode(dst); err != nil {
		WriteError(c.Response(), http.StatusBadRequest, "invalid_json", "request body must be valid JSON", map[string]any{})
		return err
	}
	return nil
}
func catalogID(c echo.Context) (int32, error) {
	n, err := strconv.ParseInt(c.Param("id"), 10, 32)
	return int32(n), err
}
func validRarity(s string) (models.Rarity, bool) {
	r := models.Rarity(s)
	switch r {
	case models.RarityCOMMON, models.RarityUNCOMMON, models.RarityRARE, models.RarityEPIC, models.RarityLEGENDARY, models.RarityMYTHICAL, models.RarityROLESPECIFIC, models.RarityUNIQUE:
		return r, true
	}
	return "", false
}
func validAlignment(s string) (models.Alignment, bool) {
	a := models.Alignment(s)
	switch a {
	case models.AlignmentGOOD, models.AlignmentNEUTRAL, models.AlignmentEVIL:
		return a, true
	}
	return "", false
}
func catalogBad(c echo.Context, msg string) error {
	WriteError(c.Response(), http.StatusBadRequest, "invalid_catalog_input", msg, map[string]any{})
	return nil
}
func catalogFailure(c echo.Context, code string) error {
	WriteError(c.Response(), http.StatusInternalServerError, code, "could not update catalog", map[string]any{})
	return nil
}

func roleDTO(r models.Role, abilities []models.AbilityInfo, perks []models.PerkInfo) catalogRoleDTO {
	d := catalogRoleDTO{ID: r.ID, Name: r.Name, Description: r.Description, Alignment: string(r.Alignment), Abilities: []catalogAbilityDTO{}, Perks: []catalogPerkDTO{}}
	for _, a := range abilities {
		d.Abilities = append(d.Abilities, catalogAbilityDTO{a.ID, a.Name, a.Description, a.DefaultCharges, a.AnyAbility, string(a.Rarity)})
	}
	for _, p := range perks {
		d.Perks = append(d.Perks, catalogPerkDTO{p.ID, p.Name, p.Description})
	}
	return d
}
func abilityDTO(a models.AbilityInfo) catalogAbilityDTO {
	return catalogAbilityDTO{a.ID, a.Name, a.Description, a.DefaultCharges, a.AnyAbility, string(a.Rarity)}
}
func itemDTO(i models.Item) catalogItemDTO {
	return catalogItemDTO{i.ID, i.Name, i.Description, string(i.Rarity), i.Cost}
}
func statusDTO(s models.Status) catalogStatusDTO {
	return catalogStatusDTO{s.ID, s.Name, s.Description, s.HourDuration}
}

func (h *CatalogHandler) ListRoles(c echo.Context) error {
	ctx, cancel := catalogContext(c)
	defer cancel()
	q := models.New(h.pool)
	var rs []models.Role
	var err error
	if term := c.QueryParam("q"); term != "" {
		rs, err = q.SearchRoleByName(ctx, term)
	} else {
		rs, err = q.Listrole(ctx)
	}
	if err != nil {
		return catalogFailure(c, "roles_unavailable")
	}
	out := make([]catalogRoleDTO, 0, len(rs))
	for _, r := range rs {
		abilities, _ := q.ListRoleAbilityForRole(ctx, r.ID)
		perks, _ := q.ListRolePerkForRole(ctx, r.ID)
		out = append(out, roleDTO(r, abilities, perks))
	}
	WriteJSON(c.Response(), http.StatusOK, out)
	return nil
}
func (h *CatalogHandler) GetRole(c echo.Context) error {
	id, err := catalogID(c)
	if err != nil {
		return catalogBad(c, "invalid role id")
	}
	ctx, cancel := catalogContext(c)
	defer cancel()
	q := models.New(h.pool)
	r, err := q.GetRole(ctx, id)
	if err != nil {
		WriteError(c.Response(), http.StatusNotFound, "role_not_found", "role not found", nil)
		return nil
	}
	a, _ := q.ListRoleAbilityForRole(ctx, id)
	p, _ := q.ListRolePerkForRole(ctx, id)
	WriteJSON(c.Response(), http.StatusOK, roleDTO(r, a, p))
	return nil
}
func (h *CatalogHandler) CreateRole(c echo.Context) error {
	var in catalogRoleInput
	if decodeCatalog(c, &in) != nil {
		return nil
	}
	a, ok := validAlignment(in.Alignment)
	if !ok {
		return catalogBad(c, "alignment must be GOOD, NEUTRAL, or EVIL")
	}
	ctx, cancel := catalogContext(c)
	defer cancel()
	r, err := models.New(h.pool).CreateRole(ctx, models.CreateRoleParams{Name: in.Name, Description: in.Description, Alignment: a})
	if err != nil {
		return catalogFailure(c, "role_create_failed")
	}
	WriteJSON(c.Response(), http.StatusCreated, roleDTO(r, nil, nil))
	return nil
}
func (h *CatalogHandler) UpdateRole(c echo.Context) error {
	id, err := catalogID(c)
	if err != nil {
		return catalogBad(c, "invalid role id")
	}
	var in catalogRoleInput
	if decodeCatalog(c, &in) != nil {
		return nil
	}
	a, ok := validAlignment(in.Alignment)
	if !ok {
		return catalogBad(c, "alignment must be GOOD, NEUTRAL, or EVIL")
	}
	ctx, cancel := catalogContext(c)
	defer cancel()
	r, err := models.New(h.pool).UpdateRole(ctx, models.UpdateRoleParams{ID: id, Name: in.Name, Description: in.Description, Alignment: a})
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return catalogFailure(c, "role_update_failed")
		}
		WriteError(c.Response(), http.StatusNotFound, "role_not_found", "role not found", nil)
		return nil
	}
	WriteJSON(c.Response(), http.StatusOK, roleDTO(r, nil, nil))
	return nil
}
func (h *CatalogHandler) DeleteRole(c echo.Context) error {
	id, err := catalogID(c)
	if err != nil {
		return catalogBad(c, "invalid role id")
	}
	ctx, cancel := catalogContext(c)
	defer cancel()
	if err := models.New(h.pool).DeleteRole(ctx, id); err != nil {
		return catalogFailure(c, "role_delete_failed")
	}
	c.NoContent(http.StatusNoContent)
	return nil
}

func (h *CatalogHandler) ListItems(c echo.Context) error {
	ctx, cancel := catalogContext(c)
	defer cancel()
	q := models.New(h.pool)
	var rows []models.Item
	var err error
	if term := c.QueryParam("q"); term != "" {
		rows, err = q.SearchItemByKeyword(ctx, "%"+term+"%")
	} else {
		rows, err = q.ListItem(ctx)
	}
	if err != nil {
		return catalogFailure(c, "items_unavailable")
	}
	out := make([]catalogItemDTO, 0, len(rows))
	for _, r := range rows {
		out = append(out, itemDTO(r))
	}
	WriteJSON(c.Response(), http.StatusOK, out)
	return nil
}
func (h *CatalogHandler) GetItem(c echo.Context) error {
	id, err := catalogID(c)
	if err != nil {
		return catalogBad(c, "invalid item id")
	}
	ctx, cancel := catalogContext(c)
	defer cancel()
	r, err := models.New(h.pool).GetItem(ctx, id)
	if err != nil {
		WriteError(c.Response(), 404, "item_not_found", "item not found", nil)
		return nil
	}
	WriteJSON(c.Response(), 200, itemDTO(r))
	return nil
}
func (h *CatalogHandler) CreateItem(c echo.Context) error {
	var in catalogItemInput
	if decodeCatalog(c, &in) != nil {
		return nil
	}
	rarity, ok := validRarity(in.Rarity)
	if !ok {
		return catalogBad(c, "invalid rarity")
	}
	ctx, cancel := catalogContext(c)
	defer cancel()
	r, err := models.New(h.pool).CreateItem(ctx, models.CreateItemParams{Name: in.Name, Description: in.Description, Rarity: rarity, Cost: in.Cost})
	if err != nil {
		return catalogFailure(c, "item_create_failed")
	}
	WriteJSON(c.Response(), 201, itemDTO(r))
	return nil
}
func (h *CatalogHandler) UpdateItem(c echo.Context) error {
	id, err := catalogID(c)
	if err != nil {
		return catalogBad(c, "invalid item id")
	}
	var in catalogItemInput
	if decodeCatalog(c, &in) != nil {
		return nil
	}
	rarity, ok := validRarity(in.Rarity)
	if !ok {
		return catalogBad(c, "invalid rarity")
	}
	ctx, cancel := catalogContext(c)
	defer cancel()
	r, err := models.New(h.pool).UpdateItem(ctx, models.UpdateItemParams{ID: id, Name: in.Name, Description: in.Description, Rarity: rarity, Cost: in.Cost})
	if err != nil {
		WriteError(c.Response(), 404, "item_not_found", "item not found", nil)
		return nil
	}
	WriteJSON(c.Response(), 200, itemDTO(r))
	return nil
}
func (h *CatalogHandler) DeleteItem(c echo.Context) error {
	id, err := catalogID(c)
	if err != nil {
		return catalogBad(c, "invalid item id")
	}
	ctx, cancel := catalogContext(c)
	defer cancel()
	if err := models.New(h.pool).DeleteItem(ctx, id); err != nil {
		return catalogFailure(c, "item_delete_failed")
	}
	c.NoContent(204)
	return nil
}

func (h *CatalogHandler) ListAbilities(c echo.Context) error {
	ctx, cancel := catalogContext(c)
	defer cancel()
	q := models.New(h.pool)
	var rows []models.AbilityInfo
	var err error
	if term := c.QueryParam("q"); term != "" {
		rows, err = q.SearchAbilityByKeyword(ctx, "%"+term+"%")
	} else {
		rows, err = q.ListAbilityInfo(ctx)
	}
	if err != nil {
		return catalogFailure(c, "abilities_unavailable")
	}
	out := make([]catalogAbilityDTO, 0, len(rows))
	for _, r := range rows {
		out = append(out, abilityDTO(r))
	}
	WriteJSON(c.Response(), 200, out)
	return nil
}
func (h *CatalogHandler) GetAbility(c echo.Context) error {
	id, err := catalogID(c)
	if err != nil {
		return catalogBad(c, "invalid ability id")
	}
	ctx, cancel := catalogContext(c)
	defer cancel()
	r, err := models.New(h.pool).GetAbilityInfo(ctx, id)
	if err != nil {
		WriteError(c.Response(), 404, "ability_not_found", "ability not found", nil)
		return nil
	}
	WriteJSON(c.Response(), 200, abilityDTO(r))
	return nil
}
func (h *CatalogHandler) CreateAbility(c echo.Context) error {
	var in catalogAbilityInput
	if decodeCatalog(c, &in) != nil {
		return nil
	}
	rarity, ok := validRarity(in.Rarity)
	if !ok {
		return catalogBad(c, "invalid rarity")
	}
	ctx, cancel := catalogContext(c)
	defer cancel()
	r, err := models.New(h.pool).CreateAbilityInfo(ctx, models.CreateAbilityInfoParams{Name: in.Name, Description: in.Description, DefaultCharges: in.DefaultCharges, AnyAbility: in.AnyAbility, Rarity: rarity})
	if err != nil {
		return catalogFailure(c, "ability_create_failed")
	}
	WriteJSON(c.Response(), 201, abilityDTO(r))
	return nil
}
func (h *CatalogHandler) UpdateAbility(c echo.Context) error {
	id, err := catalogID(c)
	if err != nil {
		return catalogBad(c, "invalid ability id")
	}
	var in catalogAbilityInput
	if decodeCatalog(c, &in) != nil {
		return nil
	}
	rarity, ok := validRarity(in.Rarity)
	if !ok {
		return catalogBad(c, "invalid rarity")
	}
	ctx, cancel := catalogContext(c)
	defer cancel()
	r, err := models.New(h.pool).UpdateAbilityInfo(ctx, models.UpdateAbilityInfoParams{ID: id, Name: in.Name, Description: in.Description, DefaultCharges: in.DefaultCharges, AnyAbility: in.AnyAbility, Rarity: rarity})
	if err != nil {
		WriteError(c.Response(), 404, "ability_not_found", "ability not found", nil)
		return nil
	}
	WriteJSON(c.Response(), 200, abilityDTO(r))
	return nil
}
func (h *CatalogHandler) DeleteAbility(c echo.Context) error {
	id, err := catalogID(c)
	if err != nil {
		return catalogBad(c, "invalid ability id")
	}
	ctx, cancel := catalogContext(c)
	defer cancel()
	if err := models.New(h.pool).DeleteAbilityInfo(ctx, id); err != nil {
		return catalogFailure(c, "ability_delete_failed")
	}
	c.NoContent(204)
	return nil
}

func (h *CatalogHandler) ListStatuses(c echo.Context) error {
	ctx, cancel := catalogContext(c)
	defer cancel()
	q := models.New(h.pool)
	var rows []models.Status
	var err error
	if term := c.QueryParam("q"); term != "" {
		rows, err = q.SearchStatusByKeyword(ctx, "%"+term+"%")
	} else {
		rows, err = q.ListStatus(ctx)
	}
	if err != nil {
		return catalogFailure(c, "statuses_unavailable")
	}
	out := make([]catalogStatusDTO, 0, len(rows))
	for _, r := range rows {
		out = append(out, statusDTO(r))
	}
	WriteJSON(c.Response(), 200, out)
	return nil
}
func (h *CatalogHandler) GetStatus(c echo.Context) error {
	id, err := catalogID(c)
	if err != nil {
		return catalogBad(c, "invalid status id")
	}
	ctx, cancel := catalogContext(c)
	defer cancel()
	rows, err := models.New(h.pool).ListStatus(ctx)
	if err != nil {
		return catalogFailure(c, "statuses_unavailable")
	}
	for _, r := range rows {
		if r.ID == id {
			WriteJSON(c.Response(), 200, statusDTO(r))
			return nil
		}
	}
	WriteError(c.Response(), 404, "status_not_found", "status not found", nil)
	return nil
}
func (h *CatalogHandler) CreateStatus(c echo.Context) error {
	var in catalogStatusInput
	if decodeCatalog(c, &in) != nil {
		return nil
	}
	ctx, cancel := catalogContext(c)
	defer cancel()
	r, err := models.New(h.pool).CreateStatus(ctx, models.CreateStatusParams{Name: in.Name, Description: in.Description})
	if err != nil {
		return catalogFailure(c, "status_create_failed")
	}
	if in.HourDuration > 0 {
		r, err = models.New(h.pool).UpdateStatus(ctx, models.UpdateStatusParams{ID: r.ID, Name: r.Name, Description: r.Description, HourDuration: in.HourDuration})
		if err != nil {
			return catalogFailure(c, "status_create_failed")
		}
	}
	WriteJSON(c.Response(), 201, statusDTO(r))
	return nil
}
func (h *CatalogHandler) UpdateStatus(c echo.Context) error {
	id, err := catalogID(c)
	if err != nil {
		return catalogBad(c, "invalid status id")
	}
	var in catalogStatusInput
	if decodeCatalog(c, &in) != nil {
		return nil
	}
	ctx, cancel := catalogContext(c)
	defer cancel()
	r, err := models.New(h.pool).UpdateStatus(ctx, models.UpdateStatusParams{ID: id, Name: in.Name, Description: in.Description, HourDuration: in.HourDuration})
	if err != nil {
		WriteError(c.Response(), 404, "status_not_found", "status not found", nil)
		return nil
	}
	WriteJSON(c.Response(), 200, statusDTO(r))
	return nil
}
func (h *CatalogHandler) DeleteStatus(c echo.Context) error {
	id, err := catalogID(c)
	if err != nil {
		return catalogBad(c, "invalid status id")
	}
	ctx, cancel := catalogContext(c)
	defer cancel()
	if err := models.New(h.pool).DeleteStatus(ctx, id); err != nil {
		return catalogFailure(c, "status_delete_failed")
	}
	c.NoContent(204)
	return nil
}
