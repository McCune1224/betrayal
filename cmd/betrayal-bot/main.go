package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/mccune1224/betrayal/internal/data"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/bwmarrin/discordgo"
	"github.com/spf13/viper"
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
	discordSession *discordgo.Session
}

func main() {
	viper.SetConfigFile(".env")
	viper.ReadInConfig()
	var cfg config
	cfg.discord.botToken = viper.GetString("DISCORD_BOT_TOKEN")
	cfg.discord.clientID = viper.GetString("DISCORD_CLIENT_ID")
	cfg.discord.clientID = viper.GetString("DISCORD_CLIENT_SECRET")

	cfg.database.dsn = viper.GetString("DATABASE_URL")

	discordSession, err := discordgo.New("Bot " + cfg.discord.botToken)
	if err != nil {
		log.Fatal("error creating Discord session,", err)
	}
	discordSession.Identify.Intents = discordgo.PermissionAdministrator

	err = discordSession.Open()
	if err != nil {
		log.Fatal("error opening connection,", err)
	}

	db, err := gorm.Open(postgres.Open(cfg.database.dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("error opening database,", err)
	}

	dbModels := data.NewModels(db, true)
	app := &app{
		conifg:         cfg,
		models:         dbModels,
		discordSession: discordSession,
	}

	betrayalCM := app.NewSlashCommandManager()
	betrayalCM.MapCommand(app.PingCommand())
	betrayalCM.MapCommand(app.GetRoleCommand())
	betrayalCM.MapCommand(app.WhoAmICommand())
	betrayalCM.MapCommand(app.EchoCommand())
	betrayalCM.MapCommand(app.InsultCommand())
	betrayalCM.MapCommand(app.RandomInsultCommand())
	registeredCommandsTally := betrayalCM.RegisterCommands(app.discordSession)

	defer app.discordSession.Close()

	log.Printf(
		"%s is now running. %d commands available. Press CTRL-C to exit.\n",
		discordSession.State.User.Username,
		registeredCommandsTally,
	)
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	if err := app.discordSession.Close(); err != nil {
		log.Fatal("error closing connection,", err)
	}

	// Commands Cleanup
	for id, name := range betrayalCM.CommandIDs {
		err := app.discordSession.ApplicationCommandDelete(
			app.discordSession.State.User.ID,
			"",
			id,
		)
		if err != nil {
			log.Printf("error deleting command %s: %s on Cleanup", name, err)
		} else {
			log.Println("deleted command", name)
		}
	}

}
