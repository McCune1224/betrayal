package web_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
)

type apiAuthState struct {
	Authenticated bool `json:"authenticated"`
}

type apiCSRFToken struct {
	Token string `json:"token"`
}

type apiErrorResponse struct {
	Error struct {
		Code    string          `json:"code"`
		Message string          `json:"message"`
		Fields  json.RawMessage `json:"fields"`
	} `json:"error"`
}

func TestAPIAuth(t *testing.T) {
	pool := mustPool(t)
	client := newTestClient(t, testServer(t, pool))

	t.Run("session reports unauthenticated JSON without redirect", func(t *testing.T) {
		resp := client.get("/api/v1/auth/session")
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("GET session: status = %d, want %d", resp.StatusCode, http.StatusOK)
		}
		assertAPIJSON(t, resp)
		if location := resp.Header.Get("Location"); location != "" {
			t.Errorf("GET session: Location = %q, want no redirect", location)
		}
		var state apiAuthState
		decodeAPIJSON(t, resp, &state)
		if state.Authenticated {
			t.Error("GET session authenticated = true, want false")
		}
	})

	t.Run("csrf returns a token suitable for the CSRF header", func(t *testing.T) {
		resp := client.get("/api/v1/auth/csrf")
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("GET csrf: status = %d, want %d", resp.StatusCode, http.StatusOK)
		}
		assertAPIJSON(t, resp)
		var body apiCSRFToken
		decodeAPIJSON(t, resp, &body)
		if body.Token == "" {
			t.Fatal("GET csrf: empty token")
		}
		if body.Token != client.csrfToken() {
			t.Errorf("GET csrf token does not match CSRF cookie")
		}
	})

	t.Run("login rejects invalid JSON with a canonical JSON error", func(t *testing.T) {
		resp := apiRequest(t, client, http.MethodPost, "/api/v1/auth/login", []byte("{"), true)
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("invalid JSON login: status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
		}
		assertAPIError(t, resp, "invalid_request")
	})

	t.Run("login rejects an empty password with a canonical JSON error", func(t *testing.T) {
		resp := apiRequest(t, client, http.MethodPost, "/api/v1/auth/login", []byte(`{"password":""}`), true)
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("empty password login: status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
		}
		assertAPIError(t, resp, "invalid_request")
	})

	t.Run("login rejects an invalid password without redirect", func(t *testing.T) {
		resp := apiRequest(t, client, http.MethodPost, "/api/v1/auth/login", []byte(`{"password":"wrong"}`), true)
		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("invalid password login: status = %d, want %d", resp.StatusCode, http.StatusUnauthorized)
		}
		if location := resp.Header.Get("Location"); location != "" {
			t.Errorf("invalid password login: Location = %q, want no redirect", location)
		}
		assertAPIError(t, resp, "invalid_credentials")
	})

	t.Run("login establishes the signed session and session reports it", func(t *testing.T) {
		resp := apiRequest(t, client, http.MethodPost, "/api/v1/auth/login", []byte(`{"password":"hunter2-test-password"}`), true)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("valid login: status = %d, want %d", resp.StatusCode, http.StatusOK)
		}
		assertAPIJSON(t, resp)
		var state apiAuthState
		decodeAPIJSON(t, resp, &state)
		if !state.Authenticated {
			t.Error("valid login authenticated = false, want true")
		}

		resp = client.get("/api/v1/auth/session")
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("authenticated session: status = %d, want %d", resp.StatusCode, http.StatusOK)
		}
		var session apiAuthState
		decodeAPIJSON(t, resp, &session)
		if !session.Authenticated {
			t.Error("authenticated session = false, want true")
		}
	})

	t.Run("logout rejects a missing CSRF token as canonical JSON", func(t *testing.T) {
		loginAPI(t, client)
		resp := apiRequest(t, client, http.MethodPost, "/api/v1/auth/logout", nil, false)
		if resp.StatusCode != http.StatusForbidden {
			t.Fatalf("logout without CSRF: status = %d, want %d", resp.StatusCode, http.StatusForbidden)
		}
		assertAPIError(t, resp, "csrf_token_invalid")
	})

	t.Run("logout requires auth and CSRF then clears the session", func(t *testing.T) {
		loginAPI(t, client)
		resp := apiRequest(t, client, http.MethodPost, "/api/v1/auth/logout", nil, true)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("logout: status = %d, want %d", resp.StatusCode, http.StatusOK)
		}
		assertAPIJSON(t, resp)
		var state apiAuthState
		decodeAPIJSON(t, resp, &state)
		if state.Authenticated {
			t.Error("logout authenticated = true, want false")
		}

		resp = client.get("/api/v1/auth/session")
		var session apiAuthState
		decodeAPIJSON(t, resp, &session)
		if session.Authenticated {
			t.Error("session after logout authenticated = true, want false")
		}
	})
}

func loginAPI(t *testing.T, client *testClient) {
	t.Helper()
	resp := apiRequest(t, client, http.MethodPost, "/api/v1/auth/login", []byte(`{"password":"hunter2-test-password"}`), true)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("API login: status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	resp.Body.Close()
}

func apiRequest(t *testing.T, client *testClient, method, path string, body []byte, csrf bool) *http.Response {
	t.Helper()
	req, err := http.NewRequest(method, client.base+path, bytes.NewReader(body))
	if err != nil {
		t.Fatalf("new API request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if csrf && client.csrfToken() != "" {
		req.Header.Set("X-CSRF-Token", client.csrfToken())
	}
	resp, err := (&http.Client{Jar: client.jar, CheckRedirect: func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}}).Do(req)
	if err != nil {
		t.Fatalf("API request %s %s: %v", method, path, err)
	}
	t.Cleanup(func() { resp.Body.Close() })
	return resp
}

func assertAPIJSON(t *testing.T, resp *http.Response) {
	t.Helper()
	if contentType := resp.Header.Get("Content-Type"); contentType != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", contentType)
	}
}

func assertAPIError(t *testing.T, resp *http.Response, wantCode string) {
	t.Helper()
	assertAPIJSON(t, resp)
	var body apiErrorResponse
	decodeAPIJSON(t, resp, &body)
	if body.Error.Code != wantCode {
		t.Errorf("error code = %q, want %q", body.Error.Code, wantCode)
	}
	if body.Error.Message == "" {
		t.Error("error message is empty")
	}
	if body.Error.Fields == nil {
		t.Error("error fields key is missing")
	}
}

func decodeAPIJSON(t *testing.T, resp *http.Response, into any) {
	t.Helper()
	if err := json.NewDecoder(resp.Body).Decode(into); err != nil {
		t.Fatalf("decode API JSON: %v", err)
	}
}
