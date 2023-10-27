package main

import (
	"fmt"
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

	db, err := sqlx.Connect("postgres", cfg.database.dsn)
	if err != nil {
		log.Fatal("error opening database,", err)
	}

	dbModels := data.NewModels(db)

	app := &app{
		conifg: cfg,
		models: dbModels,
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
			log.Printf("[CMD] %s - %s : %s\n", ctx.Command.Name(), ctx.GetEvent().Member.User.Username, err.Error())
		},
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
		new(roll.Roll),
		new(inventory.Inventory),
		new(commands.ActionFunnel),
		new(view.View),
		new(commands.Buy),
		new(commands.List),
		new(commands.Insult),
		new(commands.Ping),
		new(commands.Vote),
		new(commands.Kill),
		new(commands.Revive),
	)
	// logger setup for slash commands and ken
	// loggerChID := "1108318770138714163"
	// testLoggerID := "1140968068705701898"
	//
	//
	//

	app.betrayalManager.Session().AddHandler(logHandler)
	defer app.betrayalManager.Unregister()

	err = bot.Open()
	if err != nil {
		log.Fatal("error opening connection,", err)
	}
	defer bot.Close()

	// bad grammar

	log.Printf(
		"%s is now running with %d commands. Press CTRL-C to exit.\n",
		bot.State.User.Username,
		tally,
	)

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	if err := app.betrayalManager.Session().Close(); err != nil {
		log.Fatal("error closing connection,", err)
	}
}

// Handles logging of slash commands when invoked
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
	case discordgo.ApplicationCommandOptionRole:
		return ""
	case discordgo.ApplicationCommandOptionMentionable:
		return ""
	case discordgo.ApplicationCommandOptionSubCommand:
		return processOptions(s, o.Options)
	case discordgo.ApplicationCommandOptionSubCommandGroup:
		// subcommandGroupName
		return fmt.Sprintf("%s  %s", o.Name, processOptions(s, o.Options))
	}
}
