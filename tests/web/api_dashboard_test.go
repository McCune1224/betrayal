package web_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/mccune1224/betrayal/internal/models"
)

type dashboardResponse struct {
	Cycle struct {
		Phase  string `json:"phase"`
		Number int    `json:"number"`
	} `json:"cycle"`
	Players struct {
		Alive int `json:"alive"`
		Dead  int `json:"dead"`
		Total int `json:"total"`
	} `json:"players"`
}

func TestAPIDashboard(t *testing.T) {
	pool := mustPool(t)
	client := newTestClient(t, testServer(t, pool))

	t.Run("rejects unauthenticated requests as JSON without redirect", func(t *testing.T) {
		resp := client.get("/api/v1/dashboard")
		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("GET dashboard: status = %d, want %d", resp.StatusCode, http.StatusUnauthorized)
		}
		if location := resp.Header.Get("Location"); location != "" {
			t.Errorf("GET dashboard: Location = %q, want no redirect", location)
		}
		assertAPIError(t, resp, "unauthorized")
	})

	t.Run("returns typed cycle and player metrics for an authenticated admin", func(t *testing.T) {
		seedPlayer(t, pool, 901)
		seedPlayer(t, pool, 902)
		if _, err := models.New(pool).UpdatePlayerAlive(context.Background(), models.UpdatePlayerAliveParams{ID: 902, Alive: false}); err != nil {
			t.Fatalf("mark player dead: %v", err)
		}

		csrf := client.get("/api/v1/auth/csrf")
		if csrf.StatusCode != http.StatusOK {
			t.Fatalf("GET csrf: status = %d, want %d", csrf.StatusCode, http.StatusOK)
		}
		decodeAPIJSON(t, csrf, &apiCSRFToken{})
		loginAPI(t, client)

		resp := client.get("/api/v1/dashboard")
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("GET dashboard: status = %d, want %d", resp.StatusCode, http.StatusOK)
		}
		assertAPIJSON(t, resp)

		var body dashboardResponse
		decodeAPIJSON(t, resp, &body)
		if body.Cycle.Phase != "Day" || body.Cycle.Number != 0 {
			t.Errorf("cycle = %#v, want Day 0", body.Cycle)
		}
		if body.Players.Alive != 1 || body.Players.Dead != 1 || body.Players.Total != 2 {
			t.Errorf("players = %#v, want alive=1 dead=1 total=2", body.Players)
		}
	})
}
