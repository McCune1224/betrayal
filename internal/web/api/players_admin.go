package api

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
	"github.com/mccune1224/betrayal/internal/models"
	"github.com/mccune1224/betrayal/internal/services/inventory"
	"github.com/mccune1224/betrayal/internal/util"
)

type playerDTO struct {
	ID        int64  `json:"id"`
	Alive     bool   `json:"alive"`
	Coins     int32  `json:"coins"`
	Luck      int32  `json:"luck"`
	ItemLimit int32  `json:"item_limit"`
	Alignment string `json:"alignment"`
	Role      string `json:"role"`
}
type playerItemDTO struct {
	ID          int32  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Quantity    int32  `json:"quantity"`
	Cost        int32  `json:"cost"`
}
type playerAbilityDTO struct {
	ID          int32  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Quantity    int32  `json:"quantity"`
}
type playerStatusDTO struct {
	ID          int32  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Quantity    int32  `json:"quantity"`
}
type playerImmunityDTO struct {
	ID          int32  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	OneTime     bool   `json:"one_time"`
}
type playerPerkDTO struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
}
type playerNoteDTO struct {
	ID       int32  `json:"id"`
	Position int32  `json:"position"`
	Info     string `json:"info"`
}
type playerDetailDTO struct {
	playerDTO
	Items      []playerItemDTO     `json:"items"`
	Abilities  []playerAbilityDTO  `json:"abilities"`
	Statuses   []playerStatusDTO   `json:"statuses"`
	Immunities []playerImmunityDTO `json:"immunities"`
	Perks      []playerPerkDTO     `json:"perks"`
	Notes      []playerNoteDTO     `json:"notes"`
}
type playerCreateInput struct {
	ID   json.RawMessage `json:"id"`
	Role string          `json:"role"`
}
type playerUpdateInput struct {
	Coins     *int32  `json:"coins"`
	Luck      *int32  `json:"luck"`
	Alive     *bool   `json:"alive"`
	Alignment *string `json:"alignment"`
	ItemLimit *int32  `json:"item_limit"`
	Role      *string `json:"role"`
}
type playerMutationInput struct {
	Name     string `json:"name"`
	Quantity int32  `json:"quantity"`
	OneTime  bool   `json:"one_time"`
	Position int32  `json:"position"`
	Info     string `json:"info"`
	NoteID   int32  `json:"note_id"`
}

func parsePlayerID(raw json.RawMessage) (int64, error) {
	if len(raw) == 0 {
		return 0, fmt.Errorf("missing player ID")
	}
	var text string
	if raw[0] == '"' {
		if err := json.Unmarshal(raw, &text); err != nil {
			return 0, err
		}
	} else {
		text = string(raw)
	}
	id, err := strconv.ParseInt(strings.TrimSpace(text), 10, 64)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("invalid player ID")
	}
	return id, nil
}

func playerID(c echo.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	return id, err == nil && id > 0
}
func decodePlayer(c echo.Context, dst any) error {
	if err := json.NewDecoder(c.Request().Body).Decode(dst); err != nil {
		WriteError(c.Response(), 400, "invalid_json", "request body must be valid JSON", nil)
		return err
	}
	return nil
}
func (h *PlayersHandler) getPlayer(ctx context.Context, id int64) (models.Player, string, error) {
	q := models.New(h.pool)
	p, err := q.GetPlayer(ctx, id)
	if err != nil {
		return p, "", err
	}
	role := "Unknown"
	if p.RoleID.Valid {
		if r, e := q.GetRole(ctx, p.RoleID.Int32); e == nil {
			role = r.Name
		}
	}
	return p, role, nil
}
func playerDTOFor(p models.Player, role string) playerDTO {
	return playerDTO{p.ID, p.Alive, p.Coins, p.Luck, p.ItemLimit, string(p.Alignment), role}
}

