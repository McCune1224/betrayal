package web_test

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestWhisperSettingsLoadsForAuthenticatedAdmin(t *testing.T) {
	pool := mustPool(t)
	srv := testServer(t, pool)
	client := newTestClient(t, srv)
	client.login()

	resp := client.get("/api/v1/whisper")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("whisper settings: status = %d, want %d: %s", resp.StatusCode, http.StatusOK, client.body(resp))
	}

	var body struct {
		Groups   []json.RawMessage `json:"groups"`
		Players  []json.RawMessage `json:"players"`
		Messages []json.RawMessage `json:"messages"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode whisper settings: %v", err)
	}
	if body.Groups == nil || body.Players == nil || body.Messages == nil {
		t.Fatalf("whisper settings returned null collections: %+v", body)
	}
}
