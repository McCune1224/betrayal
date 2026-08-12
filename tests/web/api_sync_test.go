package web_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
)

func TestAPISyncRoutesRequireJSONAuth(t *testing.T) {
	pool := mustPool(t)
	client := newTestClient(t, testServer(t, pool))
	for _, path := range []string{"/api/v1/sync/sources", "/api/v1/sync/preview"} {
		resp := client.get(path)
		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("%s status=%d want 401", path, resp.StatusCode)
		}
		assertAPIError(t, resp, "unauthorized")
	}
}

func TestAPISyncApplyRejectsUnknownSourceWithout501(t *testing.T) {
	pool := mustPool(t)
	client := newTestClient(t, testServer(t, pool))
	client.get("/api/v1/auth/csrf")
	loginAPI(t, client)
	resp := apiRequest(t, client, http.MethodPost, "/api/v1/sync/apply", []byte(`{"source_id":999999}`), true)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status=%d want 400", resp.StatusCode)
	}
	if resp.StatusCode == http.StatusNotImplemented {
		t.Fatal("sync apply must not remain a placeholder")
	}
}

func TestAPISyncApplyQueuesRunAndExposesStatus(t *testing.T) {
	pool := mustPool(t)
	_, err := pool.Exec(t.Context(), `INSERT INTO sync_source (name, kind, alignment, url, enabled) VALUES ('test-roles', 'roles', 'good', 'https://example.com/roles.csv', true)`)
	if err != nil {
		t.Fatalf("seed source: %v", err)
	}
	client := newTestClient(t, testServer(t, pool))
	client.get("/api/v1/auth/csrf")
	loginAPI(t, client)
	resp := apiRequest(t, client, http.MethodPost, "/api/v1/sync/apply", []byte(`{"source_id":1}`), true)
	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("status=%d want 202", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	var envelope struct {
		Run struct {
			ID     int64  `json:"id"`
			Status string `json:"status"`
		} `json:"run"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if envelope.Run.ID == 0 || envelope.Run.Status != "pending" {
		t.Fatalf("queued run=%+v", envelope.Run)
	}
	status := client.get("/api/v1/sync/runs/" + fmt.Sprint(envelope.Run.ID))
	if status.StatusCode != http.StatusOK {
		t.Fatalf("status endpoint=%d want 200", status.StatusCode)
	}
}
