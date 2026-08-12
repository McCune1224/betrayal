// Package web provides the admin web server for the Betrayal Bot
package web

import (
	"context"
	"crypto/sha256"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/gorilla/sessions"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	dbmigrate "github.com/mccune1224/betrayal/internal/db/migrate"
	"github.com/mccune1224/betrayal/internal/services/datasync"
	"github.com/mccune1224/betrayal/internal/services/gamereset"
	"github.com/mccune1224/betrayal/internal/web/api"
	webmiddleware "github.com/mccune1224/betrayal/internal/web/middleware"
	"github.com/mccune1224/betrayal/internal/web/railway"
	"github.com/mccune1224/betrayal/internal/web/ui"
	"github.com/rs/zerolog"
	"golang.org/x/time/rate"
)

// Login rate limiting: allows a short burst of attempts, then throttles to a
// trickle per IP. Generous for a solo admin panel, hostile to brute force.
const (
	loginRateLimit    rate.Limit = 1   // requests/sec sustained
	loginRateBurst    int        = 10  // burst before throttling
	redeployRate      rate.Limit = 0.1 // 1 request / 10s
	redeployRateBurst int        = 2   // max 2 within the window
)

// Config holds the web server configuration
type Config struct {
	Port          string
	AdminPassword string

	// DatabaseURL is the DSN the server's DB pool was built from. Used by the
	// production context for the admin UI and operational logging.
	DatabaseURL string
	// Environment (ENVIRONMENT env var) is also treated as production so the UI
	// remains explicit even if the deployed DSN host changes.
	Environment string
	// Production is the intended target; this field is only used to label the UI.
	// SyncEnvURLs maps sync source names to their CSV URLs from the
	// environment, used to seed the sync_source table (empty for tests).
	SyncEnvURLs map[string]string
	// AllowUnsafeSyncURLs is intended only for localhost fixture tests. It
	// permits HTTP/private hosts; production configuration must leave it false.
	AllowUnsafeSyncURLs bool

	// Railway API configuration
	RailwayToken     string
	RailwayProjectID string
	RailwayServiceID string
	RailwayEnvID     string
}

// Server holds the Echo instance and dependencies
type Server struct {
	echo           *echo.Echo
	dbPool         *pgxpool.Pool
	discordSession *discordgo.Session
	logger         zerolog.Logger
	config         Config
	sessionStore   *sessions.CookieStore
	railwayClient  *railway.Client
	syncService    *datasync.Service
	syncHandler    *api.SyncHandler

	// migrateRunner is built lazily on first use (the embedded runner opens a
	// connection eagerly, so constructing it at startup would add a blocking
	// connect to boot and an unused connection for web-only runs).
	migrateRunner     *dbmigrate.Runner
	migrateRunnerOnce sync.Once
}

// New creates a new web server. The admin password is also the sole source for
// the signed session cookie key, keeping deployment configuration to one
// credential as intended for this small private admin panel.
func New(pool *pgxpool.Pool, discord *discordgo.Session, logger zerolog.Logger, cfg Config) (*Server, error) {
	if cfg.AdminPassword == "" {
		return nil, fmt.Errorf("ADMIN_PASSWORD must be set to enable the web admin")
	}
	signingKey := sha256.Sum256([]byte(cfg.AdminPassword))

	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	store := sessions.NewCookieStore(signingKey[:])
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7, // 7 days
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}

	// Create Railway client
	railwayClient := railway.New(
		cfg.RailwayToken,
		cfg.RailwayProjectID,
		cfg.RailwayServiceID,
		cfg.RailwayEnvID,
	)

	syncService := datasync.New(pool, cfg.SyncEnvURLs)
	syncService.SetAllowUnsafeURLs(cfg.AllowUnsafeSyncURLs)
	s := &Server{
		echo:           e,
		dbPool:         pool,
		discordSession: discord,
		logger:         logger.With().Str("component", "web").Logger(),
		config:         cfg,
		sessionStore:   store,
		railwayClient:  railwayClient,
		syncService:    syncService,
	}

	// Seed the canonical sync sources (URLs only when rows still have the
	// placeholder). Non-fatal: the panel still works for configured sources.
	if err := s.syncService.SeedSources(context.Background()); err != nil {
		s.logger.Warn().Err(err).Msg("failed to seed sync sources")
	}

	s.setupMiddleware()
	s.setupRoutes()

	return s, nil
}

