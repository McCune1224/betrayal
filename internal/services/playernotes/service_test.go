package playernotes_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/mccune1224/betrayal/internal/models"
	"github.com/mccune1224/betrayal/internal/services/playernotes"
	"github.com/mccune1224/betrayal/tests/testutil"
	"github.com/stretchr/testify/require"
)

func TestServiceRejectsMissingAuthorization(t *testing.T) {
	svc := playernotes.New(nil)
	_, err := svc.List(context.Background(), playernotes.Authorization{}, 1)
	require.ErrorIs(t, err, playernotes.ErrUnauthorized)
}

func TestServiceUpdateFailsWhenPositionIsMissing(t *testing.T) {
	pool := testutil.NewTestPool(t)
	testutil.TruncateAll(t, pool)
	q := models.New(pool)
	role, err := q.CreateRole(context.Background(), models.CreateRoleParams{
		Name: "Mafia", Description: "The mafia boss", Alignment: models.AlignmentEVIL,
	})
	require.NoError(t, err)
	player, err := q.CreatePlayer(context.Background(), models.CreatePlayerParams{
		ID: 100000000000000001, RoleID: pgtype.Int4{Int32: role.ID, Valid: true},
		Alive: true, Coins: 200, Alignment: models.AlignmentEVIL,
	})
	require.NoError(t, err)

	svc := playernotes.New(pool)
	_, err = svc.Update(context.Background(), playernotes.DiscordAdminAuthorization(), player.ID, 1, "missing")
	require.ErrorIs(t, err, pgx.ErrNoRows)
}

func TestServiceAddsConcurrentNotesAtDistinctPositions(t *testing.T) {
	pool := testutil.NewTestPool(t)
	testutil.TruncateAll(t, pool)
	q := models.New(pool)
	role, err := q.CreateRole(context.Background(), models.CreateRoleParams{
		Name: "Mafia", Description: "The mafia boss", Alignment: models.AlignmentEVIL,
	})
	require.NoError(t, err)
	player, err := q.CreatePlayer(context.Background(), models.CreatePlayerParams{
		ID: 100000000000000002, RoleID: pgtype.Int4{Int32: role.ID, Valid: true},
		Alive: true, Coins: 200, Alignment: models.AlignmentEVIL,
	})
	require.NoError(t, err)

	svc := playernotes.New(pool)
	auth := playernotes.DiscordAdminAuthorization()
	results := make(chan *models.PlayerNote, 2)
	errs := make(chan error, 2)
	for _, info := range []string{"first", "second"} {
		go func(info string) {
			note, err := svc.Add(context.Background(), auth, player.ID, info)
			results <- note
			errs <- err
		}(info)
	}
	for range 2 {
		require.NoError(t, <-errs)
	}
	positions := []int32{(<-results).Position, (<-results).Position}
	require.ElementsMatch(t, []int32{1, 2}, positions)
}

func TestServiceAddsNoteAtNextAvailablePosition(t *testing.T) {
	pool := testutil.NewTestPool(t)
	testutil.TruncateAll(t, pool)
	q := models.New(pool)
	role, err := q.CreateRole(context.Background(), models.CreateRoleParams{
		Name: "Mafia", Description: "The mafia boss", Alignment: models.AlignmentEVIL,
	})
	require.NoError(t, err)
	player, err := q.CreatePlayer(context.Background(), models.CreatePlayerParams{
		ID: 100000000000000001, RoleID: pgtype.Int4{Int32: role.ID, Valid: true},
		Alive: true, Coins: 200, Alignment: models.AlignmentEVIL,
	})
	require.NoError(t, err)

	svc := playernotes.New(pool)
	admin := playernotes.DiscordAdminAuthorization()
	first, err := svc.Add(context.Background(), admin, player.ID, "first")
	require.NoError(t, err)
	_, err = svc.Add(context.Background(), admin, player.ID, "second")
	require.NoError(t, err)
	require.NoError(t, svc.Delete(context.Background(), admin, player.ID, first.Position))

	third, err := svc.Add(context.Background(), admin, player.ID, "third")
	require.NoError(t, err)
	require.Equal(t, int32(3), third.Position)
}

