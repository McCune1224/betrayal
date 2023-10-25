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

// FIXME: this is a mess
// May god have mercy on my soul for the following function
func logHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}
	// loggerChID := "1108318770138714163"
	testLoggerID := "1140968068705701898"
	// check for subcommand and if so fetch the subcommand name from the options
	options := i.ApplicationCommandData().Options
	optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
	for _, opt := range options {
		optionMap[opt.Name] = opt
	}

	msg := ""

	for _, opt := range options {
		if o, ok := optionMap[opt.Name]; ok {
			switch o.Type {
			default:
				continue
			case discordgo.ApplicationCommandOptionString:
				msg += fmt.Sprintf("%s: %s, ", o.Name, o.StringValue())
			case discordgo.ApplicationCommandOptionInteger:
				msg += fmt.Sprintf("%s: %d, ", o.Name, o.IntValue())
			case discordgo.ApplicationCommandOptionBoolean:
				msg += fmt.Sprintf("%s: %t, ", o.Name, o.BoolValue())
			case discordgo.ApplicationCommandOptionUser:
				msg += fmt.Sprintf("%s: %s, ", o.Name, o.UserValue(s).Username)
			case discordgo.ApplicationCommandOptionChannel:
				msg += fmt.Sprintf("%s: %s, ", o.Name, o.ChannelValue(s).Name)
			case discordgo.ApplicationCommandOptionRole:
				continue
			case discordgo.ApplicationCommandOptionMentionable:
				continue
			case discordgo.ApplicationCommandOptionSubCommand:
				// do the same thing as above but for subcommands
				subOptions := o.Options
				subOptionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(subOptions))
				for _, subOpt := range subOptions {
					subOptionMap[subOpt.Name] = subOpt
				}
				for _, subOpt := range subOptions {
					if so, ok := subOptionMap[subOpt.Name]; ok {
						switch so.Type {

						case discordgo.ApplicationCommandOptionString:
							msg += fmt.Sprintf("%s: %s, ", so.Name, so.StringValue())
						case discordgo.ApplicationCommandOptionInteger:
							msg += fmt.Sprintf("%s: %d, ", so.Name, so.IntValue())
						case discordgo.ApplicationCommandOptionBoolean:
							msg += fmt.Sprintf("%s: %t, ", so.Name, so.BoolValue())
						case discordgo.ApplicationCommandOptionUser:
							msg += fmt.Sprintf("%s: %s, ", so.Name, so.UserValue(s).Username)
						case discordgo.ApplicationCommandOptionChannel:
							msg += fmt.Sprintf("%s: %s, ", so.Name, so.ChannelValue(s).Name)
						case discordgo.ApplicationCommandOptionRole:
							continue
						case discordgo.ApplicationCommandOptionMentionable:
							continue
						default:
							continue
						}
					}
				}
			}
		}
	}
	// get invoker name
	logOutput := fmt.Sprintf("%s - /%s '%s' - %s", i.Member.User.Username, i.ApplicationCommandData().Name, msg, util.GetEstTimeStamp())
	_, err := s.ChannelMessageSend(testLoggerID, discord.Code(logOutput))
	if err != nil {
		log.Println(err)
	}
}
