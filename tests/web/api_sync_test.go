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
