package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/mccune1224/betrayal/internal/models"
)

type WhisperHandler struct {
	pool      *pgxpool.Pool
	discord   *discordgo.Session
	resources *ResourceCache
}

func NewWhisperHandler(pool *pgxpool.Pool, discord *discordgo.Session, resources *ResourceCache) *WhisperHandler {
	if resources == nil {
		resources = NewResourceCache(discord, ResourcesCacheTTL)
	}
	return &WhisperHandler{pool: pool, discord: discord, resources: resources}
}

type whisperGroupDTO struct {
	ID      int64    `json:"id"`
	Name    string   `json:"name"`
	Players []string `json:"players"`
}
type whisperPlayerDTO struct {
	ID     string `json:"id"`
	Label  string `json:"label"`
	Detail string `json:"detail,omitempty"`
}
type whisperMessageDTO struct {
	ID      int64  `json:"id"`
	Message string `json:"message"`
	Enabled bool   `json:"enabled"`
}
type whisperDTO struct {
	Groups   []whisperGroupDTO   `json:"groups"`
	Players  []whisperPlayerDTO  `json:"players"`
	Messages []whisperMessageDTO `json:"messages"`
}

func (h *WhisperHandler) Get(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()
	q := models.New(h.pool)
	groups, err := q.ListWhisperGroups(ctx)
	if err != nil {
		return whisperFailure(c, "whisper_groups_unavailable")
	}
	players, err := q.ListPlayer(ctx)
	if err != nil {
		return whisperFailure(c, "whisper_players_unavailable")
	}
	messages, err := q.ListWhisperDoubtMessages(ctx)
	if err != nil {
		return whisperFailure(c, "whisper_messages_unavailable")
	}
	discordPlayers := h.discordPlayerNames()
	result := whisperDTO{Groups: make([]whisperGroupDTO, 0), Players: make([]whisperPlayerDTO, 0, len(players)), Messages: make([]whisperMessageDTO, 0, len(messages))}
	byID := make(map[int64]int)
	for _, row := range groups {
		index, ok := byID[row.ID]
		if !ok {
			result.Groups = append(result.Groups, whisperGroupDTO{ID: row.ID, Name: row.Name, Players: []string{}})
			index = len(result.Groups) - 1
			byID[row.ID] = index
		}
		if row.PlayerID.Valid {
			result.Groups[index].Players = append(result.Groups[index].Players, strconv.FormatInt(row.PlayerID.Int64, 10))
		}
	}
	for index, player := range players {
		id := strconv.FormatInt(player.ID, 10)
		pair := discordPlayers[id]
		label, detail := pair[0], pair[1]
		if label == "" {
			label = "Player " + strconv.Itoa(index+1)
			detail = "Discord identity unavailable in local mode"
		}
		result.Players = append(result.Players, whisperPlayerDTO{ID: id, Label: label, Detail: detail})
	}
	for _, message := range messages {
		result.Messages = append(result.Messages, whisperMessageDTO{ID: message.ID, Message: message.Message, Enabled: message.Enabled})
	}
	WriteJSON(c.Response(), http.StatusOK, result)
	return nil
}

func (h *WhisperHandler) discordPlayerNames() map[string][2]string {
	names := make(map[string][2]string)
	if h.discord == nil || h.discord.State == nil {
		return names
	}
	snapshot := h.resources.Get()
	for _, member := range snapshot.members {
		if member.ID == "" {
			continue
		}
		label := member.Nickname
		if label == "" {
			label = member.Username
		}
		names[member.ID] = [2]string{label, "Discord member"}
	}
	return names
}

func (h *WhisperHandler) CreateGroup(c echo.Context) error {
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil || strings.TrimSpace(req.Name) == "" || len([]rune(req.Name)) > 100 {
		return whisperBad(c, "group name is required and must be at most 100 characters")
	}
	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()
	group, err := models.New(h.pool).CreateWhisperGroup(ctx, strings.TrimSpace(req.Name))
	if err != nil {
		return whisperFailure(c, "whisper_group_create_failed")
	}
	WriteJSON(c.Response(), http.StatusCreated, whisperGroupDTO{ID: group.ID, Name: group.Name, Players: []string{}})
	return nil
}

