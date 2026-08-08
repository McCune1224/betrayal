package web_test

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/mccune1224/betrayal/internal/web"
	"github.com/rs/zerolog"
)

// TestNewRequiresExplicitSessionSecret: session signing must use a dedicated,
// explicit secret rather than deriving one from the admin password.
func TestNewRequiresExplicitSessionSecret(t *testing.T) {
	pool := mustPool(t)

	_, err := web.New(pool, nil, zerolog.Nop(), web.Config{
		Port:          "0",
		AdminPassword: testAdminPassword,
		SessionSecret: "",
	})
	if err == nil || !strings.Contains(err.Error(), "SESSION_SECRET") {
		t.Fatalf("expected explicit session secret error, got: %v", err)
	}
}

func TestNewRefusesWithoutPasswordOrSessionSecret(t *testing.T) {
	pool := mustPool(t)
	_, err := web.New(pool, nil, zerolog.Nop(), web.Config{Port: "0"})
	if err == nil || !strings.Contains(err.Error(), "ADMIN_PASSWORD") {
		t.Fatalf("expected missing password error, got: %v", err)
	}
}

// TestNewRefusesShortSessionSecret: gorilla/securecookie panics on keys shorter
// than 32 bytes — the server must refuse instead of crashing at runtime.
func TestNewRefusesShortSessionSecret(t *testing.T) {
	pool := mustPool(t)

	_, err := web.New(pool, nil, zerolog.Nop(), web.Config{
		Port:          "0",
		AdminPassword: testAdminPassword,
		SessionSecret: "too-short", // 10 bytes
	})
	if err == nil {
		t.Fatal("expected error for short SESSION_SECRET, got nil")
	}
}

// TestNewAcceptsValidConfig: a 32+ byte secret starts cleanly.
func TestNewAcceptsValidConfig(t *testing.T) {
	pool := mustPool(t)

	srv, err := web.New(pool, nil, zerolog.Nop(), web.Config{
		Port:          "0",
		AdminPassword: testAdminPassword,
		SessionSecret: testSessionSecret,
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if srv == nil {
		t.Fatal("expected non-nil server")
	}
}

// TestCSRFBlocksTokenlessPost: a state-changing POST without the CSRF token is
// rejected even with the correct password (login is the attack surface).
func TestCSRFBlocksTokenlessPost(t *testing.T) {
	pool := mustPool(t)
	client := newTestClient(t, testServer(t, pool))

	client.get("/login") // establishes the _csrf cookie but we won't send it

	req := client.doRawWithoutCSRF(http.MethodPost, "/login", url.Values{"password": {testAdminPassword}})
	if req.StatusCode != http.StatusBadRequest {
		t.Fatalf("tokenless POST /login: expected 400, got %d", req.StatusCode)
	}
}

// TestLoginFlow: wrong password redirects with error; right password logs in;
// logout clears the session.
func TestLoginFlow(t *testing.T) {
	pool := mustPool(t)
	client := newTestClient(t, testServer(t, pool))

	// Login page renders
	resp := client.get("/login")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /login: expected 200, got %d", resp.StatusCode)
	}
	if !strings.Contains(client.body(resp), "PASSWORD") {
		t.Fatal("login page should contain the password field")
	}

	// Unauthenticated dashboard redirects to login
	resp = client.get("/")
	if resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("GET / unauthenticated: expected 303, got %d", resp.StatusCode)
	}
	if loc := resp.Header.Get("Location"); loc != "/login" {
		t.Fatalf("expected redirect to /login, got %q", loc)
	}

	// Wrong password
	resp = client.do(http.MethodPost, "/login", url.Values{"password": {"wrong"}}, nil)
	if resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("wrong password: expected 303, got %d", resp.StatusCode)
	}
	if loc := resp.Header.Get("Location"); !strings.Contains(loc, "error=") {
		t.Fatalf("wrong password should redirect with error, got %q", loc)
	}

	// Right password
	resp = client.do(http.MethodPost, "/login", url.Values{"password": {testAdminPassword}}, nil)
	if resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("login: expected 303, got %d", resp.StatusCode)
	}
	if loc := resp.Header.Get("Location"); loc != "/" {
		t.Fatalf("login should redirect to /, got %q", loc)
	}

	// Now the dashboard is reachable
	resp = client.get("/")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET / authenticated: expected 200, got %d", resp.StatusCode)
	}
	if !strings.Contains(client.body(resp), "DASHBOARD") {
		t.Fatal("dashboard should render after login")
	}

	// Logout
	resp = client.do(http.MethodPost, "/logout", nil, nil)
	if resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("logout: expected 303, got %d", resp.StatusCode)
	}
	resp = client.get("/")
	if resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("GET / after logout: expected 303 to /login, got %d", resp.StatusCode)
	}
}

// TestLoginRateLimit: after the burst is exhausted the login endpoint returns
// 429 to further attempts.
func TestLoginRateLimit(t *testing.T) {
	pool := mustPool(t)
	client := newTestClient(t, testServer(t, pool))

	client.get("/login") // establish CSRF cookie

	statuses := make([]int, 0, 14)
	for i := 0; i < 14; i++ {
		resp := client.do(http.MethodPost, "/login", url.Values{"password": {"wrong"}}, nil)
		statuses = append(statuses, resp.StatusCode)
	}

	// The burst window allows the first several; the tail must be limited.
	limited := 0
	for _, s := range statuses {
		if s == http.StatusTooManyRequests {
			limited++
		}
	}
	if limited == 0 {
		t.Fatalf("expected at least one 429 after exhausting the login burst, got all: %v", statuses)
	}
}
