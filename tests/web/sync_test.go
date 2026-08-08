package web_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mccune1224/betrayal/internal/models"
	"github.com/mccune1224/betrayal/internal/web"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

// prodDSN is a fake Railway pooler DSN used to exercise the production guard.
// web.New never connects to it — it only inspects the string.
const prodDSN = "postgres://user:pass@roundhouse.proxy.rlwy.net:5432/betrayal"

// roleCSV mirrors the real Google Sheets role export format (leading "so"
// row, label row, role row, "Abilities:" marker, abilities, "Passives:"
// marker, passives, blank-row chunk separators).
const syncRoleCSV = `so,,,,,,
,Name ,Description,,,,
,RoleA,Role description A,,,,
,Abilities:,Charges,Type,Description,Categories,Rarity (if AA)
,Ability One,3,*,Does a thing,Stealth/Combat,COMMON
,Passives:,Description,,,,
,Passive One,Passive desc one,,,,
`

// syncTestServer builds a web server seeded with the given sync source URLs.
func syncTestServer(t *testing.T, pool *pgxpool.Pool, envURLs map[string]string, isProd bool) *web.Server {
	t.Helper()
	dsn := ""
	if isProd {
		dsn = prodDSN
	}
	srv, err := web.New(pool, nil, zerolog.Nop(), web.Config{
		Port:                "0",
		AdminPassword:       testAdminPassword,
		SessionSecret:       testSessionSecret,
		DatabaseURL:         dsn,
		AllowProdMutations:  false,
		SyncEnvURLs:         envURLs,
	})
	if err != nil {
		t.Fatalf("web.New: %v", err)
	}
	return srv
}

func TestSyncPage(t *testing.T) {
	pool := mustPool(t)
	srv := syncTestServer(t, pool, nil, false)
	client := newTestClient(t, srv)
	client.login()

	resp := client.get("/sync")
	require.Equal(t, http.StatusOK, resp.StatusCode)
	body := client.body(resp)
	require.Contains(t, body, "SPREADSHEET SYNC")
	require.Contains(t, body, "good_roles")
	require.Contains(t, body, "items")
	require.Contains(t, body, "once before a game starts", "run-once callout present")
}

func TestSyncUpdateSource(t *testing.T) {
	pool := mustPool(t)
	srv := syncTestServer(t, pool, nil, false)
	client := newTestClient(t, srv)
	client.login()

	sources, err := models.New(pool).ListSyncSources(context.Background())
	require.NoError(t, err)
	var good models.SyncSource
	for _, s := range sources {
		if s.Name == "good_roles" {
			good = s
		}
	}
	require.NotZero(t, good.ID)

	resp := client.do(http.MethodPost, fmt.Sprintf("/sync/sources/%d", good.ID),
		url.Values{"url": {"https://sheets.example/edited.csv"}, "enabled": {"on"}}, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	updated, err := models.New(pool).GetSyncSourceByName(context.Background(), "good_roles")
	require.NoError(t, err)
	require.Equal(t, "https://sheets.example/edited.csv", updated.Url)
	require.True(t, updated.Enabled)
}

func TestSyncPreviewAndApply(t *testing.T) {
	pool := mustPool(t)

	// Serve the fixture CSV so the full fetch → preview → apply flow runs
	// against a real HTTP endpoint (no external network).
	csvServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/csv")
		fmt.Fprint(w, syncRoleCSV)
	}))
	defer csvServer.Close()

	srv := syncTestServer(t, pool, map[string]string{"good_roles": csvServer.URL}, false)
	client := newTestClient(t, srv)
	client.login()

	// Preview: diff renders the new role + ability.
	resp := client.do(http.MethodPost, "/sync/preview", nil, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	body := client.body(resp)
	require.Contains(t, body, "RoleA")
	require.Contains(t, body, "create")
	require.Contains(t, body, "Ability One")

	// Apply: writes the role + ability + perk and records an applied run.
	sources, err := models.New(pool).ListSyncSources(context.Background())
	require.NoError(t, err)
	var good models.SyncSource
	for _, s := range sources {
		if s.Name == "good_roles" {
			good = s
		}
	}
	resp = client.do(http.MethodPost, "/sync/apply",
		url.Values{"source_id": {fmt.Sprintf("%d", good.ID)}}, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	q := models.New(pool)
	ctx := context.Background()
	role, err := q.GetRoleByName(ctx, "RoleA")
	require.NoError(t, err)
	require.Equal(t, "Role description A", role.Description)

	// Re-preview now shows the role as unchanged.
	resp = client.do(http.MethodPost, "/sync/preview", nil, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	body = client.body(resp)
	require.Contains(t, body, "unchanged")

	// Audit row recorded.
	runs, err := models.New(pool).ListSyncRuns(ctx, 10)
	require.NoError(t, err)
	var applied bool
	for _, r := range runs {
		if r.Status == "applied" && r.SourceName == "good_roles" {
			applied = true
		}
	}
	require.True(t, applied, "an applied sync_run row exists")
}

func TestSyncApplyBlockedInProd(t *testing.T) {
	pool := mustPool(t)
	srv := syncTestServer(t, pool, nil, true) // fake prod DSN
	client := newTestClient(t, srv)
	client.login()

	resp := client.do(http.MethodPost, "/sync/apply", url.Values{"source_id": {"1"}}, nil)
	require.Equal(t, http.StatusForbidden, resp.StatusCode, "apply is hard-blocked against prod")
}

func TestSyncApplyUnknownSource(t *testing.T) {
	pool := mustPool(t)
	srv := syncTestServer(t, pool, nil, false)
	client := newTestClient(t, srv)
	client.login()

	resp := client.do(http.MethodPost, "/sync/apply", url.Values{"source_id": {"99999"}}, nil)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestSyncCSRFGate(t *testing.T) {
	pool := mustPool(t)
	srv := syncTestServer(t, pool, nil, false)
	client := newTestClient(t, srv)
	client.login()

	resp := client.doRawWithoutCSRF(http.MethodPost, "/sync/preview", nil)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode, "state-changing sync POST without CSRF token is rejected")

	resp = client.doRawWithoutCSRF(http.MethodPost, "/sync/apply", url.Values{"source_id": {"1"}})
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestSyncRequiresAuth(t *testing.T) {
	pool := mustPool(t)
	srv := syncTestServer(t, pool, nil, false)
	client := newTestClient(t, srv)

	resp := client.get("/sync")
	require.Equal(t, http.StatusSeeOther, resp.StatusCode, "unauthenticated /sync redirects to login")
	require.Contains(t, resp.Header.Get("Location"), "/login")
}
