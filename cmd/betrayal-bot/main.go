package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/mccune1224/betrayal/config"
	"github.com/mccune1224/betrayal/internal/discord/commands"

	"github.com/bwmarrin/discordgo"
	"github.com/spf13/viper"
)

var botEnv struct {
	DISCORD_CLIENT_ID     string
	DISCORD_CLIENT_SECRET string
	DISCORD_BOT_TOKEN     string
}

// Initialize config and bot
func init() {
	config.LoadBetrayalConfig()
	botEnv.DISCORD_BOT_TOKEN = viper.GetString("DISCORD_BOT_TOKEN")
	botEnv.DISCORD_CLIENT_ID = viper.GetString("DISCORD_CLIENT_ID")
	botEnv.DISCORD_CLIENT_SECRET = viper.GetString("DISCORD_CLIENT_SECRET")

	for _, env := range []string{"DISCORD_BOT_TOKEN", "DISCORD_CLIENT_ID", "DISCORD_CLIENT_SECRET"} {
		if botEnv.DISCORD_BOT_TOKEN == "" {
			log.Fatalf("Environment variable %s not set", env)
		}
	}
}

func main() {
	betrayalBot, err := discordgo.New("Bot " + botEnv.DISCORD_BOT_TOKEN)
	if err != nil {
		log.Fatal("error creating Discord session,", err)
	}

	betrayalBot.Identify.Intents = discordgo.PermissionAdministrator

	err = betrayalBot.Open()
	if err != nil {
		log.Println("error opening connection,", err)
		return
	}
	defer betrayalBot.Close()

	betrayalCM := commands.NewSlashCommandManager()
	betrayalCM.MapCommand(commands.Ping)
	betrayalCM.MapCommand(commands.Whoami)

	registeredCommandsTally := betrayalCM.RegisterCommands(betrayalBot)

	log.Printf("%s is now running. %d commands available. Press CTRL-C to exit.\n", betrayalBot.State.User.Username, registeredCommandsTally)
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

}
