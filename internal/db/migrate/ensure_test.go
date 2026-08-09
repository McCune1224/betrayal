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
