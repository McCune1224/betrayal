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

// RolesHandler handles role-related requests
type RolesHandler struct {
	dbPool *pgxpool.Pool
}

// NewRolesHandler creates a new RolesHandler
func NewRolesHandler(pool *pgxpool.Pool) *RolesHandler {
	return &RolesHandler{dbPool: pool}
}

// List handles GET /roles - main roles page
func (h *RolesHandler) List(c echo.Context) error {
	return render(c, http.StatusOK, pages.Roles())
}

// Search handles GET /roles/search - HTMX partial for live search
func (h *RolesHandler) Search(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()

	q := models.New(h.dbPool)
	searchQuery := c.QueryParam("q")

	var roles []models.Role
	var err error

	if searchQuery == "" {
		// Return empty results if no search query
		roles = []models.Role{}
	} else {
		roles, err = q.SearchRoleByName(ctx, searchQuery)
		if err != nil {
			roles = []models.Role{}
		}
	}

	// Convert to view model
	roleData := make([]partials.RoleSearchRow, len(roles))
	for i, r := range roles {
		roleData[i] = partials.RoleSearchRow{
			ID:          r.ID,
			Name:        r.Name,
			Description: r.Description,
			Alignment:   string(r.Alignment),
		}
	}

	return render(c, http.StatusOK, partials.RoleSearchResults(roleData))
}

// Detail handles GET /roles/:id - role detail/edit page
func (h *RolesHandler) Detail(c echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid role ID")
	}

	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()

	q := models.New(h.dbPool)

	role, err := q.GetRole(ctx, int32(id))
	if err != nil {
		return c.String(http.StatusNotFound, "Role not found")
	}

	// Get abilities for this role
	abilities, _ := q.ListRoleAbilityForRole(ctx, role.ID)
	abilityData := make([]pages.RoleAbilityRow, len(abilities))
	for i, a := range abilities {
		abilityData[i] = pages.RoleAbilityRow{
			ID:             a.ID,
			Name:           a.Name,
			Description:    a.Description,
			DefaultCharges: a.DefaultCharges,
			AnyAbility:     a.AnyAbility,
			Rarity:         string(a.Rarity),
		}
	}

	// Get perks for this role
	perks, _ := q.ListRolePerkForRole(ctx, role.ID)
	perkData := make([]pages.RolePerkRow, len(perks))
	for i, p := range perks {
		perkData[i] = pages.RolePerkRow{
			ID:          p.ID,
			Name:        p.Name,
			Description: p.Description,
		}
	}

	data := pages.RoleDetailData{
		ID:          role.ID,
		Name:        role.Name,
		Description: role.Description,
		Alignment:   string(role.Alignment),
		Abilities:   abilityData,
		Perks:       perkData,
	}

	return render(c, http.StatusOK, pages.RoleDetail(data))
}

// Update handles PUT /roles/:id - update role fields
func (h *RolesHandler) Update(c echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid role ID")
	}

	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()

	q := models.New(h.dbPool)

	name := c.FormValue("name")
	description := c.FormValue("description")
	alignment := c.FormValue("alignment")

	// Validate alignment
	var alignmentEnum models.Alignment
	switch alignment {
	case "GOOD":
		alignmentEnum = models.AlignmentGOOD
	case "NEUTRAL":
		alignmentEnum = models.AlignmentNEUTRAL
	case "EVIL":
		alignmentEnum = models.AlignmentEVIL
	default:
		c.Response().Header().Set("HX-Trigger", `{"showToast": {"message": "Invalid alignment value", "type": "error"}}`)
		return c.String(http.StatusBadRequest, "Invalid alignment")
	}

	_, err = q.UpdateRole(ctx, models.UpdateRoleParams{
		ID:          int32(id),
		Name:        name,
		Description: description,
		Alignment:   alignmentEnum,
	})
	if err != nil {
		c.Response().Header().Set("HX-Trigger", `{"showToast": {"message": "Failed to update role", "type": "error"}}`)
		return c.String(http.StatusInternalServerError, "Failed to update role")
	}

	c.Response().Header().Set("HX-Trigger", `{"showToast": {"message": "Role updated successfully", "type": "success"}}`)
	return c.String(http.StatusOK, "")
}

// ListAbilities handles GET /roles/:id/abilities - HTMX partial
func (h *RolesHandler) ListAbilities(c echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid role ID")
	}

	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()

	q := models.New(h.dbPool)

	abilities, _ := q.ListRoleAbilityForRole(ctx, int32(id))
	abilityData := make([]partials.AbilityRow, len(abilities))
	for i, a := range abilities {
		abilityData[i] = partials.AbilityRow{
			ID:             a.ID,
			Name:           a.Name,
			Description:    a.Description,
			DefaultCharges: a.DefaultCharges,
			AnyAbility:     a.AnyAbility,
			Rarity:         string(a.Rarity),
		}
	}

	return render(c, http.StatusOK, partials.RoleAbilities(int32(id), abilityData))
}

