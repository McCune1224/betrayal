package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
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
	ID             int32    `json:"id"`
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	DefaultCharges int32    `json:"default_charges"`
	AnyAbility     bool     `json:"any_ability"`
	Rarity         string   `json:"rarity"`
	Categories     []string `json:"categories"`
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
	ID          int32    `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Rarity      string   `json:"rarity"`
	Cost        int32    `json:"cost"`
	Categories  []string `json:"categories"`
}
type catalogStatusDTO struct {
	ID           int32  `json:"id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	HourDuration int32  `json:"hour_duration"`
}
type catalogCategoryDTO struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
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
type catalogPerkInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}
type catalogCategoryInput struct {
	Name string `json:"name"`
}
type catalogCategoryAssignInput struct {
	Category string `json:"category"`
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

func (h *CatalogHandler) roleDTO(ctx context.Context, r models.Role, abilities []models.AbilityInfo, perks []models.PerkInfo) catalogRoleDTO {
	d := catalogRoleDTO{ID: r.ID, Name: r.Name, Description: r.Description, Alignment: string(r.Alignment), Abilities: []catalogAbilityDTO{}, Perks: []catalogPerkDTO{}}
	for _, a := range abilities {
		d.Abilities = append(d.Abilities, catalogAbilityDTO{a.ID, a.Name, a.Description, a.DefaultCharges, a.AnyAbility, string(a.Rarity), []string{}})
	}
	for _, p := range perks {
		d.Perks = append(d.Perks, catalogPerkDTO{p.ID, p.Name, p.Description})
	}
	return d
}
func (h *CatalogHandler) abilityDTO(ctx context.Context, a models.AbilityInfo) catalogAbilityDTO {
	d := catalogAbilityDTO{a.ID, a.Name, a.Description, a.DefaultCharges, a.AnyAbility, string(a.Rarity), []string{}}
	if names, err := models.New(h.pool).ListAbilityCategoryNames(ctx, a.ID); err == nil {
		d.Categories = names
	}
	return d
}
func (h *CatalogHandler) itemDTO(ctx context.Context, i models.Item) catalogItemDTO {
	d := catalogItemDTO{i.ID, i.Name, i.Description, string(i.Rarity), i.Cost, []string{}}
	if names, err := models.New(h.pool).ListItemCategoryNames(ctx, i.ID); err == nil {
		d.Categories = names
	}
	return d
}
func statusDTO(s models.Status) catalogStatusDTO {
	return catalogStatusDTO{s.ID, s.Name, s.Description, s.HourDuration}
}
func categoryDTO(c models.Category) catalogCategoryDTO {
	return catalogCategoryDTO{c.ID, c.Name}
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
		out = append(out, h.roleDTO(ctx, r, abilities, perks))
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
	WriteJSON(c.Response(), http.StatusOK, h.roleDTO(ctx, r, a, p))
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
	WriteJSON(c.Response(), http.StatusCreated, h.roleDTO(ctx, r, nil, nil))
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
	WriteJSON(c.Response(), http.StatusOK, h.roleDTO(ctx, r, nil, nil))
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
		out = append(out, h.itemDTO(ctx, r))
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
	WriteJSON(c.Response(), 200, h.itemDTO(ctx, r))
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
	WriteJSON(c.Response(), 201, h.itemDTO(ctx, r))
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
	WriteJSON(c.Response(), 200, h.itemDTO(ctx, r))
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
		out = append(out, h.abilityDTO(ctx, r))
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
	WriteJSON(c.Response(), 200, h.abilityDTO(ctx, r))
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
	WriteJSON(c.Response(), 201, h.abilityDTO(ctx, r))
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
	WriteJSON(c.Response(), 200, h.abilityDTO(ctx, r))
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

func perkDTO(p models.PerkInfo) catalogPerkDTO {
	return catalogPerkDTO{p.ID, p.Name, p.Description}
}

func (h *CatalogHandler) ListPerks(c echo.Context) error {
	ctx, cancel := catalogContext(c)
	defer cancel()
	rows, err := models.New(h.pool).ListPerkInfo(ctx)
	if err != nil {
		return catalogFailure(c, "perks_unavailable")
	}
	out := make([]catalogPerkDTO, 0, len(rows))
	for _, r := range rows {
		out = append(out, perkDTO(r))
	}
	WriteJSON(c.Response(), 200, out)
	return nil
}
func (h *CatalogHandler) GetPerk(c echo.Context) error {
	id, err := catalogID(c)
	if err != nil {
		return catalogBad(c, "invalid perk id")
	}
	ctx, cancel := catalogContext(c)
	defer cancel()
	r, err := models.New(h.pool).GetPerkInfo(ctx, id)
	if err != nil {
		WriteError(c.Response(), 404, "perk_not_found", "perk not found", nil)
		return nil
	}
	WriteJSON(c.Response(), 200, perkDTO(r))
	return nil
}
func (h *CatalogHandler) CreatePerk(c echo.Context) error {
	var in catalogPerkInput
	if decodeCatalog(c, &in) != nil {
		return nil
	}
	ctx, cancel := catalogContext(c)
	defer cancel()
	r, err := models.New(h.pool).CreatePerkInfo(ctx, models.CreatePerkInfoParams{Name: in.Name, Description: in.Description})
	if err != nil {
		return catalogFailure(c, "perk_create_failed")
	}
	WriteJSON(c.Response(), 201, perkDTO(r))
	return nil
}
func (h *CatalogHandler) UpdatePerk(c echo.Context) error {
	id, err := catalogID(c)
	if err != nil {
		return catalogBad(c, "invalid perk id")
	}
	var in catalogPerkInput
	if decodeCatalog(c, &in) != nil {
		return nil
	}
	ctx, cancel := catalogContext(c)
	defer cancel()
	r, err := models.New(h.pool).UpdatePerkInfo(ctx, models.UpdatePerkInfoParams{ID: id, Name: in.Name, Description: in.Description})
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return catalogFailure(c, "perk_update_failed")
		}
		WriteError(c.Response(), 404, "perk_not_found", "perk not found", nil)
		return nil
	}
	WriteJSON(c.Response(), 200, perkDTO(r))
	return nil
}
func (h *CatalogHandler) DeletePerk(c echo.Context) error {
	id, err := catalogID(c)
	if err != nil {
		return catalogBad(c, "invalid perk id")
	}
	ctx, cancel := catalogContext(c)
	defer cancel()
	if err := models.New(h.pool).DeletePerkInfo(ctx, id); err != nil {
		return catalogFailure(c, "perk_delete_failed")
	}
	c.NoContent(204)
	return nil
}

