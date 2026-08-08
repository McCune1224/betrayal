package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/mccune1224/betrayal/internal/models"
	"github.com/mccune1224/betrayal/internal/util"
	"github.com/mccune1224/betrayal/internal/web/templates/pages"
)

// Channel statuses surfaced on the /channels page
const (
	channelStatusConfigured = "configured"
	channelStatusMissing    = "missing"
	channelStatusOrphaned   = "orphaned"
	channelStatusUnverified = "unverified"
)

// ChannelsHandler implements the channel config validation page (/channels).
// This is the long-standing "/admin health" ask: a page that lists the
// configured vote/action/admin/lifeboard channels and flags problems.
type ChannelsHandler struct {
	dbPool         *pgxpool.Pool
	discordSession *discordgo.Session // nil in web-only mode (DISABLE_DISCORD=true)
}

// NewChannelsHandler creates a new ChannelsHandler
func NewChannelsHandler(pool *pgxpool.Pool, discordSession *discordgo.Session) *ChannelsHandler {
	return &ChannelsHandler{dbPool: pool, discordSession: discordSession}
}

// Update handles channel configuration mutations from the web panel.
func (h *ChannelsHandler) Update(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()
	kind, channelID := c.FormValue("kind"), c.FormValue("channel_id")
	if channelID == "" {
		return c.String(http.StatusBadRequest, "channel ID is required")
	}
	if h.discordSession != nil {
		if _, err := h.discordSession.Channel(channelID); err != nil {
			return c.String(http.StatusBadRequest, "Discord channel was not found")
		}
	}
	q := models.New(h.dbPool)
	var err error
	switch kind {
	case "vote":
		err = q.UpsertVoteChannel(ctx, channelID)
	case "action":
		err = q.WipeActionChannel(ctx)
		if err == nil {
			err = q.UpsertActionChannel(ctx, channelID)
		}
	case "log":
		_, err = q.SetCommandLogChannel(ctx, channelID)
	case "admin":
		_, err = q.CreateAdminChannel(ctx, channelID)
	case "lifeboard":
		messageID := c.FormValue("message_id")
		if messageID == "" {
			return c.String(http.StatusBadRequest, "lifeboard message ID is required")
		}
		if err = q.DeletePlayerLifeboard(ctx); err == nil {
			_, err = q.CreatePlayerLifeboard(ctx, models.CreatePlayerLifeboardParams{ChannelID: channelID, MessageID: messageID})
		}
	default:
		return c.String(http.StatusBadRequest, "unknown channel type")
	}
	if err != nil {
		c.Response().Header().Set("HX-Trigger", toastTrigger("Channel update failed", "error"))
		return c.String(http.StatusInternalServerError, "channel update failed")
	}
	c.Response().Header().Set("HX-Trigger", toastTrigger("Channel configuration updated", "success"))
	return h.Page(c)
}

func (h *ChannelsHandler) DeleteAdmin(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()
	id := c.FormValue("channel_id")
	if id == "" {
		return c.String(http.StatusBadRequest, "channel ID is required")
	}
	if err := models.New(h.dbPool).DeleteAdminChannel(ctx, id); err != nil {
		return c.String(http.StatusInternalServerError, "failed to remove admin channel")
	}
	c.Response().Header().Set("HX-Trigger", toastTrigger("Admin channel removed", "success"))
	return h.Page(c)
}

// Page handles GET /channels
func (h *ChannelsHandler) Page(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()

	q := models.New(h.dbPool)

	var entries []pages.ChannelEntry

	// Vote channel (singleton)
	entries = append(entries, h.singletonEntry(ctx, q, "Vote Channel", "vote",
		func(ctx context.Context) (string, error) {
			return q.GetVoteChannel(ctx)
		}))

	// Action channel (singleton)
	entries = append(entries, h.singletonEntry(ctx, q, "Action Channel", "action",
		func(ctx context.Context) (string, error) {
			return q.GetActionChannel(ctx)
		}))

	// Lifeboard (singleton)
	entries = append(entries, h.singletonEntry(ctx, q, "Lifeboard Channel", "lifeboard",
		func(ctx context.Context) (string, error) {
			lb, err := q.GetPlayerLifeboard(ctx)
			return lb.ChannelID, err
		}))

	// Admin channels (list; empty is a warning, not fatal)
	adminChannels, err := q.ListAdminChannel(ctx)
	if err == nil {
		if len(adminChannels) == 0 {
			entries = append(entries, pages.ChannelEntry{
				Name:   "Admin Channels",
				Kind:   "admin",
				Status: channelStatusMissing,
				Note:   "No whitelisted admin channels configured — /inv admin subcommands cannot be used from Discord",
			})
		} else {
			for _, ch := range adminChannels {
				entries = append(entries, h.checkedEntry("Admin Channel", "admin", ch))
			}
		}
	}

	// Player confessionals (list)
	confessionals, err := q.ListPlayerConfessional(ctx)
	if err == nil && len(confessionals) > 0 {
		for _, conf := range confessionals {
			entries = append(entries, h.checkedEntry("Confessional", "confessional", util.Itoa64(conf.ChannelID)))
		}
	}

	summary := summarizeEntries(entries)
	data := pages.ChannelsData{
		Entries:          entries,
		DiscordConnected: h.discordSession != nil,
		Summary:          summary,
	}

	return render(c, http.StatusOK, pages.Channels(data))
}

// singletonEntry loads a single configured channel (vote/action/lifeboard) and
// flags it as missing when the row does not exist.
func (h *ChannelsHandler) singletonEntry(ctx context.Context, q *models.Queries, name, kind string, load func(context.Context) (string, error)) pages.ChannelEntry {
	channelID, err := load(ctx)
	if err != nil {
		return pages.ChannelEntry{
			Name:   name,
			Kind:   kind,
			Status: channelStatusMissing,
			Note:   "Not configured — set it with /channel in Discord",
		}
	}
	return h.checkedEntry(name, kind, channelID)
}

// checkedEntry verifies a configured channel ID against Discord when a session
// is available. Without Discord (web-only mode) the channel is "unverified".
func (h *ChannelsHandler) checkedEntry(name, kind, channelID string) pages.ChannelEntry {
	entry := pages.ChannelEntry{
		Name:      name,
		Kind:      kind,
		ChannelID: channelID,
	}

	if h.discordSession == nil {
		entry.Status = channelStatusUnverified
		entry.Note = "Configured, but Discord is disabled — cannot verify the channel exists"
		return entry
	}

	if _, err := h.discordSession.Channel(channelID); err != nil {
		entry.Status = channelStatusOrphaned
		entry.Note = "Configured, but Discord reports this channel no longer exists"
		return entry
	}

	entry.Status = channelStatusConfigured
	return entry
}

func summarizeEntries(entries []pages.ChannelEntry) pages.ChannelSummary {
	var s pages.ChannelSummary
	s.Total = len(entries)
	for _, e := range entries {
		switch e.Status {
		case channelStatusConfigured:
			s.Configured++
		case channelStatusMissing:
			s.Missing++
		case channelStatusOrphaned:
			s.Orphaned++
		case channelStatusUnverified:
			s.Unverified++
		}
	}
	return s
}
