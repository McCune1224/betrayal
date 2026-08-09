// Package web provides the admin web server for the Betrayal Bot
package web

import (
	"context"
	"crypto/sha256"
	"fmt"
	"net/http"
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
	"github.com/mccune1224/betrayal/internal/web/handlers"
	webmiddleware "github.com/mccune1224/betrayal/internal/web/middleware"
	"github.com/mccune1224/betrayal/internal/web/railway"
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
	// production guard: destructive actions (sync apply, migrations) are
	// hard-blocked when it points at the prod pooler.
	DatabaseURL string
	// Environment (ENVIRONMENT env var) is also treated as production for the
	// guard, so a renamed pooler host cannot disable it.
	Environment string
	// AllowProdMutations (WEB_ALLOW_PROD_MUTATIONS=true) lifts that block.
	AllowProdMutations bool

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

	// CSRF protection — double-submit cookie pattern.
	// - HTMX requests carry the token in the X-CSRF-Token header (the base
	//   layout's htmx:configRequest handler reads the `_csrf` cookie and adds it).
	// - Regular HTML forms include a hidden `_csrf` input, injected by the base
	//   layout script. Token validation only applies to state-changing methods;
	//   safe methods (GET/HEAD/OPTIONS/TRACE) are skipped by Echo.
	// The cookie is intentionally NOT HttpOnly so client JS can read it — this is
	// the standard HTMX-compatible double-submit setup.
	s.echo.Use(middleware.CSRFWithConfig(middleware.CSRFConfig{
		TokenLookup:    "header:" + echo.HeaderXCSRFToken + ",form:_csrf",
		CookieName:     "_csrf",
		CookieHTTPOnly: false, // readable by JS so htmx can attach it to every request
		CookieSameSite: http.SameSiteLaxMode,
		CookiePath:     "/",
	}))

	// Static files
	s.echo.Static("/static", "web/static")
}

