package api

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/labstack/echo/v4"
)

// ResourcesCacheTTL is how long a Discord resources snapshot is served from
// memory before the next request refreshes it from the Discord REST API.
// The scrape (channels + paginated member list) costs seconds of round-trips;
// nicknames and channels rarely change mid-game, so a time-bounded cache turns
// every page load after the first into an in-memory read.
const ResourcesCacheTTL = 60 * time.Second

type DiscordGuildDTO struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type DiscordChannelDTO struct {
	ID       string `json:"id"`
	GuildID  string `json:"guild_id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Category string `json:"category,omitempty"`
}

type DiscordMemberDTO struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Nickname string `json:"nickname,omitempty"`
	Bot      bool   `json:"bot"`
}

// resourceSnapshot is one immutable scrape of Discord guild/channel/member
// data. Sections are best-effort exactly like the previous uncached behavior:
// a failed section is left empty rather than failing the whole page.
type resourceSnapshot struct {
	guilds   []DiscordGuildDTO
	channels []DiscordChannelDTO
	members  []DiscordMemberDTO
}

// ResourceCache memoizes a resourceSnapshot for a bounded window. A single
// cache is shared by every web handler that needs Discord identities, so the
// expensive REST scrape happens at most once per TTL across all pages.
type ResourceCache struct {
	mu        sync.Mutex
	fetch     func() *resourceSnapshot
	ttl       time.Duration
	now       func() time.Time
	snapshot  *resourceSnapshot
	fetchedAt time.Time
}

// NewResourceCache returns a cache backed by the given Discord session. A nil
// session yields an empty snapshot (local/web-only mode).
func NewResourceCache(session *discordgo.Session, ttl time.Duration) *ResourceCache {
	return &ResourceCache{
		fetch: func() *resourceSnapshot { return fetchDiscordSnapshot(session) },
		ttl:   ttl,
		now:   time.Now,
	}
}

// Get returns the cached snapshot if it is still fresh, otherwise it refetches
// and stores a new one. Safe for concurrent use.
func (c *ResourceCache) Get() *resourceSnapshot {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.snapshot != nil && c.now().Sub(c.fetchedAt) < c.ttl {
		return c.snapshot
	}
	c.snapshot = c.fetch()
	c.fetchedAt = c.now()
	return c.snapshot
}

type DiscordResourceHandler struct {
	discord *discordgo.Session
	cache   *ResourceCache
}

func NewDiscordResourceHandler(discord *discordgo.Session, cache *ResourceCache) *DiscordResourceHandler {
	if cache == nil {
		cache = NewResourceCache(discord, ResourcesCacheTTL)
	}
	return &DiscordResourceHandler{discord: discord, cache: cache}
}

func (h *DiscordResourceHandler) Resources(c echo.Context) error {
	if h.discord == nil {
		WriteError(c.Response(), http.StatusServiceUnavailable, "discord_unavailable", "Discord is disabled or not connected", nil)
		return nil
	}
	if h.discord.State == nil {
		WriteError(c.Response(), http.StatusServiceUnavailable, "discord_unavailable", "Discord state is not ready", nil)
		return nil
	}
	snapshot := h.cache.Get()
	WriteJSON(c.Response(), http.StatusOK, map[string]any{"guilds": snapshot.guilds, "channels": snapshot.channels, "members": snapshot.members})
	return nil
}

// fetchDiscordSnapshot scrapes the current guilds from gateway state and
// refreshes channel/member data over the Discord REST API. Each section is
// best-effort: failures fall back to an empty section, matching the historical
// behavior of the Resources endpoint.
func fetchDiscordSnapshot(session *discordgo.Session) *resourceSnapshot {
	snapshot := &resourceSnapshot{
		guilds:   make([]DiscordGuildDTO, 0),
		channels: make([]DiscordChannelDTO, 0),
		members:  make([]DiscordMemberDTO, 0),
	}
	if session == nil || session.State == nil {
		return snapshot
	}
	for _, guild := range session.State.Guilds {
		if guild == nil {
			continue
		}
		snapshot.guilds = append(snapshot.guilds, DiscordGuildDTO{ID: guild.ID, Name: guild.Name})
		if guildChannels, err := session.GuildChannels(guild.ID); err == nil {
			categoryNames := channelCategoryNames(guildChannels)
			for _, channel := range guildChannels {
				if channel == nil || channel.Type == discordgo.ChannelTypeGuildCategory {
					continue
				}
				snapshot.channels = append(snapshot.channels, DiscordChannelDTO{
					ID:       channel.ID,
					GuildID:  guild.ID,
					Name:     channel.Name,
					Type:     strconv.Itoa(int(channel.Type)),
					Category: categoryNames[channel.ParentID],
				})
			}
		}
		if guildMembers, err := allGuildMembers(session, guild.ID); err == nil {
			for _, member := range guildMembers {
				if member == nil || member.User == nil {
					continue
				}
				snapshot.members = append(snapshot.members, DiscordMemberDTO{ID: member.User.ID, Username: member.User.Username, Nickname: member.Nick, Bot: member.User.Bot})
			}
		}
	}
	return snapshot
}

// channelCategoryNames builds a channel-ID → name map in one pass. The previous
// implementation scanned every channel for every channel (O(n²)); the map makes
// category resolution O(n) and is definitionally equivalent.
func channelCategoryNames(channels []*discordgo.Channel) map[string]string {
	names := make(map[string]string, len(channels))
	for _, channel := range channels {
		if channel != nil {
			names[channel.ID] = channel.Name
		}
	}
	return names
}

// allGuildMembers walks Discord's paginated member endpoint. A single request
// is capped at 1000 members, so only fetching the first page silently hides
// valid players from admin selectors on larger guilds.
func allGuildMembers(session *discordgo.Session, guildID string) ([]*discordgo.Member, error) {
	const pageSize = 1000
	var members []*discordgo.Member
	after := ""
	for {
		page, err := session.GuildMembers(guildID, after, pageSize)
		if err != nil {
			return nil, err
		}
		members = append(members, page...)
		if len(page) < pageSize {
			return members, nil
		}
		last := page[len(page)-1]
		if last == nil || last.User == nil || last.User.ID == after {
			return members, nil
		}
		after = last.User.ID
	}
}