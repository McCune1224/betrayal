package handlers

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
	"github.com/mccune1224/betrayal/internal/web/templates/pages"
)

const (
	sessionName    = "betrayal-admin"
	sessionKeyAuth = "authenticated"
)

// AuthHandler handles authentication requests
type AuthHandler struct {
	store         *sessions.CookieStore
	adminPassword string
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(store *sessions.CookieStore, adminPassword string) *AuthHandler {
	return &AuthHandler{
		store:         store,
		adminPassword: adminPassword,
	}
}

// LoginPage handles GET /login
func (h *AuthHandler) LoginPage(c echo.Context) error {
	// Check if already authenticated
	session, _ := h.store.Get(c.Request(), sessionName)
	if auth, ok := session.Values[sessionKeyAuth].(bool); ok && auth {
		return c.Redirect(http.StatusSeeOther, "/")
	}

	errorMsg := c.QueryParam("error")
	return render(c, http.StatusOK, pages.Login(errorMsg))
}

// Login handles POST /login
func (h *AuthHandler) Login(c echo.Context) error {
	password := c.FormValue("password")

	if password != h.adminPassword {
		return c.Redirect(http.StatusSeeOther, "/login?error=Invalid+password")
	}

	session, _ := h.store.Get(c.Request(), sessionName)
	session.Values[sessionKeyAuth] = true
	if err := session.Save(c.Request(), c.Response()); err != nil {
		return c.Redirect(http.StatusSeeOther, "/login?error=Session+error")
	}

	return c.Redirect(http.StatusSeeOther, "/")
}

// Logout handles POST /logout
func (h *AuthHandler) Logout(c echo.Context) error {
	session, _ := h.store.Get(c.Request(), sessionName)
	session.Values[sessionKeyAuth] = false
	session.Options.MaxAge = -1 // Delete cookie
	session.Save(c.Request(), c.Response())

	return c.Redirect(http.StatusSeeOther, "/login")
}

// render is a helper function to render templ components
func render(c echo.Context, status int, t templ.Component) error {
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
	c.Response().WriteHeader(status)
	return t.Render(c.Request().Context(), c.Response())
}
