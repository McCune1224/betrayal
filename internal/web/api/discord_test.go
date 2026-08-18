package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/labstack/echo/v4"
)

// newFixedCache builds a ResourceCache with a fake fetcher and an injectable
// clock so cache semantics can be asserted without touching Discord.
func newFixedCache(t *testing.T, fetch func() *resourceSnapshot, ttl time.Duration, now *time.Time) *ResourceCache {
	t.Helper()
	return &ResourceCache{
		fetch: fetch,
		ttl:   ttl,
		now:   func() time.Time { return *now },
	}
}

func TestResourceCacheServesFreshSnapshotWithinTTL(t *testing.T) {
	fetchCount := 0
	now := time.Now()
	cache := newFixedCache(t, func() *resourceSnapshot {
		fetchCount++
		return &resourceSnapshot{guilds: []DiscordGuildDTO{{ID: "g1", Name: "Guild"}}}
	}, time.Hour, &now)

	first := cache.Get()
	if got := first.guilds[0].ID; got != "g1" {
		t.Fatalf("first snapshot guild id = %q, want g1", got)
	}
	second := cache.Get()
	if fetchCount != 1 {
		t.Fatalf("fetch called %d times within TTL, want 1", fetchCount)
	}
	if second != first {
		t.Fatal("cache returned a different snapshot within TTL")
	}
}

func TestResourceCacheRefetchesAfterTTLExpiry(t *testing.T) {
	fetchCount := 0
	now := time.Now()
	cache := newFixedCache(t, func() *resourceSnapshot {
		fetchCount++
		return &resourceSnapshot{members: []DiscordMemberDTO{{ID: "u1", Username: "alice"}}}
	}, 30*time.Second, &now)

	cache.Get()
	if fetchCount != 1 {
		t.Fatalf("initial fetch count = %d, want 1", fetchCount)
	}
	now = now.Add(31 * time.Second)
	cache.Get()
	if fetchCount != 2 {
		t.Fatalf("fetch count after TTL expiry = %d, want 2", fetchCount)
	}
}

func TestResourcesHandlerServesCachedSnapshot(t *testing.T) {
	fetchCount := 0
	now := time.Now()
	cache := newFixedCache(t, func() *resourceSnapshot {
		fetchCount++
		return &resourceSnapshot{
			guilds:   []DiscordGuildDTO{{ID: "g1", Name: "Guild One"}},
			channels: []DiscordChannelDTO{{ID: "c1", GuildID: "g1", Name: "general", Type: "0"}},
			members:  []DiscordMemberDTO{{ID: "u1", Username: "alice", Bot: false}},
		}
	}, time.Hour, &now)

	h := NewDiscordResourceHandler(&discordgo.Session{State: discordgo.NewState()}, cache)

	body := func() string {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/discord/resources", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		if err := h.Resources(c); err != nil {
			t.Fatalf("Resources returned error: %v", err)
		}
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rec.Code)
		}
		return rec.Body.String()
	}

	first := body()
	if fetchCount != 1 {
		t.Fatalf("fetch count after first request = %d, want 1", fetchCount)
	}
	second := body()
	if fetchCount != 1 {
		t.Fatalf("fetch count after second request within TTL = %d, want 1 (cached)", fetchCount)
	}
	if first != second {
		t.Fatal("handler returned different payloads for two requests within TTL")
	}
	for _, want := range []string{`"Guild One"`, `"general"`, `"alice"`} {
		if !strings.Contains(first, want) {
			t.Fatalf("payload missing %q: %s", want, first)
		}
	}
}

func TestResourcesHandlerRefetchesAfterTTLExpiry(t *testing.T) {
	fetchCount := 0
	now := time.Now()
	cache := newFixedCache(t, func() *resourceSnapshot {
		fetchCount++
		return &resourceSnapshot{guilds: []DiscordGuildDTO{{ID: "g1", Name: "Guild"}}}
	}, 30*time.Second, &now)

	h := NewDiscordResourceHandler(&discordgo.Session{State: discordgo.NewState()}, cache)

	request := func() {
		e := echo.New()
		rec := httptest.NewRecorder()
		c := e.NewContext(httptest.NewRequest(http.MethodGet, "/api/v1/discord/resources", nil), rec)
		if err := h.Resources(c); err != nil {
			t.Fatalf("Resources returned error: %v", err)
		}
	}
	request()
	now = now.Add(31 * time.Second)
	request()
	if fetchCount != 2 {
		t.Fatalf("fetch count across TTL boundary = %d, want 2", fetchCount)
	}
}

func TestWhisperHandlerSharesResourceCache(t *testing.T) {
	fetchCount := 0
	now := time.Now()
	cache := newFixedCache(t, func() *resourceSnapshot {
		fetchCount++
		return &resourceSnapshot{members: []DiscordMemberDTO{
			{ID: "u1", Username: "alice", Nickname: "Ali", Bot: false},
			{ID: "u2", Username: "bob", Bot: false},
		}}
	}, time.Hour, &now)

	h := NewWhisperHandler(nil, &discordgo.Session{State: discordgo.NewState()}, cache)

	names := h.discordPlayerNames()
	if fetchCount != 1 {
		t.Fatalf("fetch count after first whisper load = %d, want 1", fetchCount)
	}
	if got := names["u1"]; got != [2]string{"Ali", "Discord member"} {
		t.Fatalf("names[u1] = %v, want [Ali Discord member]", got)
	}
	if got := names["u2"]; got != [2]string{"bob", "Discord member"} {
		t.Fatalf("names[u2] = %v, want [bob Discord member]", got)
	}
	_ = h.discordPlayerNames()
	if fetchCount != 1 {
		t.Fatalf("fetch count after second whisper load within TTL = %d, want 1 (shared cache)", fetchCount)
	}
}

func TestChannelCategoryNamesMapsParentIDs(t *testing.T) {
	channels := []*discordgo.Channel{
		{ID: "cat1", Name: "Alliances", Type: discordgo.ChannelTypeGuildCategory},
		{ID: "ch1", Name: "red-alliance", ParentID: "cat1"},
		{ID: "ch2", Name: "orphan", ParentID: "missing"},
	}
	got := channelCategoryNames(channels)
	if got["cat1"] != "Alliances" {
		t.Fatalf("categoryNames[cat1] = %q, want Alliances", got["cat1"])
	}
	if got["ch1"] != "red-alliance" {
		t.Fatalf("categoryNames[ch1] = %q, want red-alliance", got["ch1"])
	}
	if _, ok := got["missing"]; ok {
		t.Fatal("categoryNames contains an entry for a channel that does not exist")
	}
}