func (h *CatalogHandler) ListCategories(c echo.Context) error {
	ctx, cancel := catalogContext(c)
	defer cancel()
	rows, err := models.New(h.pool).ListCategory(ctx)
	if err != nil {
		return catalogFailure(c, "categories_unavailable")
	}
	out := make([]catalogCategoryDTO, 0, len(rows))
	for _, r := range rows {
		out = append(out, categoryDTO(r))
	}
	WriteJSON(c.Response(), 200, out)
	return nil
}
func (h *CatalogHandler) GetCategory(c echo.Context) error {
	id, err := catalogID(c)
	if err != nil {
		return catalogBad(c, "invalid category id")
	}
	ctx, cancel := catalogContext(c)
	defer cancel()
	r, err := models.New(h.pool).GetCategory(ctx, id)
	if err != nil {
		WriteError(c.Response(), 404, "category_not_found", "category not found", nil)
		return nil
	}
	WriteJSON(c.Response(), 200, categoryDTO(r))
	return nil
}
func (h *CatalogHandler) CreateCategory(c echo.Context) error {
	var in catalogCategoryInput
	if decodeCatalog(c, &in) != nil {
		return nil
	}
	ctx, cancel := catalogContext(c)
	defer cancel()
	r, err := models.New(h.pool).CreateCategory(ctx, in.Name)
	if err != nil {
		return catalogFailure(c, "category_create_failed")
	}
	WriteJSON(c.Response(), 201, categoryDTO(r))
	return nil
}
func (h *CatalogHandler) UpdateCategory(c echo.Context) error {
	id, err := catalogID(c)
	if err != nil {
		return catalogBad(c, "invalid category id")
	}
	var in catalogCategoryInput
	if decodeCatalog(c, &in) != nil {
		return nil
	}
	ctx, cancel := catalogContext(c)
	defer cancel()
	r, err := models.New(h.pool).UpdateCategory(ctx, models.UpdateCategoryParams{ID: id, Name: in.Name})
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return catalogFailure(c, "category_update_failed")
		}
		WriteError(c.Response(), 404, "category_not_found", "category not found", nil)
		return nil
	}
	WriteJSON(c.Response(), 200, categoryDTO(r))
	return nil
}
func (h *CatalogHandler) DeleteCategory(c echo.Context) error {
	id, err := catalogID(c)
	if err != nil {
		return catalogBad(c, "invalid category id")
	}
	ctx, cancel := catalogContext(c)
	defer cancel()
	if err := models.New(h.pool).DeleteCategory(ctx, id); err != nil {
		return catalogFailure(c, "category_delete_failed")
	}
	c.NoContent(204)
	return nil
}

// resolveCategoryName resolves the category for item/ability assignment by
// fuzzy name, mirroring the player-role resolver.
func (h *CatalogHandler) resolveCategoryName(ctx context.Context, name string) (models.Category, bool) {
	if strings.TrimSpace(name) == "" {
		return models.Category{}, false
	}
	c, err := models.New(h.pool).GetCategoryByFuzzy(ctx, name)
	return c, err == nil
}

