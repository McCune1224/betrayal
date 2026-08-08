package playernotes_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/mccune1224/betrayal/internal/models"
	"github.com/mccune1224/betrayal/internal/services/playernotes"
	"github.com/mccune1224/betrayal/tests/testutil"
	"github.com/stretchr/testify/require"
)

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
	admin := playernotes.Authorization{IsAdmin: true}
	first, err := svc.Add(context.Background(), admin, player.ID, "first")
	require.NoError(t, err)
	_, err = svc.Add(context.Background(), admin, player.ID, "second")
	require.NoError(t, err)
	require.NoError(t, svc.Delete(context.Background(), admin, player.ID, first.Position))

	third, err := svc.Add(context.Background(), admin, player.ID, "third")
	require.NoError(t, err)
	require.Equal(t, int32(3), third.Position)
}
