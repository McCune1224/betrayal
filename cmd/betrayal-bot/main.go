package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	_ "github.com/joho/godotenv/autoload"
	_ "github.com/lib/pq"
	"github.com/mccune1224/betrayal/internal/commands"
	"github.com/mccune1224/betrayal/internal/commands/inventory"
	roll "github.com/mccune1224/betrayal/internal/commands/luck"
	"github.com/mccune1224/betrayal/internal/data"
	"github.com/mccune1224/betrayal/internal/middlewares"
	"github.com/zekrotja/ken"
)

// config struct to hold env variables and any other config settings
type config struct {
	discord struct {
		clientID     string
		clientSecret string
		botToken     string
	}
	database struct {
		dsn string
	}
}

// Global app struct
type app struct {
	models          data.Models
	logger          *log.Logger
	betrayalManager *ken.Ken
	conifg          config
}

// Wrapper for Ken.Command that needs DB access
// (AKA basically every command)
type BetrayalCommand interface {
	ken.Command
	SetModels(data.Models)
}

// Wrapper for ken.RegisterBetrayalCommands for inserting DB access
func (a *app) RegisterBetrayalCommands(commands ...BetrayalCommand) int {
	tally := 0
	for _, command := range commands {
		command.SetModels(a.models)
		err := a.betrayalManager.RegisterCommands(command)
		if err != nil {
			a.logger.Fatal(err)
		}
		tally += 1

	}
	return tally
}

func main() {
	var cfg config
	cfg.discord.botToken = os.Getenv("DISCORD_BOT_TOKEN")
	cfg.discord.clientID = os.Getenv("DISCORD_CLIENT_ID")
	cfg.discord.clientSecret = os.Getenv("DISCORD_CLIENT_SECRET")
	cfg.database.dsn = os.Getenv("DATABASE_URL")

	bot, err := discordgo.New("Bot " + cfg.discord.botToken)
	if err != nil {
		log.Fatal("error creating Discord session,", err)
	}

	bot.Identify.Intents = discordgo.PermissionAdministrator
	if err != nil {
		log.Fatal("error opening connection,", err)
	}

	logger := log.New(os.Stdout, "betrayal-bot ", log.Ldate|log.Ltime|log.Lshortfile)

	db, err := sqlx.Connect("postgres", cfg.database.dsn)
	if err != nil {
		log.Fatal("error opening database,", err)
	}

	dbModels := data.NewModels(db)

	app := &app{
		conifg: cfg,
		models: dbModels,
		logger: logger,
	}
	km, err := ken.New(bot)
	app.betrayalManager = km
	if err != nil {
		app.logger.Fatal(err)
	}

	// Call unregister twice to remove any lingering commands from previous runs
	app.betrayalManager.Unregister()

	tally := app.RegisterBetrayalCommands(
		new(roll.Roll),
		new(inventory.Inventory),
		// new(commands.ActionFunnel),
		// new(view.View),
		new(commands.Buy),
	// new(commands.List),
	// new(commands.Insult),
	// new(commands.Ping),
	)
	err = app.betrayalManager.RegisterMiddlewares(new(middlewares.PermissionsMiddleware))
	if err != nil {
		app.logger.Fatal(err)
	}

	defer app.betrayalManager.Unregister()

	err = bot.Open()
	if err != nil {
		app.logger.Fatal("error opening connection,", err)
	}
	defer bot.Close()

	app.logger.Printf(
		"%s is now running with %d commands. Press CTRL-C to exit.\n",
		bot.State.User.Username,
		tally,
	)

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	if err := app.betrayalManager.Session().Close(); err != nil {
		app.logger.Fatal("error closing connection,", err)
	}
}