// getMigrateRunner lazily builds the embedded migrations runner bound to the
// configured DatabaseURL. Returns nil when the DSN is unset or the runner
// fails to initialize (the page renders an "unavailable" state).
func (s *Server) getMigrateRunner() *dbmigrate.Runner {
	s.migrateRunnerOnce.Do(func() {
		if s.config.DatabaseURL == "" {
			return
		}
		r, err := dbmigrate.New(s.config.DatabaseURL)
		if err != nil {
			s.logger.Warn().Err(err).Msg("failed to init embedded migrations runner")
			return
		}
		s.migrateRunner = r
	})
	return s.migrateRunner
}

func (s *Server) setupMiddleware() {
	// Recovery middleware
	s.echo.Use(middleware.Recover())

	// Request logging
	s.echo.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:     true,
		LogStatus:  true,
		LogMethod:  true,
		LogError:   true,
		LogLatency: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			entry := s.logger.Info().
				Str("method", v.Method).
				Str("uri", v.URI).
				Int("status", v.Status).
				Dur("latency", v.Latency)
			if v.Error != nil {
				entry = entry.Err(v.Error)
			}
			entry.Msg("request")
			return nil
		},
	}))

	// CSRF protection — double-submit cookie pattern. The SvelteKit client
	// reads the cookie and sends it as X-CSRF-Token on JSON mutations.
	// Token validation applies only to state-changing methods; safe methods
	// (GET/HEAD/OPTIONS/TRACE) are skipped by Echo.
	// The cookie is intentionally not HttpOnly so client JS can read it.
	s.echo.Use(middleware.CSRFWithConfig(middleware.CSRFConfig{
		TokenLookup:    "header:" + echo.HeaderXCSRFToken + ",form:_csrf",
		CookieName:     "_csrf",
		CookieHTTPOnly: false, // readable by JS so htmx can attach it to every request
		CookieSameSite: http.SameSiteLaxMode,
		CookiePath:     "/",
		ErrorHandler: func(err error, c echo.Context) error {
			if strings.HasPrefix(c.Request().URL.Path, "/api/") {
				api.WriteError(c.Response(), http.StatusForbidden, "csrf_token_invalid", "invalid CSRF token", map[string]any{})
				return nil
			}
			return err
		},
	}))

}

