package web_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/mccune1224/betrayal/internal/models"
)

func TestGameOpsCycleAPIRequiresAuthAndReturnsDTO(t *testing.T) {
	pool := mustPool(t)
	client := newTestClient(t, testServer(t, pool))

	unauthenticated := client.get("/api/v1/ops/cycle")
	if unauthenticated.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected JSON 401, got %d", unauthenticated.StatusCode)
	}

	client.login()
	resp := client.get("/api/v1/ops/cycle")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", resp.StatusCode, client.body(resp))
	}
	var body struct {
		Day           int    `json:"day"`
		Phase         string `json:"phase"`
		IsElimination bool   `json:"is_elimination"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body.Day != 0 || body.Phase != "Day" || body.IsElimination {
		t.Fatalf("unexpected cycle DTO: %+v", body)
	}
}

func TestGameOpsChannelsAPIUsesExplicitSummaryAndDiscordStatus(t *testing.T) {
	pool := mustPool(t)
	client := newTestClient(t, testServer(t, pool))
	client.login()

	resp := client.get("/api/v1/ops/channels")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", resp.StatusCode, client.body(resp))
	}
	var body struct {
		DiscordConnected bool `json:"discord_connected"`
		Entries          []struct {
			Kind      string `json:"kind"`
			ChannelID string `json:"channel_id"`
			Status    string `json:"status"`
		} `json:"entries"`
		Summary struct {
			Total      int `json:"total"`
			Missing    int `json:"missing"`
			Unverified int `json:"unverified"`
		} `json:"summary"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body.DiscordConnected || body.Summary.Total != len(body.Entries) {
		t.Fatalf("unexpected channels DTO: %+v", body)
	}
	if body.Summary.Total == 0 || body.Summary.Missing != body.Summary.Total {
		t.Fatalf("expected missing configured-channel entries: %+v", body.Summary)
	}
}

func TestGameOpsVotesAPIReturnsCycleVotesTalliesAndStats(t *testing.T) {
	pool := mustPool(t)
	client := newTestClient(t, testServer(t, pool))
	client.login()

	seedPlayer(t, pool, 100000000000000001)
	seedPlayer(t, pool, 100000000000000002)
	_, err := models.New(pool).UpsertVote(context.Background(), models.UpsertVoteParams{
		VoterID: 100000000000000001, TargetID: 100000000000000002, CycleDay: 0,
		IsElimination: false, Weight: 2, Context: pgtype.Text{String: "test", Valid: true},
	})
	if err != nil {
		t.Fatal(err)
	}

	resp := client.get("/api/v1/ops/votes?day=0&elimination=false")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", resp.StatusCode, client.body(resp))
	}
	var body struct {
		Cycle struct {
			Day           int  `json:"day"`
			IsElimination bool `json:"is_elimination"`
		} `json:"cycle"`
		Votes []struct {
			ID int32 `json:"id"`
		} `json:"votes"`
		Tallies []struct {
			TargetID   int64 `json:"target_id"`
			TotalVotes int   `json:"total_votes"`
		} `json:"tallies"`
		Stats struct {
			MostVoted []struct {
				PlayerID int64 `json:"player_id"`
			} `json:"most_voted_players"`
		} `json:"stats"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body.Cycle.Day != 0 || body.Cycle.IsElimination || len(body.Votes) != 1 || len(body.Tallies) != 1 || len(body.Stats.MostVoted) != 1 {
		t.Fatalf("unexpected votes DTO: %+v", body)
	}
	if body.Tallies[0].TargetID != 100000000000000002 || body.Tallies[0].TotalVotes != 2 {
		t.Fatalf("unexpected tally: %+v", body.Tallies[0])
	}
}

func TestGameOpsHealthcheckAPIReturnsReadinessDTO(t *testing.T) {
	pool := mustPool(t)
	client := newTestClient(t, testServer(t, pool))
	client.login()

	resp := client.get("/api/v1/ops/healthcheck")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", resp.StatusCode, client.body(resp))
	}
	var body struct {
		Ready            bool `json:"ready"`
		DiscordConnected bool `json:"discord_connected"`
		Players          struct {
			Total int `json:"total"`
		} `json:"players"`
		Cycle struct {
			Ready bool `json:"ready"`
		} `json:"cycle"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body.Ready || body.DiscordConnected || body.Players.Total != 0 || !body.Cycle.Ready {
		t.Fatalf("unexpected readiness DTO: %+v", body)
	}
}
