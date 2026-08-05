package web_test

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/mccune1224/betrayal/internal/models"
)

// TestCyclePageAndAdvance covers GET /cycle, POST /cycle/advance and
// POST /cycle/set with the full day/elimination transitions.
func TestCyclePageAndAdvance(t *testing.T) {
	pool := mustPool(t)
	resetCycle(t, pool)
	client := newTestClient(t, testServer(t, pool))
	client.login()

	// Page renders the current phase (Day 0 after reset)
	resp := client.get("/cycle")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /cycle: expected 200, got %d", resp.StatusCode)
	}
	body := client.body(resp)
	if !strings.Contains(body, "GAME CYCLE") {
		t.Fatal("cycle page should render the header")
	}
	if !strings.Contains(body, ">0<") {
		t.Fatal("cycle page should show day 0 after reset")
	}

	q := models.New(pool)

	// Advance: Day 0 -> Day 1
	resp = client.do(http.MethodPost, "/cycle/advance", nil, nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("POST /cycle/advance: expected 200, got %d", resp.StatusCode)
	}
	if !strings.Contains(client.body(resp), "Day") {
		t.Fatal("advance partial should show the new phase")
	}
	cycle, err := q.GetCycle(context.Background())
	if err != nil {
		t.Fatalf("get cycle: %v", err)
	}
	if cycle.Day != 1 || cycle.IsElimination {
		t.Fatalf("after advance expected Day 1, got %+v", cycle)
	}

	// Advance: Day 1 -> Elimination 1
	client.do(http.MethodPost, "/cycle/advance", nil, nil)
	cycle, _ = q.GetCycle(context.Background())
	if cycle.Day != 1 || !cycle.IsElimination {
		t.Fatalf("after second advance expected Elimination 1, got %+v", cycle)
	}

	// Advance: Elimination 1 -> Day 2
	client.do(http.MethodPost, "/cycle/advance", nil, nil)
	cycle, _ = q.GetCycle(context.Background())
	if cycle.Day != 2 || cycle.IsElimination {
		t.Fatalf("after third advance expected Day 2, got %+v", cycle)
	}
}

// TestCycleSet covers hard-setting the phase from the web.
func TestCycleSet(t *testing.T) {
	pool := mustPool(t)
	resetCycle(t, pool)
	client := newTestClient(t, testServer(t, pool))
	client.login()

	resp := client.do(http.MethodPost, "/cycle/set", url.Values{
		"phase":  {"Elimination"},
		"number": {"3"},
	}, nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("POST /cycle/set: expected 200, got %d", resp.StatusCode)
	}
	if !strings.Contains(client.body(resp), "Elimination") {
		t.Fatal("set partial should show Elimination")
	}

	cycle, err := models.New(pool).GetCycle(context.Background())
	if err != nil {
		t.Fatalf("get cycle: %v", err)
	}
	if cycle.Day != 3 || !cycle.IsElimination {
		t.Fatalf("expected Elimination 3, got %+v", cycle)
	}

	// Invalid phase is rejected
	resp = client.do(http.MethodPost, "/cycle/set", url.Values{
		"phase":  {"Night"},
		"number": {"1"},
	}, nil)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("invalid phase: expected 400, got %d", resp.StatusCode)
	}

	// Invalid number is rejected
	resp = client.do(http.MethodPost, "/cycle/set", url.Values{
		"phase":  {"Day"},
		"number": {"abc"},
	}, nil)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("invalid number: expected 400, got %d", resp.StatusCode)
	}
}

// TestCycleRequiresAuth: cycle endpoints are behind the session gate.
func TestCycleRequiresAuth(t *testing.T) {
	pool := mustPool(t)
	client := newTestClient(t, testServer(t, pool))

	resp := client.get("/cycle")
	if resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("GET /cycle unauthenticated: expected 303, got %d", resp.StatusCode)
	}
	resp = client.do(http.MethodPost, "/cycle/advance", nil, nil)
	if resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("POST /cycle/advance unauthenticated: expected 303, got %d", resp.StatusCode)
	}
}
