package web_test

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestAPIShell(t *testing.T) {
	pool := mustPool(t)
	client := newTestClient(t, testServer(t, pool))

	t.Run("serves the embedded UI index for the root and client routes", func(t *testing.T) {
		for _, path := range []string{"/", "/ui/client-route"} {
			resp := client.get(path)
			if resp.StatusCode != http.StatusOK {
				t.Fatalf("GET %s: status = %d, want %d", path, resp.StatusCode, http.StatusOK)
			}
			if contentType := resp.Header.Get("Content-Type"); contentType != "text/html; charset=utf-8" {
				t.Errorf("GET %s: Content-Type = %q, want HTML", path, contentType)
			}
			if cacheControl := resp.Header.Get("Cache-Control"); cacheControl != "no-cache" {
				t.Errorf("GET %s: Cache-Control = %q, want no-cache", path, cacheControl)
			}
			if got := client.body(resp); got == "" {
				t.Errorf("GET %s: empty index body", path)
			}
		}
	})

	t.Run("serves hash-named assets with immutable caching", func(t *testing.T) {
		resp := client.get("/assets/app.12345678.js")
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("asset status = %d, want %d", resp.StatusCode, http.StatusOK)
		}
		if cacheControl := resp.Header.Get("Cache-Control"); cacheControl != "public, max-age=31536000, immutable" {
			t.Errorf("asset Cache-Control = %q", cacheControl)
		}
	})

	t.Run("serves API health as JSON", func(t *testing.T) {
		resp := client.get("/api/v1/health")
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("health status = %d, want %d", resp.StatusCode, http.StatusOK)
		}
		if contentType := resp.Header.Get("Content-Type"); contentType != "application/json" {
			t.Errorf("health Content-Type = %q, want application/json", contentType)
		}
		var body map[string]string
		if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
			t.Fatalf("decode health JSON: %v", err)
		}
		if body["status"] != "ok" {
			t.Errorf("health status body = %q, want ok", body["status"])
		}
	})

	t.Run("keeps unknown API routes as JSON errors", func(t *testing.T) {
		for _, path := range []string{"/api/v1", "/api/v1/missing"} {
			resp := client.get(path)
			if resp.StatusCode != http.StatusNotFound {
				t.Fatalf("GET %s: status = %d, want %d", path, resp.StatusCode, http.StatusNotFound)
			}
			if contentType := resp.Header.Get("Content-Type"); contentType != "application/json" {
				t.Errorf("GET %s: Content-Type = %q, want application/json", path, contentType)
			}
			var body struct {
				Error struct {
					Code    string          `json:"code"`
					Message string          `json:"message"`
					Fields  json.RawMessage `json:"fields"`
				} `json:"error"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
				t.Fatalf("decode error JSON: %v", err)
			}
			if body.Error.Code != "not_found" {
				t.Errorf("GET %s: error code = %q, want not_found", path, body.Error.Code)
			}
			if body.Error.Message == "" {
				t.Errorf("GET %s: error message is empty", path)
			}
			if body.Error.Fields == nil {
				t.Errorf("GET %s: error fields key is missing", path)
			}
		}
	})
}
