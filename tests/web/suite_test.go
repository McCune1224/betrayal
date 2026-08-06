// Package web_test contains httptest coverage for the Betrayal admin web panel.
// Requires a local Postgres reachable via DATABASE_URL (make db-up +
// make mock-migrate-up). Never points at DATABASE_POOLER_URL (production).
package web_test

import (
	"context"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mccune1224/betrayal/internal/models"
	"github.com/mccune1224/betrayal/internal/web"
	"github.com/mccune1224/betrayal/tests/testutil"
	"github.com/rs/zerolog"
)

const (
	// 48 bytes — satisfies the >= 32 byte SESSION_SECRET requirement
	testSessionSecret = "0123456789abcdef0123456789abcdef0123456789abcdef"
	testAdminPassword = "hunter2-test-password"
)

// TestMain bootstraps the suite through testutil: loads env, enforces the
// production guard, serializes against other DB suites via an advisory lock,
// and applies migrations once.
func TestMain(m *testing.M) {
	os.Exit(testutil.Bootstrap(m))
}

// mustPool connects to the local test DB via testutil (fails, not skips, when
// unavailable) and truncates all tables so each test starts from a clean
// schema with the game_cycle Day-0 seed row present.
func mustPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	pool := testutil.NewTestPool(t)
	t.Cleanup(pool.Close)
	testutil.TruncateAll(t, pool)
	return pool
}

// testServer builds a web server wired to the test pool with a fixed admin
// password + session secret. No Discord session (web-only mode).
func testServer(t *testing.T, pool *pgxpool.Pool) *web.Server {
	t.Helper()
	srv, err := web.New(pool, nil, zerolog.Nop(), web.Config{
		Port:          "0",
		AdminPassword: testAdminPassword,
		SessionSecret: testSessionSecret,
	})
	if err != nil {
		t.Fatalf("web.New: %v", err)
	}
	return srv
}

// testClient wraps an httptest server with a cookie jar so the session and
// CSRF cookies persist across requests, and attaches the CSRF token header
// automatically (mirroring what the base layout's JS does in the browser).
type testClient struct {
	t    *testing.T
	base string
	jar  http.CookieJar
}

func newTestClient(t *testing.T, srv *web.Server) *testClient {
	t.Helper()
	ts := httptest.NewServer(srv.Handler())
	t.Cleanup(ts.Close)

	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatalf("cookiejar: %v", err)
	}
	return &testClient{t: t, base: ts.URL, jar: jar}
}

// do performs a request. form, when non-nil, is sent url-encoded. extra
// headers are merged in. The X-CSRF-Token header is attached automatically
// when the jar holds a _csrf cookie (call GET /login first if needed).
func (c *testClient) do(method, path string, form url.Values, extra map[string]string) *http.Response {
	return c.doWithCSRF(method, path, form, extra, true)
}

// doRawWithoutCSRF is like do but never attaches the CSRF token — used to
// prove that state-changing requests without a token are rejected.
func (c *testClient) doRawWithoutCSRF(method, path string, form url.Values) *http.Response {
	return c.doWithCSRF(method, path, form, nil, false)
}

func (c *testClient) doWithCSRF(method, path string, form url.Values, extra map[string]string, attachCSRF bool) *http.Response {
	c.t.Helper()

	var body io.Reader
	if form != nil {
		body = strings.NewReader(form.Encode())
	}

	req, err := http.NewRequest(method, c.base+path, body)
	if err != nil {
		c.t.Fatalf("new request %s %s: %v", method, path, err)
	}
	if form != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	for k, v := range extra {
		req.Header.Set(k, v)
	}
	if attachCSRF {
		if token := c.csrfToken(); token != "" {
			req.Header.Set("X-CSRF-Token", token)
		}
	}

	client := &http.Client{Jar: c.jar, CheckRedirect: func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse // don't follow redirects; assert on them
	}}
	resp, err := client.Do(req)
	if err != nil {
		c.t.Fatalf("request %s %s: %v", method, path, err)
	}
	c.t.Cleanup(func() { resp.Body.Close() })
	return resp
}

// get is a convenience wrapper for do with no form.
func (c *testClient) get(path string) *http.Response {
	return c.do(http.MethodGet, path, nil, nil)
}

func (c *testClient) csrfToken() string {
	u, _ := url.Parse(c.base)
	for _, cookie := range c.jar.Cookies(u) {
		if cookie.Name == "_csrf" {
			return cookie.Value
		}
	}
	return ""
}

// login performs a successful admin login (requires the CSRF cookie first).
func (c *testClient) login() {
	c.t.Helper()
	c.get("/login") // sets the _csrf cookie
	resp := c.do(http.MethodPost, "/login", url.Values{"password": {testAdminPassword}}, nil)
	if resp.StatusCode != http.StatusSeeOther {
		c.t.Fatalf("login: expected 303, got %d", resp.StatusCode)
	}
}

// body reads the full response body.
func (c *testClient) body(resp *http.Response) string {
	c.t.Helper()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		c.t.Fatalf("read body: %v", err)
	}
	return string(b)
}

// ---------------------------------------------------------------------------
// DB seeding helpers (all test data is created and cleaned up per test)
// ---------------------------------------------------------------------------