type catalogAssignResponse struct {
	Categories []string `json:"categories"`
}

func (h *CatalogHandler) assignCategory(c echo.Context, entityLabel string, entityExists func(ctx context.Context, q *models.Queries) error, link func(ctx context.Context, q *models.Queries, categoryID int32) error, current func(ctx context.Context, q *models.Queries) ([]string, error)) error {
	ctx, cancel := catalogContext(c)
	defer cancel()
	q := models.New(h.pool)
	if err := entityExists(ctx, q); err != nil {
		WriteError(c.Response(), 404, entityLabel+"_not_found", entityLabel+" not found", nil)
		return nil
	}
	var in catalogCategoryAssignInput
	if decodeCatalog(c, &in) != nil {
		return nil
	}
	cat, ok := h.resolveCategoryName(ctx, in.Category)
	if !ok {
		WriteError(c.Response(), 400, "category_not_found", "category not found", nil)
		return nil
	}
	if err := link(ctx, q, cat.ID); err != nil {
		return catalogFailure(c, entityLabel+"_category_update_failed")
	}
	names, err := current(ctx, q)
	if err != nil {
		return catalogFailure(c, entityLabel+"_unavailable")
	}
	WriteJSON(c.Response(), 200, catalogAssignResponse{Categories: names})
	return nil
}

func (h *CatalogHandler) ItemAddCategory(c echo.Context) error {
	id, err := catalogID(c)
	if err != nil {
		return catalogBad(c, "invalid item id")
	}
	return h.assignCategory(c, "item",
		func(ctx context.Context, q *models.Queries) error { _, err := q.GetItem(ctx, id); return err },
		func(ctx context.Context, q *models.Queries, categoryID int32) error {
			return q.CreateItemCategoryJoin(ctx, models.CreateItemCategoryJoinParams{ItemID: id, CategoryID: categoryID})
		},
		func(ctx context.Context, q *models.Queries) ([]string, error) {
			return q.ListItemCategoryNames(ctx, id)
		},
	)
}

func (h *CatalogHandler) ItemRemoveCategory(c echo.Context) error {
	id, err := catalogID(c)
	if err != nil {
		return catalogBad(c, "invalid item id")
	}
	cid, err := strconv.ParseInt(c.Param("categoryID"), 10, 32)
	if err != nil || cid <= 0 {
		return catalogBad(c, "invalid category id")
	}
	ctx, cancel := catalogContext(c)
	defer cancel()
	q := models.New(h.pool)
	if _, err := q.GetItem(ctx, id); err != nil {
		WriteError(c.Response(), 404, "item_not_found", "item not found", nil)
		return nil
	}
	if err := q.DeleteItemCategoryJoin(ctx, models.DeleteItemCategoryJoinParams{ItemID: id, CategoryID: int32(cid)}); err != nil {
		return catalogFailure(c, "item_category_update_failed")
	}
	c.NoContent(204)
	return nil
}

func (h *CatalogHandler) AbilityAddCategory(c echo.Context) error {
	id, err := catalogID(c)
	if err != nil {
		return catalogBad(c, "invalid ability id")
	}
	return h.assignCategory(c, "ability",
		func(ctx context.Context, q *models.Queries) error { _, err := q.GetAbilityInfo(ctx, id); return err },
		func(ctx context.Context, q *models.Queries, categoryID int32) error {
			return q.CreateAbilityCategoryJoin(ctx, models.CreateAbilityCategoryJoinParams{AbilityID: id, CategoryID: categoryID})
		},
		func(ctx context.Context, q *models.Queries) ([]string, error) {
			return q.ListAbilityCategoryNames(ctx, id)
		},
	)
}

func (h *CatalogHandler) AbilityRemoveCategory(c echo.Context) error {
	id, err := catalogID(c)
	if err != nil {
		return catalogBad(c, "invalid ability id")
	}
	cid, err := strconv.ParseInt(c.Param("categoryID"), 10, 32)
	if err != nil || cid <= 0 {
		return catalogBad(c, "invalid category id")
	}
	ctx, cancel := catalogContext(c)
	defer cancel()
	q := models.New(h.pool)
	if _, err := q.GetAbilityInfo(ctx, id); err != nil {
		WriteError(c.Response(), 404, "ability_not_found", "ability not found", nil)
		return nil
	}
	if err := q.DeleteAbilityCategoryJoin(ctx, models.DeleteAbilityCategoryJoinParams{AbilityID: id, CategoryID: int32(cid)}); err != nil {
		return catalogFailure(c, "ability_category_update_failed")
	}
	c.NoContent(204)
	return nil
}
