package dbmigrate_test

import (
	"context"
	"fmt"
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
	require.NoError(t, r.DownSteps(4))

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

func TestEnsureUpToDateRejectsDirtyVersion(t *testing.T) {
	for _, version := range []uint{32, 33, 34} {
		t.Run(fmt.Sprintf("version_%d", version), func(t *testing.T) {
			dsn := scratchDB(t)
			r, err := dbmigrate.New(dsn)
			require.NoError(t, err)
			require.NoError(t, r.Up())
			require.NoError(t, r.Close())

			conn, err := pgx.Connect(context.Background(), dsn)
			require.NoError(t, err)
			_, err = conn.Exec(context.Background(), `UPDATE schema_migrations SET version = $1, dirty = true`, version)
			require.NoError(t, err)
			require.NoError(t, conn.Close(context.Background()))

			err = dbmigrate.EnsureUpToDate(dsn)
			require.EqualError(t, err, fmt.Sprintf("dbmigrate: dirty database version %d requires manual recovery", version))
		})
	}
}
