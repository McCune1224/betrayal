// Package testutil provides the shared bootstrap for Betrayal DB-backed test
// suites (tests/database, tests/inventory, tests/logger, tests/roll,
// tests/cycle, tests/vote, tests/web).
//
// Guarantees it enforces:
//  1. Tests connect ONLY to the local DATABASE_URL. A hard guard fails the run
//     if DATABASE_URL is unset or does not resolve to localhost, and fails if
//     DATABASE_POOLER_URL ever equals DATABASE_URL (the classic footgun).
//     DATABASE_POOLER_URL (the production Railway pooler) is then STRIPPED from
//     the test process so no test can ever route a connection through it.
//  2. Migrations run once per suite via TestMain (idempotent, golang-migrate).
//  3. DB suites are isolated from each other: an advisory lock serializes
//     test packages sharing the local Postgres, and TruncateAll wipes all
//     tables between tests (RESTART IDENTITY CASCADE) with the game_cycle
//     seed row re-inserted.
package testutil

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

// advisoryLockKey serializes DB test packages that share one local Postgres
// (go test ./... runs package binaries concurrently). Arbitrary constant.
const advisoryLockKey = int64(0xB374A9A1)

// allTables is the full table inventory (from internal/db/migrate/migrations) truncated
// between tests. game_cycle is included and re-seeded afterwards because its
// Day-0 row is inserted by migration 000023.
var allTables = []string{
	"sync_run",
	"sync_source",
	"vote",
	"command_audit",
	"logs",
	"player_note",
	"player_lifeboard",
	"action_channel",
	"vote_channel",
	"admin_channel",
	"player_immunity",
	"player_confessional",
	"player_ability",
	"player_perk",
	"player_status",
	"player_item",
	"player",
	"role_perk",
	"role_ability",
	"ability_category",
	"item_category",
	"item",
	"status",
	"category",
	"perk_info",
	"ability_info",
	"role",
	"game_cycle",
}

// repoRoot returns the absolute path of the repository root (parent of tests/).
func repoRoot() string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		panic("testutil: cannot determine repo root")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}

// LoadEnv loads the repo-root .env (the canonical worktree env). Variables
// already present in the environment win (CI sets DATABASE_URL itself), which
// is godotenv.Load's semantics.
func LoadEnv() {
	_ = godotenv.Load(filepath.Join(repoRoot(), ".env"))
	// Allow CWD-relative .env for standalone package runs.
	_ = godotenv.Load(".env")
}

// redactURL masks the password portion of a postgres URL for error messages.
func redactURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	if u.User != nil {
		if _, has := u.User.Password(); has {
			u.User = url.UserPassword(u.User.Username(), "***")
		}
	}
	return u.String()
}

// isLocalHost reports whether host is a loopback hostname.
func isLocalHost(host string) bool {
	host = strings.ToLower(host)
	switch host {
	case "localhost", "127.0.0.1", "::1", "[::1]":
		return true
	}
	return strings.HasPrefix(host, "127.")
}

// urlHost extracts the hostname from a postgres:// URL.
func urlHost(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	return u.Hostname()
}

// GuardProd enforces the "tests only ever use local DATABASE_URL" rule. It
// fails the run when:
//   - DATABASE_URL is unset (tests require local Postgres: `make db-up`), or
//   - DATABASE_URL resolves to anything other than localhost — this is the
//     production guard. A DATABASE_URL pointed at a remote host (e.g. the
//     Railway pooler, roundhouse.proxy.rlwy.net) aborts the run, or
//   - DATABASE_POOLER_URL equals DATABASE_URL (tests pointed at the pooler).
//
// It then strips DATABASE_POOLER_URL from the process so a buggy future test
// that reads it gets "" and fails to connect instead of touching production.
func GuardProd(t testing.TB) {
	t.Helper()
	LoadEnv()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Fatal("testutil: DATABASE_URL is not set; tests require local Postgres (run `make db-up` first)")
	}
	if host := urlHost(dbURL); !isLocalHost(host) {
		t.Fatalf("testutil: refusing to run tests against non-local database host %q (DATABASE_URL); tests must only ever use local DATABASE_URL", host)
	}

	if pooler := os.Getenv("DATABASE_POOLER_URL"); pooler != "" && pooler == dbURL {
		t.Fatalf("testutil: DATABASE_POOLER_URL equals DATABASE_URL; refusing to run tests against the production pooler")
	}

	// Belt & braces: the production pooler must not be reachable from tests.
	os.Unsetenv("DATABASE_POOLER_URL")
}

