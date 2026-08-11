package dbmigrate_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
	dbmigrate "github.com/mccune1224/betrayal/internal/db/migrate"
	"github.com/stretchr/testify/require"
)

func TestEnsureUpToDateCreatesSyncSchema(t *testing.T) {
	dsn := scratchDB(t)

	require.NoError(t, dbmigrate.EnsureUpToDate(dsn))

	conn, err := pgx.Connect(context.Background(), dsn)
	require.NoError(t, err)
	t.Cleanup(func() { conn.Close(context.Background()) })

	var sourceExists bool
	err = conn.QueryRow(context.Background(), `SELECT EXISTS (
		SELECT 1 FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'sync_source'
	)`).Scan(&sourceExists)
	require.NoError(t, err)
	require.True(t, sourceExists, "startup migration must create sync_source")
}

func TestMigration32ReconcilesDuplicateAbilityNames(t *testing.T) {
	dsn := scratchDB(t)
	r, err := dbmigrate.New(dsn)
	require.NoError(t, err)
	t.Cleanup(func() { r.Close() })
	require.NoError(t, r.Up())
	require.NoError(t, r.DownSteps(2))

	conn, err := pgx.Connect(context.Background(), dsn)
	require.NoError(t, err)
	t.Cleanup(func() { conn.Close(context.Background()) })
	_, err = conn.Exec(context.Background(), `
		INSERT INTO ability_info (name, description, default_charges, any_ability, rarity)
		VALUES ('Spatial Followings', 'canonical', 1, false, 'COMMON'),
		       ('Spatial Followings', 'duplicate', 2, false, 'COMMON')`)
	require.NoError(t, err)

	require.NoError(t, r.Up())
	var count int
	require.NoError(t, conn.QueryRow(context.Background(), `SELECT count(*) FROM ability_info WHERE name = 'Spatial Followings'`).Scan(&count))
	require.Equal(t, 1, count)
}

func TestEnsureUpToDateRecoversDirtyVersion32(t *testing.T) {
	dsn := scratchDB(t)
	r, err := dbmigrate.New(dsn)
	require.NoError(t, err)
	require.NoError(t, r.Up())

	conn, err := pgx.Connect(context.Background(), dsn)
	require.NoError(t, err)
	_, err = conn.Exec(context.Background(), `UPDATE schema_migrations SET version = 32, dirty = true`)
	require.NoError(t, err)
	conn.Close(context.Background())
	r.Close()

	require.NoError(t, dbmigrate.EnsureUpToDate(dsn))
	r2, err := dbmigrate.New(dsn)
	require.NoError(t, err)
	t.Cleanup(func() { r2.Close() })
	version, dirty, err := r2.Version()
	require.NoError(t, err)
	require.Equal(t, uint(33), version)
	require.False(t, dirty)
}

func TestEnsureUpToDateRecoversDirtyVersion33CompatibilityMarker(t *testing.T) {
	dsn := scratchDB(t)
	r, err := dbmigrate.New(dsn)
	require.NoError(t, err)
	require.NoError(t, r.Up())

	conn, err := pgx.Connect(context.Background(), dsn)
	require.NoError(t, err)
	_, err = conn.Exec(context.Background(), `UPDATE schema_migrations SET version = 33, dirty = true`)
	require.NoError(t, err)
	conn.Close(context.Background())
	require.NoError(t, r.Close())

	require.NoError(t, dbmigrate.EnsureUpToDate(dsn))
	r2, err := dbmigrate.New(dsn)
	require.NoError(t, err)
	t.Cleanup(func() { r2.Close() })
	version, dirty, err := r2.Version()
	require.NoError(t, err)
	require.Equal(t, uint(33), version)
	require.False(t, dirty)
}
