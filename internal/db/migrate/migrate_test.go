package dbmigrate_test

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	"github.com/jackc/pgx/v5"
	"github.com/mccune1224/betrayal/internal/db/migrate"
	"github.com/mccune1224/betrayal/tests/testutil"
	"github.com/stretchr/testify/require"
)

// The dbmigrate suite runs against a THROWAWAY scratch database it creates
// itself, so it never touches the shared local `betrayal` DB (or prod).
func TestMain(m *testing.M) {
	testutil.GuardProd(&testing.T{}) // fail hard if the env would hit prod
	os.Exit(m.Run())
}

// scratchDB creates a uniquely-named database on the local Postgres and
// returns its DSN. The database is dropped when the test finishes.
func scratchDB(t *testing.T) string {
	t.Helper()

	dbURL := os.Getenv("DATABASE_URL")
	require.NotEmpty(t, dbURL, "DATABASE_URL must be set (make db-up)")

	adminURL, err := replaceDBName(dbURL, "postgres")
	require.NoError(t, err)

	conn, err := pgx.Connect(context.Background(), adminURL)
	require.NoError(t, err, "connect to local postgres")
	t.Cleanup(func() { conn.Close(context.Background()) })

	name := fmt.Sprintf("betrayal_migrate_test_%d", os.Getpid())
	_, err = conn.Exec(context.Background(), "CREATE DATABASE "+name)
	require.NoError(t, err, "create scratch database")

	scratchURL, err := replaceDBName(dbURL, name)
	require.NoError(t, err)

	t.Cleanup(func() {
		// Drop the DB after releasing the runner's connection.
		if _, err := conn.Exec(context.Background(), "DROP DATABASE IF EXISTS "+name+" WITH (FORCE)"); err != nil {
			t.Errorf("drop scratch database: %v", err)
		}
	})
	return scratchURL
}

// replaceDBName swaps the database name in a postgres:// DSN.
func replaceDBName(dsn, dbname string) (string, error) {
	u, err := url.Parse(dsn)
	if err != nil {
		return "", err
	}
	u.Path = "/" + dbname
	return u.String(), nil
}

func TestScratchDBFresh(t *testing.T) {
	dsn := scratchDB(t)
	r, err := dbmigrate.New(dsn)
	require.NoError(t, err)
	defer r.Close()

	v, dirty, err := r.Version()
	require.ErrorIs(t, err, migrate.ErrNilVersion, "fresh DB has no version")
	require.Zero(t, v)
	require.False(t, dirty)

	st, err := r.Status()
	require.NoError(t, err)
	require.NotEmpty(t, st)
	for _, m := range st {
		require.False(t, m.Applied, "migration %d should be pending on fresh DB", m.Version)
	}
}

func TestUpDownCycle(t *testing.T) {
	dsn := scratchDB(t)
	r, err := dbmigrate.New(dsn)
	require.NoError(t, err)
	defer r.Close()

	// Up: all migrations applied.
	require.NoError(t, r.Up())
	v, dirty, err := r.Version()
	require.NoError(t, err)
	require.False(t, dirty)
	st, err := r.Status()
	require.NoError(t, err)
	require.NotEmpty(t, st)
	require.Equal(t, st[len(st)-1].Version, v, "version is the latest migration")
	for _, m := range st {
		require.True(t, m.Applied)
	}

	// Down one: latest migration no longer applied.
	require.NoError(t, r.DownSteps(1))
	v, _, err = r.Version()
	require.NoError(t, err)
	require.Equal(t, st[len(st)-2].Version, v, "rolled back one step")

	// Up again restores the full schema (idempotent runner).
	require.NoError(t, r.Up())
	v, _, err = r.Version()
	require.NoError(t, err)
	require.Equal(t, st[len(st)-1].Version, v)

	// Down with n <= 0 is a no-op.
	require.NoError(t, r.DownSteps(0))
}

func TestStatusNamesAndOrder(t *testing.T) {
	dsn := scratchDB(t)
	r, err := dbmigrate.New(dsn)
	require.NoError(t, err)
	defer r.Close()

	st, err := r.Status()
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(st), 31, "at least the 31 known migrations")
	// Versions ascending, names non-empty.
	for i, m := range st {
		require.NotEmpty(t, m.Name)
		if i > 0 {
			require.Greater(t, m.Version, st[i-1].Version, "ascending order")
		}
	}
	require.Equal(t, uint(31), st[len(st)-1].Version, "latest migration is 000031")
	require.Equal(t, "sync_run", st[len(st)-1].Name)
}
