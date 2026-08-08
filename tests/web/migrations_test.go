package web_test

import (
	"net/http"
	"net/url"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mccune1224/betrayal/internal/web"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

// migrationsTestServer builds a web server with the embedded migrations
// runner bound to the LOCAL test database (DatabaseURL is used for both the
// prod guard and the runner; localhost → not prod, runner connects fine).
func migrationsTestServer(t *testing.T, pool *pgxpool.Pool, dsn string) *web.Server {
	t.Helper()
	srv, err := web.New(pool, nil, zerolog.Nop(), web.Config{
		Port:               "0",
		AdminPassword:      testAdminPassword,
		SessionSecret:      testSessionSecret,
		DatabaseURL:        dsn,
		AllowProdMutations: false,
	})
	if err != nil {
		t.Fatalf("web.New: %v", err)
	}
	return srv
}

func localDSN(t *testing.T) string {
	t.Helper()
	dsn := os.Getenv("DATABASE_URL")
	require.NotEmpty(t, dsn, "DATABASE_URL must be set (make db-up)")
	return dsn
}

func TestMigrationsPage(t *testing.T) {
	pool := mustPool(t)
	srv := migrationsTestServer(t, pool, localDSN(t))
	client := newTestClient(t, srv)
	client.login()

	resp := client.get("/admin/migrations")
	require.Equal(t, http.StatusOK, resp.StatusCode)
	body := client.body(resp)
	require.Contains(t, body, "DATABASE MIGRATIONS")
	require.Contains(t, body, "catalog_name_uniqueness")
	require.Contains(t, body, "applied")
	require.Contains(t, body, "up to date", "shared local DB is fully migrated → no pending")
}

func TestMigrationsUpNoPending(t *testing.T) {
	pool := mustPool(t)
	srv := migrationsTestServer(t, pool, localDSN(t))
	client := newTestClient(t, srv)
	client.login()

	// Shared local DB is at the latest migration, so Up is a no-op success.
	resp := client.do(http.MethodPost, "/admin/migrations/up", nil, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Contains(t, client.body(resp), "MIGRATIONS")
}

func TestMigrationsDownRequiresConfirmation(t *testing.T) {
	pool := mustPool(t)
	srv := migrationsTestServer(t, pool, localDSN(t))
	client := newTestClient(t, srv)
	client.login()

	// Wrong/missing confirmation phrase → 400, nothing rolled back.
	resp := client.do(http.MethodPost, "/admin/migrations/down",
		url.Values{"steps": {"1"}, "confirm": {"not-the-name"}}, nil)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode, "rollback without the exact migration name is rejected")

	resp = client.do(http.MethodPost, "/admin/migrations/down",
		url.Values{"steps": {"1"}}, nil)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	// Happy path is deliberately NOT exercised here: a real rollback would
	// move the shared local DB off the latest schema and break other suites.
	// The runner-level DownSteps behavior is covered by the scratch-DB suite
	// in internal/db/migrate.
}

func TestMigrationsBlockedInProd(t *testing.T) {
	pool := mustPool(t)
	srv := migrationsTestServer(t, pool, prodDSN) // fake prod DSN; runner never touched (blocked first)
	client := newTestClient(t, srv)
	client.login()

	resp := client.do(http.MethodPost, "/admin/migrations/up", nil, nil)
	require.Equal(t, http.StatusForbidden, resp.StatusCode, "migrations up is hard-blocked against prod")

	resp = client.do(http.MethodPost, "/admin/migrations/down",
		url.Values{"steps": {"1"}, "confirm": {"whatever"}}, nil)
	require.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestMigrationsCSRFGate(t *testing.T) {
	pool := mustPool(t)
	srv := migrationsTestServer(t, pool, localDSN(t))
	client := newTestClient(t, srv)
	client.login()

	resp := client.doRawWithoutCSRF(http.MethodPost, "/admin/migrations/up", nil)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode, "state-changing migrations POST without CSRF token is rejected")
}

func TestMigrationsRequiresAuth(t *testing.T) {
	pool := mustPool(t)
	srv := migrationsTestServer(t, pool, localDSN(t))
	client := newTestClient(t, srv)

	resp := client.get("/admin/migrations")
	require.Equal(t, http.StatusSeeOther, resp.StatusCode, "unauthenticated /admin/migrations redirects to login")
	require.Contains(t, resp.Header.Get("Location"), "/login")
}
