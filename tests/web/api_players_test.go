package web_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/mccune1224/betrayal/internal/models"
)

type apiPlayer struct {
	ID        int64  `json:"id"`
	Alive     bool   `json:"alive"`
	Coins     int    `json:"coins"`
	Luck      int    `json:"luck"`
	ItemLimit int    `json:"item_limit"`
	Alignment string `json:"alignment"`
	Role      string `json:"role"`
}

func TestAPIPlayers(t *testing.T) {
	pool := mustPool(t)

	t.Run("rejects unauthenticated requests as JSON without redirect", func(t *testing.T) {
		client := newTestClient(t, testServer(t, pool))
		resp := client.get("/api/v1/players")
		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("GET players: status = %d, want %d", resp.StatusCode, http.StatusUnauthorized)
		}
		if location := resp.Header.Get("Location"); location != "" {
			t.Errorf("GET players: Location = %q, want no redirect", location)
		}
		assertAPIError(t, resp, "unauthorized")
	})

	t.Run("returns an empty JSON array when no players exist", func(t *testing.T) {
		client := newTestClient(t, testServer(t, pool))
		csrf := client.get("/api/v1/auth/csrf")
		if csrf.StatusCode != http.StatusOK {
			t.Fatalf("GET csrf: status = %d, want %d", csrf.StatusCode, http.StatusOK)
		}
		loginAPI(t, client)

		resp := client.get("/api/v1/players")
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("GET players: status = %d, want %d", resp.StatusCode, http.StatusOK)
		}
		assertAPIJSON(t, resp)

		var players []apiPlayer
		decodeAPIJSON(t, resp, &players)
		if len(players) != 0 {
			t.Errorf("players = %#v, want empty array", players)
		}
	})

	t.Run("returns explicit player list DTOs including role display", func(t *testing.T) {
		role, err := models.New(pool).CreateRole(context.Background(), models.CreateRoleParams{
			Name: "Oracle", Description: "test role", Alignment: models.AlignmentGOOD,
		})
		if err != nil {
			t.Fatalf("create role: %v", err)
		}
		_, err = models.New(pool).CreatePlayer(context.Background(), models.CreatePlayerParams{
			ID:        701,
			RoleID:    pgtype.Int4{Int32: role.ID, Valid: true},
			Alive:     true,
			Coins:     42,
			CoinBonus: pgtype.Numeric{},
			Luck:      7,
			ItemLimit: 3,
			Alignment: models.AlignmentGOOD,
		})
		if err != nil {
			t.Fatalf("create player: %v", err)
		}

		client := newTestClient(t, testServer(t, pool))
		csrf := client.get("/api/v1/auth/csrf")
		if csrf.StatusCode != http.StatusOK {
			t.Fatalf("GET csrf: status = %d, want %d", csrf.StatusCode, http.StatusOK)
		}
		loginAPI(t, client)
		resp := client.get("/api/v1/players")
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("GET players: status = %d, want %d", resp.StatusCode, http.StatusOK)
		}
		assertAPIJSON(t, resp)

		var players []apiPlayer
		decodeAPIJSON(t, resp, &players)
		if len(players) != 1 {
			t.Fatalf("players = %#v, want one player", players)
		}
		if got, want := players[0], (apiPlayer{ID: 701, Alive: true, Coins: 42, Luck: 7, ItemLimit: 3, Alignment: "GOOD", Role: "Oracle"}); got != want {
			t.Errorf("player = %#v, want %#v", got, want)
		}

		resp = client.get("/api/v1/players")
		var raw []map[string]json.RawMessage
		decodeAPIJSON(t, resp, &raw)
		if _, ok := raw[0]["role_id"]; ok {
			t.Error("player response leaked raw role_id")
		}
		if _, ok := raw[0]["coin_bonus"]; ok {
			t.Error("player response leaked raw coin_bonus")
		}
	})
}
