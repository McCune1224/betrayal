// Package dbmigrate runs golang-migrate over the EMBEDDED migration files.
//
// Migrations live in migrations/ inside this package so `//go:embed` can pick
// them up — the web admin panel (and tests) use this runner instead of the
// golang-migrate CLI. The Makefile migrate-* targets point at the same files
// on disk, so there is exactly one source of truth for the schema.
package dbmigrate

import (
	"embed"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

// Migration is one schema migration as surfaced to the UI.
type Migration struct {
	Version uint
	Name    string
	Applied bool
	// Dirty is only meaningful for the current version: it means the last
	// migration was applied but did not finish cleanly (schema_migrations
	// rows exist but the migration's own rows may be partial).
	Dirty bool
}

// Runner wraps a golang-migrate instance bound to a single database URL.
// New opens a connection eagerly (the postgres driver pings and ensures the
// schema_migrations table at construction); subsequent operations reuse it.
// Call Close when done. Server-side, construct lazily so a briefly-down DB
// doesn't block startup.
type Runner struct {
	m *migrate.Migrate
}

// EnsureUpToDate applies every embedded migration before the application
// starts serving requests. A deploy must not advertise a healthy web server
// while a required table such as sync_source is absent.
func EnsureUpToDate(dsn string) error {
	r, err := New(dsn)
	if err != nil {
		return err
	}
	defer r.Close()
	return r.Up()
}

// New creates a Runner for dsn (postgres://...). No connection is opened
// until the first operation.
func New(dsn string) (*Runner, error) {
	src, err := iofs.New(migrationFS, "migrations")
	if err != nil {
		return nil, fmt.Errorf("dbmigrate: load embedded migrations: %w", err)
	}
	m, err := migrate.NewWithSourceInstance("iofs", src, dsn)
	if err != nil {
		return nil, fmt.Errorf("dbmigrate: init migrate: %w", err)
	}
	return &Runner{m: m}, nil
}

// Version returns the current schema version and whether the database is in a
// dirty state. version==0 with no error means no migrations have been applied.
func (r *Runner) Version() (version uint, dirty bool, err error) {
	return r.m.Version()
}

// Up applies all pending migrations.
func (r *Runner) Up() error {
	err := r.m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("dbmigrate: up: %w", err)
	}
	return nil
}

// DownSteps rolls back n migrations (1 = the most recent). n <= 0 is a no-op.
func (r *Runner) DownSteps(n int) error {
	if n <= 0 {
		return nil
	}
	err := r.m.Steps(-n)
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("dbmigrate: down %d: %w", n, err)
	}
	return nil
}

// Status lists every embedded migration with its applied state. Migrations
// are ordered by version ascending. Dirty is reported for the current
// version only.
func (r *Runner) Status() ([]Migration, error) {
	entries, err := migrationFS.ReadDir("migrations")
	if err != nil {
		return nil, fmt.Errorf("dbmigrate: read embedded migrations: %w", err)
	}

	current, dirty, err := r.m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return nil, fmt.Errorf("dbmigrate: current version: %w", err)
	}

	seen := map[uint]string{}
	for _, e := range entries {
		name := e.Name()
		if !strings.HasSuffix(name, ".up.sql") {
			continue
		}
		versionStr, _, ok := strings.Cut(name, "_")
		if !ok {
			continue
		}
		v, err := strconv.ParseUint(versionStr, 10, 32)
		if err != nil {
			continue
		}
		seen[uint(v)] = strings.TrimSuffix(strings.TrimPrefix(name, versionStr+"_"), ".up.sql")
	}

	versions := make([]uint, 0, len(seen))
	for v := range seen {
		versions = append(versions, v)
	}
	sort.Slice(versions, func(i, j int) bool { return versions[i] < versions[j] })

	out := make([]Migration, 0, len(versions))
	for _, v := range versions {
		out = append(out, Migration{
			Version: v,
			Name:    seen[v],
			Applied: v <= current,
			Dirty:   v == current && dirty,
		})
	}
	return out, nil
}

// Close releases the underlying database connection.
func (r *Runner) Close() error {
	srcErr, dbErr := r.m.Close()
	if srcErr != nil {
		return fmt.Errorf("dbmigrate: close source: %w", srcErr)
	}
	if dbErr != nil {
		return fmt.Errorf("dbmigrate: close db: %w", dbErr)
	}
	return nil
}
