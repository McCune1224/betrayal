package middleware

import (
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
)

const (
	sessionName    = "betrayal-admin"
	sessionKeyAuth = "authenticated"
)

// AuthMiddleware provides authentication middleware
type AuthMiddleware struct {
	store *sessions.CookieStore
}

// NewAuthMiddleware creates a new AuthMiddleware
func NewAuthMiddleware(store *sessions.CookieStore) *AuthMiddleware {
	return &AuthMiddleware{store: store}
}

// RequireAuth is middleware that requires authentication
func (m *AuthMiddleware) RequireAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		session, err := m.store.Get(c.Request(), sessionName)
		if err != nil {
			return c.Redirect(http.StatusSeeOther, "/login")
		}

		auth, ok := session.Values[sessionKeyAuth].(bool)
		if !ok || !auth {
			return c.Redirect(http.StatusSeeOther, "/login")
		}

		return next(c)
	}
}
