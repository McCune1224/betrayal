package web

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mccune1224/betrayal/internal/models"
	"github.com/mccune1224/betrayal/internal/web"
	"github.com/mccune1224/betrayal/tests/testutil"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/suite"
)

const (
	testPassword = "test-admin-password"
	testSecret   = "test-session-secret"
)

// WebServerSuite drives the real Echo routes with httptest against the LOCAL
// database. Session auth is exercised through a cookie jar, so the same flow
// a browser performs is covered end-to-end (minus the templates' markup).
type WebServerSuite struct {
	suite.Suite
	DB               *pgxpool.Pool
	server           *web.Server
	ts               *httptest.Server
	client           *http.Client
	clientNoRedirect *http.Client
}

func (s *WebServerSuite) SetupSuite() {
	s.DB = testutil.NewTestPool(s.T())
	s.server = web.New(s.DB, nil, zerolog.Nop(), web.Config{
		Port:          "0",
		AdminPassword: testPassword,
		SessionSecret: testSecret,
	})
	s.ts = httptest.NewServer(s.server.Handler())
}

func (s *WebServerSuite) TearDownSuite() {
	s.ts.Close()
	s.DB.Close()
}

func (s *WebServerSuite) SetupTest() {
	testutil.TruncateAll(s.T(), s.DB)
	// Fresh cookie jar per test: sessions must not leak between tests.
	jar, err := cookiejar.New(nil)
	s.Require().NoError(err)
	s.client = &http.Client{Jar: jar}
	// A client that does NOT follow redirects, for asserting 30x statuses.
	s.clientNoRedirect = &http.Client{
		Jar: jar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

func (s *WebServerSuite) get(path string) *http.Response {
	resp, err := s.client.Get(s.ts.URL + path)
	s.Require().NoError(err)
	return resp
}

func (s *WebServerSuite) postForm(path string, form url.Values) *http.Response {
	resp, err := s.client.PostForm(s.ts.URL+path, form)
	s.Require().NoError(err)
	return resp
}

func (s *WebServerSuite) postFormNoRedirect(path string, form url.Values) *http.Response {
	resp, err := s.clientNoRedirect.PostForm(s.ts.URL+path, form)
	s.Require().NoError(err)
	return resp
}

func (s *WebServerSuite) putForm(path string, form url.Values) *http.Response {
	req, err := http.NewRequest(http.MethodPut, s.ts.URL+path, strings.NewReader(form.Encode()))
	s.Require().NoError(err)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := s.client.Do(req)
	s.Require().NoError(err)
	return resp
}

func (s *WebServerSuite) TestHealth() {
	resp := s.get("/health")
	s.Require().Equal(http.StatusOK, resp.StatusCode)
	defer resp.Body.Close()

	var body map[string]string
	s.Require().NoError(json.NewDecoder(resp.Body).Decode(&body))
	s.Equal("ok", body["status"])
	s.Equal("ok", body["database"])
}

func (s *WebServerSuite) TestHealthStatusPartialRequiresAuth() {
	resp := s.get("/health/status")
	defer resp.Body.Close()
	// Redirects to /login when unauthenticated.
	s.Require().Equal(http.StatusOK, resp.StatusCode)
	s.Require().Contains(resp.Request.URL.Path, "/login")
}

func (s *WebServerSuite) TestLoginPage() {
	resp := s.get("/login")
	defer resp.Body.Close()
	s.Require().Equal(http.StatusOK, resp.StatusCode)
}

func (s *WebServerSuite) TestLoginWrongPasswordRedirectsToError() {
	resp := s.postFormNoRedirect("/login", url.Values{"password": {"nope"}})
	defer resp.Body.Close()
	s.Require().Equal(http.StatusSeeOther, resp.StatusCode)
	s.Require().Contains(resp.Header.Get("Location"), "/login?error=")
}

func (s *WebServerSuite) TestLoginCorrectPasswordAuthenticates() {
	resp := s.postFormNoRedirect("/login", url.Values{"password": {testPassword}})
	s.Require().Equal(http.StatusSeeOther, resp.StatusCode)
	s.Require().Equal("/", resp.Header.Get("Location"))
	resp.Body.Close()

	// Session cookie now unlocks protected routes.
	resp = s.get("/players")
	defer resp.Body.Close()
	s.Require().Equal(http.StatusOK, resp.StatusCode)
}

func (s *WebServerSuite) TestPlayersRequiresAuth() {
	resp := s.get("/players")
	defer resp.Body.Close()
	s.Require().Contains(resp.Request.URL.Path, "/login")
}

func (s *WebServerSuite) TestPlayersList() {
	s.login()
	resp := s.get("/players")
	defer resp.Body.Close()
	s.Require().Equal(http.StatusOK, resp.StatusCode)
}

func (s *WebServerSuite) TestPlayersDetail() {
	s.login()
	player := s.seedPlayer()

	resp := s.get(fmt.Sprintf("/players/%d", player.ID))
	defer resp.Body.Close()
	s.Require().Equal(http.StatusOK, resp.StatusCode)

	// Unknown player -> 404.
	resp = s.get("/players/999999")
	defer resp.Body.Close()
	s.Require().Equal(http.StatusNotFound, resp.StatusCode)
}

func (s *WebServerSuite) TestRolesListAndDetail() {
	s.login()
	role := s.seedRole()

	resp := s.get("/roles")
	defer resp.Body.Close()
	s.Require().Equal(http.StatusOK, resp.StatusCode)

	resp = s.get(fmt.Sprintf("/roles/%d", role.ID))
	defer resp.Body.Close()
	s.Require().Equal(http.StatusOK, resp.StatusCode)

	resp = s.get("/roles/999999")
	defer resp.Body.Close()
	s.Require().Equal(http.StatusNotFound, resp.StatusCode)
}

func (s *WebServerSuite) TestRolesUpdate() {
	s.login()
	role := s.seedRole()

	resp := s.putForm(fmt.Sprintf("/roles/%d", role.ID), url.Values{
		"name":        {"Mafia Boss"},
		"description": {"Updated description"},
		"alignment":   {"EVIL"},
	})
	defer resp.Body.Close()
	s.Require().Equal(http.StatusOK, resp.StatusCode)
	s.Require().Contains(resp.Header.Get("HX-Trigger"), "success")

	// Persisted.
	got, err := models.New(s.DB).GetRole(context.Background(), role.ID)
	s.Require().NoError(err)
	s.Equal("Mafia Boss", got.Name)
	s.Equal("Updated description", got.Description)
}

func (s *WebServerSuite) TestRolesUpdateInvalidAlignment() {
	s.login()
	role := s.seedRole()

	resp := s.putForm(fmt.Sprintf("/roles/%d", role.ID), url.Values{
		"name":        {"Mafia Boss"},
		"description": {"Updated"},
		"alignment":   {"CHAOTIC"},
	})
	defer resp.Body.Close()
	s.Require().Equal(http.StatusBadRequest, resp.StatusCode)
	s.Require().Contains(resp.Header.Get("HX-Trigger"), "error")
}

func (s *WebServerSuite) TestLogout() {
	s.login()
	resp := s.postFormNoRedirect("/logout", nil)
	s.Require().Equal(http.StatusSeeOther, resp.StatusCode)
	resp.Body.Close()

	// Protected routes redirect to login again.
	resp = s.get("/players")
	defer resp.Body.Close()
	s.Require().Contains(resp.Request.URL.Path, "/login")
}

// --- helpers ---

func (s *WebServerSuite) login() {
	resp := s.postFormNoRedirect("/login", url.Values{"password": {testPassword}})
	s.Require().Equal(http.StatusSeeOther, resp.StatusCode)
	resp.Body.Close()
}

func (s *WebServerSuite) seedRole() models.Role {
	role, err := models.New(s.DB).CreateRole(context.Background(), models.CreateRoleParams{
		Name: "Mafia", Description: "The mafia boss", Alignment: models.AlignmentEVIL,
	})
	s.Require().NoError(err)
	return role
}

func (s *WebServerSuite) seedPlayer() models.Player {
	role := s.seedRole()
	player, err := models.New(s.DB).CreatePlayer(context.Background(), models.CreatePlayerParams{
		ID:        100000000000000001,
		RoleID:    pgtype.Int4{Int32: role.ID, Valid: true},
		Alive:     true,
		Coins:     200,
		CoinBonus: pgtype.Numeric{}, // NULL -> column DEFAULT 0
		Luck:      0,
		Alignment: models.AlignmentEVIL,
	})
	s.Require().NoError(err)
	return player
}

func (s *WebServerSuite) TestPlayersPageRendersPlayerRow() {
	s.login()
	player := s.seedPlayer()

	resp := s.get("/players/table")
	defer resp.Body.Close()
	s.Require().Equal(http.StatusOK, resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	s.Require().NoError(err)
	s.Require().Contains(string(body), fmt.Sprintf("%d", player.ID))
}

func TestWebServerSuite(t *testing.T) {
	suite.Run(t, new(WebServerSuite))
}