func seedPlayer(t *testing.T, pool *pgxpool.Pool, id int64) models.Player {
	t.Helper()
	player, err := models.New(pool).CreatePlayer(context.Background(), models.CreatePlayerParams{
		ID:        id,
		RoleID:    pgtype.Int4{},
		Alive:     true,
		Coins:     100,
		CoinBonus: pgtype.Numeric{},
		Luck:      2,
		Alignment: models.AlignmentNEUTRAL,
	})
	if err != nil {
		t.Fatalf("seed player %d: %v", id, err)
	}
	t.Cleanup(func() { cleanupPlayer(t, pool, id) })
	return player
}

func cleanupPlayer(t *testing.T, pool *pgxpool.Pool, id int64) {
	t.Helper()
	ctx := context.Background()
	q := models.New(pool)
	_ = q.DeletePlayerAbility(ctx, models.DeletePlayerAbilityParams{PlayerID: id, AbilityID: 0})
	_, _ = pool.Exec(ctx, "DELETE FROM player_perk WHERE player_id = $1", id)
	_, _ = pool.Exec(ctx, "DELETE FROM player_status WHERE player_id = $1", id)
	_, _ = pool.Exec(ctx, "DELETE FROM player_item WHERE player_id = $1", id)
	_, _ = pool.Exec(ctx, "DELETE FROM player_note WHERE player_id = $1", id)
	_ = q.DeletePlayer(ctx, id)
}

func seedItem(t *testing.T, pool *pgxpool.Pool, name string) models.Item {
	t.Helper()
	item, err := models.New(pool).CreateItem(context.Background(), models.CreateItemParams{
		Name:        name,
		Description: "test item",
		Rarity:      models.RarityCOMMON,
		Cost:        10,
	})
	if err != nil {
		t.Fatalf("seed item %s: %v", name, err)
	}
	t.Cleanup(func() { _ = models.New(pool).DeleteItem(context.Background(), item.ID) })
	return item
}

func seedAbility(t *testing.T, pool *pgxpool.Pool, name string) models.AbilityInfo {
	t.Helper()
	ability, err := models.New(pool).CreateAbilityInfo(context.Background(), models.CreateAbilityInfoParams{
		Name:           name,
		Description:    "test ability",
		DefaultCharges: 2,
		AnyAbility:     false,
		Rarity:         models.RarityRARE,
	})
	if err != nil {
		t.Fatalf("seed ability %s: %v", name, err)
	}
	t.Cleanup(func() { _ = models.New(pool).DeleteAbilityInfo(context.Background(), ability.ID) })
	return ability
}

func seedStatus(t *testing.T, pool *pgxpool.Pool, name string) models.Status {
	t.Helper()
	status, err := models.New(pool).CreateStatus(context.Background(), models.CreateStatusParams{
		Name:        name,
		Description: "test status",
	})
	if err != nil {
		t.Fatalf("seed status %s: %v", name, err)
	}
	t.Cleanup(func() { _ = models.New(pool).DeleteStatus(context.Background(), status.ID) })
	return status
}

func seedPerk(t *testing.T, pool *pgxpool.Pool, name string) models.PerkInfo {
	t.Helper()
	perk, err := models.New(pool).CreatePerkInfo(context.Background(), models.CreatePerkInfoParams{
		Name:        name,
		Description: "test perk",
	})
	if err != nil {
		t.Fatalf("seed perk %s: %v", name, err)
	}
	t.Cleanup(func() { _ = models.New(pool).DeletePerkInfo(context.Background(), perk.ID) })
	return perk
}

// resetCycle forces the game cycle back to Day 0 and restores it after the
// test, keeping the shared local DB pristine for other worktrees.
func resetCycle(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	ctx := context.Background()
	cycle, err := models.New(pool).GetCycle(ctx)
	if err != nil {
		t.Fatalf("get cycle: %v", err)
	}
	t.Cleanup(func() {
		_, _ = models.New(pool).UpdateCycle(context.Background(), models.UpdateCycleParams{
			IsElimination: cycle.IsElimination,
			Day:           cycle.Day,
			ID:            cycle.ID,
		})
	})
	_, err = pool.Exec(ctx, "UPDATE game_cycle SET day = 0, is_elimination = FALSE")
	if err != nil {
		t.Fatalf("reset cycle: %v", err)
	}
}

// seedVoteChannel inserts a vote channel row and removes it after the test.
func seedVoteChannel(t *testing.T, pool *pgxpool.Pool, channelID string) {
	t.Helper()
	q := models.New(pool)
	ctx := context.Background()
	if err := q.UpsertVoteChannel(ctx, channelID); err != nil {
		t.Fatalf("seed vote channel %s: %v", channelID, err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), "DELETE FROM vote_channel WHERE channel_id = $1", channelID)
	})
}

// seedActionChannel inserts an action channel row and removes it after the test.
func seedActionChannel(t *testing.T, pool *pgxpool.Pool, channelID string) {
	t.Helper()
	q := models.New(pool)
	ctx := context.Background()
	if err := q.UpsertActionChannel(ctx, channelID); err != nil {
		t.Fatalf("seed action channel %s: %v", channelID, err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), "DELETE FROM action_channel WHERE channel_id = $1", channelID)
	})
}