func (s *Server) setupRoutes() {
	// Create handlers
	healthHandler := handlers.NewHealthHandler(s.dbPool, s.discordSession)
	authHandler := handlers.NewAuthHandler(s.sessionStore, s.config.AdminPassword)
	dashboardHandler := handlers.NewDashboardHandler(s.dbPool)
	playersHandler := handlers.NewPlayersHandler(s.dbPool)
	adminHandler := handlers.NewAdminHandler(s.dbPool, s.railwayClient)
	votesHandler := handlers.NewVotesHandler(s.dbPool)
	rolesHandler := handlers.NewRolesHandler(s.dbPool)
	cycleHandler := handlers.NewCycleHandler(s.dbPool)
	channelsHandler := handlers.NewChannelsHandler(s.dbPool, s.discordSession)
	playerEditHandler := handlers.NewPlayerEditHandler(s.dbPool)
	catalogHandler := handlers.NewCatalogHandler(s.dbPool)
	isProd := IsProd(s.config.DatabaseURL, s.config.Environment)
	syncHandler := handlers.NewSyncHandler(s.dbPool, s.syncService, isProd, s.config.AllowProdMutations, s.logger)
	migrationsHandler := handlers.NewMigrationsHandler(s.getMigrateRunner, isProd, s.config.AllowProdMutations)
	setupHandler := handlers.NewSetupHandler(s.dbPool, s.discordSession)
	readinessHandler := handlers.NewGameReadinessHandler(s.dbPool, s.discordSession)
	playerCreateHandler := handlers.NewPlayerCreateHandler(s.dbPool)
	resetHandler := handlers.NewResetHandler(gamereset.New(s.dbPool, s.syncService), isProd)

	// Auth middleware
	authMiddleware := webmiddleware.NewAuthMiddleware(s.sessionStore)

	// Public routes
	s.echo.GET("/health", healthHandler.Health)
	s.echo.GET("/login", authHandler.LoginPage)

	// Login is the brute-force surface: rate limit by IP (a burst of attempts,
	// then ~1/sec). Note: behind a proxy this buckets by proxy IP unless a
	// trusted-proxy extractor is configured — still throttles global brute force.
	loginLimiter := middleware.NewRateLimiterMemoryStoreWithConfig(middleware.RateLimiterMemoryStoreConfig{
		Rate:      loginRateLimit,
		Burst:     loginRateBurst,
		ExpiresIn: 10 * time.Minute,
	})
	s.echo.POST("/login", authHandler.Login, middleware.RateLimiterWithConfig(middleware.RateLimiterConfig{
		Store: loginLimiter,
	}))

	// Protected routes
	protected := s.echo.Group("", authMiddleware.RequireAuth)
	protected.POST("/logout", authHandler.Logout)
	protected.GET("/", dashboardHandler.Dashboard)
	protected.GET("/health/status", healthHandler.HealthStatusPartial)
	protected.GET("/healthcheck", readinessHandler.Page)
	protected.GET("/players", playersHandler.List)
	protected.GET("/players/new", playerCreateHandler.Page)
	protected.POST("/players", playerCreateHandler.Create)
	protected.GET("/players/table", playersHandler.Table)
	protected.GET("/players/:id", playersHandler.Detail)
	protected.GET("/setup", setupHandler.Page)
	protected.POST("/setup/generate", setupHandler.Generate)
	protected.GET("/players/:id/edit", playerEditHandler.Edit)
	protected.POST("/players/:id/edit", playerEditHandler.UpdateStats)
	protected.POST("/players/:id/state", playerEditHandler.UpdateState)
	protected.POST("/players/:id/items/add", playerEditHandler.AddItem)
	protected.POST("/players/:id/items/buy", playerEditHandler.BuyItem)
	protected.POST("/players/:id/items/remove", playerEditHandler.RemoveItem)
	protected.POST("/players/:id/abilities/add", playerEditHandler.AddAbility)
	protected.POST("/players/:id/abilities/remove", playerEditHandler.RemoveAbility)
	protected.POST("/players/:id/statuses/add", playerEditHandler.AddStatus)
	protected.POST("/players/:id/statuses/remove", playerEditHandler.RemoveStatus)
	protected.POST("/players/:id/perks/add", playerEditHandler.AddPerk)
	protected.POST("/players/:id/perks/remove", playerEditHandler.RemovePerk)
	protected.POST("/players/:id/immunities/add", playerEditHandler.AddImmunity)
	protected.POST("/players/:id/immunities/remove", playerEditHandler.RemoveImmunity)
	protected.POST("/players/:id/notes/add", playerEditHandler.AddNote)
	protected.POST("/players/:id/notes/remove", playerEditHandler.RemoveNote)
	protected.GET("/votes", votesHandler.Votes)
	protected.GET("/votes/tally", votesHandler.VoteTally)
	protected.GET("/admin/audit", adminHandler.AuditLogs)

	// Migrations: destructive + schema-changing → the redeploy-style limiter.
	migrateLimiter := middleware.NewRateLimiterMemoryStoreWithConfig(middleware.RateLimiterMemoryStoreConfig{
		Rate:      redeployRate,
		Burst:     redeployRateBurst,
		ExpiresIn: 10 * time.Minute,
	})
	migrateRate := middleware.RateLimiterWithConfig(middleware.RateLimiterConfig{Store: migrateLimiter})
	protected.GET("/admin/migrations", migrationsHandler.Page)
	protected.POST("/admin/migrations/up", migrationsHandler.Up, migrateRate)
	protected.POST("/admin/migrations/down", migrationsHandler.Down, migrateRate)
	protected.GET("/admin/reset", resetHandler.Page)
	protected.POST("/admin/reset", resetHandler.Execute, migrateRate)

	// Redeploy is a state-changing, cost-incurring action: rate limit it.
	redeployLimiter := middleware.NewRateLimiterMemoryStoreWithConfig(middleware.RateLimiterMemoryStoreConfig{
		Rate:      redeployRate,
		Burst:     redeployRateBurst,
		ExpiresIn: 10 * time.Minute,
	})
	protected.POST("/admin/redeploy", adminHandler.Redeploy, middleware.RateLimiterWithConfig(middleware.RateLimiterConfig{
		Store: redeployLimiter,
	}))

	// Sync routes. Preview fetches remote sheets (slow, network-bound) and
	// apply writes to the database — both get a shared burst limiter.
	syncLimiter := middleware.NewRateLimiterMemoryStoreWithConfig(middleware.RateLimiterMemoryStoreConfig{
		Rate:      0.5, // 1 request / 2s sustained
		Burst:     4,
		ExpiresIn: 10 * time.Minute,
	})
	limiter := middleware.RateLimiterWithConfig(middleware.RateLimiterConfig{Store: syncLimiter})
	protected.GET("/sync", syncHandler.Page)
	protected.POST("/sync/preview", syncHandler.Preview, limiter)
	protected.POST("/sync/apply", syncHandler.Apply, limiter)
	protected.POST("/sync/sources/:id", syncHandler.UpdateSource)

	// Role routes
	protected.GET("/roles", rolesHandler.List)
	protected.GET("/roles/search", rolesHandler.Search)
	protected.GET("/roles/:id", rolesHandler.Detail)
	protected.PUT("/roles/:id", rolesHandler.Update)
	protected.GET("/roles/:id/abilities", rolesHandler.ListAbilities)
	protected.PUT("/roles/:id/abilities/:abilityId", rolesHandler.UpdateAbility)
	protected.DELETE("/roles/:id/abilities/:abilityId", rolesHandler.RemoveAbility)
	protected.GET("/roles/:id/perks", rolesHandler.ListPerks)
	protected.PUT("/roles/:id/perks/:perkId", rolesHandler.UpdatePerk)
	protected.DELETE("/roles/:id/perks/:perkId", rolesHandler.RemovePerk)

	// Cycle routes
	protected.GET("/cycle", cycleHandler.Page)
	protected.POST("/cycle/advance", cycleHandler.Advance)
	protected.POST("/cycle/set", cycleHandler.Set)

	// Channel config routes
	protected.GET("/channels", channelsHandler.Page)
	protected.POST("/channels/update", channelsHandler.Update)
	protected.POST("/channels/admin/delete", channelsHandler.DeleteAdmin)

	// Catalog (items / abilities / statuses) routes
	protected.GET("/items", catalogHandler.Items)
	protected.GET("/items/search", catalogHandler.SearchItems)
	protected.POST("/items", catalogHandler.CreateItem)
	protected.GET("/items/:id", catalogHandler.ItemDetail)
	protected.POST("/items/:id", catalogHandler.UpdateItem)
	protected.POST("/items/:id/delete", catalogHandler.DeleteItem)
	protected.GET("/abilities", catalogHandler.Abilities)
	protected.GET("/abilities/search", catalogHandler.SearchAbilities)
	protected.POST("/abilities", catalogHandler.CreateAbility)
	protected.GET("/abilities/:id", catalogHandler.AbilityDetail)
	protected.POST("/abilities/:id", catalogHandler.UpdateAbility)
	protected.POST("/abilities/:id/delete", catalogHandler.DeleteAbility)
	protected.GET("/statuses", catalogHandler.Statuses)
	protected.GET("/statuses/search", catalogHandler.SearchStatuses)
	protected.POST("/statuses", catalogHandler.CreateStatus)
	protected.GET("/statuses/:id", catalogHandler.StatusDetail)
	protected.POST("/statuses/:id", catalogHandler.UpdateStatus)
	protected.POST("/statuses/:id/delete", catalogHandler.DeleteStatus)
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
	return s.echo.Shutdown(ctx)
}