func TestServiceSavesConcurrentMissingPositionsAtDistinctPositions(t *testing.T) {
	pool := testutil.NewTestPool(t)
	testutil.TruncateAll(t, pool)
	q := models.New(pool)
	role, err := q.CreateRole(context.Background(), models.CreateRoleParams{
		Name: "Mafia", Description: "The mafia boss", Alignment: models.AlignmentEVIL,
	})
	require.NoError(t, err)
	player, err := q.CreatePlayer(context.Background(), models.CreatePlayerParams{
		ID: 100000000000000003, RoleID: pgtype.Int4{Int32: role.ID, Valid: true},
		Alive: true, Coins: 200, Alignment: models.AlignmentEVIL,
	})
	require.NoError(t, err)

	svc := playernotes.New(pool)
	results := make(chan *models.PlayerNote, 8)
	errs := make(chan error, 8)
	for i := 0; i < 8; i++ {
		go func(info string) {
			note, err := svc.Save(context.Background(), playernotes.WebAdminAuthorization(), player.ID, 1, info)
			results <- note
			errs <- err
		}(fmt.Sprintf("note-%d", i))
	}
	for range 8 {
		require.NoError(t, <-errs)
	}
	for range 8 {
		require.Equal(t, int32(1), (<-results).Position)
	}
	notes, err := playernotes.New(pool).List(context.Background(), playernotes.WebAdminAuthorization(), player.ID)
	require.NoError(t, err)
	require.Len(t, notes, 1)
}

func TestServiceDeleteByIDReturnsNotFoundForMissingNote(t *testing.T) {
	pool := testutil.NewTestPool(t)
	testutil.TruncateAll(t, pool)
	q := models.New(pool)
	role, err := q.CreateRole(context.Background(), models.CreateRoleParams{
		Name: "Mafia", Description: "The mafia boss", Alignment: models.AlignmentEVIL,
	})
	require.NoError(t, err)
	player, err := q.CreatePlayer(context.Background(), models.CreatePlayerParams{
		ID: 100000000000000004, RoleID: pgtype.Int4{Int32: role.ID, Valid: true},
		Alive: true, Coins: 200, Alignment: models.AlignmentEVIL,
	})
	require.NoError(t, err)

	err = playernotes.New(pool).DeleteByID(context.Background(), playernotes.WebAdminAuthorization(), player.ID, 999)
	require.ErrorIs(t, err, playernotes.ErrNotFound)
}

func TestServiceSaveUsesPlayerNotesAdvisoryLock(t *testing.T) {
	pool := testutil.NewTestPool(t)
	testutil.TruncateAll(t, pool)
	q := models.New(pool)
	role, err := q.CreateRole(context.Background(), models.CreateRoleParams{
		Name: "Mafia", Description: "The mafia boss", Alignment: models.AlignmentEVIL,
	})
	require.NoError(t, err)
	player, err := q.CreatePlayer(context.Background(), models.CreatePlayerParams{
		ID: 100000000000000005, RoleID: pgtype.Int4{Int32: role.ID, Valid: true},
		Alive: true, Coins: 200, Alignment: models.AlignmentEVIL,
	})
	require.NoError(t, err)

	lockConn, err := pool.Acquire(context.Background())
	require.NoError(t, err)
	defer lockConn.Release()
	lockTx, err := lockConn.Begin(context.Background())
	require.NoError(t, err)
	defer func() { _ = lockTx.Rollback(context.Background()) }()
	_, err = lockTx.Exec(context.Background(), "select pg_advisory_xact_lock($1)", player.ID)
	require.NoError(t, err)
	var lockHeld bool
	err = lockTx.QueryRow(context.Background(), `
		SELECT EXISTS (
			SELECT 1
			FROM pg_locks
			WHERE pid = pg_backend_pid()
			  AND locktype = 'advisory'
			  AND mode = 'ExclusiveLock'
			  AND granted
		)`).Scan(&lockHeld)
	require.NoError(t, err)
	require.True(t, lockHeld, "test transaction must hold the player notes advisory lock")

	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()
	_, err = playernotes.New(pool).Save(ctx, playernotes.WebAdminAuthorization(), player.ID, 1, "blocked")
	require.ErrorIs(t, err, context.DeadlineExceeded)
}
