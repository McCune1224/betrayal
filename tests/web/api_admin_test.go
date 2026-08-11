package web_test

import (
	"net/http"
	"testing"
)

func TestAPIAdminAuditIsJSONAndAuthenticated(t *testing.T) {
	pool := mustPool(t)
	client := newTestClient(t, testServer(t, pool))
	resp := client.get("/api/v1/admin/audit")
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", resp.StatusCode)
	}
	if got := resp.Header.Get("Location"); got != "" {
		t.Fatalf("unexpected redirect %q", got)
	}
	assertAPIError(t, resp, "unauthorized")
}

func TestAPIAdminRoutesExist(t *testing.T) {
	pool := mustPool(t)
	client := newTestClient(t, testServer(t, pool))
	client.get("/api/v1/auth/csrf")
	loginAPI(t, client)
	for _, path := range []string{"/api/v1/admin/audit", "/api/v1/admin/migrations", "/api/v1/admin/reset"} {
		resp := client.get(path)
		if resp.StatusCode == http.StatusNotFound {
			t.Fatalf("%s unexpectedly missing", path)
		}
	}
}