func (h *WhisperHandler) DeleteGroup(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		return whisperBad(c, "group id must be positive")
	}
	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()
	if err := models.New(h.pool).DeleteWhisperGroup(ctx, id); err != nil {
		return whisperFailure(c, "whisper_group_delete_failed")
	}
	c.NoContent(http.StatusNoContent)
	return nil
}

func (h *WhisperHandler) AddMember(c echo.Context) error    { return h.memberMutation(c, true) }
func (h *WhisperHandler) RemoveMember(c echo.Context) error { return h.memberMutation(c, false) }
func (h *WhisperHandler) memberMutation(c echo.Context, add bool) error {
	groupID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || groupID <= 0 {
		return whisperBad(c, "group id must be positive")
	}
	var req struct {
		PlayerID string `json:"player_id"`
	}
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return whisperBad(c, "player_id must be positive")
	}
	playerID, err := strconv.ParseInt(req.PlayerID, 10, 64)
	if err != nil || playerID <= 0 {
		return whisperBad(c, "player_id must be a positive integer")
	}
	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()
	q := models.New(h.pool)
	var mutationErr error
	if add {
		mutationErr = q.AddWhisperGroupMember(ctx, models.AddWhisperGroupMemberParams{GroupID: groupID, PlayerID: playerID})
	} else {
		mutationErr = q.RemoveWhisperGroupMember(ctx, models.RemoveWhisperGroupMemberParams{GroupID: groupID, PlayerID: playerID})
	}
	if mutationErr != nil {
		return whisperFailure(c, "whisper_group_member_update_failed")
	}
	return h.Get(c)
}

func (h *WhisperHandler) CreateMessage(c echo.Context) error {
	var req struct {
		Message string `json:"message"`
	}
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil || strings.TrimSpace(req.Message) == "" || len([]rune(req.Message)) > 1000 {
		return whisperBad(c, "message is required and must be at most 1000 characters")
	}
	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()
	message, err := models.New(h.pool).CreateWhisperDoubtMessage(ctx, strings.TrimSpace(req.Message))
	if err != nil {
		return whisperFailure(c, "whisper_message_create_failed")
	}
	WriteJSON(c.Response(), http.StatusCreated, whisperMessageDTO{ID: message.ID, Message: message.Message, Enabled: message.Enabled})
	return nil
}

func (h *WhisperHandler) UpdateMessage(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		return whisperBad(c, "message id must be positive")
	}
	var req struct {
		Message string `json:"message"`
		Enabled bool   `json:"enabled"`
	}
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil || strings.TrimSpace(req.Message) == "" || len([]rune(req.Message)) > 1000 {
		return whisperBad(c, "message is required and must be at most 1000 characters")
	}
	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()
	message, err := models.New(h.pool).UpdateWhisperDoubtMessage(ctx, models.UpdateWhisperDoubtMessageParams{ID: id, Message: strings.TrimSpace(req.Message), Enabled: req.Enabled})
	if err != nil {
		return whisperFailure(c, "whisper_message_update_failed")
	}
	WriteJSON(c.Response(), http.StatusOK, whisperMessageDTO{ID: message.ID, Message: message.Message, Enabled: message.Enabled})
	return nil
}

func (h *WhisperHandler) DeleteMessage(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		return whisperBad(c, "message id must be positive")
	}
	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()
	if err := models.New(h.pool).DeleteWhisperDoubtMessage(ctx, id); err != nil {
		return whisperFailure(c, "whisper_message_delete_failed")
	}
	c.NoContent(http.StatusNoContent)
	return nil
}

func whisperBad(c echo.Context, message string) error {
	WriteError(c.Response(), http.StatusBadRequest, "invalid_whisper_request", message, nil)
	return nil
}
func whisperFailure(c echo.Context, code string) error {
	WriteError(c.Response(), http.StatusInternalServerError, code, "whisper administration is unavailable", nil)
	return nil
}
