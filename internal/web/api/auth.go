package api

import (
	"crypto/subtle"
	"encoding/json"
	"io"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
)

const (
	sessionName    = "betrayal-admin"
	sessionKeyAuth = "authenticated"
	csrfContextKey = "csrf"
)

// AuthHandler provides JSON authentication endpoints backed by the existing
// signed browser session cookie.
type AuthHandler struct {
	store         *sessions.CookieStore
	adminPassword string
}

// NewAuthHandler creates a JSON authentication handler.
func NewAuthHandler(store *sessions.CookieStore, adminPassword string) *AuthHandler {
	return &AuthHandler{store: store, adminPassword: adminPassword}
}

// Session returns the current authentication state without redirecting.
func (h *AuthHandler) Session(c echo.Context) error {
	WriteJSON(c.Response(), http.StatusOK, map[string]bool{"authenticated": h.authenticated(c)})
	return nil
}

// CSRF returns Echo's double-submit token for use in the X-CSRF-Token header.
func (h *AuthHandler) CSRF(c echo.Context) error {
	token, _ := c.Get(csrfContextKey).(string)
	WriteJSON(c.Response(), http.StatusOK, map[string]string{"token": token})
	return nil
}

// Login validates a JSON password and establishes the signed session cookie.
func (h *AuthHandler) Login(c echo.Context) error {
	var request struct {
		Password string `json:"password"`
	}
	decoder := json.NewDecoder(c.Request().Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil || request.Password == "" || decoder.Decode(&struct{}{}) != io.EOF {
		WriteError(c.Response(), http.StatusBadRequest, "invalid_request", "password is required", map[string]any{"password": "required"})
		return nil
	}
	if subtle.ConstantTimeCompare([]byte(request.Password), []byte(h.adminPassword)) != 1 {
		WriteError(c.Response(), http.StatusUnauthorized, "invalid_credentials", "invalid password", map[string]any{})
		return nil
	}

	session, err := h.store.Get(c.Request(), sessionName)
	if err != nil {
		WriteError(c.Response(), http.StatusInternalServerError, "session_error", "could not create session", map[string]any{})
		return nil
	}
	session.Values[sessionKeyAuth] = true
	if err := session.Save(c.Request(), c.Response()); err != nil {
		WriteError(c.Response(), http.StatusInternalServerError, "session_error", "could not save session", map[string]any{})
		return nil
	}
	WriteJSON(c.Response(), http.StatusOK, map[string]bool{"authenticated": true})
	return nil
}

// Logout clears the signed session cookie. Authentication and CSRF validation
// are applied by route middleware.
func (h *AuthHandler) Logout(c echo.Context) error {
	session, err := h.store.Get(c.Request(), sessionName)
	if err != nil {
		WriteError(c.Response(), http.StatusInternalServerError, "session_error", "could not clear session", map[string]any{})
		return nil
	}
	session.Values[sessionKeyAuth] = false
	session.Options.MaxAge = -1
	if err := session.Save(c.Request(), c.Response()); err != nil {
		WriteError(c.Response(), http.StatusInternalServerError, "session_error", "could not clear session", map[string]any{})
		return nil
	}
	WriteJSON(c.Response(), http.StatusOK, map[string]bool{"authenticated": false})
	return nil
}

func (h *AuthHandler) authenticated(c echo.Context) bool {
	session, err := h.store.Get(c.Request(), sessionName)
	if err != nil {
		return false
	}
	authenticated, ok := session.Values[sessionKeyAuth].(bool)
	return ok && authenticated
}

// AuthMiddleware protects API routes with JSON 401 responses instead of HTML
// redirects, so future API endpoints can compose it directly.
type AuthMiddleware struct {
	store *sessions.CookieStore
}

// NewAuthMiddleware creates API authentication middleware.
func NewAuthMiddleware(store *sessions.CookieStore) *AuthMiddleware {
	return &AuthMiddleware{store: store}
}

// RequireAuth rejects requests without a valid signed session cookie.
func (m *AuthMiddleware) RequireAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		session, err := m.store.Get(c.Request(), sessionName)
		if err != nil {
			WriteError(c.Response(), http.StatusUnauthorized, "unauthorized", "authentication required", map[string]any{})
			return nil
		}
		authenticated, ok := session.Values[sessionKeyAuth].(bool)
		if !ok || !authenticated {
			WriteError(c.Response(), http.StatusUnauthorized, "unauthorized", "authentication required", map[string]any{})
			return nil
		}
		return next(c)
	}
}