func (h *PlayersHandler) Detail(c echo.Context) error {
	id, ok := playerID(c)
	if !ok {
		WriteError(c.Response(), 400, "invalid_player_id", "player ID must be a positive integer", nil)
		return nil
	}
	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()
	p, role, err := h.getPlayer(ctx, id)
	if err != nil {
		WriteError(c.Response(), 404, "player_not_found", "player not found", nil)
		return nil
	}
	q := models.New(h.pool)
	items, err := q.ListPlayerItemInventory(ctx, id)
	if err != nil {
		return playerFailure(c)
	}
	abilities, err := q.ListPlayerAbilityInventory(ctx, id)
	if err != nil {
		return playerFailure(c)
	}
	statuses, err := q.ListPlayerStatusInventory(ctx, id)
	if err != nil {
		return playerFailure(c)
	}
	immunities, err := q.ListPlayerImmunity(ctx, id)
	if err != nil {
		return playerFailure(c)
	}
	perks, err := q.ListPlayerPerk(ctx, id)
	if err != nil {
		return playerFailure(c)
	}
	notes, err := q.ListPlayerNote(ctx, id)
	if err != nil {
		return playerFailure(c)
	}
	d := playerDetailDTO{playerDTO: playerDTOFor(p, role), Items: make([]playerItemDTO, 0), Abilities: make([]playerAbilityDTO, 0), Statuses: make([]playerStatusDTO, 0), Immunities: make([]playerImmunityDTO, 0), Perks: make([]playerPerkDTO, 0), Notes: make([]playerNoteDTO, 0)}
	for _, x := range items {
		d.Items = append(d.Items, playerItemDTO{x.ID, x.Name, x.Description, x.Quantity, x.Cost})
	}
	for _, x := range abilities {
		d.Abilities = append(d.Abilities, playerAbilityDTO{x.ID, x.Name, x.Description, x.Quantity})
	}
	for _, x := range statuses {
		d.Statuses = append(d.Statuses, playerStatusDTO{x.ID, x.Name, x.Description, x.Quantity})
	}
	for _, x := range immunities {
		d.Immunities = append(d.Immunities, playerImmunityDTO{x.ID, x.Name, x.Description, x.OneTime})
	}
	for _, x := range perks {
		d.Perks = append(d.Perks, playerPerkDTO{x.ID, x.Name})
	}
	for _, x := range notes {
		d.Notes = append(d.Notes, playerNoteDTO{x.NoteID, x.Position, x.Info})
	}
	WriteJSON(c.Response(), 200, d)
	return nil
}
func playerFailure(c echo.Context) error {
	WriteError(c.Response(), 500, "player_unavailable", "could not load player", nil)
	return nil
}
func (h *PlayersHandler) Create(c echo.Context) error {
	var in playerCreateInput
	if decodePlayer(c, &in) != nil {
		return nil
	}
	id, err := parsePlayerID(in.ID)
	if err != nil || strings.TrimSpace(in.Role) == "" {
		WriteError(c.Response(), 400, "invalid_player_input", "Discord member and role are required", nil)
		return nil
	}
	ctx, cn := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cn()
	q := models.New(h.pool)
	r, err := q.GetRoleByFuzzy(ctx, in.Role)
	if err != nil {
		WriteError(c.Response(), 400, "role_not_found", "role not found", nil)
		return nil
	}
	n, _ := util.Numeric(0)
	coins, luck, limit := int32(200), int32(0), int32(4)
	for key, dst := range map[string]*int32{"default_coins": &coins, "default_luck": &luck, "default_item_limit": &limit} {
		if v, e := q.GetGameConfig(ctx, key); e == nil {
			if x, e := strconv.ParseInt(v, 10, 32); e == nil {
				*dst = int32(x)
			}
		}
	}
	p, err := q.CreatePlayer(ctx, models.CreatePlayerParams{ID: id, RoleID: pgtype.Int4{Int32: r.ID, Valid: true}, Alive: true, Coins: coins, CoinBonus: n, Luck: luck, ItemLimit: limit, Alignment: r.Alignment})
	if err != nil {
		WriteError(c.Response(), 400, "player_create_failed", "could not create player", nil)
		return nil
	}
	WriteJSON(c.Response(), 201, playerDTOFor(p, r.Name))
	return nil
}
func (h *PlayersHandler) Update(c echo.Context) error {
	id, ok := playerID(c)
	if !ok {
		WriteError(c.Response(), 400, "invalid_player_id", "invalid player ID", nil)
		return nil
	}
	var in playerUpdateInput
	if decodePlayer(c, &in) != nil {
		return nil
	}
	ctx, cn := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cn()
	p, role, err := h.getPlayer(ctx, id)
	if err != nil {
		WriteError(c.Response(), 404, "player_not_found", "player not found", nil)
		return nil
	}
	q := models.New(h.pool)
	if in.Coins != nil {
		p.Coins = *in.Coins
	}
	if in.Luck != nil {
		p.Luck = *in.Luck
	}
	if in.Alive != nil {
		p.Alive = *in.Alive
	}
	if in.ItemLimit != nil {
		if *in.ItemLimit < 0 {
			WriteError(c.Response(), 400, "invalid_player_input", "item limit must be non-negative", nil)
			return nil
		}
		p.ItemLimit = *in.ItemLimit
	}
	if in.Alignment != nil {
		a := models.Alignment(*in.Alignment)
		if a != models.AlignmentGOOD && a != models.AlignmentEVIL && a != models.AlignmentNEUTRAL {
			WriteError(c.Response(), 400, "invalid_player_input", "invalid alignment", nil)
			return nil
		}
		p.Alignment = a
	}
	if in.Role != nil {
		r, e := q.GetRoleByFuzzy(ctx, *in.Role)
		if e != nil {
			WriteError(c.Response(), 400, "role_not_found", "role not found", nil)
			return nil
		}
		p.RoleID = pgtype.Int4{Int32: r.ID, Valid: true}
		role = r.Name
	}
	p, err = q.UpdatePlayer(ctx, models.UpdatePlayerParams{ID: p.ID, RoleID: p.RoleID, Alive: p.Alive, Coins: p.Coins, CoinBonus: p.CoinBonus, Luck: p.Luck, ItemLimit: p.ItemLimit, Alignment: p.Alignment})
	if err != nil {
		WriteError(c.Response(), 400, "player_update_failed", "could not update player", nil)
		return nil
	}
	WriteJSON(c.Response(), 200, playerDTOFor(p, role))
	return nil
}

