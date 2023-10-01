package main

import (
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/joho/godotenv/autoload"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/lib/pq"
	"github.com/mccune1224/betrayal/internal/data"
)

type config struct {
	discord struct {
		clientID     string
		clientSecret string
	}
	database struct {
		dsn string
	}
}

type app struct {
	config config
	models data.Models
}

func getPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	return ":" + port
}

func main() {

	var cfg config
	var webapp app
	cfg.discord.clientID = os.Getenv("DISCORD_CLIENT_ID")
	cfg.discord.clientSecret = os.Getenv("DISCORD_CLIENT_SECRET")
	cfg.database.dsn = os.Getenv("DATABASE_URL")
	db, err := sqlx.Connect("postgres", cfg.database.dsn)
	if err != nil {
		panic(err)
	}
	models := data.NewModels(db)

	webapp = app{
		config: cfg,
		models: models,
	}
	e := echo.New()
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "method=${method}, uri=${uri}, status=${status}, latency=${latency_human}, error=${error}\n",
	}))
	e.Use(middleware.StaticWithConfig(middleware.StaticConfig{
		Skipper: nil,
		// Root directory from where the static content is served.
		Root: "./www/dist",
		// Index file for serving a directory.
		// Optional. Default value "index.html".
		Index: "index.html",
		// Enable HTML5 mode by forwarding all not-found requests to root so that
		// SPA (single-page application) can handle the routing.
		HTML5:      true,
		Browse:     false,
		IgnoreBase: false,
		Filesystem: nil,
	}))
	e.Use(middleware.Recover())

	webapp.AttachRoutes(e)
	e.Logger.Fatal(e.Start(getPort()))

}
