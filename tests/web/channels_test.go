package web_test

import (
	"net/http"
	"strings"
	"testing"
)

// TestChannelsPage: the config validation page renders the known channel
// sections. We seed a vote + action channel so the page has something to show;
// with Discord disabled they must be listed as UNVERIFIED rather than falsely
// "OK", and the empty lifeboard must be flagged MISSING.
func TestChannelsPage(t *testing.T) {
	pool := mustPool(t)
	seedVoteChannel(t, pool, "9000000000000000001")
	seedActionChannel(t, pool, "9000000000000000002")

	client := newTestClient(t, testServer(t, pool))
	client.login()

	resp := client.get("/channels")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /channels: expected 200, got %d", resp.StatusCode)
	}

	body := client.body(resp)
	if !strings.Contains(body, "CHANNEL CONFIG") {
		t.Fatal("channels page should render the header")
	}
	if !strings.Contains(body, "Vote Channel") {
		t.Fatal("channels page should list the vote channel")
	}
	if !strings.Contains(body, "Action Channel") {
		t.Fatal("channels page should list the action channel")
	}
	if !strings.Contains(body, "9000000000000000001") {
		t.Fatal("channels page should show the configured vote channel ID")
	}
	if !strings.Contains(body, "UNVERIFIED") {
		t.Fatal("with Discord disabled, configured channels should be UNVERIFIED (not falsely OK)")
	}
	if !strings.Contains(body, "MISSING") {
		t.Fatal("the unconfigured lifeboard should be flagged MISSING")
	}
	if !strings.Contains(body, "Discord is disabled") {
		t.Fatal("channels page should explain Discord is disabled in web-only mode")
	}
}

// TestChannelsRequiresAuth: the channels page is behind the session gate.
func TestChannelsRequiresAuth(t *testing.T) {
	pool := mustPool(t)
	client := newTestClient(t, testServer(t, pool))

	resp := client.get("/channels")
	if resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("GET /channels unauthenticated: expected 303, got %d", resp.StatusCode)
	}
}