// UpdateStats and UpdateState expose named endpoints for clients that prefer
// the legacy edit form's conceptual split while sharing one explicit DTO.
func (h *PlayersHandler) UpdateStats(c echo.Context) error { return h.Update(c) }
func (h *PlayersHandler) UpdateState(c echo.Context) error { return h.Update(c) }

func (h *PlayersHandler) mutate(c echo.Context, op string) error {
	id, ok := playerID(c)
	if !ok {
		WriteError(c.Response(), 400, "invalid_player_id", "invalid player ID", nil)
		return nil
	}
	var in playerMutationInput
	if decodePlayer(c, &in) != nil {
		return nil
	}
	ctx, cn := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cn()
	p, _, err := h.getPlayer(ctx, id)
	if err != nil {
		WriteError(c.Response(), 404, "player_not_found", "player not found", nil)
		return nil
	}
	q := models.New(h.pool)
	name := strings.TrimSpace(in.Name)
	if name == "" {
		name = strings.TrimSpace(in.Info)
	}
	ih := inventory.NewManualInventoryHandler(p, h.pool)
	var opErr error
	switch op {
	case "item_add":
		_, opErr = ih.AddItem(name, maxQuantity(in.Quantity, 1))
	case "item_remove":
		_, opErr = ih.RemoveItem(name, 1)
	case "item_buy":
		item, e := q.GetItemByFuzzy(ctx, name)
		if e != nil {
			opErr = fmt.Errorf("item not found")
		} else if p.Coins < item.Cost {
			opErr = fmt.Errorf("not enough coins")
		} else {
			_, opErr = ih.AddItem(item.Name, 1)
			if opErr == nil {
				_, opErr = q.UpdatePlayerCoins(ctx, models.UpdatePlayerCoinsParams{ID: id, Coins: p.Coins - item.Cost})
			}
		}
	case "ability_add":
		_, opErr = ih.AddAbility(name, in.Quantity)
	case "ability_remove":
		_, opErr = ih.RemoveAbility(name)
	case "status_add":
		_, opErr = ih.AddStatus(name, maxQuantity(in.Quantity, 1))
	case "status_remove":
		_, opErr = ih.RemoveStatus(name, 1)
	case "immunity_add":
		st, e := q.GetStatusByFuzzy(ctx, name)
		if e != nil {
			opErr = fmt.Errorf("immunity not found")
		} else {
			_, opErr = q.CreateOneTimePlayerImmunityJoin(ctx, models.CreateOneTimePlayerImmunityJoinParams{PlayerID: id, StatusID: st.ID, OneTime: in.OneTime})
		}
	case "immunity_remove":
		st, e := q.GetStatusByFuzzy(ctx, name)
		if e != nil {
			opErr = e
		} else {
			opErr = q.DeletePlayerImmunity(ctx, models.DeletePlayerImmunityParams{PlayerID: id, StatusID: st.ID})
		}
	case "note_add":
		if in.Position < 1 || strings.TrimSpace(in.Info) == "" {
			opErr = fmt.Errorf("position and info are required")
		} else {
			_, opErr = q.CreatePlayerNote(ctx, models.CreatePlayerNoteParams{PlayerID: id, Position: in.Position, Info: in.Info})
		}
	case "note_remove":
		opErr = q.DeletePlayerNote(ctx, models.DeletePlayerNoteParams{PlayerID: id, NoteID: in.NoteID})
	}
	if opErr != nil {
		WriteError(c.Response(), 400, "player_mutation_failed", opErr.Error(), nil)
		return nil
	}
	return h.Detail(c)
}
func maxQuantity(n, def int32) int32 {
	if n < 1 {
		return def
	}
	return n
}
func (h *PlayersHandler) ItemAdd(c echo.Context) error        { return h.mutate(c, "item_add") }
func (h *PlayersHandler) ItemRemove(c echo.Context) error     { return h.mutate(c, "item_remove") }
func (h *PlayersHandler) ItemBuy(c echo.Context) error        { return h.mutate(c, "item_buy") }
func (h *PlayersHandler) AbilityAdd(c echo.Context) error     { return h.mutate(c, "ability_add") }
func (h *PlayersHandler) AbilityRemove(c echo.Context) error  { return h.mutate(c, "ability_remove") }
func (h *PlayersHandler) StatusAdd(c echo.Context) error      { return h.mutate(c, "status_add") }
func (h *PlayersHandler) StatusRemove(c echo.Context) error   { return h.mutate(c, "status_remove") }
func (h *PlayersHandler) ImmunityAdd(c echo.Context) error    { return h.mutate(c, "immunity_add") }
func (h *PlayersHandler) ImmunityRemove(c echo.Context) error { return h.mutate(c, "immunity_remove") }
func (h *PlayersHandler) NoteAdd(c echo.Context) error        { return h.mutate(c, "note_add") }
func (h *PlayersHandler) NoteRemove(c echo.Context) error     { return h.mutate(c, "note_remove") }
