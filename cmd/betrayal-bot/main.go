package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/go-co-op/gocron"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	_ "github.com/joho/godotenv/autoload"
	_ "github.com/lib/pq"
	"github.com/mccune1224/betrayal/internal/commands"
	"github.com/mccune1224/betrayal/internal/commands/inventory"
	roll "github.com/mccune1224/betrayal/internal/commands/luck"
	"github.com/mccune1224/betrayal/internal/commands/view"
	"github.com/mccune1224/betrayal/internal/data"
	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/mccune1224/betrayal/internal/util"
	"github.com/zekrotja/ken"
	"github.com/zekrotja/ken/state"
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
	scheduler       *gocron.Scheduler
	betrayalManager *ken.Ken
	conifg          config
}

// Wrapper for Ken.Command that needs DB access
// (AKA basically every command)
type BetrayalCommand interface {
	ken.Command
	Initialize(data.Models, *gocron.Scheduler)
}

// Wrapper for ken.RegisterBetrayalCommands for inserting DB access
func (a *app) RegisterBetrayalCommands(commands ...BetrayalCommand) int {
	tally := 0
	for _, command := range commands {
		command.Initialize(a.models, a.scheduler)
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

	// Spin up Bot and give it admin permissions
	bot, err := discordgo.New("Bot " + cfg.discord.botToken)
	if err != nil {
		log.Fatal("error creating Discord session,", err)
	}
	bot.Identify.Intents = discordgo.PermissionAdministrator
	if err != nil {
		log.Fatal("error opening connection,", err)
	}

	// Connect to DB
	db, err := sqlx.Connect("postgres", cfg.database.dsn)
	if err != nil {
		log.Fatal("error opening database,", err)
	}
	dbModels := data.NewModels(db)

	scheduler := gocron.NewScheduler(time.UTC)

	// Create central app struct and attach ken framework to it
	app := &app{
		conifg:    cfg,
		models:    dbModels,
		scheduler: scheduler,
	}
	km, err := ken.New(bot, ken.Options{
		State: state.NewInternal(),
		EmbedColors: ken.EmbedColors{
			Default: discord.ColorThemeOrange,
			Error:   discord.ColorThemeRuby,
		},
		DisableCommandInfoCache: false,
		OnSystemError: func(ctx string, err error, args ...interface{}) {
			log.Printf("[STM] {%s} - %s\n", ctx, err.Error())
		},
		OnCommandError: func(err error, ctx *ken.Ctx) {
			// get the command name, options and args
			// TODO: Make this show full argument details like logHandler?
			log.Printf("[CMD] %s - %s : %s\n", ctx.Command.Name(), ctx.GetEvent().Member.User.Username, err.Error())
		},
		// Not really doing events but keeping this in just in case...
		OnEventError: func(context string, err error) {
			log.Printf("[EVT] %s : %s\n", context, err.Error())
		},
	})
	app.betrayalManager = km
	if err != nil {
		log.Fatal(err)
	}

	// Call unregister twice to remove any lingering commands from previous runs
	app.betrayalManager.Unregister()

	tally := app.RegisterBetrayalCommands(
		new(commands.Test),
		new(inventory.Inventory),
		new(roll.Roll),
		new(commands.ActionFunnel),
		new(view.View),
		new(commands.Buy),
		new(commands.List),
		new(commands.Insult),
		new(commands.Ping),
		new(commands.Vote),
		new(commands.Kill),
		new(commands.Revive),
		new(commands.Setup),
		new(commands.Alliance),
	)

	app.betrayalManager.Session().AddHandler(logHandler)
	defer app.betrayalManager.Unregister()

	err = bot.Open()
	if err != nil {
		log.Fatal("error opening connection,", err)
	}
	defer bot.Close()

	log.Printf(
		"%s is now running with %d commands. Press CTRL-C to exit.\n",
		bot.State.User.Username,
		tally,
	)

	// start the scheduler
	app.scheduler.StartAsync()
	log.Printf("Scheduler started at %s EST\n", util.GetEstTimeStamp())

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	if err := app.betrayalManager.Session().Close(); err != nil {
		log.Fatal("error closing connection,", err)
	}
}

// TODO: Make Log Channel configurable with a slash command maybe?

// Handles logging all slash commands to a dedicated channel
func logHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}

	testLoggerID := "1108318770138714163"
	options := i.ApplicationCommandData().Options
	msg := processOptions(s, options)

	logOutput := fmt.Sprintf("%s - /%s %s - %s", i.Member.User.Username, i.ApplicationCommandData().Name, msg, util.GetEstTimeStamp())
	_, err := s.ChannelMessageSend(testLoggerID, discord.Code(logOutput))
	if err != nil {
		log.Println(err)
	}
}

// Primary helper for logHandler to process options that a user inputted for a slash command to get invoked (including value arguments)
func processOptions(s *discordgo.Session, options []*discordgo.ApplicationCommandInteractionDataOption) string {
	var msg string
	optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))

	for _, opt := range options {
		optionMap[opt.Name] = opt
	}

	for _, opt := range options {
		if o, ok := optionMap[opt.Name]; ok {
			msg += formatOption(s, o)
		}
	}

	return msg
}

// Helper to handle parsing argument options to a string format (really should just be apart of processOptions but...Too Bad!)
// Also this function is a mess, but it works even if I'm using recursion :)
func formatOption(s *discordgo.Session, o *discordgo.ApplicationCommandInteractionDataOption) string {
	switch o.Type {
	default:
		return ""
	case discordgo.ApplicationCommandOptionString:
		return fmt.Sprintf("%s:%s, ", o.Name, o.StringValue())
	case discordgo.ApplicationCommandOptionInteger:
		return fmt.Sprintf("%s:%d, ", o.Name, o.IntValue())
	case discordgo.ApplicationCommandOptionBoolean:
		return fmt.Sprintf("%s:%t, ", o.Name, o.BoolValue())
	case discordgo.ApplicationCommandOptionUser:
		return fmt.Sprintf("%s:%s, ", o.Name, o.UserValue(s).Username)
	case discordgo.ApplicationCommandOptionChannel:
		return fmt.Sprintf("%s:%s, ", o.Name, o.ChannelValue(s).Name)
	case discordgo.ApplicationCommandOptionSubCommand:
		return processOptions(s, o.Options)
	case discordgo.ApplicationCommandOptionSubCommandGroup:
		return fmt.Sprintf("%s  %s", o.Name, processOptions(s, o.Options))

		// I don't think there's ever going to be a case where I'm using these...
	case discordgo.ApplicationCommandOptionRole:
		return ""
	case discordgo.ApplicationCommandOptionMentionable:
		return ""
	}
}