func (s *Server) setupRoutes() {
	apiAuthHandler := api.NewAuthHandler(s.sessionStore, s.config.AdminPassword)
	apiDashboardHandler := api.NewDashboardHandler(s.dbPool)
	apiPlayersHandler := api.NewPlayersHandler(s.dbPool)
	apiPlayersAdminHandler := api.NewPlayersHandler(s.dbPool)
	apiCatalogHandler := api.NewCatalogHandler(s.dbPool)
	apiCycleHandler := api.NewCycleHandler(s.dbPool)
	apiChannelsHandler := api.NewChannelsHandler(s.dbPool, s.discordSession)
	apiSetupHandler := api.NewSetupHandler(s.dbPool)
	apiVotesHandler := api.NewVotesHandler(s.dbPool)
	apiReadinessHandler := api.NewReadinessHandler(s.dbPool, s.discordSession)
	apiAdminHandler := api.NewAdminHandler(s.dbPool, s.railwayClient, s.getMigrateRunner, gamereset.New(s.dbPool, s.syncService))
	apiSyncHandler := api.NewSyncHandler(s.dbPool, s.syncService)
	s.syncHandler = apiSyncHandler
	apiDiscordResourceHandler := api.NewDiscordResourceHandler(s.discordSession)
	apiAuthMiddleware := api.NewAuthMiddleware(s.sessionStore)
	browserAuth := webmiddleware.NewAuthMiddleware(s.sessionStore)

	s.echo.GET("/health", func(c echo.Context) error {
		api.Health(c.Response(), c.Request())
		return nil
	})
	s.echo.HEAD("/health", func(c echo.Context) error {
		api.Health(c.Response(), c.Request())
		return nil
	})
	apiV1 := s.echo.Group("/api/v1")
	apiMigrateLimiter := middleware.NewRateLimiterMemoryStoreWithConfig(middleware.RateLimiterMemoryStoreConfig{Rate: redeployRate, Burst: redeployRateBurst, ExpiresIn: 10 * time.Minute})
	apiMigrateRate := middleware.RateLimiterWithConfig(middleware.RateLimiterConfig{Store: apiMigrateLimiter})
	apiNotFound := func(c echo.Context) error { api.NotFound(c.Response(), c.Request()); return nil }
	apiV1.GET("", apiNotFound)
	apiV1.RouteNotFound("/*", apiNotFound)
	apiV1.GET("/health", func(c echo.Context) error { api.Health(c.Response(), c.Request()); return nil })
	apiV1.HEAD("/health", func(c echo.Context) error { api.Health(c.Response(), c.Request()); return nil })
	apiV1.GET("/auth/session", apiAuthHandler.Session)
	apiV1.GET("/auth/csrf", apiAuthHandler.CSRF)
	loginLimiter := middleware.NewRateLimiterMemoryStoreWithConfig(middleware.RateLimiterMemoryStoreConfig{Rate: loginRateLimit, Burst: loginRateBurst, ExpiresIn: 10 * time.Minute})
	apiV1.POST("/auth/login", apiAuthHandler.Login, middleware.RateLimiterWithConfig(middleware.RateLimiterConfig{Store: loginLimiter}))
	apiV1.POST("/auth/logout", apiAuthHandler.Logout, apiAuthMiddleware.RequireAuth)
	apiV1.GET("/discord/resources", apiDiscordResourceHandler.Resources, apiAuthMiddleware.RequireAuth)
	apiV1.GET("/dashboard", apiDashboardHandler.Dashboard, apiAuthMiddleware.RequireAuth)
	apiV1.GET("/players", apiPlayersHandler.List, apiAuthMiddleware.RequireAuth)
	apiV1.GET("/players/:id", apiPlayersAdminHandler.Detail, apiAuthMiddleware.RequireAuth)
	apiV1.POST("/players", apiPlayersAdminHandler.Create, apiAuthMiddleware.RequireAuth)
	apiV1.PUT("/players/:id", apiPlayersAdminHandler.Update, apiAuthMiddleware.RequireAuth)
	apiV1.PATCH("/players/:id", apiPlayersAdminHandler.Update, apiAuthMiddleware.RequireAuth)
	apiV1.PUT("/players/:id/stats", apiPlayersAdminHandler.UpdateStats, apiAuthMiddleware.RequireAuth)
	apiV1.PUT("/players/:id/state", apiPlayersAdminHandler.UpdateState, apiAuthMiddleware.RequireAuth)
	apiPlayers := apiV1.Group("/players/:id", apiAuthMiddleware.RequireAuth)
	apiPlayers.POST("/items/add", apiPlayersAdminHandler.ItemAdd)
	apiPlayers.POST("/items/remove", apiPlayersAdminHandler.ItemRemove)
	apiPlayers.POST("/items/buy", apiPlayersAdminHandler.ItemBuy)
	apiPlayers.POST("/abilities/add", apiPlayersAdminHandler.AbilityAdd)
	apiPlayers.POST("/abilities/remove", apiPlayersAdminHandler.AbilityRemove)
	apiPlayers.POST("/statuses/add", apiPlayersAdminHandler.StatusAdd)
	apiPlayers.POST("/statuses/remove", apiPlayersAdminHandler.StatusRemove)
	apiPlayers.POST("/immunities/add", apiPlayersAdminHandler.ImmunityAdd)
	apiPlayers.POST("/immunities/remove", apiPlayersAdminHandler.ImmunityRemove)
	apiPlayers.POST("/notes/add", apiPlayersAdminHandler.NoteAdd)
	apiPlayers.POST("/notes/remove", apiPlayersAdminHandler.NoteRemove)

	apiCatalog := apiV1.Group("/catalog", apiAuthMiddleware.RequireAuth)
	apiCatalog.GET("/roles", apiCatalogHandler.ListRoles)
	apiCatalog.GET("/roles/search", apiCatalogHandler.ListRoles)
	apiCatalog.GET("/roles/:id", apiCatalogHandler.GetRole)
	apiCatalog.POST("/roles", apiCatalogHandler.CreateRole)
	apiCatalog.PUT("/roles/:id", apiCatalogHandler.UpdateRole)
	apiCatalog.DELETE("/roles/:id", apiCatalogHandler.DeleteRole)
	apiCatalog.GET("/items", apiCatalogHandler.ListItems)
	apiCatalog.GET("/items/search", apiCatalogHandler.ListItems)
	apiCatalog.GET("/items/:id", apiCatalogHandler.GetItem)
	apiCatalog.POST("/items", apiCatalogHandler.CreateItem)
	apiCatalog.PUT("/items/:id", apiCatalogHandler.UpdateItem)
	apiCatalog.DELETE("/items/:id", apiCatalogHandler.DeleteItem)
	apiCatalog.GET("/abilities", apiCatalogHandler.ListAbilities)
	apiCatalog.GET("/abilities/search", apiCatalogHandler.ListAbilities)
	apiCatalog.GET("/abilities/:id", apiCatalogHandler.GetAbility)
	apiCatalog.POST("/abilities", apiCatalogHandler.CreateAbility)
	apiCatalog.PUT("/abilities/:id", apiCatalogHandler.UpdateAbility)
	apiCatalog.DELETE("/abilities/:id", apiCatalogHandler.DeleteAbility)
	apiCatalog.GET("/statuses", apiCatalogHandler.ListStatuses)
	apiCatalog.GET("/statuses/search", apiCatalogHandler.ListStatuses)
	apiCatalog.GET("/statuses/:id", apiCatalogHandler.GetStatus)
	apiCatalog.POST("/statuses", apiCatalogHandler.CreateStatus)
	apiCatalog.PUT("/statuses/:id", apiCatalogHandler.UpdateStatus)
	apiCatalog.DELETE("/statuses/:id", apiCatalogHandler.DeleteStatus)

	s.echo.GET("/api/v1/ops/cycle", apiCycleHandler.Get, apiAuthMiddleware.RequireAuth)
	s.echo.POST("/api/v1/ops/cycle/advance", apiCycleHandler.Advance, apiAuthMiddleware.RequireAuth)
	s.echo.POST("/api/v1/ops/cycle/set", apiCycleHandler.Set, apiAuthMiddleware.RequireAuth)
	s.echo.GET("/api/v1/ops/setup", apiSetupHandler.Get, apiAuthMiddleware.RequireAuth)
	s.echo.POST("/api/v1/ops/setup", apiSetupHandler.Generate, apiAuthMiddleware.RequireAuth)
	s.echo.GET("/api/v1/ops/channels", apiChannelsHandler.Get, apiAuthMiddleware.RequireAuth)
	s.echo.POST("/api/v1/ops/channels", apiChannelsHandler.Mutate, apiAuthMiddleware.RequireAuth)
	s.echo.POST("/api/v1/ops/channels/update", apiChannelsHandler.Mutate, apiAuthMiddleware.RequireAuth)
	s.echo.DELETE("/api/v1/ops/channels/:kind/:id", apiChannelsHandler.Delete, apiAuthMiddleware.RequireAuth)
	s.echo.GET("/api/v1/ops/votes", apiVotesHandler.Get, apiAuthMiddleware.RequireAuth)
	s.echo.GET("/api/v1/ops/healthcheck", apiReadinessHandler.Get, apiAuthMiddleware.RequireAuth)

	apiAdmin := apiV1.Group("/admin", apiAuthMiddleware.RequireAuth)
	apiAdmin.GET("/audit", apiAdminHandler.Audit)
	apiAdmin.GET("/migrations", apiAdminHandler.Migrations)
	apiAdmin.POST("/migrations/up", apiAdminHandler.MigrationUp, apiMigrateRate)
	apiAdmin.POST("/migrations/down", apiAdminHandler.MigrationDown, apiMigrateRate)
	apiAdmin.GET("/reset", apiAdminHandler.ResetPreview)
	apiAdmin.POST("/reset", apiAdminHandler.ResetExecute, apiMigrateRate)
	apiAdmin.POST("/redeploy", apiAdminHandler.Redeploy, apiMigrateRate)
	apiSync := apiV1.Group("/sync", apiAuthMiddleware.RequireAuth)
	apiSync.GET("/sources", apiSyncHandler.Sources)
	apiSync.POST("/preview", apiSyncHandler.Preview, apiMigrateRate)
	apiSync.POST("/apply", apiSyncHandler.Apply, apiMigrateRate)
	apiSync.GET("/runs/:id", apiSyncHandler.Run)
	apiSync.PUT("/sources/:id", apiSyncHandler.UpdateSource, apiMigrateRate)

	serveUI := func(c echo.Context) error { ui.Handler(c.Response(), c.Request()); return nil }
	s.echo.GET("/login", serveUI)
	s.echo.HEAD("/login", serveUI)
	// SvelteKit's hashed JavaScript and CSS assets must remain publicly
	// readable so the unauthenticated login shell can bootstrap in the browser.
	// The page routes themselves stay behind the browser session guard below.
	s.echo.GET("/_app/*", serveUI)
	s.echo.HEAD("/_app/*", serveUI)
	s.echo.GET("/", serveUI, browserAuth.RequireAuth)
	s.echo.HEAD("/", serveUI, browserAuth.RequireAuth)
	s.echo.GET("/*", serveUI, browserAuth.RequireAuth)
	s.echo.HEAD("/*", serveUI, browserAuth.RequireAuth)
}

// Handler exposes the underlying Echo router as a plain http.Handler.
// Used by httptest-based handler tests; the running server uses Start().
func (s *Server) Handler() http.Handler {
	return s.echo
}

// Start begins listening on the configured port (blocking)
func (s *Server) Start() error {
	addr := fmt.Sprintf(":%s", s.config.Port)
	s.logger.Info().Str("addr", addr).Msg("Starting web server")
	return s.echo.Start(addr)
}

// Shutdown gracefully stops the server
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info().Msg("Shutting down web server")
	if s.syncHandler != nil {
		s.syncHandler.Shutdown()
	}
	return s.echo.Shutdown(ctx)
}
