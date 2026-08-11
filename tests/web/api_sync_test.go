package web_test

import (
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
