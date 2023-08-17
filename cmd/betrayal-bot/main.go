package main

import (
	"fmt"
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

	// Open a websocket connection to Discord and begin listening.
	err = betrayalBot.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	betrayalCM := commands.NewSlashCommandManager()
	betrayalCM.AddCommand(commands.Ping)
	betrayalCM.AddCommand(commands.Whoami)

	// betrayalBot.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// 	log.Print(i.ApplicationCommandData().Name)
	// 	if cmd, ok := betrayalCM.MappedCommands[i.ApplicationCommandData().Name]; ok {
	// 		log.Printf("Command %s found, attempting to add", i.ApplicationCommandData().Name)
	// 		cmd.Handler(s, i)
	// 	} else {
	// 		log.Printf("Command %s not found", i.ApplicationCommandData().Name)
	// 	}
	// })
	registeredCommandsTally := betrayalCM.RegisterCommands(betrayalBot)

	fmt.Printf("%s is now running. %d commands available. Press CTRL-C to exit.\n", betrayalBot.State.User.Username, registeredCommandsTally)
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	// Cleanly close down the Discord session.
	betrayalBot.Close()
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the authenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}
	// If the message is "ping" reply with "Pong!"
	if m.Content == "ping" {
		s.ChannelMessageSend(m.ChannelID, "Pong!")
	}

	// If the message is "pong" reply with "Ping!"
	if m.Content == "pong" {
		s.ChannelMessageSend(m.ChannelID, "Ping!")
	}
}

func messagePapa(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Content == "hello my son" {
		s.ChannelMessageSend(m.ChannelID, "Hello Papa!")
	}
}