// guardProdErr is GuardProd for TestMain (no *testing.T available).
func guardProdErr() error {
	if dbURL := os.Getenv("DATABASE_URL"); dbURL == "" {
		return fmt.Errorf("DATABASE_URL is not set; tests require local Postgres (run `make db-up` first)")
	} else if host := urlHost(dbURL); !isLocalHost(host) {
		return fmt.Errorf("refusing to run tests against non-local database host %q (DATABASE_URL); tests must only ever use local DATABASE_URL", host)
	}
	if pooler := os.Getenv("DATABASE_POOLER_URL"); pooler != "" && pooler == os.Getenv("DATABASE_URL") {
		return fmt.Errorf("DATABASE_POOLER_URL equals DATABASE_URL; refusing to run tests against the production pooler")
	}
	os.Unsetenv("DATABASE_POOLER_URL")
	return nil
}

// acquireAdvisoryLock serializes this test process against other DB test
// packages sharing the local Postgres. The lock is held on a dedicated
// connection for the lifetime of the process and released by the returned
// function.
func acquireAdvisoryLock(dbURL string) (func(), error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	conn, err := pgx.Connect(ctx, dbURL)
	if err != nil {
		return nil, fmt.Errorf("connect for suite lock: %w", err)
	}
	if _, err := conn.Exec(ctx, "SELECT pg_advisory_lock($1)", advisoryLockKey); err != nil {
		conn.Close(context.Background())
		return nil, fmt.Errorf("acquire suite lock: %w", err)
	}
	return func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, _ = conn.Exec(ctx, "SELECT pg_advisory_unlock($1)", advisoryLockKey)
		conn.Close(ctx)
	}, nil
}

// Bootstrap is the TestMain entrypoint for every DB-backed test package:
//
//	func TestMain(m *testing.M) { os.Exit(testutil.Bootstrap(m)) }
//
// It loads env, enforces the production guard (exit 1 with a clear message on
// violation), serializes against other DB suites via an advisory lock, applies
// migrations once (idempotent), runs the suite, and returns the exit code.
func Bootstrap(m *testing.M) int {
	LoadEnv()
	if err := guardProdErr(); err != nil {
		fmt.Fprintln(os.Stderr, "testutil: "+err.Error())
		return 1
	}

	dbURL := os.Getenv("DATABASE_URL")
	release, err := acquireAdvisoryLock(dbURL)
	if err != nil {
		fmt.Fprintln(os.Stderr, "testutil:", err, "— is local Postgres up? (run `make db-up`)")
		return 1
	}
	defer release()

	if err := migrateUp(dbURL); err != nil {
		fmt.Fprintln(os.Stderr, "testutil: migrations failed:", err)
		return 1
	}

	return m.Run()
}

// migrateUp applies all migrations in internal/db/migrate/migrations.
// Idempotent: it is a no-op when the schema is already at the latest version.
func migrateUp(dbURL string) error {
	sourceURL := "file://" + filepath.ToSlash(filepath.Join(repoRoot(), "internal", "db", "migrate", "migrations"))
	m, err := migrate.New(sourceURL, dbURL)
	if err != nil {
		return err
	}
	defer m.Close()
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}

// NewTestPool connects to the local DATABASE_URL and fails the test (instead
// of silently skipping — the old /tmp/.s.PGSQL.5432 behavior) if it is
// unreachable. Call from SetupSuite; use TruncateAll in SetupTest.
func NewTestPool(t testing.TB) *pgxpool.Pool {
	t.Helper()
	GuardProd(t)

	dbURL := os.Getenv("DATABASE_URL")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		t.Fatalf("testutil: failed to create pool for DATABASE_URL: %v", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		t.Fatalf("testutil: DATABASE_URL (%s) unreachable: %v — run `make db-up` first", redactURL(dbURL), err)
	}
	return pool
}

// TruncateAll wipes every table (RESTART IDENTITY CASCADE) so each test starts
// from a clean schema, then re-inserts the game_cycle Day-0 seed row that
// migration 000023 creates. Call from SetupTest.
func TruncateAll(t testing.TB, pool *pgxpool.Pool) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	stmt := "TRUNCATE " + strings.Join(allTables, ", ") + " RESTART IDENTITY CASCADE"
	if _, err := pool.Exec(ctx, stmt); err != nil {
		t.Fatalf("testutil: truncate failed: %v", err)
	}
	// game_cycle is seeded by migration 000023 (Day 0); TRUNCATE removes it.
	if _, err := pool.Exec(ctx, "INSERT INTO game_cycle (is_elimination, day) VALUES (FALSE, 0)"); err != nil {
		t.Fatalf("testutil: re-seed game_cycle failed: %v", err)
	}
}