// UpdateAbility handles PUT /roles/:id/abilities/:abilityId - update ability details
func (h *RolesHandler) UpdateAbility(c echo.Context) error {
	abilityIdStr := c.Param("abilityId")
	abilityId, err := strconv.ParseInt(abilityIdStr, 10, 32)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid ability ID")
	}

	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()

	q := models.New(h.dbPool)

	name := c.FormValue("name")
	description := c.FormValue("description")
	chargesStr := c.FormValue("default_charges")
	anyAbilityStr := c.FormValue("any_ability")
	rarity := c.FormValue("rarity")

	charges, err := strconv.ParseInt(chargesStr, 10, 32)
	if err != nil {
		charges = 0
	}

	anyAbility := anyAbilityStr == "true" || anyAbilityStr == "on"

	// Validate rarity
	var rarityEnum models.Rarity
	switch rarity {
	case "COMMON":
		rarityEnum = models.RarityCOMMON
	case "UNCOMMON":
		rarityEnum = models.RarityUNCOMMON
	case "RARE":
		rarityEnum = models.RarityRARE
	case "EPIC":
		rarityEnum = models.RarityEPIC
	case "LEGENDARY":
		rarityEnum = models.RarityLEGENDARY
	case "MYTHICAL":
		rarityEnum = models.RarityMYTHICAL
	case "ROLE_SPECIFIC":
		rarityEnum = models.RarityROLESPECIFIC
	case "UNIQUE":
		rarityEnum = models.RarityUNIQUE
	default:
		c.Response().Header().Set("HX-Trigger", `{"showToast": {"message": "Invalid rarity value", "type": "error"}}`)
		return c.String(http.StatusBadRequest, "Invalid rarity")
	}

	_, err = q.UpdateAbilityInfo(ctx, models.UpdateAbilityInfoParams{
		ID:             int32(abilityId),
		Name:           name,
		Description:    description,
		DefaultCharges: int32(charges),
		AnyAbility:     anyAbility,
		Rarity:         rarityEnum,
	})
	if err != nil {
		c.Response().Header().Set("HX-Trigger", `{"showToast": {"message": "Failed to update ability", "type": "error"}}`)
		return c.String(http.StatusInternalServerError, "Failed to update ability")
	}

	c.Response().Header().Set("HX-Trigger", `{"showToast": {"message": "Ability updated successfully", "type": "success"}}`)
	return c.String(http.StatusOK, "")
}

// RemoveAbility handles DELETE /roles/:id/abilities/:abilityId - remove ability from role
func (h *RolesHandler) RemoveAbility(c echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid role ID")
	}

	abilityIdStr := c.Param("abilityId")
	abilityId, err := strconv.ParseInt(abilityIdStr, 10, 32)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid ability ID")
	}

	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()

	q := models.New(h.dbPool)

	err = q.DeleteRoleAbilityJoin(ctx, models.DeleteRoleAbilityJoinParams{
		RoleID:    int32(id),
		AbilityID: int32(abilityId),
	})
	if err != nil {
		c.Response().Header().Set("HX-Trigger", `{"showToast": {"message": "Failed to remove ability", "type": "error"}}`)
		return c.String(http.StatusInternalServerError, "Failed to remove ability")
	}

	c.Response().Header().Set("HX-Trigger", `{"showToast": {"message": "Ability removed from role", "type": "success"}}`)
	return c.String(http.StatusOK, "")
}

// ListPerks handles GET /roles/:id/perks - HTMX partial
func (h *RolesHandler) ListPerks(c echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid role ID")
	}

	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()

	q := models.New(h.dbPool)

	perks, _ := q.ListRolePerkForRole(ctx, int32(id))
	perkData := make([]partials.PerkRow, len(perks))
	for i, p := range perks {
		perkData[i] = partials.PerkRow{
			ID:          p.ID,
			Name:        p.Name,
			Description: p.Description,
		}
	}

	return render(c, http.StatusOK, partials.RolePerks(int32(id), perkData))
}

// UpdatePerk handles PUT /roles/:id/perks/:perkId - update perk details
func (h *RolesHandler) UpdatePerk(c echo.Context) error {
	perkIdStr := c.Param("perkId")
	perkId, err := strconv.ParseInt(perkIdStr, 10, 32)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid perk ID")
	}

	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()

	q := models.New(h.dbPool)

	name := c.FormValue("name")
	description := c.FormValue("description")

	_, err = q.UpdatePerkInfo(ctx, models.UpdatePerkInfoParams{
		ID:          int32(perkId),
		Name:        name,
		Description: description,
	})
	if err != nil {
		c.Response().Header().Set("HX-Trigger", `{"showToast": {"message": "Failed to update perk", "type": "error"}}`)
		return c.String(http.StatusInternalServerError, "Failed to update perk")
	}

	c.Response().Header().Set("HX-Trigger", `{"showToast": {"message": "Perk updated successfully", "type": "success"}}`)
	return c.String(http.StatusOK, "")
}

// RemovePerk handles DELETE /roles/:id/perks/:perkId - remove perk from role
func (h *RolesHandler) RemovePerk(c echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid role ID")
	}

	perkIdStr := c.Param("perkId")
	perkId, err := strconv.ParseInt(perkIdStr, 10, 32)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid perk ID")
	}

	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()

	q := models.New(h.dbPool)

	err = q.DeleteRolePerkJoin(ctx, models.DeleteRolePerkJoinParams{
		RoleID: int32(id),
		PerkID: int32(perkId),
	})
	if err != nil {
		c.Response().Header().Set("HX-Trigger", `{"showToast": {"message": "Failed to remove perk", "type": "error"}}`)
		return c.String(http.StatusInternalServerError, "Failed to remove perk")
	}

	c.Response().Header().Set("HX-Trigger", `{"showToast": {"message": "Perk removed from role", "type": "success"}}`)
	return c.String(http.StatusOK, "")
}
