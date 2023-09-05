package main

import (
	"log"
	"os"
	"os/signal"
	"syscall" // New import

	// New import

	"github.com/bwmarrin/discordgo"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	_ "github.com/joho/godotenv/autoload"
	_ "github.com/lib/pq"
	"github.com/mccune1224/betrayal/internal/data"
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
	conifg         config
	models         data.Models
	logger         *log.Logger
	discordSession *discordgo.Session
	commandHandler *SlashCommandManager
}

func main() {
	var cfg config
	cfg.discord.botToken = os.Getenv("DISCORD_BOT_TOKEN")
	cfg.discord.clientID = os.Getenv("DISCORD_CLIENT_ID")
	cfg.discord.clientSecret = os.Getenv("DISCORD_CLIENT_SECRET")
	cfg.database.dsn = os.Getenv("DATABASE_URL")

	discordSession, err := discordgo.New("Bot " + cfg.discord.botToken)
	if err != nil {
		log.Fatal("error creating Discord session,", err)
	}
	discordSession.Identify.Intents = discordgo.PermissionAdministrator

	err = discordSession.Open()
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
		conifg:         cfg,
		models:         dbModels,
		logger:         logger,
		discordSession: discordSession,
	}

	betrayalCM := app.NewSlashCommandManager()

	errs := betrayalCM.RemoveCached(app.discordSession)

	for _, err := range errs {
		app.logger.Println("error deleting command", err)
	}

	betrayalCM.MapCommand(app.GetRoleCommand())
	betrayalCM.MapCommand(app.PingCommand())
	betrayalCM.MapCommand(app.EchoCommand())
	betrayalCM.MapCommand(app.HelpCommand())
	betrayalCM.MapCommand(app.ChannelDetailsCommand())
	betrayalCM.MapCommand(app.UserDetailsCommand())
	betrayalCM.MapCommand(app.FunnelCommand())

	for _, insultCommand := range app.InsultCommandBundle() {
		betrayalCM.MapCommand(insultCommand)
	}

	registeredCommandsTally := betrayalCM.RegisterCommands(app.discordSession)

	app.commandHandler = betrayalCM

	app.logger.Printf(
		"%s is now running. %d commands available. Press CTRL-C to exit.\n",
		discordSession.State.User.Username,
		registeredCommandsTally,
	)

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	if err := app.discordSession.Close(); err != nil {
		app.logger.Fatal("error closing connection,", err)
	}
	errs = betrayalCM.RemoveCached(app.discordSession)
	for _, err := range errs {
		app.logger.Println("error deleting command", err)
	}

}
