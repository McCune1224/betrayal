// Package web provides the admin web server for the Betrayal Bot
package web

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bwmarrin/discordgo"
	"github.com/gorilla/sessions"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/mccune1224/betrayal/internal/web/handlers"
	webmiddleware "github.com/mccune1224/betrayal/internal/web/middleware"
	"github.com/mccune1224/betrayal/internal/web/railway"
	"github.com/rs/zerolog"
)

// Config holds the web server configuration
type Config struct {
	Port          string
	AdminPassword string
	SessionSecret string // For cookie encryption

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
}

// New creates a new web server
func New(pool *pgxpool.Pool, discord *discordgo.Session, logger zerolog.Logger, cfg Config) *Server {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	// Use session secret or default to admin password (not ideal but works for simple case)
	sessionSecret := cfg.SessionSecret
	if sessionSecret == "" {
		sessionSecret = cfg.AdminPassword
	}

	store := sessions.NewCookieStore([]byte(sessionSecret))
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

	s := &Server{
		echo:           e,
		dbPool:         pool,
		discordSession: discord,
		logger:         logger.With().Str("component", "web").Logger(),
		config:         cfg,
		sessionStore:   store,
		railwayClient:  railwayClient,
	}

	s.setupMiddleware()
	s.setupRoutes()

	return s
}

func (s *Server) setupMiddleware() {
	// Recovery middleware
	s.echo.Use(middleware.Recover())

	// Request logging
	s.echo.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:    true,
		LogStatus: true,
		LogMethod: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			s.logger.Info().
				Str("method", v.Method).
				Str("uri", v.URI).
				Int("status", v.Status).
				Msg("request")
			return nil
		},
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

	// Auth middleware
	authMiddleware := webmiddleware.NewAuthMiddleware(s.sessionStore)

	// Public routes
	s.echo.GET("/health", healthHandler.Health)
	s.echo.GET("/login", authHandler.LoginPage)
	s.echo.POST("/login", authHandler.Login)

	// Protected routes
	protected := s.echo.Group("", authMiddleware.RequireAuth)
	protected.POST("/logout", authHandler.Logout)
	protected.GET("/", dashboardHandler.Dashboard)
	protected.GET("/health/status", healthHandler.HealthStatusPartial)
	protected.GET("/players", playersHandler.List)
	protected.GET("/players/table", playersHandler.Table)
	protected.GET("/players/:id", playersHandler.Detail)
	protected.GET("/votes", votesHandler.Votes)
	protected.GET("/votes/tally", votesHandler.VoteTally)
	protected.POST("/admin/redeploy", adminHandler.Redeploy)
	protected.GET("/admin/audit", adminHandler.AuditLogs)

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
