package web_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/mccune1224/betrayal/internal/models"
)

func TestAPIPlayerDelete(t *testing.T) {
	pool := mustPool(t)
	ctx := context.Background()
	q := models.New(pool)

	role, err := q.CreateRole(ctx, models.CreateRoleParams{Name: "Oracle", Description: "delete test role", Alignment: models.AlignmentGOOD})
	if err != nil {
		t.Fatalf("create role: %v", err)
	}
	const playerID int64 = 802
	_, err = q.CreatePlayer(ctx, models.CreatePlayerParams{
		ID:        playerID,
		RoleID:    pgtype.Int4{Int32: role.ID, Valid: true},
		Alive:     true,
		Coins:     200,
		CoinBonus: pgtype.Numeric{},
		Luck:      0,
		ItemLimit: 4,
		Alignment: models.AlignmentGOOD,
	})
	if err != nil {
		t.Fatalf("create player: %v", err)
	}
	if _, err := q.CreatePlayerConfessional(ctx, models.CreatePlayerConfessionalParams{
		PlayerID:     playerID,
		ChannelID:    9001,
		PinMessageID: 9002,
	}); err != nil {
		t.Fatalf("create confessional: %v", err)
	}
	item, err := q.CreateItem(ctx, models.CreateItemParams{Name: "Test Blade", Description: "delete test item", Rarity: models.RarityCOMMON, Cost: 10})
	if err != nil {
		t.Fatalf("create item: %v", err)
	}
	if err := q.UpsertPlayerItemJoin(ctx, models.UpsertPlayerItemJoinParams{PlayerID: playerID, ItemID: item.ID, Quantity: 1}); err != nil {
		t.Fatalf("grant item: %v", err)
	}

	client := newTestClient(t, testServer(t, pool))
	del := func(path string) *http.Response { return client.do(http.MethodDelete, path, nil, nil) }

	t.Run("rejects unauthenticated delete as JSON unauthorized", func(t *testing.T) {
		client.get("/api/v1/auth/csrf") // sets the CSRF cookie so the request reaches the auth gate
		resp := del("/api/v1/players/802")
		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("DELETE status = %d, want %d", resp.StatusCode, http.StatusUnauthorized)
		}
		assertAPIError(t, resp, "unauthorized")
	})

	client.login()

	t.Run("rejects malformed player id", func(t *testing.T) {
		resp := del("/api/v1/players/not-a-number")
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("DELETE status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
		}
	})

	t.Run("deletes the player and cascades dependent rows", func(t *testing.T) {
		resp := del("/api/v1/players/802")
		if resp.StatusCode != http.StatusNoContent {
			t.Fatalf("DELETE status = %d, want %d: %s", resp.StatusCode, http.StatusNoContent, client.body(resp))
		}

		if _, err := q.GetPlayer(ctx, playerID); err == nil {
			t.Error("player still exists after delete")
		}
		if _, err := q.GetPlayerConfessional(ctx, playerID); err == nil {
			t.Error("confessional survived cascade delete")
		}
		rows, err := q.ListPlayerItem(ctx, playerID)
		if err != nil {
			t.Fatalf("list player items: %v", err)
		}
		if len(rows) != 0 {
			t.Errorf("player items survived cascade delete: %#v", rows)
		}
	})

	t.Run("returns 404 for a second or unknown delete", func(t *testing.T) {
		resp := del("/api/v1/players/802")
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("second DELETE status = %d, want %d", resp.StatusCode, http.StatusNotFound)
		}
		assertAPIError(t, resp, "player_not_found")

		resp = del("/api/v1/players/999999")
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("unknown DELETE status = %d, want %d", resp.StatusCode, http.StatusNotFound)
		}
	})
}
